package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

type Item struct {
	*object.Object

	Image *ebiten.Image `json:"-"`

	Name      string `json:"name"`
	Important bool   `json:"important"`
	Collected bool   `json:"collected"`
}

func New(origin *geometry.Point, width, height float64, name string, important bool) *Item {
	img := ebiten.NewImage(int(width), int(height))
	img.Fill(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff})

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
