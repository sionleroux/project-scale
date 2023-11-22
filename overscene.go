package main

import (
	"errors"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SceneOver struct {
	// TODO: some high score count here
}

func (s *SceneOver) Update() (SceneIndex, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return gameRunning, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return gameOver, errors.New("game quit by player")
	}
	return gameOver, nil
}

func (s *SceneOver) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "You died\nPress space to restart\nPress Esc to quit")
}

func (s *SceneOver) Load() {
	log.Println("Game over")
}
