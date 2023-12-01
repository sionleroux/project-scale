package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// OverScene is shown when the player dies and the game is over
type OverScene struct {
	BaseScene
	Menu *Menu
}

func (s *OverScene) Update() error {
	s.Menu.Update()
	if s.State.Input.ActionIsJustPressed(ActionPrimary) {
		if s.Menu.Active == 0 {
			s.State.ResetNeeded = true
			s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
		} else if s.Menu.Active == 1 {
			s.SceneManager.SwitchTo(s.State.Scenes[gameStart])
		}
	}
	return nil
}

func (s *OverScene) Draw(screen *ebiten.Image) {
	screen.DrawImage(s.State.lastRender, &ebiten.DrawImageOptions{})

	s.State.TextRenderer.Draw(screen, "You died!", color.White, 8, 50, 10)

	s.Menu.Draw(screen)

	if s.State.Stat.HighestPoint == s.State.Stat.LastHighestPoint {
		s.State.BoldTextRenderer.Draw(screen, fmt.Sprintf(
			"NEW HIGH SCORE!\n\nYou reached %d m",
			s.State.Stat.HighestPoint,
		), color.RGBA{255, 255, 0, 255}, 8, 50, 40)
	} else {
		if s.State.Stat.FastestRound > 0 {
			s.State.TextRenderer.Draw(screen, fmt.Sprintf(
				"Your last climb: %d m\nYour best climb so far: %d m\nYour fastest victory: %d min %d sec",
				s.State.Stat.LastHighestPoint, s.State.Stat.HighestPoint, int(s.State.Stat.FastestRound/60), int(s.State.Stat.FastestRound)%60,
			), color.White, 8, 50, 40)

		} else {
			s.State.TextRenderer.Draw(screen, fmt.Sprintf(
				"Your last climb: %d m\nYour best climb so far: %d m",
				s.State.Stat.LastHighestPoint, s.State.Stat.HighestPoint,
			), color.White, 8, 50, 40)
		}
	}
}
