package kvstore

import (
	"context"

	cmtpc "github.com/cometbft/cometbft/proto/tendermint/crypto"
	abci "github.com/fluentum-chain/fluentum/abci/types"
	tmrand "github.com/fluentum-chain/fluentum/libs/rand"
)

// RandVal creates one random validator, with a key derived
// from the input value
func RandVal(i int) abci.ValidatorUpdate {
	pubkey := tmrand.Bytes(32)
	power := tmrand.Uint16() + 1
	v := abci.ValidatorUpdate{
		PubKey: cmtpc.PublicKey{
			Sum: &cmtpc.PublicKey_Ed25519{
				Ed25519: pubkey,
			},
		},
		Power: int64(power),
	}
	return v
}

// RandVals returns a list of cnt validators for initializing
// the application. Note that the keys are deterministically
// derived from the index in the array, while the power is
// random (Change this if not desired)
func RandVals(cnt int) []abci.ValidatorUpdate {
	res := make([]abci.ValidatorUpdate, cnt)
	for i := 0; i < cnt; i++ {
		v := RandVal(i)
		res[i] = v
	}
	return res
}

// InitKVStore initializes the kvstore app with some data,
// which allows tests to pass and is fine as long as you
// don't make any tx that modify the validator state
func InitKVStore(app *PersistentKVStoreApplication) {
	app.InitChain(context.Background(), &abci.InitChainRequest{
		Validators: RandVals(1),
	})
}
