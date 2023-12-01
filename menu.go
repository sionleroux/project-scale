package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
)

type Menu struct {
	Items         []string
	Active        int
	X             float64
	Y             float64
	color         color.Color
	selectedColor color.Color
	textRenderer  *TextRenderer
}

func (m *Menu) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		m.Active -= 1
		if m.Active == -1 {
			m.Active = len(m.Items) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		m.Active += 1
		if m.Active == len(m.Items) {
			m.Active = 0
		}
	}
}

func (m *Menu) Draw(screen *ebiten.Image) {
	for i, mi := range m.Items {
		menuColor := m.color
		txt := mi
		if i == m.Active {
			menuColor = m.selectedColor
			txt = fmt.Sprintf("» %s «", mi)
		}
		m.textRenderer.DrawXY(screen, txt, menuColor, 8, int(m.X), int(m.Y)+i*12, etxt.XCenter)
	}
}
