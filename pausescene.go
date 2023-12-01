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
	Menu *Menu
}

func (p *PauseScreen) Update() error {
	p.Menu.Update()
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if p.Menu.Active == 0 {
			p.SceneManager.SwitchTo(p.State.Scenes[gameRunning])
		} else if p.Menu.Active == 1 {
			p.SceneManager.SwitchTo(p.State.Scenes[gameStart])
		}
	}
	return nil
}

func (p *PauseScreen) Draw(screen *ebiten.Image) {
	screen.DrawImage(p.State.lastRender, &ebiten.DrawImageOptions{})
	vector.DrawFilledRect(screen, 0, 0, float32(p.State.Width), float32(p.State.Height), color.RGBA{0, 0, 0, 128}, false)

	p.Menu.Draw(screen)
}
