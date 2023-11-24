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
	ButtonSprite     *SpriteAnimation
	TextRenderer     *StartTextRenderer
	TransitionPhase  int
}

type StartTextRenderer struct {
	*etxt.Renderer
	alpha uint8
}

func (s *StartScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.TransitionPhase = 1
	}

	if s.TransitionPhase == 1 {
		if s.ButtonSprite.Update(1) {
			s.TransitionPhase = 2
		}
	}
	if s.TransitionPhase == 2 {

		if s.BackgroundSprite.Update(1) {
			s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
		}
	}
	return nil
}

func (s *StartScene) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.BackgroundSprite.GetImage(), &ebiten.DrawImageOptions{})

	if s.TransitionPhase < 2 {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(s.State.Width/2), float64(s.State.Height/2))
		screen.DrawImage(s.ButtonSprite.GetImage(), op)
	}

	if s.TransitionPhase == 0 {
		s.TextRenderer.Draw(screen, "Press SPACE to start")
	}
}

func NewStartScene() *StartScene {
	return &StartScene{
		BackgroundSprite: NewSpriteAnimation("Menu"),
		ButtonSprite:     NewSpriteAnimation("Start button"),
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
