package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	StaticTileType ObjectType = iota
	PlayerType
)

type StaticTile struct {
	width  int
	height int
	x      int
	y      int
	img    *ebiten.Image
}

func (s StaticTile) Type() ObjectType {
	return StaticTileType
}

func (s StaticTile) Rectangle() Rectangle {
	return Rectangle{
		LeftX:   float64(s.x),
		TopY:    float64(s.y),
		RightX:  float64(s.x + s.width),
		BottomY: float64(s.y + s.height),
	}
}

func (s StaticTile) Image() *ebiten.Image {
	return s.img
}

type Player struct {
	img    *ebiten.Image
	x      float64
	y      float64
	width  int
	height int
}

func (p *Player) Type() ObjectType {
	return PlayerType
}

func (p *Player) Rectangle() Rectangle {
	return Rectangle{
		LeftX:   p.x,
		TopY:    p.y,
		RightX:  p.x + float64(p.width),
		BottomY: p.y + float64(p.height),
	}
}

func (p *Player) SetRectangle(r Rectangle) {
	p.x = r.LeftX
	p.y = r.TopY
}

func (p *Player) Image() *ebiten.Image {
	return p.img
}
