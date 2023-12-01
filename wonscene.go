package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// WonScreen is shown when the game is won
type WonScene struct {
	BaseScene
	Menu *Menu
}

func (s *WonScene) Update() error {
	s.Menu.Update()
	if s.State.Input.ActionIsJustPressed(ActionPrimary) {
		if s.Menu.Active == 0 {
			s.State.ResetNeeded = true
			s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
		} else if s.Menu.Active == 1 {
			s.SceneManager.SwitchTo(s.State.Scenes[gameStart])
		}
	}
	s.State.Water.Update(false)

	return nil
}

func (s *WonScene) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.State.lastRender, &ebiten.DrawImageOptions{})

	s.State.TextRenderer.Draw(screen, "CONGRATS!", color.White, 8, 50, 10)
	s.State.TextRenderer.Draw(screen, fmt.Sprintf(
		"Your last round: %d min %d sec\nYour fastest round: %d min %d sec",
		int(s.State.Stat.LastRound/60), int(s.State.Stat.LastRound)%60, int(s.State.Stat.FastestRound/60), int(s.State.Stat.FastestRound)%60,
	), color.White, 8, 50, 40)

	s.Menu.Draw(screen)
}
