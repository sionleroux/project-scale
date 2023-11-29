package main

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// OverScene is shown when the player dies and the game is over
type OverScene struct {
	BaseScene
	// TODO: some high score count here
}

func (s *OverScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.State.ResetNeeded = true
		s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	return nil
}

func (s *OverScene) Draw(screen *ebiten.Image) {
	s.State.TextRenderer.Draw(screen, "You died!", color.White, 8, 50, 10)
	s.State.TextRenderer.Draw(screen, "Press space to restart\nPress Esc to quit", color.White, 8, 50, 80)
	s.State.TextRenderer.Draw(screen, s.State.Stat.GetText(), color.White, 8, 50, 50)
}
