package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	mu "github.com/c4t-but-s4d/cbs-go/multiproto"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/logging"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/server"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/sprites"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

func main() {
	logging.Init()

	// TODO: bind to viper.
	listen := pflag.StringP("listen", "s", ":8080", "address to listen on")
	level := pflag.StringP("level", "l", "", "level to load")
	snapshotsDir := pflag.String("snapshots-dir", "snapshots", "directory to save snapshots to")
	pflag.Parse()

	game := server.NewGame(*snapshotsDir)
	smng := sprites.NewManager()

	gs := server.New(game, func() (*engine.Engine, error) {
		files, err := os.ReadDir(*snapshotsDir)
		if err != nil {
			return nil, fmt.Errorf("listing snapshots directory: %w", err)
		}

		var snapshotFilename string

		for _, f := range files {
			if !f.Type().IsRegular() || !strings.HasPrefix(f.Name(), "snapshot") {
				continue
			}

			snapshotFilename = f.Name()
		}

		engineConfig := engine.Config{
			SnapshotsDir: *snapshotsDir,
			Level:        *level,
		}

		if snapshotFilename != "" {
			data, err := os.ReadFile(filepath.Join(*snapshotsDir, snapshotFilename))
			if err != nil {
				return nil, fmt.Errorf("reading snapshot file: %w", err)
			}

			e, err := engine.NewFromSnapshot(engineConfig, &engine.Snapshot{Data: data}, smng)
			if err != nil {
				return nil, fmt.Errorf("creating engine from snapshot: %w", err)
			}
			return e, nil
		}

		e, err := engine.New(engineConfig, smng)
		if err != nil {
			return nil, fmt.Errorf("creating engine without snapshot: %w", err)
		}
		return e, nil
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
