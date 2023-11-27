package main

import (
	"github.com/joelschutz/stagehand"
)

type BaseScene struct {
	State        State
	SceneManager *stagehand.SceneManager[State]
}

func (s *BaseScene) Layout(w, h int) (int, int) {
	return s.State.Width, s.State.Height
}

func (s *BaseScene) Load(st State, sm *stagehand.SceneManager[State]) {
	s.State = st
	s.SceneManager = sm
}

func (s *BaseScene) Unload() State {
	return s.State
}
