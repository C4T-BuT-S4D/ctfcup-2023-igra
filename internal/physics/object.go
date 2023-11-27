package physics

import (
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

type Object struct {
	*object.Object

	Speed        *geometry.Vector
	Acceleration *geometry.Vector
}

const GravityAcceleration = 1.0 / 6.0

func NewObject(o *object.Object) *Object {
	return &Object{
		Object:       o,
		Speed:        &geometry.Vector{},
		Acceleration: &geometry.Vector{},
	}
}

func (o *Object) ApplyAcceleration() *Object {
	o.Speed = o.Speed.Add(o.Acceleration)
	return o
}
