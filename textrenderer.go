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

// xRatio is where to align horizontally: 0 = left, 100 = right
// yRatio is where to align vertiocally: 0 = top, 100 = bottom
func (r *TextRenderer) Draw(screen *ebiten.Image, text string, color color.Color, size int, xRatio int, yRatio int) {
	r.SetTarget(screen)
	r.SetAlign(etxt.YCenter, etxt.XCenter)
	r.SetColor(color)
	r.SetSizePx(size)
	r.Renderer.Draw(text, screen.Bounds().Dx()*xRatio/100, screen.Bounds().Dy()*yRatio/100)
}

func (r *TextRenderer) DrawXY(screen *ebiten.Image, text string, color color.Color, size int, x int, y int, align etxt.HorzAlign) {
	r.SetTarget(screen)
	r.SetAlign(etxt.Top, align)
	r.SetColor(color)
	r.SetSizePx(size)
	r.Renderer.Draw(text, x, y)
}
