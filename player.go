package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	camera "github.com/melonfunction/ebiten-camera"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/resolv"
)

//go:generate ./tools/gen_sprite_tags.sh assets/sprites/Nanobot.json player_anim.go player

const MinJumpTime = 13
const MaxJumpTime = 21

const (
	ActionMoveUp input.Action = iota
	ActionMoveLeft
	ActionMoveDown
	ActionMoveRight
	ActionJump
)

// Player is the player character in the game
type Player struct {
	*resolv.Object
	Input    *input.Handler
	State    playerAnimationTags
	Sprite   *SpriteSheet
	Frame    int
	Tick     int
	Jumping  bool
	Falling  bool
	Slipping bool
	Standing bool
	JumpTime int
	WhatTile string
}

func NewPlayer(position []int) *Player {
	object := resolv.NewObject(
		float64(position[0]), float64(position[1]),
		16, 16,
	)
	object.SetShape(resolv.NewRectangle(
		0, 0, // origin
		8, 8,
	))

	return &Player{
		Object: object,
		Sprite: loadSprite("Nanobot"),
	}
}

func (p *Player) Update() {
	p.Tick++
	if p.State == playerJumploop {
		p.JumpTime++
	}
	p.collisionChecks()
	p.updateMovement()
	p.animate()
	p.Object.Update()
}

func (p *Player) updateMovement() {
	speed := 0.6

	if !p.Falling && !p.Jumping && p.Input.ActionIsJustPressed(ActionJump) {
		p.Jumping = true
		p.State = playerJumpstart
	}

	if p.Jumping {
		switch p.State {
		case playerJumploop:
			speed = 2.0
		case playerJumpstart:
			speed = -0.3
		case playerJumpendfloor:
			speed = 0.2
		}
		p.move(+0, -speed)
	} else if p.Falling {
		switch p.State {
		case playerFallloop:
			speed = 3.0
		case playerFallstart, playerFallendwall:
			speed = 0.5
		}
		p.move(+0, speed)
	} else if p.Slipping {
		speed = 1.0
		p.move(+0, speed)
	} else {
		p.State = playerIdle
		if p.Input.ActionIsPressed(ActionMoveLeft) {
			p.move(-speed, +0)
			p.State = playerClimbleft
		} else if p.Input.ActionIsPressed(ActionMoveRight) {
			p.move(+speed, +0)
			p.State = playerClimbright
		} else if p.Input.ActionIsPressed(ActionMoveUp) {
			p.move(+0, -speed)
			p.State = playerClimbup
		} else if p.Input.ActionIsPressed(ActionMoveDown) {
			p.move(+0, +speed)
			p.State = playerClimbdown
		}
	}

}

func (p *Player) collisionChecks() {
	if collision := p.Check(0, 0); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(0, 0, o.Shape) != nil {
				p.WhatTile = o.Tags()[0]
			}
		}
	}

	// Start falling if you're stepping on a chasm
	if p.State != playerJumploop && !p.Falling && !p.Slipping {
		if collision := p.Check(0, 0, TagChasm, TagSlippery); collision != nil {
			for _, o := range collision.Objects {
				if p.Shape.Intersection(0, 0, o.Shape) != nil || p.insideOf(o) {
					p.Jumping = false // XXX: this is not the right place for this
					p.JumpTime = 0
					switch o.Tags()[0] {
					case TagChasm:
						p.State = playerFallstart
						p.Falling = true
					case TagSlippery:
						p.State = playerSlipstart
						p.Slipping = true
					}
				}
			}
		}
	}
}

func (p *Player) move(dx, dy float64) {
	if collision := p.Check(dx, 0, TagWall); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(dx, 0, o.Shape) != nil {
				dx = 0
			}
		}
	}
	p.X += dx

	if collision := p.Check(0, dy, TagWall, TagClimbable); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(0, dy, o.Shape) != nil {
				switch o.Tags()[0] {
				case TagWall:
					dy = 0
					if p.Falling {
						p.State = playerFallendwall
					}
					if p.Slipping {
						p.State = playerSlipend
					}
					if p.Jumping {
						p.State = playerJumpendwall
					}
				case TagClimbable:
					if p.State == playerFallloop {
						p.State = playerFallendfloor
					}
					if p.State == playerSliploop {
						p.State = playerSlipend
					}
				}
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

	case playerJumpstart, playerJumploop:
		if (p.Input.ActionIsPressed(ActionJump) || p.JumpTime < MinJumpTime) && p.JumpTime < MaxJumpTime {
			p.State = playerJumploop
		} else {
			p.State = playerJumpendfloor
		}

	case playerJumpendwall, playerJumpendfloor, playerJumpendmantle:
		p.State = playerIdle
		p.Jumping = false
		p.JumpTime = 0

	case playerFallstart:
		p.State = playerFallloop

	case playerFallendwall:
		p.State = playerStand
		p.Falling = false

	case playerFallendfloor:
		p.State = playerIdle
		p.Falling = false

	case playerSlipstart:
		p.State = playerSliploop

	case playerSlipend:
		p.State = playerIdle
		p.Slipping = false

	}
}

func (p *Player) insideOf(o *resolv.Object) bool {
	if o.Shape == nil {
		return false
	}

	verts := p.Shape.(*resolv.ConvexPolygon).Transformed()
	for _, v := range verts {
		if !o.Shape.(*resolv.ConvexPolygon).PointInside(v) {
			return false
		}
	}
	return true
}

func (p *Player) Draw(camera *camera.Camera) {
	op := &ebiten.DrawImageOptions{}

	s := p.Sprite
	frame := s.Sprite[p.Frame]
	img := s.Image.SubImage(image.Rect(
		frame.Position.X,
		frame.Position.Y,
		frame.Position.X+frame.Position.W,
		frame.Position.Y+frame.Position.H,
	)).(*ebiten.Image)

	// Centre sprite
	op.GeoM.Translate(
		float64(-frame.Position.W/4),
		float64(-frame.Position.H/4),
	)

	camera.Surface.DrawImage(img, camera.GetTranslation(op, p.X, p.Y))
}
