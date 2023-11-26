package camera

import (
	"math"

	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

type Shaker struct {
	Ease   *gween.Tween
	Done   bool
	time   float64
	period float64 // loops every this much
}

func NewShaker(maxMagnitude, maxTime float32, period float64) *Shaker {
	s := &Shaker{
		period: period,
	}
	s.Ease = gween.New(maxMagnitude, 0, maxTime, ease.OutExpo)
	s.Ease.Set(maxTime)
	return s
}

func (s *Shaker) calcShake() (x, y float64) {
	magnitude, done := s.Ease.Update(1)
	s.Done = done
	if s.Done {
		return 0, 0
	}
	s.time++
	return math.Sin(s.time*2*math.Pi/s.period) * (float64(magnitude) / 2), 0
}
