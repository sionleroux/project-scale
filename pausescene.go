package main

import (
	"errors"

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
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return gameOver, errors.New("game quit by player")
	}
	return gamePaused, nil
}

func (p *PauseScreen) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Game paused\nPress P to unpause\nPress Esc to quit")
}

func (p *PauseScreen) Load(prev SceneIndex) {
}
