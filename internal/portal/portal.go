package portal

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

type Portal struct {
	*object.Object
	Image      *ebiten.Image
	PortalTo   string
	TeleportTo *geometry.Point
}

func (p *Portal) Type() object.Type {
	return object.Portal
}

func New(origin *geometry.Point, width, height float64, portalTo string, teleportTo *geometry.Point) *Portal {
	img := ebiten.NewImage(int(width), int(height))
	img.Fill(color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff})
	return &Portal{
		Object: &object.Object{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		PortalTo:   portalTo,
		TeleportTo: teleportTo,
		Image:      img,
	}
}
