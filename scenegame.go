// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"log"

	"github.com/sinisterstuf/project-scale/camera"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

func NewGameScene(game *Game) *GameScene {

	g := &GameScene{
		Width:     game.Width,
		Height:    game.Height,
		Camera:    camera.NewCamera(game.Width, game.Height),
		Debuggers: debuggers,
	}

	// Input setup
	g.InputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})
	keymap := input.Keymap{
		ActionMoveUp:    {input.KeyUp, input.KeyW, input.KeyGamepadUp, input.KeyGamepadLStickUp},
		ActionMoveLeft:  {input.KeyLeft, input.KeyA, input.KeyGamepadLeft, input.KeyGamepadLStickLeft},
		ActionMoveDown:  {input.KeyDown, input.KeyS, input.KeyGamepadDown, input.KeyGamepadLStickDown},
		ActionMoveRight: {input.KeyRight, input.KeyD, input.KeyGamepadRight, input.KeyGamepadLStickRight},
		ActionJump:      {input.KeySpace, input.KeyGamepadA},
	}

	// Pre-render map
	g.LDTKProject = loadMaps("assets/maps/Project scale.ldtk")
	g.TileRenderer = NewTileRenderer(&EmbedLoader{"assets/maps"})
	level := g.LDTKProject.Levels[g.Level]
	fg := ebiten.NewImage(level.Width, level.Height)
	bg := ebiten.NewImage(level.Width, level.Height)
	bg.Fill(level.BGColor)
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
	g.Player = NewPlayer(startCenter, g.Camera)
	g.Player.Input = g.InputSystem.NewHandler(0, keymap)
	g.Space.Add(g.Player.Object)

	return g
}

// GameScene represents the main game state
type GameScene struct {
	Width        int
	Height       int
	Player       *Player
	InputSystem  input.System
	Space        *resolv.Space
	TileRenderer *TileRenderer
	LDTKProject  *ldtkgo.Project
	Background   *ebiten.Image
	Foreground   *ebiten.Image
	Level        int
	Camera       *camera.Camera
	Debuggers    Debuggers
}

// Update calculates game logic
func (g *GameScene) Update() (SceneIndex, error) {

	// Pressing F toggles full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
		} else {
			ebiten.SetFullscreen(true)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		return gamePaused, nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		wx, wy := g.Camera.GetWorldCoords(float64(x), float64(y))
		g.Player.X = wx
		g.Player.Y = wy
	}

	// Movement controls
	g.InputSystem.Update()
	g.Player.Update()

	if g.CheckFinish() {
		return gameRunning, errors.New("Player hit the finish line")
	}

	// Position camera
	g.Camera.SetPosition(g.Player.X, g.Player.Y)
	g.Camera.Update()

	return gameRunning, nil
}

// Draw draws the game screen by one frame
func (g *GameScene) Draw(screen *ebiten.Image) {
	g.Camera.Surface.Clear()
	cameraOrigin := g.Camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0)

	g.Camera.Surface.DrawImage(g.Background, cameraOrigin)
	g.Player.Draw(g.Camera)
	g.Camera.Surface.DrawImage(g.Foreground, cameraOrigin)
	g.Camera.Blit(screen)

	g.Debuggers.Debug(g, screen)
}

func (g *GameScene) Load() {
	// TODO: put some game reset logic here, unpause music etc.
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

type Entity interface {
	Update()
	Draw(cam *camera.Camera)
}
