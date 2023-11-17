package main

import (
	"github.com/solarlune/ldtkgo"
	"github.com/solarlune/resolv"
)

const (
	// TileClimbable is basic climbable terrain
	TileClimbable int8 = iota

	// TileWall is an impassable wall, cannot be jumped or grappled over
	TileWall

	// TileChasm is a chasm, passable but causes you to fall to first passable
	// tile below
	TileChasm

	// TileSlippery  is slippery terrain, player guaranteed to slip until they
	// reach bottom but can be jumped or grappled off of
	TileSlippery
)

const (
	TagClimbable = "climbable"
	TagWall      = "wall"
	TagChasm     = "chasm"
	TagSlippery  = "slippery"
	TagFinish    = "finish"
	TagDeco1     = "decoration"
	TagDeco2     = "decoration"
	TagDeco3     = "decoration"
	TagDeco4     = "decoration"
	TagTrans     = "chasm"
)

var TileTags = []string{
	TagClimbable,
	TagWall,
	TagDeco1,
	TagDeco2,
	TagChasm,
	TagSlippery,
	TagDeco3,
	TagDeco4,
	TagTrans,
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

			space.Add(object)
		}
	}
}

const (
	EntityPlayerStart = "Player_start"
	EntityFinish      = "Finish"
)

const (
	LayerEntities = "Entities"
	LayerFloor    = "Floor"
	LayerWalls    = "Walls"
)
