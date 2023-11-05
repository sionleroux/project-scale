package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/resolv"
)

//go:generate ./tools/gen_sprite_tags.sh assets/sprites/Nanobot.json player_anim.go player

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
	*resolv.Object
	Input  *input.Handler
	State  PlayerState
	Sprite *SpriteSheet
	Frame  int
}

func NewPlayer(position []int) *Player {
	object := resolv.NewObject(
		float64(position[0]), float64(position[1]),
		16, 16,
	)
	object.SetShape(resolv.NewRectangle(
		0, 0, // origin
		16, 16,
	))
	object.Shape.(*resolv.ConvexPolygon).RecenterPoints()

	return &Player{
		Object: object,
		Sprite: loadSprite("Nanobot"),
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
		speed = 10.0
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
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(p.X, p.Y)

	s := p.Sprite
	frame := s.Sprite[p.Frame]
	img := s.Image.SubImage(image.Rect(
		frame.Position.X,
		frame.Position.Y,
		frame.Position.X+frame.Position.W,
		frame.Position.Y+frame.Position.H,
	)).(*ebiten.Image)

	screen.DrawImage(img, op)
}
