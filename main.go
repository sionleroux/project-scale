package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const gameWidth, gameHeight = 320, 240
const screenScaleFactor = 4
const gridSize = 16

func main() {

	ebiten.SetWindowSize(gameWidth*screenScaleFactor, gameHeight*screenScaleFactor)
	ebiten.SetWindowTitle("project-scale")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	stageManager := NewStageManager()

	go loadGame(stageManager)

	if err := ebiten.RunGame(stageManager); err != nil {
		log.Fatal(err)
	}
}
