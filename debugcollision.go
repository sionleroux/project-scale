// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

//go:build !release && debugcol

package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/solarlune/resolv"
)

func init() {
	debuggers.Add(DebugFunc(DebugCollision))
}

// DebugCollision draws boxes around objects in collision space to easily
// visualise how they move and collide
func DebugCollision(g *GameScene, screen *ebiten.Image) {
	p := g.Player
	if collision := p.Check(0, 0); collision != nil {
		for _, o := range collision.Objects {
			if p.Shape.Intersection(0, 0, o.Shape) != nil {
				debugPosition(g, screen, o)
			}
		}
	}
}

func debugPosition(g *GameScene, screen *ebiten.Image, o *resolv.Object) {
	if o.Shape == nil {
		return
	}

	lineColor := color.NRGBA{255, 255, 255, 255}
	if tags := o.Tags(); len(tags) > 0 {
		switch tags[0] {
		case TagWall:
			lineColor = color.NRGBA{255, 0, 0, 255}
		case TagChasm:
			lineColor = color.NRGBA{0, 255, 0, 255}
		case TagSlippery:
			lineColor = color.NRGBA{0, 0, 255, 255}
		}
	}

	verts := o.Shape.(*resolv.ConvexPolygon).Transformed()
	for i := 0; i < len(verts); i++ {
		vert := verts[i]
		next := verts[0]
		if i < len(verts)-1 {
			next = verts[i+1]
		}
		vX, vY := g.State.Camera.GetScreenCoords(vert.X(), vert.Y())
		nX, nY := g.State.Camera.GetScreenCoords(next.X(), next.Y())
		ebitenutil.DrawLine(screen, vX, vY, nX, nY, lineColor)
	}
}
