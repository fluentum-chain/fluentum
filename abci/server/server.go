/*
Package server is used to start a new ABCI server.

It contains two server implementation:
  - gRPC server
  - socket server
*/
package server

import (
	"fmt"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/libs/service"
)

func NewServer(protoAddr, transport string, app cmtabci.Application) (service.Service, error) {
	var s service.Service
	var err error
	switch transport {
	case "socket":
		s = NewSocketServer(protoAddr, app)
	case "grpc":
		// Note: gRPC server is currently disabled in CometBFT v0.38+
		// The app parameter is not used in the current implementation
		s = NewGRPCServer(protoAddr)
	default:
		err = fmt.Errorf("unknown server type %s", transport)
	}
	return s, err
}
