package main

import (
	"image/color"
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
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}

	s.State.Water.Update(false)

	return nil
}

func (s *WonScene) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.State.lastRender, &ebiten.DrawImageOptions{})

	s.State.TextRenderer.Draw(screen, "CONGRATS!", color.White, 8, 50, 10)
	s.State.TextRenderer.Draw(screen, "Press space to restart\nPress Esc to quit", color.White, 8, 50, 80)
}
