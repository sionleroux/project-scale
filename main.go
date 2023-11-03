// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/resolv"
)

const (
	ActionMoveUp input.Action = iota
	ActionMoveLeft
	ActionMoveDown
	ActionMoveRight
)

func main() {
	gameWidth, gameHeight := 640, 480

	ebiten.SetWindowSize(gameWidth, gameHeight)
	ebiten.SetWindowTitle("project-scale")

	g := &Game{
		Width:  gameWidth,
		Height: gameHeight,
		Space:  resolv.NewSpace(gameWidth, gameHeight, 20, 20),
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
	}

	// Player setup
	g.Player = NewPlayer([]int{gameWidth / 2, gameHeight / 2})
	g.Player.Input = g.InputSystem.NewHandler(0, keymap)
	g.Space.Add(g.Player.Object)

	// Obstacles
	obstacle := resolv.NewObject(
		float64(gameWidth/2), float64(gameHeight/2-80),
		20, 20,
	)
	obstacle.SetShape(resolv.NewRectangle(
		0, 0, // origin
		20, 20,
	))
	obstacle.Shape.(*resolv.ConvexPolygon).RecenterPoints()
	g.Space.Add(obstacle)
	g.Obstacle = obstacle

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Width       int
	Height      int
	Player      *Player
	InputSystem input.System
	Space       *resolv.Space
	Obstacle    *resolv.Object
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

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(
		screen,
		float64(g.Player.Object.X),
		float64(g.Player.Object.Y),
		20,
		20,
		color.White,
	)

	ebitenutil.DrawRect(
		screen,
		float64(g.Obstacle.X),
		float64(g.Obstacle.Y),
		20,
		20,
		color.NRGBA{255, 0, 0, 255},
	)
}

// Player is the player character in the game
type Player struct {
	Input  *input.Handler
	Object *resolv.Object
}

func NewPlayer(position []int) *Player {
	object := resolv.NewObject(
		float64(position[0]), float64(position[1]),
		20, 20,
	)
	object.SetShape(resolv.NewRectangle(
		0, 0, // origin
		20, 20,
	))
	object.Shape.(*resolv.ConvexPolygon).RecenterPoints()

	return &Player{
		Object: object,
	}
}

func (p *Player) Update() {
	p.updateMovement()
	p.Object.Update()
}

func (p *Player) updateMovement() {
	if p.Input.ActionIsPressed(ActionMoveUp) {
		p.move(+0, -1)
	}
	if p.Input.ActionIsPressed(ActionMoveDown) {
		p.move(+0, +1)
	}
	if p.Input.ActionIsPressed(ActionMoveLeft) {
		p.move(-1, +0)
	}
	if p.Input.ActionIsPressed(ActionMoveRight) {
		p.move(+1, +0)
	}
}

func (p *Player) move(dx, dy float64) {
	if collision := p.Object.Check(dx, 0); collision != nil {
		for _, o := range collision.Objects {
			if p.Object.Shape.Intersection(dx, 0, o.Shape) != nil {
				dx = 0
			}
		}
	}
	p.Object.X += dx

	if collision := p.Object.Check(0, dy); collision != nil {
		for _, o := range collision.Objects {
			if p.Object.Shape.Intersection(0, dy, o.Shape) != nil {
				dy = 0
			}
		}
	}
	p.Object.Y += dy
}
