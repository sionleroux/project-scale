// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// How long at least to show the loading screen even if everything loads very
// fast so that it isn't just a black flash
var loadingSceneMinTime = 2 * 60

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
	BaseScene
	LoadingState LoadingState // what is being loaded
	Tick         int
}

func NewLoadingScene() *LoadingScene {
	return &LoadingScene{
		LoadingState: LoadingState{counter: 0, loaded: false, stateLock: &sync.RWMutex{}},
	}
}

// Update handles player input to update the start screen
func (s *LoadingScene) Update() error {
	s.Tick++
	loaded := s.LoadingState.GetLoaded()
	if loaded && s.Tick > loadingSceneMinTime {
		s.SceneManager.SwitchTo(s.State.Scenes[gameStart])
	}
	return nil
}

// Draw renders the start screen to the screen
func (s *LoadingScene) Draw(screen *ebiten.Image) {
	var whatTxt string
	counter := s.LoadingState.GetCounterValue()
	if counter < len(loadingWhat) {
		whatTxt = loadingWhat[counter]
	}
	s.State.TextRenderer.Draw(screen, "Loading..."+whatTxt, 8, 50, 85)

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
