package main

import (
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

const (
	TileClimbable int8 = iota
	TileWall
	TileChasm
	TileSlippery
)

const (
	TagClimbable = "climbable"
	TagWall      = "wall"
	TagChasm     = "chasm"
	TagSlippery  = "slippery"
)

var TileTags = []string{
	TagClimbable,
	TagWall,
	TagChasm,
	TagSlippery,
}

func tilesToObstacles(layer *ldtkgo.Layer, space *resolv.Space) {
	if tiles := layer.AllTiles(); len(tiles) > 0 {
		for _, tileData := range tiles {
			size := float64(layer.Tileset.GridSize)
			x, y := tileData.Position[0], tileData.Position[1]

			object := resolv.NewObject(
				float64(x+layer.OffsetX), float64(y+layer.OffsetY),
				size, size,
				TileTags[tileData.ID],
			)
			object.SetShape(resolv.NewRectangle(
				0, 0, // origin
				size, size,
			))
			object.Shape.(*resolv.ConvexPolygon).RecenterPoints()

			space.Add(object)
		}
	}
}
