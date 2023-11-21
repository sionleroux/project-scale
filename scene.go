package main

import "github.com/hajimehoshi/ebiten/v2"

// Scene is a full-screen UI Screen for some part of the game like a menu or a
// game level
type Scene interface {
	Update() (SceneIndex, error)
	Draw(screen *ebiten.Image)
	Load()
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
