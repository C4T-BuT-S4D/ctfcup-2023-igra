package input

import "github.com/hajimehoshi/ebiten/v2"

type Key ebiten.Key

func (k Key) String() string {
	switch ebiten.Key(k) {
	case ebiten.KeyPeriod:
		return "."
	case ebiten.KeySpace:
		return " "
	case ebiten.KeyComma:
		return ","
	default:
		A := 'A'
		return string(A + rune(ebiten.Key(k)-ebiten.KeyA))
	}
}
