package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type StartScene struct {
}

func (s *StartScene) Update() (SceneIndex, error) {
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		return gameRunning, nil
	}
	return gameStart, nil
}

func (s *StartScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Press space to start")
}

func (s *StartScene) Load() {
	log.Println("Game start screen")
}
