package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Fog struct {
	Image  *ebiten.Image
	Offset float64
	Tick   float64
}

func NewFog() *Fog {
	return &Fog{Image: loadImage("assets/backdrop/Project-scale-parallax-backdrop_0001_Smog-1-cloud.png"), Tick: 0}
}

func (f *Fog) Update() {
	f.Tick++
	f.Offset = math.Sin(f.Tick*2*math.Pi/20000) * 500
}

func (f *Fog) GetDrawImageOptions() *ebiten.DrawImageOptions {
	fogOp := &ebiten.DrawImageOptions{}
	fogOp.GeoM.Translate(f.Offset, 0)
	fogOp.Blend = ebiten.BlendLighter
	return fogOp
}
