package main

import (
	"flag"
	stdlog "log"
	"os"

	"github.com/cometbft/cometbft/abci/server"
	tmlog "github.com/cometbft/cometbft/libs/log"
	"github.com/fluentum-chain/fluentum/fluentum/abci/example/counter"
)

func main() {
	addr := flag.String("address", "tcp://0.0.0.0:26658", "Listen address")
	serial := flag.Bool("serial", false, "Enable serial mode")
	flag.Parse()

	app := counter.NewApplication(*serial)
	srv, err := server.NewServer(*addr, "socket", app)
	if err != nil {
		stdlog.Fatalf("failed to create ABCI server: %v", err)
	}
	srv.SetLogger(tmlog.NewTMLogger(tmlog.NewSyncWriter(os.Stdout)))
	if err := srv.Start(); err != nil {
		stdlog.Fatalf("failed to start ABCI server: %v", err)
	}
	defer srv.Stop()

	// Block forever
	select {}
}
