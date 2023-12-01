package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// PauseScreen is shown when the game is paused
type PauseScreen struct {
	BaseScene
}

func (p *PauseScreen) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		p.SceneManager.SwitchTo(p.State.Scenes[gameRunning])
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		p.SceneManager.SwitchTo(p.State.Scenes[gameStart])
	}
	return nil
}

func (p *PauseScreen) Draw(screen *ebiten.Image) {
	screen.DrawImage(p.State.lastRender, &ebiten.DrawImageOptions{})
	vector.DrawFilledRect(screen, 0, 0, float32(p.State.Width), float32(p.State.Height), color.RGBA{0, 0, 0, 128}, false)

	p.State.TextRenderer.Draw(screen, "Game paused\nPress P to unpause\nPress Esc to quit", color.White, 8, 50, 50)
}
