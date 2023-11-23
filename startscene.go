package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type StartScene struct {
	BaseScene
}

func (s *StartScene) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.SceneManager.SwitchTo(s.State.Game.Scenes[gameRunning])
	}
	return nil
}

func (s *StartScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Press space to start")
}
