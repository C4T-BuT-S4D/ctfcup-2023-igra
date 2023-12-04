package boss

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

const (
	BossV1Width  = 128
	BossV1Height = 128
)

type V1 struct {
	*object.Object
	StartPos    *geometry.Point
	MoveVector  *geometry.Vector
	Image       *ebiten.Image `msgpack:"-"`
	bulletImage *ebiten.Image
	RotateAngle float64
	Speed       float64
	Length      float64
	Ticks       int
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

func NewV1(origin *geometry.Point, img *ebiten.Image, bulletImage *ebiten.Image, speed float64, length float64) *V1 {
	return &V1{
		Object: &object.Object{
			Origin: origin,
			Width:  BossV1Width,
			Height: BossV1Height,
		},
		StartPos:    origin,
		MoveVector:  &geometry.Vector{X: -speed},
		bulletImage: bulletImage,
		Speed:       speed,
		Length:      length,
		Image:       img,
	}
}

func (v *V1) Reset() {
	v.RotateAngle = 0
	v.MoveVector = &geometry.Vector{X: -v.Speed}
	v.Ticks = 0
}

func (v *V1) Rotate() {
	v.RotateAngle += math.Pi / 60
}

func (v *V1) GetNextMove() *geometry.Vector {
	if v.Origin.X < v.StartPos.X-v.Length {
		v.MoveVector = &geometry.Vector{X: v.Speed}
	} else if v.Origin.X > v.StartPos.X {
		v.MoveVector = &geometry.Vector{X: -v.Speed}
	}
	return v.MoveVector
}

func (v *V1) CreateBullets() []*damage.Bullet {
	v.Ticks = (v.Ticks + 1) % 8
	if v.Ticks != 0 {
		return nil
	}

	var bullets []*damage.Bullet

	const bulletDamage = 5

	vv := []*geometry.Vector{
		{X: math.Cos(v.RotateAngle), Y: math.Sin(v.RotateAngle)},
		{X: math.Sin(v.RotateAngle), Y: math.Cos(v.RotateAngle)},
	}

	for _, mult := range []float64{2, 4, 8} {
		for _, vec := range vv {
			bullets = append(bullets, damage.NewBullet(
				v.Origin.Add(&geometry.Vector{X: BossV1Width / 2, Y: BossV1Height / 2}),
				v.bulletImage,
				bulletDamage,
				vec.Multiply(1/vec.Length()).Multiply(mult),
			))
		}
	}
	return bullets
}
