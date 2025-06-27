package main

import (
	"os"

	"github.com/fluentum-chain/fluentum/abci/example/counterlib"
)

func main() {
	// Example: create and use the counterlib Application
	_ = counterlib.NewApplication(false)
	// You can add code here to run the app as needed, or leave this as a stub
	os.Exit(0)
}
