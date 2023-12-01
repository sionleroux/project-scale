package main

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	s.State.Backdrops.Draw(s.State.Camera, s.State.Water.Level)

	s.State.Camera.Blit(screen)

	vector.DrawFilledRect(screen, 0, 0, float32(s.State.Width), float32(s.State.Height), color.RGBA{0, 0, 0, 200}, false)

	s.State.TextRenderer.Draw(screen, "You WON!", color.White, 8, 50, 10)
	s.State.TextRenderer.Draw(screen, "Press space to restart\nPress Esc to quit", color.White, 8, 50, 80)
	s.State.TextRenderer.Draw(screen, s.State.Stat.GetText(), color.White, 8, 50, 50)
}
