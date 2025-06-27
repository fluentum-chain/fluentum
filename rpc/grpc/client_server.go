package coregrpc

import (
	"net"

	"golang.org/x/net/context"
	grpcserver "google.golang.org/grpc"

	tmnet "github.com/fluentum-chain/fluentum/libs/net"
	grpc "github.com/fluentum-chain/fluentum/proto/tendermint/rpc/grpc"
)

// Config is an gRPC server configuration.
type Config struct {
	MaxOpenConnections int
}

// StartGRPCServer starts a new gRPC BroadcastAPIServer using the given
// net.Listener.
// NOTE: This function blocks - you may want to call it in a go-routine.
func StartGRPCServer(ln net.Listener) error {
	server := grpcserver.NewServer()
	grpc.RegisterBroadcastAPIServer(server, &broadcastAPI{})
	return server.Serve(ln)
}

// StartGRPCClient dials the gRPC server using protoAddr and returns a new
// BroadcastAPIClient.
func StartGRPCClient(protoAddr string) grpc.BroadcastAPIClient {
	//nolint:staticcheck // SA1019 Existing use of deprecated but supported dial option.
	conn, err := grpcserver.Dial(protoAddr, grpcserver.WithInsecure(), grpcserver.WithContextDialer(dialerFunc))
	if err != nil {
		panic(err)
	}
	return grpc.NewBroadcastAPIClient(conn)
}

func dialerFunc(ctx context.Context, addr string) (net.Conn, error) {
	return tmnet.Connect(addr)
}
