package world

import "github.com/hajimehoshi/ebiten/v2"

const (
	StaticTileType ObjectType = iota
	PlayerType
)

type StaticTile struct {
	Width  int
	Height int
	X      int
	Y      int

	Img *ebiten.Image
}

func NewStaticTile(width, height, x, y int, img *ebiten.Image) *StaticTile {
	return &StaticTile{
		Width:  width,
		Height: height,
		X:      x,
		Y:      y,
		Img:    img,
	}
}

func (s StaticTile) Type() ObjectType {
	return StaticTileType
}

func (s StaticTile) Rectangle() Rectangle {
	return Rectangle{
		LeftX:   float64(s.X),
		TopY:    float64(s.Y),
		RightX:  float64(s.X + s.Width),
		BottomY: float64(s.Y + s.Height),
	}
}

func (s StaticTile) Image() *ebiten.Image {
	return s.Img
}
