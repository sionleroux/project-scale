package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"path"

	"github.com/sinisterstuf/project-scale/camera"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quartercastle/vector"
	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/resolv"
)

//go:generate ./tools/gen_sprite_tags.sh assets/sprites/Nanobot.json player_anim.go player

const MinJumpDist = 32 - 4 // I it's because 4 is the distance from the player sprite origin to the collision object or maybe it's because 4 is the current jump movement distance and a fencepost error means it has already moved once by 4 before the check happens
const MaxJumpDist = 48 - 4 // either way 4 is the value that seems to take you the right distance to the next tile in practice

const (
	ActionMoveUp input.Action = iota
	ActionMoveLeft
	ActionMoveDown
	ActionMoveRight
	ActionJump
)

type Direction int8

const (
	directionUp Direction = iota
	directionRight
	directionDown
	directionLeft
)

// Player is the player character in the game
type Player struct {
	*resolv.Object
	Input        *input.Handler
	State        playerAnimationTags
	Sprite       *SpriteSheet
	Frame        int
	Tick         int
	Jumping      bool
	Falling      bool
	Slipping     bool
	Standing     bool
	JumpFrom     vector.Vector
	WhatTile     string
	Camera       *camera.Camera
	Light        *Light
	Facing       Direction
	ControlHints []*ControlHint
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

	hints := make([]*ControlHint, 1)
	hints[0] = &ControlHint{Sprite: NewSpriteAnimation("Controls"), FrameTag: 0, From: 2536, To: 2500, Dx: -8, Dy: -8}
	// Space hint can be added with proper From/To values
	// hints[1] = &ControlHint{Sprite: NewSpriteAnimation("Controls"), FrameTag: 1, From: 2400, To: 2300, Dx: -8, Dy: 8}

	return &Player{
		Object:       object,
		Sprite:       loadSpriteWithOSOverride("Nanobot"),
		ControlHints: hints,
		Camera:       camera,
		Light:        NewLight(),
	}
}

func (p *Player) Update() {
	p.Tick++
	p.collisionChecks()
	p.updateMovement()
	p.Light.SetPos(p.X, p.Y)
	p.Light.SetColor(p.State)
	p.animate()
	p.Object.Update()
	for _, hint := range p.ControlHints {
		hint.Update(p.Y)
	}
}

func (p *Player) updateMovement() {
	speed := 0.0

	if !p.Falling && !p.Jumping && p.Input.ActionIsJustPressed(ActionJump) {
		p.Jumping = true
		p.State = playerJumpstart
		p.JumpFrom = vector.Vector{p.X, p.Y}
	}

	if p.Jumping {
		switch p.State {
		case playerJumploop:
			if (p.Input.ActionIsPressed(ActionJump) || !p.jumpedMin()) && !p.jumpedMax() {
				p.State = playerJumploop
			} else {
				p.State = playerJumpendfloor
			}
			speed = 4.0
		}
		if p.Facing == directionLeft {
			p.move(-speed, +0)
		} else if p.Facing == directionRight {
			p.move(+speed, +0)
		} else if p.Facing == directionUp {
			p.move(+0, -speed)
		} else if p.Facing == directionDown {
			p.move(+0, +speed)
		} else {
			p.move(+0, -speed)
		}

	} else if p.Falling {
		switch p.State {
		case playerFallloop:
			speed = 6.0
		}
		p.move(+0, speed)

	} else if p.Slipping {
		if p.State == playerSliploop {
			speed = 2.0
		} else {
			speed = 0.5
		}
		p.move(+0, speed)

	} else {
		p.State = playerIdle
		speed = 1.2
		if p.Input.ActionIsPressed(ActionMoveLeft) {
			p.move(-speed, +0)
			p.State = playerClimbup
			p.Facing = directionLeft
		} else if p.Input.ActionIsPressed(ActionMoveRight) {
			p.move(+speed, +0)
			p.State = playerClimbup
			p.Facing = directionRight
		} else if p.Input.ActionIsPressed(ActionMoveUp) {
			p.move(+0, -speed)
			p.State = playerClimbup
			p.Facing = directionUp
		} else if p.Input.ActionIsPressed(ActionMoveDown) {
			p.move(+0, +speed)
			p.State = playerClimbup
			p.Facing = directionDown
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
					switch o.Tags()[0] {
					case TagChasm:
						p.State = playerFallstart
						p.Falling = true
						p.Facing = directionUp
					case TagSlippery:
						p.State = playerSlipstart
						p.Slipping = true
						p.Facing = directionUp
					}
				}
			}
		}
	}
}

func (p *Player) move(dx, dy float64) {
	if collision := p.Check(dx, 0, TagWall); collision != nil {
		for _, o := range collision.Objects {
			if intersection := p.Shape.Intersection(dx, 0, o.Shape); intersection != nil {
				dx = 0
				if intersection.MTV.X() != 0 {
					log.Println("MTV X:", intersection.MTV.X())
				}
			}
		}
	}
	p.X += dx

	if collision := p.Check(0, dy, TagWall, TagClimbable, TagChasm); collision != nil {
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
						p.Camera.Shake(camera.NewShaker(10, 40, 10))
						dy -= intersection.MTV.Y()
						if intersection.MTV.Y() != 0 {
							log.Println("MTV Y:", intersection.MTV.Y())
						}
					}
				case TagChasm:
					if dy < 0 && !p.Jumping { // Don't climb up into chasm
						dy = 0
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

	case playerJumpstart:
		p.State = playerJumploop

	case playerJumploop:
		p.State = playerJumploop

	case playerJumpendwall, playerJumpendfloor, playerJumpendmantle:
		p.Jumping = false

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

func (p *Player) jumpedMax() bool {
	return math.Abs(p.jumpDistance().Magnitude()) >= MaxJumpDist
}

func (p *Player) jumpedMin() bool {
	return math.Abs(p.jumpDistance().Magnitude()) >= MinJumpDist
}

func (p *Player) jumpDistance() vector.Vector {
	return vector.Vector{p.X, p.Y}.Sub(p.JumpFrom)
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

	// Rotate
	op.GeoM.Translate(
		float64(-frame.Position.W/2),
		float64(-frame.Position.H/2),
	)
	op.GeoM.Rotate(math.Pi / 2 * float64(p.Facing))
	op.GeoM.Translate(
		float64(+frame.Position.W/2),
		float64(+frame.Position.H/2),
	)

	// Centre sprite on object centre
	op.GeoM.Translate(
		float64(-frame.Position.W/4),
		float64(-frame.Position.H/4),
	)

	camera.Surface.DrawImage(img, camera.GetTranslation(op, p.X, p.Y))

	for _, hint := range p.ControlHints {
		hint.Draw(p.X, p.Y, camera)
	}
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
