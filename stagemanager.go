package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
	"github.com/sinisterstuf/project-scale/camera"
)

type State *Game

// SceneIndex is global state for the whole game
type SceneIndex int

const (
	gameStart   = iota // Game start screen is shown
	gameRunning        // The game is running the main game code
	gamePaused         // The game is paused temporarily
	gameOver           // The game has ended because you died
	gameWon            // The game has ended because you won
)

type StageManager struct {
	loadingScene *LoadingScene
	sceneManager *stagehand.SceneManager[State]
	loaded       bool
}

type Game struct {
	Width, Height    int
	Scenes           []stagehand.Scene[State]
	ResetNeeded      bool
	TextRenderer     *TextRenderer
	BoldTextRenderer *TextRenderer
	Stat             *Stat
	StartPos         []int
	Fog              *Fog
	Backdrops        Backdrops
	Water            *Water
	Camera           *camera.Camera
	minScale         float64
	lastRender       *ebiten.Image
}

func NewStageManager() *StageManager {
	return &StageManager{loadingScene: NewLoadingScene(), loaded: false}
}

func (s *StageManager) Layout(w, h int) (int, int) {
	if s.loaded {
		return s.sceneManager.Layout(w, h)
	} else {
		return s.loadingScene.Layout(w, h)
	}
}

func (s *StageManager) Update() error {
	if s.loaded {
		return s.sceneManager.Update()
	} else {
		if s.loadingScene.IsLoaded() {
			s.loaded = true
		}
		return s.loadingScene.Update()
	}
}

func (s *StageManager) Draw(screen *ebiten.Image) {
	if s.loaded {
		s.sceneManager.Draw(screen)
	} else {
		s.loadingScene.Draw(screen)
	}

}

func loadGame(s *StageManager) {

	game := &Game{
		Width:            gameWidth,
		Height:           gameHeight,
		TextRenderer:     NewTextRenderer("assets/fonts/PixelOperator8.ttf"),
		BoldTextRenderer: NewTextRenderer("assets/fonts/PixelOperator8-Bold.ttf"),
		Stat:             &Stat{},
		Camera:           camera.NewCamera(gameWidth, gameHeight),
		lastRender:       ebiten.NewImage(gameWidth, gameHeight),
	}

	game.Stat.Load()

	game.Scenes = []stagehand.Scene[State]{
		NewStartScene(),
		&GameScene{},
		&PauseScreen{},
		&OverScene{},
		&WonScene{},
	}

	s.sceneManager = stagehand.NewSceneManager[State](game.Scenes[gameStart], game)

	NewGameScene(game, &s.loadingScene.LoadingState)
}
