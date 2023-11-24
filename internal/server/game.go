package server

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

type Game struct {
	engine *atomic.Pointer[engine.Engine]
	events chan *gameserverpb.ClientEvent
}

func NewGame() *Game {
	return &Game{
		engine: atomic.NewPointer[engine.Engine](nil),
	}
}

func (g *Game) Update() error {
	eng := g.engine.Load()
	if eng == nil {
		return nil
	}

	event, ok := <-g.events
	logrus.Debugf("new update from client: %v, ok: %v", event, ok)
	if !ok {
		return nil
	}
	inp := input.NewFromProto(event.KeysPressed)

	if err := eng.Update(inp); err != nil {
		return fmt.Errorf("updating engine state: %w", err)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if eng := g.engine.Load(); eng != nil {
		eng.Draw(screen)
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return 640, 480
}
