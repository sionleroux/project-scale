package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/resolv"
)

const (
	ActionMoveUp input.Action = iota
	ActionMoveLeft
	ActionMoveDown
	ActionMoveRight
	ActionJump
)

type PlayerState int8

const (
	StateIdle PlayerState = iota
	StateMoving
	StateJumping
)

// Player is the player character in the game
type Player struct {
	Input *input.Handler
	*resolv.Object
	State PlayerState
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
	speed := 1.0
	if p.Input.ActionIsJustPressed(ActionJump) {
		p.State = StateJumping
		speed = 20.0
	}

	if p.Input.ActionIsPressed(ActionMoveUp) {
		p.move(+0, -speed)
	}
	if p.Input.ActionIsPressed(ActionMoveDown) {
		p.move(+0, +speed)
	}
	if p.Input.ActionIsPressed(ActionMoveLeft) {
		p.move(-speed, +0)
	}
	if p.Input.ActionIsPressed(ActionMoveRight) {
		p.move(+speed, +0)
	}
}

func (p *Player) move(dx, dy float64) {
	if collision := p.Check(dx, 0); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(dx, 0, o.Shape) != nil {
				dx = 0
			}
		}
	}
	p.X += dx

	if collision := p.Check(0, dy); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(0, dy, o.Shape) != nil {
				dy = 0
			}
		}
	}
	p.Y += dy
}

func (p *Player) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(
		screen,
		float64(p.X),
		float64(p.Y),
		20,
		20,
		color.White,
	)
}
