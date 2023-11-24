package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/joelschutz/stagehand"
	"github.com/tinne26/etxt"
)

type StartScene struct {
	BaseScene
	BackgroundSprite *SpriteAnimation
	ButtonSprite     *SpriteAnimation
	TransitionPhase  int
	Music            *MusicLoop
	Voice            Sound
}

type StartTextRenderer struct {
	*etxt.Renderer
	alpha uint8
}

func (s *StartScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.TransitionPhase = 1
		s.Music.Pause()
		s.Voice.Play()
	}

	if s.TransitionPhase == 1 {
		if s.ButtonSprite.Update(1) {
			s.TransitionPhase = 2
		}
	}
	if s.TransitionPhase == 2 {

		if s.BackgroundSprite.Update(1) {
			s.State.ResetNeeded = true
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
		s.State.TextRenderer.Draw(screen, "Press SPACE to start", 8, 50, 85)
	}
}

func NewStartScene() *StartScene {
	voice := Sound{Volume: 0.5}
	voice.AddSound("assets/voices/Start_button", sampleRate, context)

	bgMusic := NewMusicPlayer(loadSoundFile("assets/music/Start_menu_idle.ogg", sampleRate))
	bgMusic.SetVolume(0.7)

	return &StartScene{
		BackgroundSprite: NewSpriteAnimation("Menu"),
		ButtonSprite:     NewSpriteAnimation("Start button"),
		Music:            bgMusic,
		Voice:            voice,
	}
}

func (s *StartScene) Load(st State, sm *stagehand.SceneManager[State]) {
	s.BaseScene.Load(st, sm)
	s.Music.Play()
}
