package main

import (
	"net"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/logging"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/server"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

func main() {
	logging.Init()

	game := server.NewGame()
	gs := server.New(game, engine.New)

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
