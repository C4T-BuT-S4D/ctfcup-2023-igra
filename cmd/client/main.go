package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"

	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

func NewGame(ctx context.Context, client gameserverpb.GameServerClient) (*Game, error) {
	e, err := engine.New()
	if err != nil {
		return nil, fmt.Errorf("initializing engine: %w", err)
	}

	g := &Game{
		Engine: e,
	}
	g.recvErrChan = make(chan error, 1)
	g.serverEventChan = make(chan *gameserverpb.ServerEvent)

	if client != nil {
		es, err := client.ProcessEvent(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating event stream: %w", err)
		}
		g.cli = es

		go func() {
			se, err := es.Recv()
			if err != nil {
				g.recvErrChan <- err
				return
			}
			g.serverEventChan <- se
		}()
	}

	return g, nil
}

type Game struct {
	Engine *engine.Engine
	cli    gameserverpb.GameServer_ProcessEventClient

	serverEventChan chan *gameserverpb.ServerEvent
	recvErrChan     chan error
}

func (g *Game) Update() error {
	var i input.Input

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		i.WPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		i.APressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		i.SPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		i.DPressed = true
	}

	select {
	case err := <-g.recvErrChan:
		return fmt.Errorf("server returned error: %w", err)
	default:
	}

	// TODO(b1r1b1r1): Provide interface for checksum calculation.
	checksum := ""
	if g.cli != nil {
		if err := g.cli.Send(&gameserverpb.ClientEventRequest{
			Checksum: checksum,
			Event:    input.ToProtoEvent(&i),
		}); err != nil {
			return fmt.Errorf("failed to send event to the server: %w", err)
		}
	}

	if err := g.Engine.Update(&i); err != nil {
		return fmt.Errorf("updating engine state: %w", err)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Engine.Draw(screen)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	serverHost := ""
	if len(os.Args) > 1 {
		serverHost = os.Args[1]
	}

	var client gameserverpb.GameServerClient
	if serverHost != "" {
		conn, err := grpc.Dial(serverHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to server: %v", err)
		}
		defer conn.Close()
		client = gameserverpb.NewGameServerClient(conn)
	}

	g, err := NewGame(context.Background(), client)
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatalf("Failed to run game: %v", err)
	}
}
