// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/joelschutz/stagehand"
	"github.com/sinisterstuf/project-scale/camera"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/tinne26/etxt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

// Length of the fading animation
const fadeOutTime = 360
const maxScore = 1000
const minMinScale = 0.18

const (
	ActionMoveUp input.Action = iota
	ActionMoveLeft
	ActionMoveDown
	ActionMoveRight
	ActionPrimary
	ActionMenu
)

func NewGameScene(game *Game, loadingState *LoadingState) {

	g := &GameScene{
		FadeTween: gween.New(0, 255, fadeOutTime, ease.Linear),
		Debuggers: debuggers,
	}

	loadingState.IncreaseCounter(1)
	// Pre-render map
	g.LDTKProject = loadMaps("assets/maps/Project scale.ldtk")
	g.TileRenderer = NewTileRenderer(&EmbedLoader{"assets/maps"})
	level := g.LDTKProject.Levels[g.Level]
	fg := ebiten.NewImage(level.Width, level.Height)
	bg := ebiten.NewImage(level.Width, level.Height)
	g.TileRenderer.Render(level)
	for _, layer := range g.TileRenderer.RenderedLayers {
		log.Println("Pre-rendering layer:", layer.Layer.Identifier)
		switch layer.Layer.Identifier {
		case LayerInvisible:
			continue
		case LayerWalls:
			fg.DrawImage(layer.Image, &ebiten.DrawImageOptions{})
		default:
			bg.DrawImage(layer.Image, &ebiten.DrawImageOptions{})
		}
	}
	g.Background = bg
	g.Foreground = fg
	game.Fog = NewFog(float64(level.Height))

	// Backdrop
	game.Backdrops = NewBackdrops(float64(level.Height))

	// Create space for collision detection
	g.Space = resolv.NewSpace(level.Width, level.Height, 16, 16)

	// Obstacles
	for _, layerName := range []string{
		LayerFloor,
		LayerWalls,
		LayerInvisible,
	} {
		tilesToObstacles(level.LayerByIdentifier(layerName), g.Space)
	}

	// SoundLoops
	loadingState.IncreaseCounter(1)
	g.Sounds = make(Sounds, 5)
	g.Sounds[backgroundMusic] = &Sound{Volume: 0.5}
	g.Sounds[backgroundMusic].AddSound("assets/music/game-music", sampleRate, context, 7)

	// Sounds
	loadingState.IncreaseCounter(1)
	g.Sounds[sfxSubmerge] = &Sound{Volume: 0.7}
	g.Sounds[sfxSubmerge].AddSound("assets/sfx/submerge", sampleRate, context, 1)
	g.Sounds[sfxSplash] = &Sound{Volume: 0.7}
	g.Sounds[sfxSplash].AddSound("assets/sfx/splash", sampleRate, context, 1)
	g.Sounds[sfxUnderwater] = &Sound{Volume: 1}
	g.Sounds[sfxUnderwater].AddSound("assets/sfx/underwater", sampleRate, context, 1)
	g.Sounds[voiceGameWon] = &Sound{Volume: 0.5}
	g.Sounds[voiceGameWon].AddSound("assets/voices/game-won", sampleRate, context, 1)

	// Entities
	loadingState.IncreaseCounter(1)

	// Finish point
	entities := level.LayerByIdentifier(LayerEntities)
	finishPos := entities.EntityByIdentifier(EntityFinish)
	finish := resolv.NewObject(
		float64(finishPos.Position[0]), float64(finishPos.Position[1]),
		float64(finishPos.Width), float64(finishPos.Height),
		TagFinish,
	)
	finish.SetShape(resolv.NewRectangle(
		0, 0, // origin
		float64(finishPos.Width), float64(finishPos.Height),
	))
	g.Space.Add(finish)

	// Player setup
	startPos := entities.EntityByIdentifier(EntityPlayerStart)
	startCenter := []int{
		startPos.Position[0] + (startPos.Width / 2),
		startPos.Position[1] + (startPos.Height / 2),
	}
	game.StartPos = startCenter
	g.Player = NewPlayer(startCenter, game.Camera)
	g.Player.Input = game.Input
	g.Space.Add(g.Player.Object)

	game.Water = NewWater(float64(level.Height) + 4*g.Player.Size.Y)

	// Done
	loadingState.IncreaseCounter(1)
	game.Scenes[gameRunning] = g
	loadingState.SetLoaded(true)
}

// GameScene represents the main game state
type GameScene struct {
	BaseScene
	Player       *Player
	Space        *resolv.Space
	TileRenderer *TileRenderer
	LDTKProject  *ldtkgo.Project
	Background   *ebiten.Image
	Foreground   *ebiten.Image
	Level        int
	Debuggers    Debuggers
	Sounds       Sounds
	Alpha        uint8
	FadeTween    *gween.Tween
}

// Update calculates game logic
func (g *GameScene) Update() error {
	g.State.InputSystem.Update()

	if g.State.Input.ActionIsJustPressed(ActionMenu) {
		g.SaveLastRender(true)
		g.SceneManager.SwitchTo(g.State.Scenes[gamePaused])
		return nil
	}

	if CheatsAllowed && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		wx, wy := g.State.Camera.GetWorldCoords(float64(x), float64(y))
		g.Player.Position.X = wx
		g.Player.Position.Y = wy
	}

	// Movement controls
	g.Player.Update()

	pos := GetScoreFromY(int(g.Player.Position.Y), g.State.StartPos[1])
	if pos > g.State.Stat.LastHighestPoint {
		g.State.Stat.LastHighestPoint = pos
	}

	if g.Player.State != stateWinning && g.CheckFinish() {
		g.State.Stat.GameEnd = time.Now()
		g.State.Stat.LastRound = int(g.State.Stat.GameEnd.Sub(g.State.Stat.GameStart).Seconds())
		if g.State.Stat.FastestRound <= 0 || g.State.Stat.FastestRound > g.State.Stat.LastRound {
			g.State.Stat.FastestRound = g.State.Stat.LastRound
			g.State.Stat.Save()
		}
		g.Player.State = stateWinning
		g.State.minScale = float64(g.State.Camera.Width) / float64(g.State.Backdrops.Backdrops[0].Image.Bounds().Dx()-int(math.Abs(g.Player.Position.X))*2)
		if g.State.minScale < minMinScale {
			g.State.minScale = minMinScale
		}
		g.Sounds[backgroundMusic].FadeOut(1)
		g.Sounds[voiceGameWon].Play()
	}

	if g.Player.State == stateWinning {
		if g.State.Camera.Scale > g.State.minScale {
			g.State.Camera.Zoom(0.99)
			g.State.Camera.SetPosition(g.Player.Position.X, float64(g.State.Camera.Height/2)/g.State.Camera.Scale)
			g.State.Camera.Update()
		}
	} else {
		// Position camera and clamp in to the Map dimensions
		maxHeight := g.LDTKProject.Levels[g.Level].Height
		g.State.Camera.SetPosition(g.Player.Position.X, math.Min(
			math.Max(g.Player.Position.Y, float64(g.State.Camera.Height/2)),
			float64(maxHeight-g.State.Camera.Height/2),
		))
		g.State.Camera.Update()
	}

	if g.Player.State == stateWinning || g.Player.State == stateDying {
		g.Sounds[backgroundMusic].Update()
	}

	g.State.Water.Update(g.Player.State != stateWinning)

	g.State.Fog.Update()

	switch g.Player.State {
	case stateDying:
		alpha, _ := g.FadeTween.Update(1)
		g.Alpha = uint8(alpha)
		if g.Alpha == 128 {
			g.Player.State = stateDead
			g.Sounds[backgroundMusic].Pause()
			g.Sounds[backgroundMusic].LowPass(false)
			g.SaveLastRender(true)
			g.SceneManager.SwitchTo(g.State.Scenes[gameOver])
			return nil
		}

	case stateWinning:
		if g.State.Camera.Scale <= g.State.minScale {
			alpha, _ := g.FadeTween.Update(1)
			g.Alpha = uint8(alpha)
			if g.Alpha == 200 {
				g.SaveLastRender(false)
				g.State.Stat.LastHighestPoint = maxScore
				g.State.Stat.HighestPoint = maxScore
				g.State.Stat.Save()
				g.Player.State = gameWon
				g.SceneManager.SwitchTo(g.State.Scenes[gameWon])
				return nil
			}
		}
	default:
		if !g.Sounds[backgroundMusic].IsPlaying() {
			g.Sounds[backgroundMusic].PlayNext()
		}
		if g.CheckDeath() {
			g.Sounds[backgroundMusic].LowPass(true)
			g.Sounds[backgroundMusic].FadeOut(2)
			if g.Player.State != stateFalling {
				g.Sounds[sfxSubmerge].Play()
			} else {
				g.Sounds[sfxSplash].Play()
			}
			g.Sounds[sfxUnderwater].Play()
			g.Player.State = stateDying
			g.Player.AnimState = playerFallloop
			if g.State.Stat.LastHighestPoint > g.State.Stat.HighestPoint {
				g.State.Stat.HighestPoint = g.State.Stat.LastHighestPoint
				g.State.Stat.Save()
			}
		}

	}

	return nil
}

// Draw draws the game screen by one frame
func (g *GameScene) Draw(screen *ebiten.Image) {
	g.State.Camera.Surface.Clear()
	cameraOrigin := g.State.Camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0)

	g.State.Backdrops.Draw(g.State.Camera, g.State.Water.Level)
	if g.Player.State == stateDying {
		g.State.Camera.Surface.DrawImage(g.Background, cameraOrigin)
		g.State.Camera.Surface.DrawImage(g.Foreground, cameraOrigin)
		g.Player.Draw(g.State.Camera)
	} else {
		g.State.Camera.Surface.DrawImage(g.Background, cameraOrigin)
		g.Player.Draw(g.State.Camera)
		g.State.Camera.Surface.DrawImage(g.Foreground, cameraOrigin)
		for _, hint := range g.Player.ControlHints {
			hint.Draw(g.Player.Position.X, g.Player.Position.Y, g.State.Camera)
		}
	}
	g.State.Water.Draw(g.State.Camera)

	fogOp := g.State.Fog.GetDrawImageOptions()
	fogOp = g.State.Camera.GetTranslation(fogOp, -float64(g.State.Fog.Image.Bounds().Dx())/2, 0)
	g.State.Camera.Surface.DrawImage(g.State.Fog.Image, fogOp)

	g.State.Camera.Blit(screen)

	if g.Player.State == stateDying || g.Player.State == stateDead || g.Player.State == stateWinning || g.Player.State == stateWon {
		vector.DrawFilledRect(screen, 0, 0, float32(g.State.Width), float32(g.State.Height), color.RGBA{0, 0, 0, g.Alpha}, false)
	}

	if g.Player.State != stateWinning && g.Player.State != stateWon {
		g.DrawMinimap(screen)
	}
	g.Debuggers.Debug(g, screen)
}

func (g *GameScene) Load(st State, sm *stagehand.SceneManager[State]) {
	g.BaseScene.Load(st, sm)
	if g.State.ResetNeeded {
		g.State.ResetNeeded = false
		g.Reset()
		g.Sounds[backgroundMusic].PlayNext()
	} else {
		g.Sounds[backgroundMusic].Resume()
	}

}

func (g *GameScene) Unload() State {
	g.Sounds[backgroundMusic].Pause()
	g.Sounds[sfxUnderwater].Pause()

	return g.BaseScene.Unload()
}

func (g *GameScene) CheckFinish() bool {
	if collision := g.Player.Check(0, 0, TagFinish); collision != nil {
		for _, o := range collision.Objects {
			if g.Player.Shape.Intersection(0, 0, o.Shape) != nil {
				return true
			}
		}
	}
	return false
}

func (g *GameScene) CheckDeath() bool {
	// Death by water (water covers the top of you)
	if g.State.Water.Level < g.Player.Position.Y-g.Player.Size.Y/4 {
		return true
	}

	return false
}

func (g *GameScene) Reset() {
	level := g.LDTKProject.Levels[g.Level]
	g.Player.Position.X, g.Player.Position.Y = float64(g.State.StartPos[0]), float64(g.State.StartPos[1])
	g.Player.Facing = directionUp
	g.Player.AnimState = playerIdle
	g.Player.State = stateIdle
	g.Player.Rotation = 0
	g.State.Water = NewWater(float64(level.Height) + 4*g.Player.Size.Y)
	g.Sounds[backgroundMusic].SetVolume(0.5)
	g.Alpha = 0
	g.FadeTween.Reset()
	g.State.Camera.Zoom(1 / g.State.Camera.Scale)
	g.State.Stat.GameStart = time.Now()
	g.State.Stat.LastHighestPoint = 0
}

type Entity interface {
	Update()
	Draw(cam *camera.Camera)
}

func blurImage(image *ebiten.Image) *ebiten.Image {
	result := ebiten.NewImage(image.Bounds().Dx(), image.Bounds().Dy())

	layers := 0
	for j := -3; j <= 3; j++ {
		for i := -3; i <= 3; i++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(i), float64(j))
			layers++
			op.ColorScale.ScaleAlpha(1 / float32(layers))
			result.DrawImage(image, op)
		}
	}
	return result
}

func (g *GameScene) SaveLastRender(blur bool) {
	g.Draw(g.State.lastRender)
	if blur {
		g.State.lastRender = blurImage(g.State.lastRender)
	}
}

func (g *GameScene) DrawMinimap(screen *ebiten.Image) {
	// v1
	// scale := float64(g.State.Height) / float64(g.Background.Bounds().Dy())
	// vector.DrawFilledRect(screen, 0, 0, 20, float32(g.State.Height), color.RGBA{0, 0, 0, 128}, false)

	// op := &ebiten.DrawImageOptions{}
	// op.GeoM.Scale(scale, scale)
	// screen.DrawImage(g.Background, op)

	// v2
	// Draw building
	scale := float64(g.State.Height) / float64(g.Background.Bounds().Dy())
	minimapWidth := float64(g.Background.Bounds().Dx()) * scale
	vector.DrawFilledRect(screen, 0, 0, float32(minimapWidth), float32(g.State.Height), color.RGBA{40, 40, 40, 128}, false)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.ColorScale.Scale(0, 0, 0, 255)
	screen.DrawImage(g.Background, op)

	// Draw high score
	hsColor := color.RGBA{255, 0, 0, 255}
	hsYPosition := GetYFromScore(g.State.Stat.HighestPoint, g.State.StartPos[1]) * scale
	vector.StrokeLine(screen, 0, float32(hsYPosition), 30, float32(hsYPosition), 1, hsColor, false)
	g.State.TextRenderer.DrawXY(screen, fmt.Sprintf("%d", g.State.Stat.HighestPoint), hsColor, 8, int(minimapWidth+1), int(hsYPosition-8), etxt.Left)

	// Draw player
	playerColor := color.RGBA{255, 255, 0, 255}
	playerXPosition := g.Player.Position.X * scale
	playerYPosition := g.Player.Position.Y * scale
	playerHeightValue := GetScoreFromY(int(g.Player.Position.Y), g.State.StartPos[1])
	vector.StrokeLine(screen, float32(playerXPosition+3), float32(playerYPosition), 30, float32(playerYPosition), 1, playerColor, false)
	vector.StrokeLine(screen, float32(playerXPosition-1), float32(playerYPosition), float32(playerXPosition+1), float32(playerYPosition), 1, playerColor, false)
	g.State.TextRenderer.DrawXY(screen, fmt.Sprintf("%d", playerHeightValue), playerColor, 8, int(minimapWidth+1), int(playerYPosition-8), etxt.Left)

	// Draw water
	vector.DrawFilledRect(screen, 0, float32(g.State.Water.Level*scale), float32(minimapWidth), float32(float64(g.State.Height)-g.State.Water.Level*scale), color.RGBA{58, 79, 118, 204}, false)
}

func GetScoreFromY(position, levelHeight int) int {
	return int(float64(levelHeight-position) / float64(levelHeight) * float64(maxScore))
}

func GetYFromScore(score, levelHeight int) float64 {
	return float64(levelHeight) - float64(score)/float64(maxScore)*float64(levelHeight)
}
