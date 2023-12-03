package damage

import (
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	BulletWidth  = 1
	BulletHeight = 1
)

type Bullet struct {
	*object.Object
	Image      *ebiten.Image `msgpack:"-"`
	Damageable `msgpack:"-"`
	Direction  *geometry.Vector
	Triggered  bool
}

func (e *Bullet) Type() object.Type {
	return object.EnemyBullet
}

func NewBullet(origin *geometry.Point, img *ebiten.Image, damage int, direction *geometry.Vector) *Bullet {
	return &Bullet{
		Object: &object.Object{
			Origin: origin,
			Width:  BulletWidth,
			Height: BulletHeight,
		},
		Image:      img,
		Damageable: NewDamageable(damage),
		Direction:  direction,
	}
}
