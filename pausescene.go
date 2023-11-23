package main

import (
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PauseScreen is shown when the game is paused
type PauseScreen struct {
	BaseScene
}

func (p *PauseScreen) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		p.SceneManager.SwitchTo(p.State.Scenes[gameRunning])
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	return nil
}

func (p *PauseScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Game paused\nPress P to unpause\nPress Esc to quit")
}
