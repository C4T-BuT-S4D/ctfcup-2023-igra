package geometry

import "math"

type Vector struct {
	X, Y float64
}

func (v *Vector) Add(other *Vector) *Vector {
	v.X += other.X
	v.Y += other.Y
	return v
}

func (v *Vector) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v *Vector) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}
