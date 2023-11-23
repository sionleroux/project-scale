package main

import (
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// WonScreen is shown when the game is won
type WonScene struct {
	BaseScene
	// TODO: maybe a lap time?
}

func (s *WonScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.SceneManager.SwitchTo(s.State.Scenes[gameRunning])
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	return nil
}

func (s *WonScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "You died\nPress space to restart\nPress Esc to quit")
}
