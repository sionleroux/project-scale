// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// AnimationSkipTicks sets how many ticks to skip before stepping to the next
// frame in the animation
const AnimationSkipTicks = 5

// Animate determines the next animation frame for a sprite
func Animate(frame, tick int, ft FrameTags) int {
	from, to := ft.From, ft.To

	// Instantly start animation if state changed
	if frame < from || frame >= to {
		return from
	}

	// Update only in every 5th cycle
	if tick%AnimationSkipTicks != 0 {
		return frame
	}

	// Continuously increase the Frame counter between from and to
	return frame + 1
}

type SpriteAnimation struct {
	Sprite   *SpriteSheet
	FrameTag int
	Frame    int
	Tick     int
}

func NewSpriteAnimation(name string) *SpriteAnimation {
	return &SpriteAnimation{
		Sprite: loadSprite(name),
	}

}

// Returns TRUE if the endframe of the frametag is reached
func (s *SpriteAnimation) Update(frameTag int) bool {
	ft := s.Sprite.Meta.FrameTags[frameTag]
	from, to := ft.From, ft.To

	if s.FrameTag != frameTag {
		s.FrameTag = frameTag
		s.Frame = from
		s.Tick = 0
	}

	s.Tick++

	// Instantly start animation if state changed
	if s.Frame < from || s.Frame > to {
		s.Frame = from
	}

	// Update only in every 5th cycle
	if s.Tick%AnimationSkipTicks == 0 {
		s.Frame++
	}

	return s.Frame == to
}

func (s *SpriteAnimation) GetImage() *ebiten.Image {
	frame := s.Sprite.Sprite[s.Frame]

	return s.Sprite.Image.SubImage(image.Rect(
		frame.Position.X,
		frame.Position.Y,
		frame.Position.X+frame.Position.W,
		frame.Position.Y+frame.Position.H,
	)).(*ebiten.Image)
}
