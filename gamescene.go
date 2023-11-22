// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"log"
	"math"

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

	g.Water = NewWater(float64(level.Height) + 4*g.Player.H)

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
	Water        *Water
	Backdrops    Backdrops
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
		return gameWon, nil
	}

	// Position camera and clamp in to the Map dimensions
	maxHeight := g.LDTKProject.Levels[g.Level].Height
	g.Camera.SetPosition(g.Player.X, math.Min(
		math.Max(g.Player.Y, float64(g.Camera.Height/2)),
		float64(maxHeight-g.Camera.Height/2),
	))

	g.Camera.Update()

	g.Water.Update()

	if g.CheckDeath() {
		return gameOver, nil
	}

	return gameRunning, nil
}

// Draw draws the game screen by one frame
func (g *GameScene) Draw(screen *ebiten.Image) {
	g.Camera.Surface.Clear()
	cameraOrigin := g.Camera.GetTranslation(&ebiten.DrawImageOptions{}, 0, 0)

	g.Backdrops.Draw(g.Camera, g.Water.Level)
	g.Camera.Surface.DrawImage(g.Background, cameraOrigin)
	g.Player.Draw(g.Camera)
	g.Camera.Surface.DrawImage(g.Foreground, cameraOrigin)
	g.Water.Draw(g.Camera)
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

func (g *GameScene) CheckDeath() bool {
	// Death by water (water covers the top of you)
	if g.Water.Level < g.Player.Y-g.Player.H/4 {
		return true
	}

	return false
}

type Entity interface {
	Update()
	Draw(cam *camera.Camera)
}

func NewBackdrops(bottomOfMap float64) Backdrops {
	return Backdrops{
		bottomOfMap,
		[]Backdrop{
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0015_Background.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0014_Sky.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0013_Smog-5.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0012_Water-5.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0011_City-4.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0010_Smog-4.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0009_Water-4.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0008_City-3.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0007_Smog-3.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0006_Water-3.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0005_City-2.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0004_Smog-2.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0003_Water-2.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0002_City-1.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0001_Smog-1.png"), false, 0.0},
		},
	}
}

type Backdrops struct {
	bottomOfMap float64
	Backdrops   []Backdrop
}

func (bs Backdrops) Draw(cam *camera.Camera, waterLevel float64) {
	const waterSpacing = 16.0
	const howManyWaters = 5.0 // I counted them by hand.0
	backdropCenter := -float64(bs.Backdrops[0].Image.Bounds().Dx()) / 2
	watersDone := 0.0

	backdropPos := cam.GetTranslation(
		&ebiten.DrawImageOptions{},
		backdropCenter,
		-float64(bs.Backdrops[0].Image.Bounds().Dy())+bs.bottomOfMap,
	)

	for _, b := range bs.Backdrops {
		if b.Water {
			cam.Surface.DrawImage(b.Image, cam.GetTranslation(
				&ebiten.DrawImageOptions{},
				backdropCenter,
				waterLevel-(howManyWaters-watersDone)*waterSpacing,
			))
			watersDone++
		} else {
			cam.Surface.DrawImage(b.Image, backdropPos)
		}
	}
}

type Backdrop struct {
	Image *ebiten.Image
	Water bool
	Speed float64
}
