package state_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"

	abci "github.com/cometbft/cometbft/abci/types"
	cfg "github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/crypto"
	"github.com/fluentum-chain/fluentum/crypto/ed25519"
	tmrand "github.com/fluentum-chain/fluentum/libs/rand"
	tmproto "github.com/fluentum-chain/fluentum/proto/tendermint/types"
	sm "github.com/fluentum-chain/fluentum/state"
	"github.com/fluentum-chain/fluentum/types"
)

func TestStoreLoadValidators(t *testing.T) {
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	val, _ := types.RandValidator(true, 10)
	vals := types.NewValidatorSet([]*types.Validator{val})

	// 1) LoadValidators loads validators using a height where they were last changed
	err := sm.SaveValidatorsInfo(stateDB, 1, 1, vals)
	require.NoError(t, err)
	err = sm.SaveValidatorsInfo(stateDB, 2, 1, vals)
	require.NoError(t, err)
	loadedVals, err := stateStore.LoadValidators(2)
	require.NoError(t, err)
	assert.NotZero(t, loadedVals.Size())

	// 2) LoadValidators loads validators using a checkpoint height

	err = sm.SaveValidatorsInfo(stateDB, sm.ValSetCheckpointInterval, 1, vals)
	require.NoError(t, err)

	loadedVals, err = stateStore.LoadValidators(sm.ValSetCheckpointInterval)
	require.NoError(t, err)
	assert.NotZero(t, loadedVals.Size())
}

func BenchmarkLoadValidators(b *testing.B) {
	const valSetSize = 100

	config := cfg.ResetTestRoot("state_")
	defer os.RemoveAll(config.RootDir)
	dbType := dbm.BackendType(config.DBBackend)
	stateDB, err := dbm.NewDB("state", dbType, config.DBDir())
	require.NoError(b, err)
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	state, err := stateStore.LoadFromDBOrGenesisFile(config.GenesisFile())
	if err != nil {
		b.Fatal(err)
	}

	state.Validators = genValSet(valSetSize)
	state.NextValidators = state.Validators.CopyIncrementProposerPriority(1)
	err = stateStore.Save(state)
	require.NoError(b, err)

	for i := 10; i < 10000000000; i *= 10 { // 10, 100, 1000, ...
		i := i
		if err := sm.SaveValidatorsInfo(stateDB,
			int64(i), state.LastHeightValidatorsChanged, state.NextValidators); err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("height=%d", i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, err := stateStore.LoadValidators(int64(i))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestPruneStates(t *testing.T) {
	testcases := map[string]struct {
		makeHeights  int64
		pruneFrom    int64
		pruneTo      int64
		expectErr    bool
		expectVals   []int64
		expectParams []int64
		expectABCI   []int64
	}{
		"error on pruning from 0":      {100, 0, 5, true, nil, nil, nil},
		"error when from > to":         {100, 3, 2, true, nil, nil, nil},
		"error when from == to":        {100, 3, 3, true, nil, nil, nil},
		"error when to does not exist": {100, 1, 101, true, nil, nil, nil},
		"prune all":                    {100, 1, 100, false, []int64{93, 100}, []int64{95, 100}, []int64{100}},
		"prune some": {10, 2, 8, false, []int64{1, 3, 8, 9, 10},
			[]int64{1, 5, 8, 9, 10}, []int64{1, 8, 9, 10}},
		"prune across checkpoint": {100001, 1, 100001, false, []int64{99993, 100000, 100001},
			[]int64{99995, 100001}, []int64{100001}},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			db := dbm.NewMemDB()
			stateStore := sm.NewStore(db, sm.StoreOptions{
				DiscardABCIResponses: false,
			})
			pk := ed25519.GenPrivKey().PubKey()

			// Generate a bunch of state data. Validators change for heights ending with 3, and
			// parameters when ending with 5.
			validator := &types.Validator{Address: tmrand.Bytes(crypto.AddressSize), VotingPower: 100, PubKey: pk}
			validatorSet := &types.ValidatorSet{
				Validators: []*types.Validator{validator},
				Proposer:   validator,
			}
			valsChanged := int64(0)
			paramsChanged := int64(0)

			for h := int64(1); h <= tc.makeHeights; h++ {
				if valsChanged == 0 || h%10 == 2 {
					valsChanged = h + 1 // Have to add 1, since NextValidators is what's stored
				}
				if paramsChanged == 0 || h%10 == 5 {
					paramsChanged = h
				}

				state := sm.State{
					InitialHeight:   1,
					LastBlockHeight: h - 1,
					Validators:      validatorSet,
					NextValidators:  validatorSet,
					ConsensusParams: tmproto.ConsensusParams{
						Block: tmproto.BlockParams{MaxBytes: 10e6},
					},
					LastHeightValidatorsChanged:      valsChanged,
					LastHeightConsensusParamsChanged: paramsChanged,
				}

				if state.LastBlockHeight >= 1 {
					state.LastValidators = state.Validators
				}

				err := stateStore.Save(state)
				require.NoError(t, err)

				err = stateStore.SaveABCIResponses(h, &ABCIResponses{
					DeliverTxs: []*abci.ExecTxResult{
						{Data: []byte{1}},
						{Data: []byte{2}},
						{Data: []byte{3}},
					},
				})
				require.NoError(t, err)
			}

			// Test assertions
			err := stateStore.PruneStates(tc.pruneFrom, tc.pruneTo)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			expectVals := sliceToMap(tc.expectVals)
			expectParams := sliceToMap(tc.expectParams)
			expectABCI := sliceToMap(tc.expectABCI)

			for h := int64(1); h <= tc.makeHeights; h++ {
				vals, err := stateStore.LoadValidators(h)
				if expectVals[h] {
					require.NoError(t, err, "validators height %v", h)
					require.NotNil(t, vals)
				} else {
					require.Error(t, err, "validators height %v", h)
					require.Equal(t, sm.ErrNoValSetForHeight{Height: h}, err)
				}

				params, err := stateStore.LoadConsensusParams(h)
				if expectParams[h] {
					require.NoError(t, err, "params height %v", h)
					require.False(t, params.Equal(&tmproto.ConsensusParams{}))
				} else {
					require.Error(t, err, "params height %v", h)
				}

				abci, err := stateStore.LoadABCIResponses(h)
				if expectABCI[h] {
					require.NoError(t, err, "abci height %v", h)
					require.NotNil(t, abci)
				} else {
					require.Error(t, err, "abci height %v", h)
					require.Equal(t, sm.ErrNoABCIResponsesForHeight{Height: h}, err)
				}
			}
		})
	}
}

func TestABCIResponsesResultsHash(t *testing.T) {
	responses := &ABCIResponses{
		DeliverTxs: []*abci.ExecTxResult{
			{Code: 32, Data: []byte("Hello"), Log: "Huh?"},
		},
	}

	root := sm.ABCIResponsesResultsHash(responses)

	// root should be Merkle tree root of DeliverTxs responses
	results := types.NewResults(responses.DeliverTxs)
	assert.Equal(t, root, results.Hash())

	// test we can prove first DeliverTx
	proof := results.ProveResult(0)
	bz, err := results[0].Marshal()
	require.NoError(t, err)
	assert.NoError(t, proof.Verify(root, bz))
}

func sliceToMap(s []int64) map[int64]bool {
	m := make(map[int64]bool, len(s))
	for _, i := range s {
		m[i] = true
	}
	return m
}

func TestLastABCIResponses(t *testing.T) {
	// create an empty state store.
	t.Run("Not persisting responses", func(t *testing.T) {
		stateDB := dbm.NewMemDB()
		stateStore := sm.NewStore(stateDB, sm.StoreOptions{
			DiscardABCIResponses: false,
		})
		responses, err := stateStore.LoadABCIResponses(1)
		require.Error(t, err)
		require.Nil(t, responses)
		// stub the abciresponses.
		response1 := &ABCIResponses{
			BeginBlock: &abci.ResponseBeginBlock{},
			DeliverTxs: []*abci.ExecTxResult{
				{Code: 32, Data: []byte("Hello"), Log: "Huh?"},
			},
			EndBlock: &abci.ResponseEndBlock{},
		}
		// create new db and state store and set discard abciresponses to false.
		stateDB = dbm.NewMemDB()
		stateStore = sm.NewStore(stateDB, sm.StoreOptions{DiscardABCIResponses: false})
		height := int64(10)
		// save the last abci response.
		err = stateStore.SaveABCIResponses(height, response1)
		require.NoError(t, err)
		// search for the last abciresponse and check if it has saved.
		lastResponse, err := stateStore.LoadLastABCIResponse(height)
		require.NoError(t, err)
		// check to see if the saved response height is the same as the loaded height.
		assert.Equal(t, lastResponse, response1)
		// use an incorret height to make sure the state store errors.
		_, err = stateStore.LoadLastABCIResponse(height + 1)
		assert.Error(t, err)
		// check if the abci response didnt save in the abciresponses.
		responses, err = stateStore.LoadABCIResponses(height)
		require.NoError(t, err, responses)
		require.Equal(t, response1, responses)
	})

	t.Run("persisting responses", func(t *testing.T) {
		stateDB := dbm.NewMemDB()
		height := int64(10)
		// stub the second abciresponse.
		response2 := &ABCIResponses{
			BeginBlock: &abci.ResponseBeginBlock{},
			DeliverTxs: []*abci.ExecTxResult{
				{Code: 44, Data: []byte("Hello again"), Log: "????"},
			},
			EndBlock: &abci.ResponseEndBlock{},
		}
		// create a new statestore with the responses on.
		stateStore := sm.NewStore(stateDB, sm.StoreOptions{
			DiscardABCIResponses: true,
		})
		// save an additional response.
		err := stateStore.SaveABCIResponses(height+1, response2)
		require.NoError(t, err)
		// check to see if the response saved by calling the last response.
		lastResponse2, err := stateStore.LoadLastABCIResponse(height + 1)
		require.NoError(t, err)
		// check to see if the saved response height is the same as the loaded height.
		assert.Equal(t, response2, lastResponse2)
		// should error as we are no longer saving the response.
		_, err = stateStore.LoadABCIResponses(height + 1)
		assert.Equal(t, sm.ErrABCIResponsesNotPersisted, err)
	})

}
