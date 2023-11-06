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

type PlayerAxis int8

const (
	AxisVertical PlayerAxis = iota
	AxisHorizontal
	AxisBoth
)

// Player is the player character in the game
type Player struct {
	*resolv.Object
	Input   *input.Handler
	State   playerAnimationTags
	Sprite  *SpriteSheet
	Frame   int
	Tick    int
	Jumping bool
	Axis    PlayerAxis
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
	p.Tick++
	p.updateMovement()
	p.animate()
	p.Object.Update()
}

func (p *Player) updateMovement() {
	speed := 1.0

	if p.Input.ActionIsPressed(ActionJump) {
		speed = 2.0
		if p.Input.ActionIsJustPressed(ActionJump) {
			p.Jumping = true
			p.State = playerJumpingstart
		}
		if p.State == playerJumpingmidair {
			p.move(+0, -speed)
		}
	}

	if !p.Jumping { // XXX: I don't like this
		p.State = playerIdle
		if p.Input.ActionIsPressed(ActionMoveUp) {
			p.move(+0, -speed)
			p.State = playerClimpingupdown
			p.Axis = AxisVertical
		}
		if p.Input.ActionIsPressed(ActionMoveDown) {
			p.move(+0, +speed)
			p.State = playerClimpingupdown
			p.Axis = AxisVertical
		}
		if p.Input.ActionIsPressed(ActionMoveLeft) {
			p.move(-speed, +0)
			p.State = playerClimbingleftright
			if p.Axis == AxisVertical {
				p.Axis = AxisBoth
			} else {
				p.Axis = AxisHorizontal
			}
		}
		if p.Input.ActionIsPressed(ActionMoveRight) {
			p.move(+speed, +0) // TODO: cancel movement when pressing opposite directions
			p.State = playerClimbingleftright
			if p.Axis == AxisVertical {
				p.Axis = AxisBoth
			} else {
				p.Axis = AxisHorizontal
			}
		}
		if p.Axis == AxisBoth {
			p.State = playerClimbingdiagonally
		}
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

func (p *Player) animate() {
	if p.Frame == p.Sprite.Meta.FrameTags[p.State].To {
		p.animationBasedStateChanges()
	}
	p.Frame = Animate(p.Frame, p.Tick, p.Sprite.Meta.FrameTags[p.State])
}

// Animation-trigged state changes
func (p *Player) animationBasedStateChanges() {
	switch p.State {
	case playerJumpingstart, playerJumpingmidair:
		if p.Input.ActionIsPressed(ActionJump) {
			p.State = playerJumpingmidair
		} else {
			p.State = playerJumpingend
		}
	case playerJumpingend:
		p.State = playerIdle
		p.Jumping = false
	}
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
