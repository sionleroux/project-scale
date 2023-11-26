package camera

import (
	ebicam "github.com/melonfunction/ebiten-camera"
)

func NewCamera(w, h int) *Camera {
	return &Camera{
		ebicam.NewCamera(w, h, 0, 0, 0, 1), nil,
	}
}

type Camera struct {
	*ebicam.Camera
	shaker *Shaker
}

func (cam *Camera) Update() {
	x, y := 0.0, 0.0
	if cam.shaker != nil && !cam.shaker.Done {
		x, y = cam.shaker.calcShake()
	}
	cam.MovePosition(x, y)
}

func (cam *Camera) Shake(shaker *Shaker) {
	cam.shaker = shaker
	cam.shaker.Ease.Reset()
}
