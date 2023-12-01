package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
	input "github.com/quasilyte/ebitengine-input"
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
	InputSystem      input.System
	Keymap           input.Keymap
	Input            *input.Handler
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

	// Input setup
	game.InputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})
	game.Keymap = input.Keymap{
		ActionMoveUp:    {input.KeyUp, input.KeyW, input.KeyGamepadUp, input.KeyGamepadLStickUp},
		ActionMoveLeft:  {input.KeyLeft, input.KeyA, input.KeyGamepadLeft, input.KeyGamepadLStickLeft},
		ActionMoveDown:  {input.KeyDown, input.KeyS, input.KeyGamepadDown, input.KeyGamepadLStickDown},
		ActionMoveRight: {input.KeyRight, input.KeyD, input.KeyGamepadRight, input.KeyGamepadLStickRight},
		ActionPrimary:   {input.KeySpace, input.KeyGamepadA},
		ActionMenu:      {input.KeyEscape, input.KeyGamepadStart},
	}

	game.Input = game.InputSystem.NewHandler(0, game.Keymap)

	game.Stat.Load()

	game.Scenes = []stagehand.Scene[State]{
		NewStartScene(game),
		&GameScene{},
		&PauseScreen{
			Menu: &Menu{
				Items:         []string{"Continue", "Back to main menu"},
				X:             gameWidth / 2,
				Y:             190,
				color:         color.RGBA{255, 255, 255, 255},
				selectedColor: color.RGBA{255, 255, 0, 255},
				textRenderer:  game.TextRenderer,
				Input:         game.Input,
			},
		},
		&OverScene{
			Menu: &Menu{
				Items:         []string{"Restart", "Back to main menu"},
				X:             gameWidth / 2,
				Y:             190,
				color:         color.RGBA{255, 255, 255, 255},
				selectedColor: color.RGBA{255, 255, 0, 255},
				textRenderer:  game.TextRenderer,
				Input:         game.Input,
			},
		},
		&WonScene{
			Menu: &Menu{
				Items:         []string{"Restart", "Back to main menu"},
				X:             gameWidth / 2,
				Y:             190,
				color:         color.RGBA{255, 255, 255, 255},
				selectedColor: color.RGBA{255, 255, 0, 255},
				textRenderer:  game.TextRenderer,
				Input:         game.Input,
			},
		},
	}

	s.sceneManager = stagehand.NewSceneManager[State](game.Scenes[gameStart], game)

	NewGameScene(game, &s.loadingScene.LoadingState)
}
