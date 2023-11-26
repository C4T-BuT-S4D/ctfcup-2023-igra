package geometry

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p *Point) Add(v *Vector) *Point {
	p.X += v.X
	p.Y += v.Y
	return p
}
