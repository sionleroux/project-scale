// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// How long at least to show the loading screen even if everything loads very
// fast so that it isn't just a black flash
var loadingSceneMinTime = 2 * 60

// LoadingCounter is for tracking how much of the assets have been loaded
type LoadingCounter *uint8

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
	Counter LoadingCounter // what is being loaded
	Tick    int
	Loaded  bool
}

func NewLoadingScene() *LoadingScene {
	return &LoadingScene{
		Counter: new(uint8),
	}
}

// Update handles player input to update the start screen
func (s *LoadingScene) Update() error {
	s.Tick++
	if s.Loaded && s.Tick > loadingSceneMinTime {
		s.SceneManager.SwitchTo(s.State.Scenes[gameStart])
		return nil
	}
	return nil
}

// Draw renders the start screen to the screen
func (s *LoadingScene) Draw(screen *ebiten.Image) {
	var whatTxt string
	if int(*s.Counter) < len(loadingWhat) {
		whatTxt = loadingWhat[*s.Counter]
	}
	s.State.TextRenderer.Draw(screen, "Loading..."+whatTxt, 8, 50, 85)

}
