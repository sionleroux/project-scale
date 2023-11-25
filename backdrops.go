package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sinisterstuf/project-scale/camera"
)

func NewBackdrops(bottomOfMap float64) Backdrops {
	return Backdrops{
		bottomOfMap,
		[]Backdrop{
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0015_Background.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0014_Sky.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0013_Smog-5.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0012_Water-5.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0011_City-4.png"), false, 0.5},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0010_Smog-4.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0009_Water-4.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0008_City-3.png"), false, 0.35},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0007_Smog-3.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0006_Water-3.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0005_City-2.png"), false, 0.2},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0004_Smog-2.png"), false, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0003_Water-2.png"), true, 0.0},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0002_City-1.png"), false, 0.05},
			{loadImage("assets/backdrop/Project-scale-parallax-backdrop_0001_Smog-1.png"), false, 0.0},
		},
	}
}

type Backdrops struct {
	bottomOfMap float64
	Backdrops   []Backdrop
}

func (bs Backdrops) Draw(cam *camera.Camera, waterLevel float64) {
	const waterSpacing = 8.0
	const howManyWaters = 5.0 // I counted them by hand.0
	backdropCenter := -float64(bs.Backdrops[0].Image.Bounds().Dx()) / 2
	watersDone := 0.0

	for _, b := range bs.Backdrops {
		if b.Water {
			cam.Surface.DrawImage(b.Image, cam.GetTranslation(
				&ebiten.DrawImageOptions{},
				backdropCenter,
				waterLevel-(howManyWaters-watersDone)*waterSpacing,
			))
			watersDone++
		} else {

			// Due to parallax effect, the further the backdrop is the slower it moves.
			// Thus in order to reach the top of all the backdrops at the top,
			// they needs to be scaled down according to their speed

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-float64(b.Image.Bounds().Dx())/2, 0)
			op.GeoM.Scale(1-b.Speed, 1-b.Speed)
			op.GeoM.Translate(float64(b.Image.Bounds().Dx())/2, 0)
			backdropPos := cam.GetTranslation(
				op,
				// Use the actual horizontal position of the backgrop, and
				// change it based on the camera position and the speed of the backdrop
				-float64(b.Image.Bounds().Dx())/2*(1-b.Speed)+(cam.X-float64(cam.Width/2))*b.Speed,
				// Use the actual height of the backdrop after scaling it down
				-float64(b.Image.Bounds().Dy())*(1-b.Speed)+bs.bottomOfMap+
					// The position is changed based on the camera position and the speed of the backdrop
					(float64(cam.Height/2)+cam.Y-bs.bottomOfMap)*b.Speed,
			)

			cam.Surface.DrawImage(b.Image, backdropPos)
		}
	}
}

type Backdrop struct {
	Image *ebiten.Image
	Water bool
	Speed float64
}
