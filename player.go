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
const MaxJumpDist = 64 - 4 // either way 4 is the value that seems to take you the right distance to the next tile in practice

const (
	speedClimb     = 1.2
	speedJump      = 4.0
	speedFall      = 6.0
	speedSliploop  = 2.0
	speedSlip      = 0.2
	speedDeathFall = 0.5
)

type Direction int8

const (
	directionUp Direction = iota
	directionRight
	directionDown
	directionLeft
)

type PlayerState int8

const (
	stateIdle = iota
	stateJumping
	stateFalling
	stateSlipping
	stateStanding
	stateDying
	stateDead
	stateWinning
	stateWon
)

var playerStateNames = []string{
	"Idle",
	"Jumping",
	"Falling",
	"Slipping",
	"Standing",
	"Dying",
	"Dead",
	"Winning",
	"Won",
}

// Player is the player character in the game
type Player struct {
	*resolv.Object
	Input        *input.Handler
	AnimState    playerAnimationTags
	State        PlayerState
	Sprite       *SpriteSheet
	Frame        int
	Tick         int
	JumpFrom     vector.Vector
	WhatTiles    []string
	Camera       *camera.Camera
	Light        *Light
	Facing       Direction
	Rotation     float64
	SpeedX       float64
	SpeedY       float64
	ControlHints []*ControlHint
}

func NewPlayer(position []int, camera *camera.Camera) *Player {
	object := resolv.NewObject(
		float64(position[0]), float64(position[1]),
		8, 8,
	)
	object.SetShape(resolv.NewRectangle(
		0, 0, // origin
		8, 8,
	))

	hints := make([]*ControlHint, 2)
	hints[0] = &ControlHint{Sprite: NewSpriteAnimation("Controls"), FrameTag: 0, From: 3232, To: 3120, Dx: -8, Dy: -8}
	hints[1] = &ControlHint{Sprite: NewSpriteAnimation("Controls"), FrameTag: 1, From: 155 * 16, To: 148 * 16, Dx: -8, Dy: 8}

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

	// Early return on death
	if p.State == stateDying || p.State == stateDead {
		p.updateDeath()
		p.animate()
		return
	}
	if p.State == stateWinning || p.State == stateWon {
		return
	}

	p.updateMovement()
	p.collisionChecks()
	p.Light.SetPos(p.X, p.Y)
	p.Light.SetColor(p.AnimState)
	p.animate()
	for _, hint := range p.ControlHints {
		hint.Update(p.Y)
	}
}

func (p *Player) updateDeath() {
	p.Y += speedDeathFall
	p.Rotation += 0.02
	p.Object.Update()
}

func (p *Player) updateMovement() {

	// State-based continued movement
	switch p.AnimState {

	case playerJumploop:
		if (p.Input.ActionIsPressed(ActionPrimary) || !p.jumpedMin()) && !p.jumpedMax() {
			p.AnimState = playerJumploop
		} else {
			p.AnimState = playerJumpendfloor
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
	case playerSlipend, playerSlipstart:
		p.SpeedX, p.SpeedY = 0, speedSlip // XXX: why don't you slip without this?!
	default:
		p.SpeedX, p.SpeedY = 0, 0
	}

	// Walking in 1D input
	if p.State == stateStanding {
		if p.AnimState == playerSwitchtotopview {
			return // no walking while switching
		}

		if p.Input.ActionIsPressed(ActionMoveLeft) {
			p.SpeedX, p.SpeedY = -speedClimb, 0
			p.AnimState = playerWalkleft
		} else if p.Input.ActionIsPressed(ActionMoveRight) {
			p.SpeedX, p.SpeedY = +speedClimb, 0
			p.AnimState = playerWalkright
		} else if p.Input.ActionIsPressed(ActionMoveUp) {
			p.SpeedX, p.SpeedY = 0, -speedClimb
			p.AnimState = playerSwitchtotopview // intent to move up
			p.Facing = directionUp
		} else {
			p.AnimState = playerStand
		}
		return
	}

	// Jump input
	if p.State != stateFalling && (p.State != stateJumping || p.AnimState == playerJumpendfloor) && p.Input.ActionIsJustPressed(ActionPrimary) {
		p.State = stateJumping
		p.AnimState = playerJumpstart
		p.JumpFrom = vector.Vector{p.X, p.Y}
	}

	// Climbing input
	if p.State != stateJumping && p.State != stateFalling && p.State != stateSlipping {
		if p.Input.ActionIsPressed(ActionMoveLeft) {
			p.SpeedX, p.SpeedY = -speedClimb, 0
			p.AnimState = playerClimb
			p.Facing = directionLeft
		} else if p.Input.ActionIsPressed(ActionMoveRight) {
			p.SpeedX, p.SpeedY = +speedClimb, 0
			p.AnimState = playerClimb
			p.Facing = directionRight
		} else if p.Input.ActionIsPressed(ActionMoveUp) {
			p.SpeedX, p.SpeedY = 0, -speedClimb
			p.AnimState = playerClimb
			p.Facing = directionUp
		} else if p.Input.ActionIsPressed(ActionMoveDown) {
			p.SpeedX, p.SpeedY = 0, +speedClimb
			p.AnimState = playerClimb
			p.Facing = directionDown
		} else {
			p.AnimState = playerIdle
			p.SpeedX, p.SpeedY = 0, 0
		}
	}

}

func (p *Player) collisionChecks() {
	if collision := p.Check(p.SpeedX, p.SpeedY); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(0, 0, o.Shape) != nil {
				p.WhatTiles = o.Tags()
			}
		}
	}

	dx := p.SpeedX
	switch p.State {

	case stateStanding:

		// Don't walk into walls
		if collision := p.Check(dx, 0, TagWall); collision != nil {
			for _, o := range collision.Objects {
				if intersection := p.Shape.Intersection(dx, 0, o.Shape); intersection != nil {
					dx = 0
				}
			}
		}

		// // Fall down if there is no more wall beneath you // TODO: fixme!!!
		// // XXX: this messes up the whole standing state!!!
		// if collision := p.Check(p.H, dx, TagWall); collision == nil {
		// 	p.AnimState = playerFallstart
		// 	p.State = stateFalling
		// 	p.Facing = directionUp
		// }

	case stateIdle: // Don't climb into a chasm
		if collision := p.Check(dx, 0, TagChasm); collision != nil {
			for _, o := range collision.Objects {
				if intersection := p.Shape.Intersection(dx, 0, o.Shape); intersection != nil {
					dx = 0
				}
			}
		}
		fallthrough

	default:
		if collision := p.Check(dx, 0, TagWall); collision != nil {
			for _, o := range collision.Objects {
				if intersection := p.Shape.Intersection(dx, 0, o.Shape); intersection != nil {
					dx = intersection.MTV.X()
				}
			}
		}
	}
	p.X += dx

	dy := p.SpeedY
	switch p.State {

	case stateStanding:

		// Successfully climb up if there's a climbable tile behind you
		if dy < 0 { // attempting to climb up
			if collision := p.Check(0, dy, TagClimbable); collision != nil {
				canClimbUp := false
				for _, o := range collision.ObjectsByTags(TagClimbable) {
					if p.Overlaps(o) { // XXX: maybe this isn't even needed?!
						canClimbUp = true
					}
				}
				if !canClimbUp {
					dy = 0 // no climbing for you today
					p.AnimState = playerStand
				}
			}
		}

	default:
		if collision := p.Check(0, dy, TagWall, TagClimbable, TagChasm); collision != nil {
			for _, o := range collision.Objects {
				if intersection := p.Shape.Intersection(0, dy, o.Shape); intersection != nil {

					switch o.Tags()[0] {

					case TagWall:
						dy = 0
						if p.State == stateFalling {
							p.AnimState = playerFallendwall
							log.Println("Avoid wall clipping after fall:", intersection.MTV.X(), intersection.MTV.Y())
							dy -= intersection.MTV.Y()
						}
						if p.State == stateSlipping {
							p.AnimState = playerSlipend
						}
						if p.State == stateJumping && p.AnimState != playerJumpendwall {
							p.AnimState = playerJumpendwall
							p.Camera.Shake(camera.NewShaker(10, 40, 10))
							dy -= intersection.MTV.Y()
							if intersection.MTV.Y() != 0 {
								log.Println("MTV Y:", intersection.MTV.Y())
							}
						}
					case TagChasm:
						if p.State == stateIdle { // Don't climb into chasm
							dy = 0
						}
					case TagClimbable:
						// only recover onto tiles below you, that means the MTV to
						// get out of them will be negative, i.e. upwards
						// log.Println("MTV WOOP:", intersection.MTV.Y())
						if intersection.MTV.Y() < 0 {
							// log.Println("AAAAAAAAAAAA")
							if p.AnimState == playerFallloop {
								p.AnimState = playerFallendfloor
							}
							if p.AnimState == playerSliploop {
								p.AnimState = playerSlipend
							}
						}
					}
				}
			}
		}
		p.Y += dy

	}

	// Start falling if you're stepping on a chasm
	if p.AnimState != playerJumploop && p.State != stateFalling && p.State != stateSlipping {
		if collision := p.Check(dx, dy, TagChasm, TagSlippery); collision != nil {
			for _, o := range collision.Objects {
				if p.Shape.Intersection(dx, dy, o.Shape) != nil || p.insideOf(o) {
					switch o.Tags()[0] {
					case TagChasm:
						p.AnimState = playerFallstart
						p.State = stateFalling
						p.Facing = directionUp
					case TagSlippery:
						p.AnimState = playerSlipstart
						p.State = stateSlipping
						p.Facing = directionUp
					}
				}
			}
		}
	}

	p.Object.Update()
}

func (p *Player) animate() {
	if p.Frame == p.Sprite.Meta.FrameTags[p.AnimState].To {
		p.animationBasedStateChanges()
	}
	p.Frame = Animate(p.Frame, p.Tick, p.Sprite.Meta.FrameTags[p.AnimState])
}

// Animation-trigged state changes
func (p *Player) animationBasedStateChanges() {
	switch p.AnimState {

	case playerJumpstart:
		p.AnimState = playerJumploop

	case playerJumploop:
		p.AnimState = playerJumploop

	case playerJumpendwall, playerJumpendfloor, playerJumpendmantle:
		p.State = stateIdle

	case playerFallstart:
		p.AnimState = playerFallloop

	case playerFallendwall:
		p.AnimState = playerStand
		p.State = stateStanding

	case playerFallendfloor:
		p.AnimState = playerIdle
		p.State = stateIdle

	case playerSwitchtotopview:
		p.AnimState = playerIdle
		p.State = stateIdle

	case playerSlipstart:
		p.AnimState = playerSliploop

	case playerSlipend:
		p.AnimState = playerIdle
		p.State = stateIdle

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

	switch p.State {
	case stateIdle, stateFalling, stateSlipping, stateJumping:
		p.Light.Draw(camera, p.Facing, p.Tick)
	}

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
	if p.State == stateDying {
		op.GeoM.Rotate(math.Pi / 2 * float64(p.Rotation))
	} else {
		op.GeoM.Rotate(math.Pi / 2 * float64(p.Facing))
	}
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
}
