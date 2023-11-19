package player

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/physics"
)

type Player struct {
	*physics.Object

	Image    *ebiten.Image
	OnGround bool
}

func New(origin *geometry.Point) *Player {
	img := ebiten.NewImage(16, 16)
	img.Fill(color.White)

	return &Player{
		Object: physics.NewObject(&object.Object{
			Origin: origin,
			Width:  16,
			Height: 16,
		}),
		Image: img,
	}
}

func (p *Player) Type() object.Type {
	return object.PlayerType
}
