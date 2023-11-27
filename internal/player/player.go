package player

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/physics"
)

type Player struct {
	*physics.Object `json:"-"`

	Image *ebiten.Image `json:"-"`

	Inventory *Inventory `json:"inventory"`

	OnGround bool `json:"-"`
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
		Image:     img,
		Inventory: &Inventory{},
	}
}

func (p *Player) Type() object.Type {
	return object.PlayerType
}

func (p *Player) Collect(it *item.Item) {
	it.Collected = true
	p.Inventory.Items = append(p.Inventory.Items, it)
}
