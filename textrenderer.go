// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
)

type TextRenderer struct {
	*etxt.Renderer
	alpha uint8
}

func NewTextRenderer(fontName string) *TextRenderer {
	font := loadFont(fontName)
	r := etxt.NewStdRenderer()
	r.SetFont(font)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	return &TextRenderer{r, 0xff}
}

func (r *TextRenderer) Draw(screen *ebiten.Image, text string, size int, x int, y int) {
	r.SetTarget(screen)
	r.SetColor(color.RGBA{0xff, 0xff, 0xff, r.alpha})
	r.SetSizePx(size)
	r.Renderer.Draw(text, x, y)
}
