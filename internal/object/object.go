package object

import "github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"

type Type int

const (
	StaticTileType Type = iota
	PlayerType
)

func (t Type) String() string {
	switch t {
	case StaticTileType:
		return "StaticTileType"
	case PlayerType:
		return "PlayerType"
	default:
		panic("unknown type")
	}
}

type Object struct {
	Origin        *geometry.Point
	Width, Height float64
}

func (o *Object) Rectangle() *geometry.Rectangle {
	return &geometry.Rectangle{
		LeftX:   o.Origin.X,
		TopY:    o.Origin.Y,
		RightX:  o.Origin.X + o.Width,
		BottomY: o.Origin.Y + o.Height,
	}
}

func (o *Object) Move(d *geometry.Vector) *Object {
	o.Origin.Add(d)
	return o
}

type GenericObject interface {
	Rectangle() *geometry.Rectangle
	Type() Type
}
