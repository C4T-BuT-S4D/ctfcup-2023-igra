package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/logging"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

type Config struct {
	SignKey string `json:"sign_key"`
}

func NewGame(ctx context.Context, cfg *Config, client gameserverpb.GameServerServiceClient) (*Game, error) {
	e, err := engine.New([]byte(cfg.SignKey))
	if err != nil {
		return nil, fmt.Errorf("initializing engine: %w", err)
	}

	g := &Game{
		Engine: e,
		ctx:    ctx,

		recvErrChan:     make(chan error, 1),
		serverEventChan: make(chan *gameserverpb.ServerEvent),
	}

	if client != nil {
		eventStream, err := client.ProcessEvent(ctx)
		if err != nil {
			return nil, fmt.Errorf("opening event stream: %w", err)
		}
		g.stream = eventStream

		go func() {
			serverEvent, err := eventStream.Recv()
			if err != nil {
				g.recvErrChan <- err
				return
			}
			g.serverEventChan <- serverEvent
		}()
	}

	return g, nil
}

type Game struct {
	Engine *engine.Engine
	stream gameserverpb.GameServerService_ProcessEventClient
	ctx    context.Context

	serverEventChan chan *gameserverpb.ServerEvent
	recvErrChan     chan error
}

func (g *Game) Update() error {
	if err := g.ctx.Err(); err != nil {
		return err
	}

	inp := input.New()
	inp.Update()

	select {
	case err := <-g.recvErrChan:
		return fmt.Errorf("server returned error: %w", err)
	default:
	}

	checksum, err := g.Engine.Checksum()
	if err != nil {
		return fmt.Errorf("calculating checksum: %w", err)
	}

	if g.stream != nil {
		if err := g.stream.Send(&gameserverpb.ClientEventRequest{
			Checksum: checksum,
			Event:    &gameserverpb.ClientEvent{KeysPressed: inp.ToProto()},
		}); err != nil {
			return fmt.Errorf("failed to send event to the server: %w", err)
		}
	}

	if err := g.Engine.Update(inp); err != nil {
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
	logging.Init()

	// TODO: bind to viper.
	configPath := pflag.StringP("config", "c", "configs/client.json", "path to config file")
	serverAddr := pflag.StringP("server", "s", "127.0.0.1:8080", "server address")
	pflag.Parse()

	cfgFile, err := os.Open(*configPath)
	if err != nil {
		logrus.Fatalf("opening config file: %v", err)
	}

	var cfg Config
	if err := json.NewDecoder(cfgFile).Decode(&cfg); err != nil {
		logrus.Fatalf("decoding config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var client gameserverpb.GameServerServiceClient
	if *serverAddr != "" {
		conn, err := grpc.DialContext(ctx, *serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logrus.Fatalf("Failed to connect to server: %v", err)
		}
		client = gameserverpb.NewGameServerServiceClient(conn)
	}

	g, err := NewGame(ctx, &cfg, client)
	if err != nil {
		logrus.Fatalf("Failed to create game: %v", err)
	}

	ebiten.SetWindowTitle("ctfcup-2023-igra client")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil && !errors.Is(err, context.Canceled) {
		logrus.Fatalf("Failed to run game: %v", err)
	}
}
