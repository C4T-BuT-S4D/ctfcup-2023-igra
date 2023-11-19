package geometry

type Point struct {
	X, Y float64
}

func (p *Point) Add(v *Vector) *Point {
	p.X += v.X
	p.Y += v.Y
	return p
}
