package server

import (
	"fmt"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/camera"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/sirupsen/logrus"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

type Game struct {
	engine *engine.Engine

	lock sync.Mutex

	snapshotsDir string
}

func NewGame(snapshotsDir string) *Game {
	return &Game{snapshotsDir: snapshotsDir}
}

func (g *Game) processEvent(event *gameserverpb.ClientEvent) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.engine == nil {
		return nil
	}

	logrus.Debugf("new update from client: %v", event)
	inp := input.NewFromProto(event.KeysPressed)

	if inp.IsKeyNewlyPressed(ebiten.KeySlash) {
		s, err := g.engine.MakeSnapshot()
		if err != nil {
			return fmt.Errorf("making snapshot: %w", err)
		}

		if err := g.engine.SaveSnapshot(s); err != nil {
			return fmt.Errorf("saving snapshot: %w", err)
		}
	}

	if err := g.engine.Update(inp); err != nil {
		return fmt.Errorf("updating engine state: %w", err)
	}

	return nil
}

func (g *Game) setEngine(eng *engine.Engine) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.engine = eng
}

func (g *Game) getEngine() *engine.Engine {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.engine
}

func (g *Game) resetEngine() {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.engine = nil
}

// Update doesn't do anything, because the game state is updated by the server.
func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.engine != nil {
		g.engine.Draw(screen)
	} else {
		// Draw a "client disconnected" text over all screen.
		img := ebiten.NewImageFromImage(screen)
		img.Fill(color.RGBA{0x80, 0x80, 0x80, 0xff})
		text := "Client disconnected"
		ebitenutil.DebugPrintAt(img, text, 0, 0)
		screen.DrawImage(img, nil)
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return camera.WIDTH, camera.HEIGHT
}
