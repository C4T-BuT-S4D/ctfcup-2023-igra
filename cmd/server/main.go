package main

import (
	"encoding/json"
	"net"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/logging"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/server"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

type Config struct {
	SignKey string `json:"sign_key"`
}

func main() {
	logging.Init()

	// TODO: use flags library

	cfgPath := "configs/server.json"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfgFile, err := os.Open(cfgPath)
	if err != nil {
		logrus.Fatalf("opening config file %s: %v", cfgPath, err)
	}

	var cfg Config
	if err := json.NewDecoder(cfgFile).Decode(&cfg); err != nil {
		logrus.Fatalf("decoding config: %v", err)
	}

	game := server.NewGame()
	gs := server.New(game, func() (*engine.Engine, error) {
		return engine.New([]byte(cfg.SignKey))
	})

	s := grpc.NewServer()
	gameserverpb.RegisterGameServerServiceServer(s, gs)
	reflection.Register(s)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		logrus.Fatalf("creating listener: %v", err)
	}

	go func() {
		logrus.Infof("starting server on %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			logrus.Fatalf("failed to serve: %v", err)
		}
	}()

	ebiten.SetWindowTitle("ctfcup-2023-igra server")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeOnlyFullscreenEnabled)

	logrus.Infof("starting game")
	if err := ebiten.RunGame(game); err != nil {
		logrus.Fatalf("failed to run game: %v", err)
	}
}
