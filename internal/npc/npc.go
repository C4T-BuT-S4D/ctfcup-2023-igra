package npc

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/dialog"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

type NPC struct {
	*object.Object
	Dialog      dialog.Dialog `msgpack:"-"`
	Image       *ebiten.Image `msgpack:"-"`
	DialogImage *ebiten.Image `msgpack:"-"`
	LinkedItem  *item.Item    `msgpack:"-"`
	ReturnsItem string
}

func (n *NPC) Type() object.Type {
	return object.NPC
}

func New(origin *geometry.Point, img *ebiten.Image, dialogImage *ebiten.Image, width, height float64, dialog dialog.Dialog, item string) *NPC {
	return &NPC{
		Object: &object.Object{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		Image:       img,
		DialogImage: dialogImage,
		Dialog:      dialog,
		ReturnsItem: item,
	}
}
