package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
)

type StartScene struct {
	BaseScene
	BackgroundSprite *SpriteAnimation
	TextRenderer     *StartTextRenderer
	Transition       bool
}

type StartTextRenderer struct {
	*etxt.Renderer
	alpha uint8
}

func (s *StartScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.Transition = true
	}

	if s.Transition {

		if s.BackgroundSprite.Update(1) {
			s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
		}
	}
	return nil
}

func (s *StartScene) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.BackgroundSprite.GetImage(), &ebiten.DrawImageOptions{})

	if !s.Transition {
		s.TextRenderer.Draw(screen, "Press SPACE to start")
	}
}

func NewStartScene() *StartScene {
	return &StartScene{
		BackgroundSprite: NewSpriteAnimation("Menu"),
		TextRenderer:     NewStartTextRenderer(),
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
