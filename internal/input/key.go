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
		ebase, base, ek := ebiten.KeyA, 'A', ebiten.Key(k)
		if ek >= ebiten.KeyDigit0 && ek <= ebiten.KeyDigit9 {
			ebase, base = ebiten.KeyDigit0, '0'
		}
		return string(base + rune(ek-ebase))
	}
}
