package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
)

type StartScene struct {
	BaseScene
	BackgroundSprite *SpriteSheet
	TextRenderer     *StartTextRenderer
	Frame            int
	Tick             int
	Transition       bool
}

type StartTextRenderer struct {
	*etxt.Renderer
	alpha uint8
}

func (s *StartScene) Update() error {
	s.Tick++
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.Transition = true
	}

	if s.Transition {
		s.Frame = Animate(s.Frame, s.Tick, s.BackgroundSprite.Meta.FrameTags[1])
		if s.Frame == s.BackgroundSprite.Meta.FrameTags[1].To {
			s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
		}
	}
	return nil
}

func (s *StartScene) Draw(screen *ebiten.Image) {
	sprite := s.BackgroundSprite
	frame := sprite.Sprite[s.Frame]

	screen.DrawImage(sprite.Image.SubImage(image.Rect(
		frame.Position.X,
		frame.Position.Y,
		frame.Position.X+frame.Position.W,
		frame.Position.Y+frame.Position.H,
	)).(*ebiten.Image), &ebiten.DrawImageOptions{})

	if !s.Transition {
		s.TextRenderer.Draw(screen, "Press SPACE to start")
	}
}

func NewStartScene() *StartScene {
	return &StartScene{
		BackgroundSprite: loadSprite("Menu"),
		TextRenderer:     NewStartTextRenderer(),
		Frame:            0,
	}

}

func NewStartTextRenderer() *StartTextRenderer {
	font := loadFont("assets/fonts/PixelOperator8-Bold.ttf")
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetSizePx(8)
	return &StartTextRenderer{r, 0xff}
}

func (r StartTextRenderer) Draw(screen *ebiten.Image, text string) {
	r.SetTarget(screen)
	r.SetColor(color.RGBA{0xff, 0xff, 0xff, r.alpha})
	r.Renderer.Draw(text, screen.Bounds().Dx()/2, screen.Bounds().Dy()/8*7)
}
