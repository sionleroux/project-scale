package main

import (
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// WonScreen is shown when the game is won
type WonScene struct {
	BaseScene
	// TODO: maybe a lap time?
}

func (s *WonScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.State.ResetNeeded = true
		s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	return nil
}

func (s *WonScene) Draw(screen *ebiten.Image) {
	s.State.TextRenderer.Draw(screen, "You WON!", 8, 50, 10)
	s.State.TextRenderer.Draw(screen, "Press space to restart\nPress Esc to quit", 8, 50, 80)
	s.State.TextRenderer.Draw(screen, s.State.Stat.GetText(), 8, 50, 50)
}
