// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"
	"log"
	"math"
	"time"

	"github.com/joelschutz/stagehand"
	"github.com/sinisterstuf/project-scale/camera"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

// Length of the fading animation
const fadeOutTime = 240

func NewGameScene(game *Game, loadingState *LoadingState) {

	g := &GameScene{
		Camera:    camera.NewCamera(game.Width, game.Height),
		FadeTween: gween.New(0, 255, fadeOutTime, ease.Linear),
		Debuggers: debuggers,
	}

	// Input setup
	g.InputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})
	g.Keymap = input.Keymap{
		ActionMoveUp:    {input.KeyUp, input.KeyW, input.KeyGamepadUp, input.KeyGamepadLStickUp},
		ActionMoveLeft:  {input.KeyLeft, input.KeyA, input.KeyGamepadLeft, input.KeyGamepadLStickLeft},
		ActionMoveDown:  {input.KeyDown, input.KeyS, input.KeyGamepadDown, input.KeyGamepadLStickDown},
		ActionMoveRight: {input.KeyRight, input.KeyD, input.KeyGamepadRight, input.KeyGamepadLStickRight},
		ActionJump:      {input.KeySpace, input.KeyGamepadA},
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
	g.Fog = loadImage("assets/backdrop/Project-scale-parallax-backdrop_0001_Smog-1-cloud.png")

	// Backdrop
	g.Backdrops = NewBackdrops(float64(level.Height))

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
	g.Music = &Sound{Volume: 0.5}
	g.Music.AddSound("assets/music/game-music", sampleRate, context, 7)
	g.Heartbeat = NewMusicPlayer(loadSoundFile("assets/sfx/heartbeat.ogg", sampleRate))

	// Sounds
	loadingState.IncreaseCounter(1)

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
	g.StartPos = startCenter
	g.Player = NewPlayer(startCenter, g.Camera)
	g.Player.Input = g.InputSystem.NewHandler(0, g.Keymap)
	g.Space.Add(g.Player.Object)

	g.Water = NewWater(float64(level.Height) + 4*g.Player.H)

	// Done
	loadingState.IncreaseCounter(1)
	game.Scenes[gameRunning] = g
	loadingState.SetLoaded(true)
}

// GameScene represents the main game state
type GameScene struct {
	BaseScene
	Player       *Player
	InputSystem  input.System
	Keymap       input.Keymap
	Space        *resolv.Space
	TileRenderer *TileRenderer
	LDTKProject  *ldtkgo.Project
	Background   *ebiten.Image
	Foreground   *ebiten.Image
	Fog          *ebiten.Image
	FogAngle     float64
	Level        int
	Camera       *camera.Camera
	Debuggers    Debuggers
	Water        *Water
	StartPos     []int
	Backdrops    Backdrops
	Music        *Sound
	Heartbeat    *MusicLoop
	Alpha        uint8
	FadeTween    *gween.Tween
}

// Update calculates game logic
func (g *GameScene) Update() error {

	// Pressing F toggles full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
		} else {
			ebiten.SetFullscreen(true)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.SceneManager.SwitchTo(g.State.Scenes[gamePaused])
		return nil
	}

	if CheatsAllowed && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		wx, wy := g.Camera.GetWorldCoords(float64(x), float64(y))
		g.Player.X = wx
		g.Player.Y = wy
	}

	// Movement controls
	g.InputSystem.Update()
	g.Player.Update()

	if g.CheckFinish() {
		g.SceneManager.SwitchTo(g.State.Scenes[gameWon])
		return nil
	}

	// Position camera and clamp in to the Map dimensions
	maxHeight := g.LDTKProject.Levels[g.Level].Height
	g.Camera.SetPosition(g.Player.X, math.Min(
		math.Max(g.Player.Y, float64(g.Camera.Height/2)),
		float64(maxHeight-g.Camera.Height/2),
	))

	g.Camera.Update()

	g.Water.Update()

	if !g.Player.Dying && !g.Player.Dead {
		if !g.Music.IsPlaying() {
			g.Music.PlayNext()
		}

		if g.CheckDeath() {
			g.Player.Dying = true
			g.State.Stat.LastHighestPoint = (g.StartPos[1] - int(g.Player.Y)) / gridSize
			if g.State.Stat.LastHighestPoint > g.State.Stat.HighestPoint {
				g.State.Stat.HighestPoint = g.State.Stat.LastHighestPoint
				g.State.Stat.Save()
			}
		}
	} else if g.Player.Dying {
		if !g.Heartbeat.IsPlaying() {
			g.Music.LowPass(true)
			g.Heartbeat.Play()
		}
		alpha, _ := g.FadeTween.Update(1)
		g.Alpha = uint8(alpha)
		if g.Alpha == 255 {
			g.Player.Dying = false
			g.Player.Dead = true
		}
	} else if g.Player.Dead {
		g.Music.Pause()
		g.Music.LowPass(false)
		g.Heartbeat.Pause()
		g.SceneManager.SwitchTo(g.State.Scenes[gameOver])
		return nil
	}

	g.FogAngle += 0.0001

	return nil
}

// Draw draws the game screen by one frame
func (g *GameScene) Draw(screen *ebiten.Image) {
	g.Camera.Surface.Clear()
	cameraOrigin := g.Camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0)

	g.Backdrops.Draw(g.Camera, g.Water.Level)
	g.Camera.Surface.DrawImage(g.Background, cameraOrigin)
	if g.Player.Dying {
		g.Camera.Surface.DrawImage(g.Foreground, cameraOrigin)
		g.Player.Draw(g.Camera)
	} else {
		g.Player.Draw(g.Camera)
		g.Camera.Surface.DrawImage(g.Foreground, cameraOrigin)
	}
	g.Water.Draw(g.Camera)
	fogOp := &ebiten.DrawImageOptions{}
	fogOp.GeoM.Translate(-float64(g.Fog.Bounds().Dx())/2, -float64(g.Fog.Bounds().Dy())/2)
	fogOp.GeoM.Rotate(g.FogAngle)
	fogOp.GeoM.Translate(+float64(g.Fog.Bounds().Dx())/2, +float64(g.Fog.Bounds().Dy())/2)
	fogOp.Blend = ebiten.BlendLighter
	g.Camera.Surface.DrawImage(g.Fog, g.Camera.GetTranslation(fogOp, -float64(g.Fog.Bounds().Dx())/2, 0))

	g.Camera.Blit(screen)

	if g.Player.Dying {
		vector.DrawFilledRect(screen, 0, 0, float32(g.State.Width), float32(g.State.Height), color.RGBA{0, 0, 0, g.Alpha}, false)
	}
	g.Debuggers.Debug(g, screen)
}

func (g *GameScene) Load(st State, sm *stagehand.SceneManager[State]) {
	g.BaseScene.Load(st, sm)
	if g.State.ResetNeeded {
		g.State.ResetNeeded = false
		g.State.Stat.GameStart = time.Now()
		g.Reset()
		g.Music.PlayNext()
	} else {
		g.Music.Resume()
	}

}

func (g *GameScene) Unload() State {
	g.Music.Pause()

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
	if g.Water.Level < g.Player.Y-g.Player.H/4 {
		return true
	}

	return false
}

func (g *GameScene) Reset() {
	level := g.LDTKProject.Levels[g.Level]
	g.Player.X, g.Player.Y = float64(g.StartPos[0]), float64(g.StartPos[1])
	g.Player.Facing = directionUp
	g.Player.State = playerIdle
	g.Player.Dead = false
	g.Player.Dying = false
	g.Player.Jumping = false
	g.Player.Falling = false
	g.Player.Standing = false
	g.Player.Slipping = false // :-(
	g.Player.Rotation = 0
	g.Player.Input = g.InputSystem.NewHandler(0, g.Keymap)
	g.Water = NewWater(float64(level.Height) + 4*g.Player.H)
	g.FadeTween.Reset()
}

type Entity interface {
	Update()
	Draw(cam *camera.Camera)
}
