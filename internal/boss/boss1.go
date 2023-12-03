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
	startPos    *geometry.Point
	moveVector  *geometry.Vector
	Image       *ebiten.Image
	bulletImage *ebiten.Image
	rotateAngle float64
	speed       float64
	length      float64
	ticks       int
}

func (v *V1) Type() object.Type {
	return object.BossV1
}

func NewV1(origin *geometry.Point, img *ebiten.Image, bulletImage *ebiten.Image, speed float64, length float64) *V1 {
	return &V1{
		Object: &object.Object{
			Origin: origin,
			Width:  BossV1Width,
			Height: BossV1Height,
		},
		startPos:    origin,
		moveVector:  &geometry.Vector{X: -speed},
		bulletImage: bulletImage,
		speed:       speed,
		length:      length,
		Image:       img,
	}
}

func (v *V1) Reset() {
	v.rotateAngle = 0
	v.moveVector = &geometry.Vector{X: -v.speed}
	v.ticks = 0
}

func (v *V1) RotateAngle() float64 {
	result := v.rotateAngle
	v.rotateAngle += math.Pi / 60
	return result
}

func (v *V1) GetNextMove() *geometry.Vector {
	if v.Origin.X < v.startPos.X-v.length {
		v.moveVector = &geometry.Vector{X: v.speed}
	} else if v.Origin.X > v.startPos.X {
		v.moveVector = &geometry.Vector{X: -v.speed}
	}
	return v.moveVector
}

func (v *V1) CreateBullets() []*damage.Bullet {
	v.ticks = (v.ticks + 1) % 8
	if v.ticks != 0 {
		return nil
	}

	var bullets []*damage.Bullet

	const bulletDamage = 5

	vv := []*geometry.Vector{
		{X: math.Cos(v.rotateAngle), Y: math.Sin(v.rotateAngle)},
		{X: math.Sin(v.rotateAngle), Y: math.Cos(v.rotateAngle)},
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
