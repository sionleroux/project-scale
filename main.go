package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
)

const screenScaleFactor = 4

func main() {
	const gameWidth, gameHeight = 320, 240

	ebiten.SetWindowSize(gameWidth*screenScaleFactor, gameHeight*screenScaleFactor)
	ebiten.SetWindowTitle("project-scale")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := &Game{
		Width:  gameWidth,
		Height: gameHeight,
	}

	game.Scenes = []stagehand.Scene[State]{
		&StartScene{},
		NewGameScene(game),
		&PauseScreen{},
		&OverScene{},
		&WonScene{},
	}

	sceneManager := stagehand.NewSceneManager[State](game.Scenes[gameStart], game)

	if err := ebiten.RunGame(sceneManager); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	Width, Height int
	Scenes        []stagehand.Scene[State]
	ResetNeeded   bool
}
