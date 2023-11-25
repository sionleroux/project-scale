package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
)

const screenScaleFactor = 4
const gridSize = 16

func main() {
	const gameWidth, gameHeight = 320, 240

	ebiten.SetWindowSize(gameWidth*screenScaleFactor, gameHeight*screenScaleFactor)
	ebiten.SetWindowTitle("project-scale")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := &Game{
		Width:        gameWidth,
		Height:       gameHeight,
		TextRenderer: NewTextRenderer("assets/fonts/PixelOperator8-Bold.ttf"),
		Stat:         &Stat{},
	}

	game.Stat.Load()

	loadingScene := NewLoadingScene()

	game.Scenes = []stagehand.Scene[State]{
		loadingScene,
		NewStartScene(),
		&GameScene{},
		&PauseScreen{},
		&OverScene{},
		&WonScene{},
	}

	sceneManager := stagehand.NewSceneManager[State](game.Scenes[gameLoading], game)

	go NewGameScene(game, game.Scenes[gameLoading].(*LoadingScene).Counter)

	if err := ebiten.RunGame(sceneManager); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	Width, Height int
	Scenes        []stagehand.Scene[State]
	ResetNeeded   bool
	TextRenderer  *TextRenderer
	Stat          *Stat
}
