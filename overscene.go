package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// OverScene is shown when the player dies and the game is over
type OverScene struct {
	BaseScene
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
	screen.DrawImage(s.State.lastRender, &ebiten.DrawImageOptions{})

	s.State.TextRenderer.Draw(screen, "You died!", color.White, 8, 50, 10)
	s.State.TextRenderer.Draw(screen, "Press space to restart\nPress Esc to quit", color.White, 8, 50, 80)

	if s.State.Stat.HighestPoint == s.State.Stat.LastHighestPoint {
		s.State.BoldTextRenderer.Draw(screen, fmt.Sprintf(
			"NEW HIGH SCORE!\n\nYou reached %d m",
			s.State.Stat.HighestPoint,
		), color.RGBA{255, 255, 0, 255}, 8, 50, 40)
	} else {
		s.State.TextRenderer.Draw(screen, fmt.Sprintf(
			"Your last climb: %d m\nYour best climb so far: %d m",
			s.State.Stat.LastHighestPoint, s.State.Stat.HighestPoint,
		), color.White, 8, 50, 40)
	}
}
