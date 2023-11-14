package camera

import (
	ebicam "github.com/melonfunction/ebiten-camera"
)

func NewCamera(w, h int) *Camera {
	return &Camera{
		ebicam.NewCamera(w, h, 0, 0, 0, 1),
		NewShaker(),
	}
}

type Camera struct {
	*ebicam.Camera
	shaker *Shaker
}

func (cam *Camera) Update() {
	cam.MovePosition(cam.shaker.calcShake())
}

func (cam *Camera) Shake() {
	cam.shaker.Ease.Reset()
}
