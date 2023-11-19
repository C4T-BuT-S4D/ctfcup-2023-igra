package geometry

type Rectangle struct {
	LeftX   float64
	TopY    float64
	RightX  float64
	BottomY float64
}

func (a *Rectangle) Intersects(b *Rectangle) bool {
	return !(a.RightX <= b.LeftX || b.RightX <= a.LeftX || a.BottomY <= b.TopY || b.BottomY <= a.TopY)
}
