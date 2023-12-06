package boss

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shopspring/decimal"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/portal"
)

const (
	BossV1Width  = 128
	BossV1Height = 128
)

type V1 struct {
	*object.Object `json:"-"`
	Name           string          `json:"-"`
	MoveX          decimal.Decimal `json:"-"`
	Image          *ebiten.Image   `json:"-" msgpack:"-"`
	bulletImage    *ebiten.Image
	RotateAngle    decimal.Decimal `json:"-"`
	Speed          decimal.Decimal `json:"-"`
	Length         decimal.Decimal `json:"-"`
	X              decimal.Decimal `json:"-"`
	StartX         decimal.Decimal `json:"-"`
	Ticks          int             `json:"-"`
	StartHealth    int64           `json:"-"`
	Health         int64           `json:"-"`
	Dead           bool            `json:"dead"`
	WinPoint       *geometry.Point `json:"-"`
	PortalName     string          `json:"-"`
	ItemName       string          `json:"-"`
	Portal         *portal.Portal  `json:"-"`
	Item           *item.Item      `json:"-"`
}

func (v *V1) Type() object.Type {
	return object.BossV1
}

func (v *V1) GetOrigin() *geometry.Point {
	if v == nil {
		return nil
	}
	return v.Object.GetOrigin()
}

func NewV1(name string, origin *geometry.Point, img *ebiten.Image, bulletImage *ebiten.Image, speed float64, length float64, health int64, portalName string, itemName string) *V1 {
	s := decimal.NewFromFloat(speed)
	return &V1{
		Object: &object.Object{
			Origin: origin,
			Width:  BossV1Width,
			Height: BossV1Height,
		},
		Name:        name,
		MoveX:       s.Neg(),
		bulletImage: bulletImage,
		Speed:       s,
		Length:      decimal.NewFromFloat(length),
		RotateAngle: decimal.Zero,
		X:           decimal.NewFromFloat(origin.X),
		StartX:      decimal.NewFromFloat(origin.X),
		Image:       img,
		StartHealth: health,
		Health:      health,
		PortalName:  portalName,
		ItemName:    itemName,
	}
}

func (v *V1) Reset() {
	v.RotateAngle = decimal.Zero
	v.MoveX = v.Speed.Neg()
	v.Ticks = 0
	v.Health = v.StartHealth
}

func (v *V1) GetNextMove() float64 {
	if v.X.LessThan(v.StartX.Sub(v.Length)) {
		v.MoveX = v.Speed
	} else if v.X.GreaterThan(v.StartX) {
		v.MoveX = v.Speed.Neg()
	}
	v.X = v.X.Add(v.MoveX)
	x, _ := v.X.Float64()
	return x
}

func (v *V1) Tick() {
	v.Ticks = (v.Ticks + 1) % 8
	v.Health--
	if v.Health == 0 {
		v.Dead = true
		return
	}
	v.RotateAngle = v.RotateAngle.Add(decimal.NewFromFloat(3.1415).Div(decimal.NewFromInt(60)))
}

type vec struct {
	X, Y decimal.Decimal
}

func (v *V1) CreateBullets() []*damage.Bullet {
	if v.Ticks != 0 {
		return nil
	}

	var bullets []*damage.Bullet

	const bulletDamage = 6

	vv := []vec{
		{X: v.RotateAngle.Cos(), Y: v.RotateAngle.Sin()},
		{X: v.RotateAngle.Sin(), Y: v.RotateAngle.Cos()},
	}

	for _, mult := range []int64{2, 4, 8} {
		for _, vec := range vv {
			x, _ := vec.X.Mul(decimal.NewFromInt(mult)).Float64()
			y, _ := vec.Y.Mul(decimal.NewFromInt(mult)).Float64()
			bullets = append(bullets, damage.NewBullet(
				v.Origin.Add(&geometry.Vector{X: BossV1Width / 2, Y: BossV1Height / 2}),
				v.bulletImage,
				bulletDamage,
				&geometry.Vector{X: x, Y: y},
			))
		}
	}
	return bullets
}
