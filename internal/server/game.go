package server

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/sirupsen/logrus"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/camera"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

var ErrGameShutdown = errors.New("game is shut down")

type Game struct {
	IsWin  bool
	engine *engine.Engine

	lock sync.Mutex

	snapshotsDir string
	shutdown     chan struct{}
}

func NewGame(snapshotsDir string) *Game {
	return &Game{
		snapshotsDir: snapshotsDir,
		shutdown:     make(chan struct{}),
	}
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
	select {
	case <-g.shutdown:
		return ErrGameShutdown
	default:
		return nil
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.engine != nil {
		g.engine.Draw(screen)
	} else {
		team := strings.Split(os.Getenv("AUTH_TOKEN"), ":")[0]
		action := "disconnected"
		if g.IsWin {
			action = "won"
		}

		text := fmt.Sprintf("Team %s: %s", team, action)
		ebitenutil.DebugPrintAt(screen, text, 32, 32)
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return camera.WIDTH, camera.HEIGHT
}

func (g *Game) Shutdown() {
	close(g.shutdown)
}
