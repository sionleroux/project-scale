package main

import (
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	ebitenutil.DebugPrint(screen, "You died\nPress space to restart\nPress Esc to quit")
}
