// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// How long at least to show the loading screen even if everything loads very
// fast so that it isn't just a black flash
var loadingSceneMinTime = 3 * 60

var loadingWhat = []string{
	"",
	"map",
	"music",
	"sounds",
	"entities",
	"done",
}

// LoadingScene is shown while all the assets are loading.
// When loading is ready it switches to Intro screen
type LoadingScene struct {
	Width        int
	Height       int
	LoadingState LoadingState // what is being loaded
	Tick         int
	TextRenderer *TextRenderer
}

func NewLoadingScene() *LoadingScene {
	return &LoadingScene{
		Width:        gameWidth,
		Height:       gameHeight,
		LoadingState: LoadingState{counter: 0, loaded: false, stateLock: &sync.RWMutex{}},
		TextRenderer: NewTextRenderer("assets/fonts/PixelOperator8.ttf"),
	}
}

func (s *LoadingScene) Layout(w, h int) (int, int) {
	return s.Width, s.Height
}

// Update handles player input to update the start screen
func (s *LoadingScene) Update() error {
	s.Tick++
	return nil
}

func (s *LoadingScene) IsLoaded() bool {
	loaded := s.LoadingState.GetLoaded()
	return loaded && s.Tick > loadingSceneMinTime
}

// Draw renders the start screen to the screen
func (s *LoadingScene) Draw(screen *ebiten.Image) {
	s.TextRenderer.Draw(
		screen,
		"An action adventure story by:\nRowan Lindeque\nTristan Le Roux\nSiôn Le Roux\nPéter Kertész",
		color.White, 8, 50, 50,
	)

	var whatTxt string
	counter := s.LoadingState.GetCounterValue()
	if counter < len(loadingWhat) {
		whatTxt = loadingWhat[counter]
	}
	s.TextRenderer.Draw(screen, "Loading..."+whatTxt, color.White, 8, 50, 85)

}

type LoadingState struct {
	stateLock *sync.RWMutex
	counter   int
	loaded    bool
}

func (state *LoadingState) IncreaseCounter(x int) {
	state.stateLock.Lock()
	state.counter += x
	state.stateLock.Unlock()
}

func (state *LoadingState) GetCounterValue() (x int) {
	state.stateLock.Lock()
	x = state.counter
	state.stateLock.Unlock()
	return
}

func (state *LoadingState) SetLoaded(x bool) {
	state.stateLock.Lock()
	state.loaded = x
	state.stateLock.Unlock()
}

func (state *LoadingState) GetLoaded() (x bool) {
	state.stateLock.Lock()
	x = state.loaded
	state.stateLock.Unlock()
	return
}
