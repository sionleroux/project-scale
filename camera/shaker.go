package camera

import (
	"math"

	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

type Shaker struct {
	Ease         *gween.Tween
	time         float64
	maxMagnitude float32
	maxTime      float32
	period       float64
}

func NewShaker() *Shaker {
	s := &Shaker{
		maxMagnitude: 10,
		maxTime:      40,
		period:       10, // loops every this much
	}
	s.Ease = gween.New(s.maxMagnitude, 0, s.maxTime, ease.OutExpo)
	s.Ease.Set(s.maxTime)
	return s
}

func (s *Shaker) calcShake() (x, y float64) {
	magnitude, _ := s.Ease.Update(1)
	s.time++
	return math.Sin(s.time*2*math.Pi/s.period) * (float64(magnitude) / 2), 0
}
