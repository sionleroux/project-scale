// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	camera "github.com/melonfunction/ebiten-camera"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

func main() {
	const gameWidth, gameHeight = 320, 240
	const screenScaleFactor = 4

	ebiten.SetWindowSize(gameWidth*screenScaleFactor, gameHeight*screenScaleFactor)
	ebiten.SetWindowTitle("project-scale")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := &Game{
		Width:  gameWidth,
		Height: gameHeight,
		Space:  resolv.NewSpace(gameWidth, gameHeight, 16, 16),
		Camera: camera.NewCamera(gameWidth, gameHeight, 0, 0, 0, 1),
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

	// Player setup
	g.Player = NewPlayer([]int{gameWidth / 2, gameHeight / 2})
	g.Player.Input = g.InputSystem.NewHandler(0, keymap)
	g.Space.Add(g.Player.Object)

	// Pre-render map
	g.LDTKProject = loadMaps("assets/maps/Project scale.ldtk")
	g.TileRenderer = NewTileRenderer(&EmbedLoader{"assets/maps"})
	level := g.LDTKProject.Levels[g.Level]
	bg := ebiten.NewImage(level.Width, level.Height)
	bg.Fill(level.BGColor)
	g.TileRenderer.Render(level)
	for _, layer := range g.TileRenderer.RenderedLayers {
		log.Println("Pre-rendering layer:", layer.Layer.Identifier)
		bg.DrawImage(layer.Image, &ebiten.DrawImageOptions{})
	}
	g.Background = bg

	// Obstacles
	tilesToObstacles(level.Layers[0], g.Space)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Width        int
	Height       int
	Player       *Player
	InputSystem  input.System
	Space        *resolv.Space
	TileRenderer *TileRenderer
	LDTKProject  *ldtkgo.Project
	Background   *ebiten.Image
	Level        int
	Camera       *camera.Camera
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

// Update calculates game logic
func (g *Game) Update() error {

	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return errors.New("game quit by player")
	}

	// Pressing F toggles full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
		} else {
			ebiten.SetFullscreen(true)
		}
	}

	// Movement controls
	g.InputSystem.Update()
	g.Player.Update()

	// Position camera
	g.Camera.SetPosition(g.Player.X, g.Player.Y)

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	g.Camera.Surface.Clear()
	cameraOrigin := g.Camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0)

	g.Camera.Surface.DrawImage(g.Background, cameraOrigin)
	g.Player.Draw(g.Camera)
	g.Camera.Blit(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintln("Tag:", g.Player.WhatTile))
}
