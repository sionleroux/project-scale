package main

import (
	"time"
)

// Stat stores the game statistics
type Stat struct {
	GameStart        time.Time
	GameEnd          time.Time
	LastLevel        int
	LastHighestPoint int
	HighestPoint     int
}
