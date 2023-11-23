package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
	"io"
	"sync/atomic"
)

func New(factory func() (*engine.Engine, error)) *GameServer {
	return &GameServer{factory: factory}
}

type GameServer struct {
	factory func() (*engine.Engine, error)
	gameserverpb.UnimplementedGameServerServer
	numStreams atomic.Int64
}

func (g *GameServer) Ping(ctx context.Context, request *gameserverpb.PingRequest) (*gameserverpb.PingResponse, error) {
	return &gameserverpb.PingResponse{}, nil
}

func (g *GameServer) ProcessEvent(stream gameserverpb.GameServer_ProcessEventServer) error {
	defer g.numStreams.Add(-1)
	if g.numStreams.Add(1) > 1 {
		return errors.New("only one game session is allowed per client")
	}

	eng, err := g.factory()
	if err != nil {
		return fmt.Errorf("creating engine: %w", err)
	}

	for {
		in, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to read from stream: %w", err)
		}

		if !g.isValidChecksum(in) {
			return fmt.Errorf("invalid checksum")
		}

		inp := input.FromProtoEvent(in.Event)
		fmt.Printf("input: %+v\n", inp)
		if err := eng.Update(inp); err != nil {
			return fmt.Errorf("updating engine state: %w", err)
		}

		//out := &gameserverpb.ServerEvent{}
		//if err := stream.Send(out); err != nil {
		//	return fmt.Errorf("failed to send : %w", err)
		//}
	}
}

func (g *GameServer) isValidChecksum(request *gameserverpb.ClientEventRequest) bool {
	// TODO(jnovikov): Implement checksum validation.
	return true
}
