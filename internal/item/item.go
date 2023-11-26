package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

type Item struct {
	*object.Object

	Image *ebiten.Image

	Name      string
	Important bool
	Collected bool
}

func New(origin *geometry.Point, width, height int, name string, important bool) *Item {
	img := ebiten.NewImage(width, height)
	img.Fill(color.RGBA{R: 255, G: 0, B: 0, A: 255})

	return &Item{
		Object: &object.Object{
			Origin: origin,
			Width:  float64(width),
			Height: float64(height),
		},
		Image:     img,
		Name:      name,
		Important: important,
	}
}

func (it *Item) Type() object.Type {
	return object.Item
}
