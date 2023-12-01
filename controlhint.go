package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sinisterstuf/project-scale/camera"
)

type HintState int8

const (
	hintHidden HintState = iota
	hintVisible
	hintFading
	hintFaded
)

type ControlHint struct {
	Sprite   *SpriteAnimation
	FrameTag int
	From     float64
	To       float64
	Dx       float64
	Dy       float64
	State    HintState
	Alpha    float64
}

func (c *ControlHint) Update(y float64) error {
	if c.State == hintHidden {
		if y <= c.From {
			c.State = hintVisible
			c.Alpha = 1
		}
	} else if c.State == hintVisible {
		c.Sprite.Update(c.FrameTag)
		if y < c.To || y >= c.From {
			c.State = hintFading
		}
	} else if c.State == hintFading {
		c.Sprite.Update(c.FrameTag)
		c.Alpha -= 0.1
		if c.Alpha <= 0 {
			if y >= c.From {
				c.State = hintHidden
			} else {
				c.State = hintFaded
			}
		}
	}

	return nil
}

func (c *ControlHint) Draw(x, y float64, camera *camera.Camera) {
	if c.State == hintVisible || c.State == hintFading {
		cOp := &ebiten.DrawImageOptions{}
		cOp.GeoM.Scale(0.5, 0.5)
		cOp.ColorScale.ScaleAlpha(float32(c.Alpha))
		camera.Surface.DrawImage(c.Sprite.GetImage(), camera.GetTranslation(cOp, x+(c.Dx), y+c.Dy))
	}
}
