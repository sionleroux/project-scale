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
)

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

func (p *Player) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(
		screen,
		float64(p.Object.X),
		float64(p.Object.Y),
		20,
		20,
		color.White,
	)
}
