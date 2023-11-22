package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	const gameWidth, gameHeight = 320, 240
	const screenScaleFactor = 4

	ebiten.SetWindowSize(gameWidth*screenScaleFactor, gameHeight*screenScaleFactor)
	ebiten.SetWindowTitle("project-scale")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := &Game{
		Width:  gameWidth,
		Height: gameHeight,
	}

	game.Scenes = []Scene{
		&StartScene{},
		NewGameScene(game),
		&PauseScreen{},
		&SceneOver{},
		&SceneWon{},
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	Width, Height int
	Scenes        []Scene
	CurrentScene  SceneIndex
}

// Update updates the inner game scene by one tick
func (g *Game) Update() error {
	prevScene := g.CurrentScene
	nextScene, err := g.Scenes[g.CurrentScene].Update()
	if err != nil {
		return err
	}
	if prevScene != nextScene {
		g.Scenes[nextScene].Load(prevScene)
	}
	g.CurrentScene = nextScene
	return nil
}

// Draw delegates drawing to the inner game scene
func (g *Game) Draw(screen *ebiten.Image) {
	g.Scenes[g.CurrentScene].Draw(screen)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}
