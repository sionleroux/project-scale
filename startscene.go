package main

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/joelschutz/stagehand"
)

type StartScene struct {
	BaseScene
	BackgroundSprite *SpriteAnimation
	ButtonSprite     *SpriteAnimation
	TransitionPhase  int
	Heartbeat        Sound
	Voice            Sound
	Menu             *Menu
}

func (s *StartScene) Update() error {

	if s.TransitionPhase == 0 {

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			if s.Menu.Active == 0 {
				s.TransitionPhase = 1
				s.Heartbeat.Pause()
				s.Voice.Play()
			} else if s.Menu.Active == 1 {
				if ebiten.IsFullscreen() {
					ebiten.SetFullscreen(false)
					s.Menu.Items[1] = "Fullscreen: OFF"
				} else {
					ebiten.SetFullscreen(true)
					s.Menu.Items[1] = "Fullscreen: ON"
				}
			} else if s.Menu.Active == 2 {
				os.Exit(0)
			}

		}

		s.Menu.Update()

		if f, _ := s.ButtonSprite.Update(0); f {
			s.Heartbeat.Play()
		}
	} else if s.TransitionPhase == 1 {
		if _, l := s.ButtonSprite.Update(1); l {
			s.TransitionPhase = 2
		}
	} else if s.TransitionPhase == 2 {

		if _, l := s.BackgroundSprite.Update(1); l {
			s.State.ResetNeeded = true
			s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
		}
	}
	s.State.Fog.Update()

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
		// s.State.BoldTextRenderer.Draw(screen, "Press SPACE to start\nPress ESC to quit", color.Black, 8, 50, 85)
		s.Menu.Draw(screen)
	}

	fogOp := s.State.Fog.GetDrawImageOptions()
	fogOp.GeoM.Translate(float64(-s.State.Fog.Image.Bounds().Dx()+s.State.StartPos[0])/2, -float64(s.State.Fog.Image.Bounds().Dy())+gameHeight)
	screen.DrawImage(s.State.Fog.Image, fogOp)

}

func (s *StartScene) Load(st State, sm *stagehand.SceneManager[State]) {
	s.BaseScene.Load(st, sm)
	s.TransitionPhase = 0
	s.BackgroundSprite.Update(0)
}

func NewStartScene(game *Game) *StartScene {
	voice := Sound{Volume: 0.5}
	voice.AddSound("assets/voices/game-start", sampleRate, context)

	heartbeat := Sound{Volume: 0.7}
	heartbeat.AddSound("assets/sfx/heartbeat", sampleRate, context)

	return &StartScene{
		BackgroundSprite: NewSpriteAnimation("Menu"),
		ButtonSprite:     NewSpriteAnimation("Start button"),
		Heartbeat:        heartbeat,
		Voice:            voice,
		Menu: &Menu{
			Items:         []string{"Start game", "Fullscreen: OFF", "Quit"},
			X:             gameWidth / 2,
			Y:             190,
			color:         color.RGBA{0, 0, 0, 255},
			selectedColor: color.RGBA{255, 255, 0, 255},
			textRenderer:  game.TextRenderer,
		},
	}
}
