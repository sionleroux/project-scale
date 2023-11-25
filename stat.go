package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/quasilyte/gdata"
)

// Stat stores the game statistics
type Stat struct {
	GameStart        time.Time
	GameEnd          time.Time
	LastHighestPoint int
	HighestPoint     int
}

func (s *Stat) Load() {
	s.HighestPoint = 0
	m, err := gdata.Open(gdata.Config{
		AppName: "project_scale",
	})
	if err != nil {
		return
	}

	result, err := m.LoadItem("Stat.HighestPoint")
	if err != nil {
		return
	}

	s.HighestPoint, _ = strconv.Atoi(string(result))
}

func (s *Stat) Save() {
	m, err := gdata.Open(gdata.Config{
		AppName: "project_scale",
	})
	if err != nil {
		return
	}

	m.SaveItem("Stat.HighestPoint", []byte(strconv.Itoa(s.HighestPoint)))
}

func (s *Stat) GetText() string {
	return fmt.Sprintf(
		"Your last climb: %d m\nYour best climb so far: %d m",
		s.LastHighestPoint, s.HighestPoint,
	)
}
