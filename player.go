package main

import (
	"image"
	"image/color"
	"path"

	"github.com/sinisterstuf/project-scale/camera"

	"github.com/hajimehoshi/ebiten/v2"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/resolv"
)

//go:generate ./tools/gen_sprite_tags.sh assets/sprites/Nanobot.json player_anim.go player

const MinJumpTime = 6
const MaxJumpTime = 10

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
	Camera   *camera.Camera
	Light    *Light
}

func NewPlayer(position []int, camera *camera.Camera) *Player {
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
		Camera: camera,
		Light:  NewLight(),
	}
}

func (p *Player) Update() {
	p.Tick++
	if p.State == playerJumploop {
		p.JumpTime++
	}
	p.collisionChecks()
	p.updateMovement()
	p.Light.SetPos(p.X, p.Y)
	p.Light.SetColor(p.State)
	p.animate()
	p.Object.Update()
}

func (p *Player) updateMovement() {
	speed := 1.2

	if !p.Falling && !p.Jumping && p.Input.ActionIsJustPressed(ActionJump) {
		p.Jumping = true
		p.State = playerJumpstart
	}

	if p.Jumping {
		switch p.State {
		case playerJumploop:
			speed = 4.0
		case playerJumpstart:
			speed = -0.3
		case playerJumpendfloor:
			speed = 0.2
		}
		p.move(+0, -speed)
	} else if p.Falling {
		switch p.State {
		case playerFallloop:
			speed = 6.0
		case playerFallstart:
			speed = 0.2
		case playerFallendfloor:
			speed = -0.1
		case playerFallendwall:
			speed = 0.1
		}
		p.move(+0, speed)
	} else if p.Slipping {
		speed = 2.0
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
			if intersection := p.Shape.Intersection(0, dy, o.Shape); intersection != nil {
				switch o.Tags()[0] {
				case TagWall:
					dy = 0
					if p.Falling {
						p.State = playerFallendwall
					}
					if p.Slipping {
						p.State = playerSlipend
					}
					if p.Jumping && p.State != playerJumpendwall {
						p.State = playerJumpendwall
						p.Camera.Shake()
					}
				case TagClimbable:
					// only recover onto tiles below you, that means the MTV to
					// get out of them will be negative, i.e. upwards
					if intersection.MTV.Y() < 0 {
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
	p.Light.Draw(camera)

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

var (
	lightGood = color.NRGBA{0, 255, 0, 100}
	lightWarn = color.NRGBA{255, 255, 0, 100}
	lightBad  = color.NRGBA{255, 0, 0, 100}
)

func NewLight() *Light {
	sprite := loadImage(path.Join("assets", "light.png"))
	const lightWidth = 32       // the PNG is 32px wide, trust me
	const playerCenter = 16 / 4 // 🙄 I didn't feel like passing in player
	return &Light{
		Sprite: sprite,
		Offset: -lightWidth/2 + playerCenter, // un-offset by the player centre
		Color:  lightGood,
	}
}

type Light struct {
	On     bool
	Sprite *ebiten.Image
	X, Y   float64
	Offset float64
	Color  color.Color
}

func (l *Light) SetPos(x, y float64) {
	l.X, l.Y = x, y
}

func (l *Light) SetColor(state playerAnimationTags) {
	switch state {
	case playerFallstart,
		playerFallloop,
		playerFallendwall,
		playerFallendfloor,
		playerJumpendwall:
		l.Color = lightBad
	case playerSlipend,
		playerSlipstart,
		playerSliploop:
		l.Color = lightWarn
	default:
		l.Color = lightGood
	}
}

func (l *Light) Draw(cam *camera.Camera) {
	op := cam.GetTranslation(&ebiten.DrawImageOptions{}, l.X, l.Y)
	op.GeoM.Translate(l.Offset, l.Offset) // centring
	op.ColorScale.ScaleWithColor(l.Color)
	op.Blend = ebiten.BlendLighter
	cam.Surface.DrawImage(l.Sprite, op)
}
