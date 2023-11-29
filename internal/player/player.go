package player

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/physics"
)

const (
	DefaultHealth = 100
)

type Player struct {
	*object.Object    `json:"-"`
	*physics.Physical `json:"-"`

	Image *ebiten.Image `json:"-"`

	Inventory *Inventory `json:"inventory"`

	OnGround bool `json:"-"`
	Health   int  `json:"health"`
}

func New(origin *geometry.Point, img *ebiten.Image) *Player {

	return &Player{
		Object: &object.Object{
			Origin: origin,
			Width:  32,
			Height: 32,
		},
		Physical:  physics.NewPhysical(),
		Image:     img,
		Inventory: &Inventory{},
		Health:    DefaultHealth,
	}
}

func (p *Player) IsDead() bool {
	return p.Health <= 0
}

func (p *Player) Type() object.Type {
	return object.PlayerType
}

func (p *Player) Collect(it *item.Item) {
	it.Collected = true
	p.Inventory.Items = append(p.Inventory.Items, it)
}
