package server

import (
	"net"

	"google.golang.org/grpc"

	tmnet "github.com/fluentum-chain/fluentum/libs/net"
	"github.com/fluentum-chain/fluentum/libs/service"
	// "github.com/cometbft/cometbft/abci/types"
)

type GRPCServer struct {
	service.BaseService

	proto    string
	addr     string
	listener net.Listener
	server   *grpc.Server

	// app types.ABCIApplicationServer // Disabled
}

// NewGRPCServer returns a new gRPC ABCI server (disabled)
func NewGRPCServer(protoAddr string) service.Service {
	proto, addr := tmnet.ProtocolAndAddress(protoAddr)
	s := &GRPCServer{
		proto:    proto,
		addr:     addr,
		listener: nil,
		// app:      app, // Disabled
	}
	s.BaseService = *service.NewBaseService(nil, "ABCIServer", s)
	return s
}

// OnStart starts the gRPC service.
func (s *GRPCServer) OnStart() error {
	// TODO: Native gRPC ABCI server support is not available in CometBFT v0.38+.
	// The previous implementation using types.ABCIApplicationServer and types.RegisterABCIApplicationServer is not supported.
	// If gRPC support is needed, implement a custom gRPC server or use a compatible proxy.
	// For now, this file is disabled to allow the build to proceed.

	ln, err := net.Listen(s.proto, s.addr)
	if err != nil {
		return err
	}

	s.listener = ln
	s.server = grpc.NewServer()
	// types.RegisterABCIApplicationServer(s.server, s.app)

	s.Logger.Info("Listening", "proto", s.proto, "addr", s.addr)
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			s.Logger.Error("Error serving gRPC server", "err", err)
		}
	}()
	return nil
}

// OnStop stops the gRPC server.
func (s *GRPCServer) OnStop() {
	s.server.Stop()
}
