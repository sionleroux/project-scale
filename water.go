package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/sinisterstuf/project-scale/camera"
)

const WaterSpeed = 0.35

type Water struct {
	Level float64
}

func NewWater(startLevel float64) *Water {
	return &Water{
		Level: startLevel,
	}
}

func (w *Water) Update() {
	w.Level -= WaterSpeed
}

func (w *Water) Draw(cam *camera.Camera) {
	_, camY := cam.GetScreenCoords(0, w.Level)
	ebitenutil.DrawRect(cam.Surface, 0, camY, float64(cam.Width), float64(cam.Height), color.Black)
	ebitenutil.DrawLine(cam.Surface, 0, camY, float64(cam.Width), camY, color.White)
}
