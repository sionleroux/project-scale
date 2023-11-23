package main

import (
	"github.com/joelschutz/stagehand"
)

type State struct {
	Game *Game
}

type BaseScene struct {
	State        State
	SceneManager *stagehand.SceneManager[State]
}

func (s *BaseScene) Layout(w, h int) (int, int) {
	return s.State.Game.Width, s.State.Game.Height
}

func (s *BaseScene) Load(st State, sm *stagehand.SceneManager[State]) {
	s.State = st
	s.SceneManager = sm
}

func (s *BaseScene) Unload() State {
	return s.State
}

// SceneIndex is global state for the whole game
type SceneIndex int

const (
	gameStart   SceneIndex = iota // Game start screen is shown
	gameRunning                   // The game is running the main game code
	gamePaused                    // The game is paused temporarily
	gameOver                      // The game has ended because you died
	gameWon                       // The game has ended because you won
)
