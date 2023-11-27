package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	mu "github.com/c4t-but-s4d/cbs-go/multiproto"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
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

	// TODO: bind to viper.
	configPath := pflag.StringP("config", "c", "configs/server.json", "path to config file")
	listen := pflag.StringP("listen", "l", ":8080", "address to listen on")
	pflag.Parse()

	cfgFile, err := os.Open(*configPath)
	if err != nil {
		logrus.Fatalf("error opening config file: %v", err)
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

	multiProtoHandler := mu.NewHandler(s)
	httpServer := &http.Server{
		Addr:    *listen,
		Handler: multiProtoHandler,
	}

	go func() {
		logrus.Infof("starting server on %v", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("error running http server: %v", err)
		}
	}()

	ebiten.SetWindowTitle("ctfcup-2023-igra server")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeOnlyFullscreenEnabled)

	logrus.Infof("starting game")
	if err := ebiten.RunGame(game); err != nil {
		logrus.Fatalf("failed to run game: %v", err)
	}
}
