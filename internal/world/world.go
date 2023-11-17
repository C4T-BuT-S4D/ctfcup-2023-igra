package world

import "github.com/hajimehoshi/ebiten/v2"

type ObjectType int

type Point struct {
	X int
	Y int
}

func NewPoint(x, y int) Point {
	return Point{
		X: x,
		Y: y,
	}
}

type Rectangle struct {
	LeftX   float64
	TopY    float64
	RightX  float64
	BottomY float64
}

func (a Rectangle) Intersects(b Rectangle) bool {
	return !(a.RightX <= b.LeftX || b.RightX <= a.LeftX || a.BottomY <= b.TopY || b.BottomY <= a.TopY)
}

type GameObject interface {
	Type() ObjectType
	Rectangle() Rectangle
	Image() *ebiten.Image
}

type World struct {
	Objects []GameObject
	Player  *Player
}

func (w *World) AddObject(o GameObject) *World {
	w.Objects = append(w.Objects, o)
	return w
}

func (w *World) Intersects(x Rectangle) []GameObject {
	var result []GameObject
	for _, o := range w.Objects {
		// fmt.Printf("Checking %+v and %+v\n", o.Rectangle(), x)
		if o.Rectangle().Intersects(x) {
			result = append(result, o)
		}
	}
	return result
}
