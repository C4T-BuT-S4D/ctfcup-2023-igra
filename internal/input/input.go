package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samber/lo"

	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

var interestingKeys = []ebiten.Key{
	ebiten.KeyW,
	ebiten.KeyA,
	ebiten.KeyS,
	ebiten.KeyD,
	ebiten.KeySpace,
	ebiten.KeySlash,
	ebiten.KeyR,
	ebiten.KeyEscape,
	ebiten.KeyEnter,
	ebiten.KeyE,
	ebiten.KeyP,
}

type Input struct {
	pressedKeys      map[ebiten.Key]struct{}
	newlyPressedKeys map[ebiten.Key]struct{}
}

func New() *Input {
	return &Input{
		pressedKeys:      make(map[ebiten.Key]struct{}),
		newlyPressedKeys: make(map[ebiten.Key]struct{}),
	}
}

func NewFromProto(p *gameserverpb.ClientEvent_KeysPressed) *Input {
	i := New()
	for _, key := range p.KeysPressed {
		i.pressedKeys[ebiten.Key(key)] = struct{}{}
	}
	for _, key := range p.NewKeysPressed {
		i.newlyPressedKeys[ebiten.Key(key)] = struct{}{}
	}
	return i
}

func (i *Input) Update() {
	oldPressedKeys := i.pressedKeys

	i.pressedKeys = make(map[ebiten.Key]struct{})
	i.newlyPressedKeys = make(map[ebiten.Key]struct{})
	for _, key := range interestingKeys {
		if ebiten.IsKeyPressed(key) {
			i.pressedKeys[key] = struct{}{}
			if _, ok := oldPressedKeys[key]; !ok {
				i.newlyPressedKeys[key] = struct{}{}
			}
		}
	}
}

func (i *Input) IsKeyPressed(key ebiten.Key) bool {
	_, ok := i.pressedKeys[key]
	return ok
}

func (i *Input) IsKeyNewlyPressed(key ebiten.Key) bool {
	_, ok := i.newlyPressedKeys[key]
	return ok
}

func (i *Input) ToProto() *gameserverpb.ClientEvent_KeysPressed {
	return &gameserverpb.ClientEvent_KeysPressed{
		KeysPressed: lo.Map(lo.Keys(i.pressedKeys), func(key ebiten.Key, _ int) int32 {
			return int32(key)
		}),
		NewKeysPressed: lo.Map(lo.Keys(i.newlyPressedKeys), func(key ebiten.Key, _ int) int32 {
			return int32(key)
		}),
	}
}
