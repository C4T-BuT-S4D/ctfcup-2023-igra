package server

import (
	"context"
	"errors"
	"io"

	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
)

func New(game *Game, factory engine.Factory) *GameServer {
	return &GameServer{
		factory:    factory,
		game:       game,
		numStreams: atomic.NewInt64(0),
	}
}

type GameServer struct {
	gameserverpb.UnimplementedGameServerServiceServer

	factory    engine.Factory
	numStreams *atomic.Int64
	game       *Game
}

func (g *GameServer) Ping(context.Context, *gameserverpb.PingRequest) (*gameserverpb.PingResponse, error) {
	return &gameserverpb.PingResponse{}, nil
}

func (g *GameServer) ProcessEvent(stream gameserverpb.GameServerService_ProcessEventServer) error {
	defer g.numStreams.Dec()
	if g.numStreams.Inc() > 1 {
		return status.Error(codes.ResourceExhausted, "only one client connection allowed")
	}

	p, _ := peer.FromContext(stream.Context())
	if p == nil {
		return status.Error(codes.FailedPrecondition, "failed to get peer info")
	}
	logrus.Infof("new connection from %v", p.Addr)

	eng, err := g.factory()
	if err != nil {
		return status.Errorf(codes.Internal, "creating engine: %v", err)
	}
	g.game.engine.Store(eng)
	g.game.events = make(chan *gameserverpb.ClientEvent)

	defer func() {
		g.game.engine.Store(nil)
		close(g.game.events)
	}()

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logrus.Info("client disconnected")
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to read from stream: %v", err)
		}
		logrus.Debugf("received event: %v", req)

		if req.Event == nil {
			return status.Error(codes.InvalidArgument, "event is nil")
		}
		if err := eng.ValidateChecksum(req.Checksum); err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid checksum: %v", err)
		}
		g.game.events <- req.Event
	}
}
