package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/fluentum-chain/fluentum/crypto/ed25519"
	"github.com/fluentum-chain/fluentum/p2p"
	tmp2p "github.com/fluentum-chain/fluentum/proto/fluentum/p2p"
	"github.com/gogo/protobuf/proto"
)

func main() {
	baseDir := flag.String("base", ".", `where the "corpus" directory will live`)
	flag.Parse()

	initCorpus(*baseDir)
}

//nolint:gosec
func initCorpus(rootDir string) {
	log.SetFlags(0)

	corpusDir := filepath.Join(rootDir, "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		log.Fatalf("Creating %q err: %v", corpusDir, err)
	}
	sizes := []int{0, 1, 2, 17, 5, 31}

	// Make the PRNG predictable
	rand.Seed(10)

	for _, n := range sizes {
		var addrs []*p2p.NetAddress

		// IPv4 addresses
		for i := 0; i < n; i++ {
			privKey := ed25519.GenPrivKey()
			addr := fmt.Sprintf(
				"%s@%v.%v.%v.%v:26656",
				p2p.PubKeyToID(privKey.PubKey()),
				rand.Int()%256,
				rand.Int()%256,
				rand.Int()%256,
				rand.Int()%256,
			)
			netAddr, _ := p2p.NewNetAddressString(addr)
			addrs = append(addrs, netAddr)
		}

		// IPv6 addresses
		privKey := ed25519.GenPrivKey()
		ipv6a, err := p2p.NewNetAddressString(
			fmt.Sprintf("%s@[ff02::1:114]:26656", p2p.PubKeyToID(privKey.PubKey())))
		if err != nil {
			log.Fatalf("can't create a new netaddress: %v", err)
		}
		addrs = append(addrs, ipv6a)

		addrsProto := p2p.NetAddressesToProto(addrs)
		msg := tmp2p.Message{
			Sum: &tmp2p.Message_PexAddrs{
				PexAddrs: &tmp2p.PexAddrs{Addrs: addrsProto},
			},
		}
		bz, err := proto.Marshal(&msg)
		if err != nil {
			log.Fatalf("unable to marshal: %v", err)
		}

		filename := filepath.Join(rootDir, "corpus", fmt.Sprintf("%d", n))

		if err := os.WriteFile(filename, bz, 0o644); err != nil {
			log.Fatalf("can't write %X to %q: %v", bz, filename, err)
		}

		log.Printf("wrote %q", filename)
	}
}
