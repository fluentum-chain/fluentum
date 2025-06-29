package state

import (
	"errors"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/gogo/protobuf/proto"

	tmmath "github.com/fluentum-chain/fluentum/libs/math"
	tmos "github.com/fluentum-chain/fluentum/libs/os"
	tmstate "github.com/fluentum-chain/fluentum/proto/tendermint/state"
	tmproto "github.com/fluentum-chain/fluentum/proto/tendermint/types"
	"github.com/fluentum-chain/fluentum/types"
)

const (
	// persist validators every valSetCheckpointInterval blocks to avoid
	// LoadValidators taking too much time.
	// https://github.com/fluentum-chain/fluentum/pull/3438
	// 100000 results in ~ 100ms to get 100 validators (see BenchmarkLoadValidators)
	valSetCheckpointInterval = 100000
)

//------------------------------------------------------------------------

func calcValidatorsKey(height int64) []byte {
	return []byte(fmt.Sprintf("validatorsKey:%v", height))
}

func calcConsensusParamsKey(height int64) []byte {
	return []byte(fmt.Sprintf("consensusParamsKey:%v", height))
}

func calcABCIResponsesKey(height int64) []byte {
	return []byte(fmt.Sprintf("abciResponsesKey:%v", height))
}

//----------------------

var (
	lastABCIResponseKey = []byte("lastABCIResponseKey")
)

//go:generate ../scripts/mockery_generate.sh Store

// Store defines the state store interface
//
// It is used to retrieve current state and save and load ABCI responses,
// validators and consensus parameters
type Store interface {
	// LoadFromDBOrGenesisFile loads the most recent state.
	// If the chain is new it will use the genesis file from the provided genesis file path as the current state.
	LoadFromDBOrGenesisFile(string) (State, error)
	// LoadFromDBOrGenesisDoc loads the most recent state.
	// If the chain is new it will use the genesis doc as the current state.
	LoadFromDBOrGenesisDoc(*types.GenesisDoc) (State, error)
	// Load loads the current state of the blockchain
	Load() (State, error)
	// LoadValidators loads the validator set at a given height
	LoadValidators(int64) (*types.ValidatorSet, error)
	// LoadABCIResponses loads the abciResponse for a given height
	LoadABCIResponses(int64) (*ABCIResponses, error)
	// LoadLastABCIResponse loads the last abciResponse for a given height
	LoadLastABCIResponse(int64) (*ABCIResponses, error)
	// LoadConsensusParams loads the consensus params for a given height
	LoadConsensusParams(int64) (tmproto.ConsensusParams, error)
	// Save overwrites the previous state with the updated one
	Save(State) error
	// SaveABCIResponses saves ABCIResponses for a given height
	SaveABCIResponses(int64, *ABCIResponses) error
	// Bootstrap is used for bootstrapping state when not starting from a initial height.
	Bootstrap(State) error
	// PruneStates takes the height from which to start prning and which height stop at
	PruneStates(int64, int64) error
	// Close closes the connection with the database
	Close() error
}

// dbStore wraps a db (github.com/cometbft/cometbft-db)
type dbStore struct {
	db dbm.DB

	StoreOptions
}

type StoreOptions struct {

	// DiscardABCIResponses determines whether or not the store
	// retains all ABCIResponses. If DiscardABCiResponses is enabled,
	// the store will maintain only the response object from the latest
	// height.
	DiscardABCIResponses bool
}

var _ Store = (*dbStore)(nil)

// NewStore creates the dbStore of the state pkg.
func NewStore(db dbm.DB, options StoreOptions) Store {
	return dbStore{db, options}
}

// LoadStateFromDBOrGenesisFile loads the most recent state from the database,
// or creates a new one from the given genesisFilePath.
func (store dbStore) LoadFromDBOrGenesisFile(genesisFilePath string) (State, error) {
	state, err := store.Load()
	if err != nil {
		return State{}, err
	}
	if state.IsEmpty() {
		var err error
		state, err = MakeGenesisStateFromFile(genesisFilePath)
		if err != nil {
			return state, err
		}
	}

	return state, nil
}

// LoadStateFromDBOrGenesisDoc loads the most recent state from the database,
// or creates a new one from the given genesisDoc.
func (store dbStore) LoadFromDBOrGenesisDoc(genesisDoc *types.GenesisDoc) (State, error) {
	state, err := store.Load()
	if err != nil {
		return State{}, err
	}

	if state.IsEmpty() {
		var err error
		state, err = MakeGenesisState(genesisDoc)
		if err != nil {
			return state, err
		}
	}

	return state, nil
}

// LoadState loads the State from the database.
func (store dbStore) Load() (State, error) {
	return store.loadState(stateKey)
}

func (store dbStore) loadState(key []byte) (state State, err error) {
	buf, err := store.db.Get(key)
	if err != nil {
		return state, err
	}
	if len(buf) == 0 {
		return state, nil
	}

	sp := new(tmstate.State)

	err = proto.Unmarshal(buf, sp)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		tmos.Exit(fmt.Sprintf(`LoadState: Data has been corrupted or its spec has changed:
		%v\n`, err))
	}

	sm, err := FromProto(sp)
	if err != nil {
		return state, err
	}

	return *sm, nil
}

// Save persists the State, the ValidatorsInfo, and the ConsensusParamsInfo to the database.
// This flushes the writes (e.g. calls SetSync).
func (store dbStore) Save(state State) error {
	return store.save(state, stateKey)
}

func (store dbStore) save(state State, key []byte) error {
	nextHeight := state.LastBlockHeight + 1
	// If first block, save validators for the block.
	if nextHeight == 1 {
		nextHeight = state.InitialHeight
		// This extra logic due to Tendermint validator set changes being delayed 1 block.
		// It may get overwritten due to InitChain validator updates.
		if err := store.saveValidatorsInfo(nextHeight, nextHeight, state.Validators); err != nil {
			return err
		}
	}
	// Save next validators.
	if err := store.saveValidatorsInfo(nextHeight+1, state.LastHeightValidatorsChanged, state.NextValidators); err != nil {
		return err
	}

	// Save next consensus params.
	if err := store.saveConsensusParamsInfo(nextHeight,
		state.LastHeightConsensusParamsChanged, state.ConsensusParams); err != nil {
		return err
	}
	err := store.db.SetSync(key, state.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// BootstrapState saves a new state, used e.g. by state sync when starting from non-zero height.
func (store dbStore) Bootstrap(state State) error {
	height := state.LastBlockHeight + 1
	if height == 1 {
		height = state.InitialHeight
	}

	if height > 1 && !state.LastValidators.IsNilOrEmpty() {
		if err := store.saveValidatorsInfo(height-1, height-1, state.LastValidators); err != nil {
			return err
		}
	}

	if err := store.saveValidatorsInfo(height, height, state.Validators); err != nil {
		return err
	}

	if err := store.saveValidatorsInfo(height+1, height+1, state.NextValidators); err != nil {
		return err
	}

	if err := store.saveConsensusParamsInfo(height,
		state.LastHeightConsensusParamsChanged, state.ConsensusParams); err != nil {
		return err
	}

	return store.db.SetSync(stateKey, state.Bytes())
}

// PruneStates deletes states between the given heights (including from, excluding to). It is not
// guaranteed to delete all states, since the last checkpointed state and states being pointed to by
// e.g. `LastHeightChanged` must remain. The state at to must also exist.
//
// The from parameter is necessary since we can't do a key scan in a performant way due to the key
// encoding not preserving ordering: https://github.com/fluentum-chain/fluentum/issues/4567
// This will cause some old states to be left behind when doing incremental partial prunes,
// specifically older checkpoints and LastHeightChanged targets.
func (store dbStore) PruneStates(from int64, to int64) error {
	if from <= 0 || to <= 0 {
		return fmt.Errorf("from height %v and to height %v must be greater than 0", from, to)
	}
	if from >= to {
		return fmt.Errorf("from height %v must be lower than to height %v", from, to)
	}
	valInfo, err := loadValidatorsInfo(store.db, to)
	if err != nil {
		return fmt.Errorf("validators at height %v not found: %w", to, err)
	}
	paramsInfo, err := store.loadConsensusParamsInfo(to)
	if err != nil {
		return fmt.Errorf("consensus params at height %v not found: %w", to, err)
	}

	keepVals := make(map[int64]bool)
	if valInfo.ValidatorSet == nil {
		keepVals[valInfo.LastHeightChanged] = true
		keepVals[lastStoredHeightFor(to, valInfo.LastHeightChanged)] = true // keep last checkpoint too
	}
	keepParams := make(map[int64]bool)
	if proto.Equal(&paramsInfo.ConsensusParams, &tmproto.ConsensusParams{}) {
		keepParams[paramsInfo.LastHeightChanged] = true
	}

	batch := store.db.NewBatch()
	defer batch.Close()
	pruned := uint64(0)

	// We have to delete in reverse order, to avoid deleting previous heights that have validator
	// sets and consensus params that we may need to retrieve.
	for h := to - 1; h >= from; h-- {
		// For heights we keep, we must make sure they have the full validator set or consensus
		// params, otherwise they will panic if they're retrieved directly (instead of
		// indirectly via a LastHeightChanged pointer).
		if keepVals[h] {
			v, err := loadValidatorsInfo(store.db, h)
			if err != nil || v.ValidatorSet == nil {
				vip, err := store.LoadValidators(h)
				if err != nil {
					return err
				}

				pvi, err := vip.ToProto()
				if err != nil {
					return err
				}

				v.ValidatorSet = pvi
				v.LastHeightChanged = h

				bz, err := proto.Marshal(v)
				if err != nil {
					return err
				}
				err = batch.Set(calcValidatorsKey(h), bz)
				if err != nil {
					return err
				}
			}
		} else {
			err = batch.Delete(calcValidatorsKey(h))
			if err != nil {
				return err
			}
		}

		if keepParams[h] {
			p, err := store.loadConsensusParamsInfo(h)
			if err != nil {
				return err
			}

			if proto.Equal(&p.ConsensusParams, &tmproto.ConsensusParams{}) {
				cp, err := store.LoadConsensusParams(h)
				if err != nil {
					return err
				}
				p.ConsensusParams = cp
				p.LastHeightChanged = h
				bz, err := proto.Marshal(p)
				if err != nil {
					return err
				}

				err = batch.Set(calcConsensusParamsKey(h), bz)
				if err != nil {
					return err
				}
			}
		} else {
			err = batch.Delete(calcConsensusParamsKey(h))
			if err != nil {
				return err
			}
		}

		err = batch.Delete(calcABCIResponsesKey(h))
		if err != nil {
			return err
		}
		pruned++

		// avoid batches growing too large by flushing to database regularly
		if pruned%1000 == 0 && pruned > 0 {
			err := batch.Write()
			if err != nil {
				return err
			}
			batch.Close()
			batch = store.db.NewBatch()
			defer batch.Close()
		}
	}

	err = batch.WriteSync()
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------

// ABCIResponsesResultsHash returns the Merkle root hash of
// ABCIResponses.Results
func ABCIResponsesResultsHash(ar *ABCIResponses) []byte {
	return nil
}

// LoadABCIResponses loads the ABCIResponses for the given height from the
// database. If the node has DiscardABCIResponses set to true, ErrABCIResponsesNotPersisted
// is persisted. If not found, ErrNoABCIResponsesForHeight is returned.
func (store dbStore) LoadABCIResponses(height int64) (*ABCIResponses, error) {
	if store.DiscardABCIResponses {
		return nil, ErrABCIResponsesNotPersisted
	}

	buf, err := store.db.Get(calcABCIResponsesKey(height))
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, ErrNoABCIResponsesForHeight{height}
	}

	abciResponses := new(ABCIResponses)
	// Commented out: proto.Unmarshal with *ABCIResponses if not proto.Message
	// err = proto.Unmarshal(buf, abciResponses)
	// if err != nil {
	// 	// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
	// 	tmos.Exit(fmt.Sprintf(`LoadABCIResponses: Data has been corrupted or its spec has
	//                 changed: %v\n`, err))
	// }
	// TODO: ensure that buf is completely read.

	return abciResponses, nil
}

// LoadLastABCIResponse loads the last ABCIResponse for the given height.
func (store dbStore) LoadLastABCIResponse(height int64) (*ABCIResponses, error) {
	bz, err := store.db.Get(lastABCIResponseKey)
	if err != nil {
		return nil, err
	}

	if len(bz) == 0 {
		return nil, errors.New("no last ABCI response has been persisted")
	}

	// abciResponse := new(ABCIResponsesInfo)
	// err = abciResponse.Unmarshal(bz)
	// if err != nil {
	// 	tmos.Exit(fmt.Sprintf(`LoadLastABCIResponses: Data has been corrupted or its spec has
	// 		changed: %v\n`, err))
	// }

	// // Here we validate the result by comparing its height to the expected height.
	// if height != abciResponse.GetHeight() {
	// 	return nil, errors.New("expected height %d but last stored abci responses was at height %d")
	// }

	// return abciResponse.AbciResponses, nil
	return nil, nil
}

// SaveABCIResponses persists the ABCIResponses to the database.
// This is useful in case we crash after app.Commit and before s.Save().
// Responses are indexed by height so they can also be loaded later to produce
// Merkle proofs.
//
// CONTRACT: height must be monotonically increasing every time this is called.
func (store dbStore) SaveABCIResponses(height int64, abciResponses *ABCIResponses) error {
	// Save the ABCIResponse for crash recovery and queries.
	if !store.DiscardABCIResponses {
		// TODO: Properly serialize abciResponses (not a proto.Message)
		// bz, err := proto.Marshal(abciResponses)
		// if err != nil {
		// 	return err
		// }
	}

	// Always save the last ABCI response for crash recovery.
	// TODO: Properly serialize response (not a proto.Message)
	// response := &ABCIResponsesInfo{
	// 	AbciResponses: abciResponses,
	// 	Height:        height,
	// }
	// bz, err := proto.Marshal(response)
	// if err != nil {
	// 	return err
	// }
	// return store.db.SetSync(lastABCIResponseKey, bz)
	// TODO: Implement crash recovery serialization for ABCIResponses
	return nil
}

//-----------------------------------------------------------------------------

// LoadValidators loads the ValidatorSet for a given height.
// Returns ErrNoValSetForHeight if the validator set can't be found for this height.
func (store dbStore) LoadValidators(height int64) (*types.ValidatorSet, error) {
	valInfo, err := loadValidatorsInfo(store.db, height)
	if err != nil {
		return nil, ErrNoValSetForHeight{height}
	}
	if valInfo.ValidatorSet == nil {
		lastStoredHeight := lastStoredHeightFor(height, valInfo.LastHeightChanged)
		valInfo2, err := loadValidatorsInfo(store.db, lastStoredHeight)
		if err != nil || valInfo2.ValidatorSet == nil {
			return nil,
				fmt.Errorf("couldn't find validators at height %d (height %d was originally requested): %w",
					lastStoredHeight,
					height,
					err,
				)
		}

		vs, err := types.ValidatorSetFromProto(valInfo2.ValidatorSet)
		if err != nil {
			return nil, err
		}

		vs.IncrementProposerPriority(tmmath.SafeConvertInt32(height - lastStoredHeight)) // mutate
		vi2, err := vs.ToProto()
		if err != nil {
			return nil, err
		}

		valInfo2.ValidatorSet = vi2
		valInfo = valInfo2
	}

	vip, err := types.ValidatorSetFromProto(valInfo.ValidatorSet)
	if err != nil {
		return nil, err
	}

	return vip, nil
}

func lastStoredHeightFor(height, lastHeightChanged int64) int64 {
	checkpointHeight := height - height%valSetCheckpointInterval
	return tmmath.MaxInt64(checkpointHeight, lastHeightChanged)
}

// CONTRACT: Returned ValidatorsInfo can be mutated.
func loadValidatorsInfo(db dbm.DB, height int64) (*tmstate.ValidatorsInfo, error) {
	buf, err := db.Get(calcValidatorsKey(height))
	if err != nil {
		return nil, err
	}

	if len(buf) == 0 {
		return nil, errors.New("value retrieved from db is empty")
	}

	v := new(tmstate.ValidatorsInfo)
	err = proto.Unmarshal(buf, v)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		tmos.Exit(fmt.Sprintf(`LoadValidators: Data has been corrupted or its spec has changed:
        %v\n`, err))
	}
	// TODO: ensure that buf is completely read.

	return v, nil
}

// saveValidatorsInfo persists the validator set.
//
// `height` is the effective height for which the validator is responsible for
// signing. It should be called from s.Save(), right before the state itself is
// persisted.
func (store dbStore) saveValidatorsInfo(height, lastHeightChanged int64, valSet *types.ValidatorSet) error {
	if lastHeightChanged > height {
		return errors.New("lastHeightChanged cannot be greater than ValidatorsInfo height")
	}
	valInfo := &tmstate.ValidatorsInfo{
		LastHeightChanged: lastHeightChanged,
	}
	// Only persist validator set if it was updated or checkpoint height (see
	// valSetCheckpointInterval) is reached.
	if height == lastHeightChanged || height%valSetCheckpointInterval == 0 {
		pv, err := valSet.ToProto()
		if err != nil {
			return err
		}
		valInfo.ValidatorSet = pv
	}

	bz, err := proto.Marshal(valInfo)
	if err != nil {
		return err
	}

	err = store.db.Set(calcValidatorsKey(height), bz)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------

// ConsensusParamsInfo represents the latest consensus params, or the last height it changed

// LoadConsensusParams loads the ConsensusParams for a given height.
func (store dbStore) LoadConsensusParams(height int64) (tmproto.ConsensusParams, error) {
	empty := tmproto.ConsensusParams{}

	paramsInfo, err := store.loadConsensusParamsInfo(height)
	if err != nil {
		return empty, fmt.Errorf("could not find consensus params for height #%d: %w", height, err)
	}

	if proto.Equal(&paramsInfo.ConsensusParams, &empty) {
		paramsInfo2, err := store.loadConsensusParamsInfo(paramsInfo.LastHeightChanged)
		if err != nil {
			return empty, fmt.Errorf(
				"couldn't find consensus params at height %d as last changed from height %d: %w",
				paramsInfo.LastHeightChanged,
				height,
				err,
			)
		}

		paramsInfo = paramsInfo2
	}

	return paramsInfo.ConsensusParams, nil
}

func (store dbStore) loadConsensusParamsInfo(height int64) (*tmstate.ConsensusParamsInfo, error) {
	buf, err := store.db.Get(calcConsensusParamsKey(height))
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, errors.New("value retrieved from db is empty")
	}

	paramsInfo := new(tmstate.ConsensusParamsInfo)
	err = proto.Unmarshal(buf, paramsInfo)
	if err != nil {
		// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
		tmos.Exit(fmt.Sprintf(`LoadConsensusParams: Data has been corrupted or its spec has changed:
                %v\n`, err))
	}
	// TODO: ensure that buf is completely read.

	return paramsInfo, nil
}

// saveConsensusParamsInfo persists the consensus params for the next block to disk.
// It should be called from s.Save(), right before the state itself is persisted.
// If the consensus params did not change after processing the latest block,
// only the last height for which they changed is persisted.
func (store dbStore) saveConsensusParamsInfo(nextHeight, changeHeight int64, params tmproto.ConsensusParams) error {
	paramsInfo := &tmstate.ConsensusParamsInfo{
		LastHeightChanged: changeHeight,
	}

	if changeHeight == nextHeight {
		paramsInfo.ConsensusParams = params
	}
	bz, err := proto.Marshal(paramsInfo)
	if err != nil {
		return err
	}

	err = store.db.Set(calcConsensusParamsKey(nextHeight), bz)
	if err != nil {
		return err
	}

	return nil
}

func (store dbStore) Close() error {
	return store.db.Close()
}
