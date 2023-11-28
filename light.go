package main

import (
	"image/color"
	"path"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sinisterstuf/project-scale/camera"
)

var (
	lightGood = color.NRGBA{0, 255, 0, 100}
	lightWarn = color.NRGBA{255, 255, 0, 100}
	lightBad  = color.NRGBA{255, 0, 0, 100}
)

type Vec struct {
	X float64
	Y float64
}

var lightFacing []Vec = []Vec{
	{0, -1},
	{1, 0},
	{0, 1},
	{-1, 0},
}

// distance to offset sprite from under player
const lightOffset = 8

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

func (l *Light) Draw(cam *camera.Camera, dir Direction) {
	op := &ebiten.DrawImageOptions{}
	op = cam.GetTranslation(op, l.X, l.Y)
	op.GeoM.Translate(l.Offset, l.Offset) // centring
	op.GeoM.Translate(
		lightFacing[dir].X*lightOffset,
		lightFacing[dir].Y*lightOffset,
	)
	op.ColorScale.ScaleWithColor(l.Color)
	op.Blend = ebiten.BlendLighter
	cam.Surface.DrawImage(l.Sprite, op)
}
