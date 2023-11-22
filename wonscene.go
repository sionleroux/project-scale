package main

import (
	"errors"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type SceneWon struct {
	// TODO: maybe a lap time?
}

func (s *SceneWon) Update() (SceneIndex, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return gameRunning, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return gameWon, errors.New("game quit by player")
	}
	return gameWon, nil
}

func (s *SceneWon) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "You died\nPress space to restart\nPress Esc to quit")
}

func (s *SceneWon) Load(prev SceneIndex) {
	log.Println("Game over")
}
