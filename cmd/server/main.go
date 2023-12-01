package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/fonts"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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
	headless := pflag.BoolP("headless", "h", false, "disable GUI")
	pflag.Parse()

	game := server.NewGame(*snapshotsDir)
	smng := sprites.NewManager()
	fntmng := fonts.NewManager()

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

			e, err := engine.NewFromSnapshot(engineConfig, &engine.Snapshot{Data: data}, smng, fntmng)
			if err != nil {
				return nil, fmt.Errorf("creating engine from snapshot: %w", err)
			}
			return e, nil
		}

		e, err := engine.New(engineConfig, smng, fntmng)
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-ctx.Done()

		logrus.Info("stopping game")
		game.Shutdown()

		logrus.Info("stopping server")

		shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logrus.Fatalf("error stopping http server: %v", err)
		}
	}()

	ebiten.SetWindowTitle("ctfcup-2023-igra server")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeOnlyFullscreenEnabled)

	if *headless {
		logrus.Info("running in headless mode")
		<-ctx.Done()
	} else {
		logrus.Info("starting game")
		if err := ebiten.RunGame(game); err != nil && !errors.Is(err, server.ErrGameShutdown) {
			logrus.Fatalf("failed to run game: %v", err)
		}
	}

	logrus.Info("finished")
}
