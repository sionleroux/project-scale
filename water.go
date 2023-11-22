package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/sinisterstuf/project-scale/camera"
)

const WaterSpeed = 0.35

type Water struct {
	Level float64
	Image *ebiten.Image
}

func NewWater(startLevel float64) *Water {
	return &Water{
		Level: startLevel,
		Image: loadImage("assets/backdrop/Project-scale-parallax-backdrop_0000_Water-1.png"),
	}
}

func (w *Water) Update() {
	if !ebiten.IsKeyPressed(ebiten.KeyM) {
		w.Level -= WaterSpeed
	}
}

func (w *Water) Draw(cam *camera.Camera) {
	// backdropPos := cam.GetTranslation(&ebiten.DrawImageOptions{}, -float64(bs[0].Image.Bounds().Dx())/2, 0)
	_, camY := cam.GetScreenCoords(0, w.Level)
	ebitenutil.DrawRect(cam.Surface, 0, camY, float64(cam.Width), float64(cam.Height), color.Black)
	ebitenutil.DrawLine(cam.Surface, 0, camY, float64(cam.Width), camY, color.White)
}
