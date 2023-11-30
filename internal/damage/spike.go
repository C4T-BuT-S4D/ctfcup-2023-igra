package damage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

type Spike struct {
	*object.Object
	Damageable

	Image *ebiten.Image `json:"-"`
}

func (s *Spike) Type() object.Type {
	return object.Spike
}

func NewSpike(origin *geometry.Point, img *ebiten.Image, width, height float64) *Spike {
	return &Spike{
		Object: &object.Object{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		Image:      img,
		Damageable: NewDamageable(100),
	}
}
