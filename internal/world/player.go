package world

import "github.com/hajimehoshi/ebiten/v2"

type Player struct {
	Width  int
	Height int
	X      float64
	Y      float64

	Img *ebiten.Image
}

func NewPlayer(width, height int, x, y float64, img *ebiten.Image) *Player {
	return &Player{
		Width:  width,
		Height: height,
		X:      x,
		Y:      y,
		Img:    img,
	}
}

func (p *Player) Type() ObjectType {
	return PlayerType
}

func (p *Player) Rectangle() Rectangle {
	return Rectangle{
		LeftX:   p.X,
		TopY:    p.Y,
		RightX:  p.X + float64(p.Width),
		BottomY: p.Y + float64(p.Height),
	}
}

func (p *Player) SetRectangle(r Rectangle) {
	p.X = r.LeftX
	p.Y = r.TopY
}

func (p *Player) Image() *ebiten.Image {
	return p.Img
}
