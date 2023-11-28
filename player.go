package main

import (
	"image"
	"log"
	"math"

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
	speedClimb     = 1.2
	speedJump      = 4.0
	speedFall      = 6.0
	speedSliploop  = 2.0
	speedSlip      = 0.2
	speedDeathFall = 0.5
)

const (
	ActionMoveUp input.Action = iota
	ActionMoveLeft
	ActionMoveDown
	ActionMoveRight
	ActionJump
)

type Direction float64

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
	SpeedX       float64
	SpeedY       float64
	ControlHints []*ControlHint
	Dying        bool
	Dead         bool
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
	if !p.Dying {
		p.updateMovement()
		p.collisionChecks()
	} else {
		p.updateDeath()
	}
	p.Light.SetPos(p.X, p.Y)
	p.Light.SetColor(p.State)
	p.animate()
	for _, hint := range p.ControlHints {
		hint.Update(p.Y)
	}
}

func (p *Player) updateDeath() {
	p.Y += speedDeathFall
	p.Facing += 0.02
	p.Object.Update()
}

func (p *Player) updateMovement() {

	// State-based continued movement
	switch p.State {

	case playerJumploop:
		if (p.Input.ActionIsPressed(ActionJump) || !p.jumpedMin()) && !p.jumpedMax() {
			p.State = playerJumploop
		} else {
			p.State = playerJumpendfloor
		}
		if p.Facing == directionLeft {
			p.SpeedX, p.SpeedY = -speedJump, 0
		} else if p.Facing == directionRight {
			p.SpeedX, p.SpeedY = +speedJump, 0
		} else if p.Facing == directionUp {
			p.SpeedX, p.SpeedY = 0, -speedJump
		} else if p.Facing == directionDown {
			p.SpeedX, p.SpeedY = 0, +speedJump
		}

	case playerFallloop:
		p.SpeedX, p.SpeedY = 0, speedFall
	case playerSliploop:
		p.SpeedX, p.SpeedY = 0, speedSliploop
	// case playerSlipend, playerSlipstart:
	// 	p.SpeedX, p.SpeedY = 0, speedSlip
	default:
		p.SpeedX, p.SpeedY = 0, 0
	}

	// Jump input
	if !p.Falling && !p.Jumping && p.Input.ActionIsJustPressed(ActionJump) {
		p.Jumping = true
		p.State = playerJumpstart
		p.JumpFrom = vector.Vector{p.X, p.Y}
	}

	// Climbing input
	if !p.Jumping && !p.Falling && !p.Slipping {
		if p.Input.ActionIsPressed(ActionMoveLeft) {
			p.SpeedX, p.SpeedY = -speedClimb, 0
			p.State = playerClimb
			p.Facing = directionLeft
		} else if p.Input.ActionIsPressed(ActionMoveRight) {
			p.SpeedX, p.SpeedY = +speedClimb, 0
			p.State = playerClimb
			p.Facing = directionRight
		} else if p.Input.ActionIsPressed(ActionMoveUp) {
			p.SpeedX, p.SpeedY = 0, -speedClimb
			p.State = playerClimb
			p.Facing = directionUp
		} else if p.Input.ActionIsPressed(ActionMoveDown) {
			p.SpeedX, p.SpeedY = 0, +speedClimb
			p.State = playerClimb
			p.Facing = directionDown
		} else {
			p.State = playerIdle
			p.SpeedX, p.SpeedY = 0, 0
		}
	}

}

func (p *Player) collisionChecks() {
	if collision := p.Check(p.SpeedX, p.SpeedY); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(0, 0, o.Shape) != nil {
				p.WhatTile = o.Tags()[0]
			}
		}
	}

	dx := p.SpeedX
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

	dy := p.SpeedY
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
					log.Println("MTV WOOP:", intersection.MTV.Y())
					if intersection.MTV.Y() < 0 {
						log.Println("AAAAAAAAAAAA")
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

	// Start falling if you're stepping on a chasm
	if p.State != playerJumploop && !p.Falling && !p.Slipping {
		if collision := p.Check(dx, dy, TagChasm, TagSlippery); collision != nil {
			for _, o := range collision.Objects {
				if p.Shape.Intersection(dx, dy, o.Shape) != nil || p.insideOf(o) {
					p.Jumping = false
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

	p.Object.Update()
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
