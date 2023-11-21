package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PauseScreen is shown when the game is paused
type PauseScreen struct {
}

func (p *PauseScreen) Update() (SceneIndex, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		return gameRunning, nil
	}
	return gamePaused, nil
}

func (p *PauseScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Game paused\nPress P to unpause")
}

func (p *PauseScreen) Load() {
}
