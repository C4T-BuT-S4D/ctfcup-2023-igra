package item

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

type Item struct {
	*object.Object `json:"-"`

	Image *ebiten.Image `json:"-" msgpack:"-"`

	Name      string `json:"name"`
	Important bool   `json:"important"`
	Collected bool   `json:"collected"`
}

func New(origin *geometry.Point, width, height float64, img *ebiten.Image, name string, important bool) *Item {
	return &Item{
		Object: &object.Object{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		Image:     img,
		Name:      name,
		Important: important,
	}
}

func (it *Item) Type() object.Type {
	return object.Item
}
