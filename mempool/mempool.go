package mempool

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"sync"

	cmabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proxy"
	"github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/libs/clist"
	"github.com/fluentum-chain/fluentum/libs/log"
	"github.com/fluentum-chain/fluentum/types"
)

const (
	MempoolChannel = byte(0x30)

	// PeerCatchupSleepIntervalMS defines how much time to sleep if a peer is behind
	PeerCatchupSleepIntervalMS = 100

	// UnknownPeerID is the peer ID to use when running CheckTx when there is
	// no peer (e.g. RPC)
	UnknownPeerID uint16 = 0

	MaxActiveIDs = math.MaxUint16
)

// Mempool defines the mempool interface.
//
// Updates to the mempool need to be synchronized with committing a block so
// applications can reset their transient state on Commit.
type Mempool interface {
	// CheckTx executes a new transaction against the application to determine
	// its validity and whether it should be added to the mempool.
	CheckTx(tx types.Tx, callback func(*cmabci.Response), txInfo TxInfo) error

	// RemoveTxByKey removes a transaction, identified by its key,
	// from the mempool.
	RemoveTxByKey(txKey types.TxKey) error

	// ReapMaxBytesMaxGas reaps transactions from the mempool up to maxBytes
	// bytes total with the condition that the total gasWanted must be less than
	// maxGas.
	//
	// If both maxes are negative, there is no cap on the size of all returned
	// transactions (~ all available transactions).
	ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs

	// ReapMaxTxs reaps up to max transactions from the mempool. If max is
	// negative, there is no cap on the size of all returned transactions
	// (~ all available transactions).
	ReapMaxTxs(max int) types.Txs

	// Lock locks the mempool. The consensus must be able to hold lock to safely
	// update.
	Lock()

	// Unlock unlocks the mempool.
	Unlock()

	// Update informs the mempool that the given txs were committed and can be
	// discarded.
	//
	// NOTE:
	// 1. This should be called *after* block is committed by consensus.
	// 2. Lock/Unlock must be managed by the caller.
	Update(
		blockHeight int64,
		blockTxs types.Txs,
		deliverTxResponses []*cmabci.ExecTxResult,
		newPreFn PreCheckFunc,
		newPostFn PostCheckFunc,
	) error

	// FlushAppConn flushes the mempool connection to ensure async callback calls
	// are done, e.g. from CheckTx.
	//
	// NOTE:
	// 1. Lock/Unlock must be managed by caller.
	FlushAppConn() error

	// Flush removes all transactions from the mempool and caches.
	Flush()

	// TxsAvailable returns a channel which fires once for every height, and only
	// when transactions are available in the mempool.
	//
	// NOTE:
	// 1. The returned channel may be nil if EnableTxsAvailable was not called.
	TxsAvailable() <-chan struct{}

	// EnableTxsAvailable initializes the TxsAvailable channel, ensuring it will
	// trigger once every height when transactions are available.
	EnableTxsAvailable()

	// Size returns the number of transactions in the mempool.
	Size() int

	// SizeBytes returns the total size of all txs in the mempool.
	SizeBytes() int64
}

// PreCheckFunc is an optional filter executed before CheckTx and rejects
// transaction if false is returned. An example would be to ensure that a
// transaction doesn't exceeded the block size.
type PreCheckFunc func(types.Tx) error

// PostCheckFunc is an optional filter executed after CheckTx and rejects
// transaction if false is returned. An example would be to ensure a
// transaction doesn't require more gas than available for the block.
type PostCheckFunc func(types.Tx, *cmabci.ResponseCheckTx) error

// PreCheckMaxBytes checks that the size of the transaction is smaller or equal
// to the expected maxBytes.
func PreCheckMaxBytes(maxBytes int64) PreCheckFunc {
	return func(tx types.Tx) error {
		txSize := types.ComputeProtoSizeForTxs([]types.Tx{tx})

		if txSize > maxBytes {
			return fmt.Errorf("tx size is too big: %d, max: %d", txSize, maxBytes)
		}

		return nil
	}
}

// PostCheckMaxGas checks that the wanted gas is smaller or equal to the passed
// maxGas. Returns nil if maxGas is -1.
func PostCheckMaxGas(maxGas int64) PostCheckFunc {
	return func(tx types.Tx, res *cmabci.ResponseCheckTx) error {
		if maxGas == -1 {
			return nil
		}
		if res.GasWanted < 0 {
			return fmt.Errorf("gas wanted %d is negative",
				res.GasWanted)
		}
		if res.GasWanted > maxGas {
			return fmt.Errorf("gas wanted %d is greater than max gas %d",
				res.GasWanted, maxGas)
		}

		return nil
	}
}

// ErrTxInCache is returned to the client if we saw tx earlier
var ErrTxInCache = errors.New("tx already exists in cache")

// TxKey is the fixed length array key used as an index.
type TxKey [sha256.Size]byte

// ErrTxTooLarge defines an error when a transaction is too big to be sent in a
// message to other peers.
type ErrTxTooLarge struct {
	Max    int
	Actual int
}

func (e ErrTxTooLarge) Error() string {
	return fmt.Sprintf("Tx too large. Max size is %d, but got %d", e.Max, e.Actual)
}

// ErrMempoolIsFull defines an error where Tendermint and the application cannot
// handle that much load.
type ErrMempoolIsFull struct {
	NumTxs      int
	MaxTxs      int
	TxsBytes    int64
	MaxTxsBytes int64
}

func (e ErrMempoolIsFull) Error() string {
	return fmt.Sprintf(
		"mempool is full: number of txs %d (max: %d), total txs bytes %d (max: %d)",
		e.NumTxs,
		e.MaxTxs,
		e.TxsBytes,
		e.MaxTxsBytes,
	)
}

// ErrPreCheck defines an error where a transaction fails a pre-check.
type ErrPreCheck struct {
	Reason error
}

func (e ErrPreCheck) Error() string {
	return e.Reason.Error()
}

// IsPreCheckError returns true if err is due to pre check failure.
func IsPreCheckError(err error) bool {
	return errors.As(err, &ErrPreCheck{})
}

// CListMempool is an ordered in-memory pool for transactions before they are proposed in a consensus
// round. Transaction validity is checked using the CheckTx abci message before the transaction is
// added to the pool. The mempool uses a concurrent list structure for storing transactions that can
// be efficiently accessed by multiple concurrent readers.
type CListMempool struct {
	// Atomic integers
	height   int64 // the latest height passed to Update
	txsBytes int64 // total size of mempool, in bytes

	// notify listeners (ie. consensus) when txs are available
	notifiedTxsAvailable bool
	txsAvailable         chan struct{} // fires once for each height, when the mempool is not empty

	config *config.MempoolConfig

	// Exclusive mutex for Update method to prevent concurrent execution of it and ReapMaxBytesMaxGas.
	updateMtx sync.Mutex
	preCheck  PreCheckFunc
	postCheck PostCheckFunc

	txs          *clist.CList // concurrent linked-list of good txs
	proxyAppConn proxy.AppConnMempool

	logger log.Logger
}

// NewCListMempool returns a new mempool with the given configuration and connection to an application.
func NewCListMempool(
	config *config.MempoolConfig,
	proxyAppConn proxy.AppConnMempool,
	height int64,
	preCheck PreCheckFunc,
	postCheck PostCheckFunc,
) *CListMempool {
	mempool := &CListMempool{
		config:               config,
		proxyAppConn:         proxyAppConn,
		txs:                  clist.New(),
		height:               height,
		preCheck:             preCheck,
		postCheck:            postCheck,
		notifiedTxsAvailable: false,
		txsAvailable:         make(chan struct{}, 1),
	}
	return mempool
}

// CheckTx implements Mempool.CheckTx: validate the transaction for the mempool.
func (mem *CListMempool) CheckTx(tx types.Tx) error {
	mem.updateMtx.Lock()
	// use defer to unlock mutex because application (*local client*) might panic
	defer mem.updateMtx.Unlock()

	// Check if the transaction is valid
	if err := mem.preCheck(tx); err != nil {
		return err
	}

	// Check if the transaction is already in the mempool
	// TODO: Implement proper duplicate checking
	// if mem.txs.Has(tx.Key()) {
	// 	return ErrTxInCache
	// }

	// Check if the transaction is valid according to the application
	reqRes, err := mem.proxyAppConn.CheckTxAsync(context.Background(), &cmabci.RequestCheckTx{Tx: tx})
	if err != nil {
		return err
	}
	<-reqRes.Done
	res := reqRes.Response.GetCheckTx()
	if res.Code != cmabci.CodeTypeOK {
		return fmt.Errorf("transaction rejected: %s", res.Log)
	}

	// Add the transaction to the mempool
	mem.txs.PushBack(tx)
	mem.txsBytes += int64(len(tx))

	// Notify that new transactions are available
	if !mem.notifiedTxsAvailable {
		mem.notifiedTxsAvailable = true
		select {
		case mem.txsAvailable <- struct{}{}:
		default:
		}
	}

	return nil
}

// Update implements Mempool.Update: remove all transactions from the mempool that were included in the block.
func (mem *CListMempool) Update(
	height int64,
	txs []types.Tx,
	deliverTxResponses []*cmabci.ExecTxResult,
	preCheck PreCheckFunc,
	postCheck PostCheckFunc,
) error {
	mem.updateMtx.Lock()
	defer mem.updateMtx.Unlock()

	mem.height = height
	mem.notifiedTxsAvailable = false
	mem.preCheck = preCheck
	mem.postCheck = postCheck

	// Remove transactions that were included in the block
	for i, tx := range txs {
		if deliverTxResponses[i].Code == cmabci.CodeTypeOK {
			// TODO: Implement proper transaction removal
			// mem.txs.Remove(tx.Key())
			mem.txsBytes -= int64(len(tx))
		}
	}

	return nil
}

// ReapMaxBytesMaxGas returns a list of transactions that fit within the given size and gas limits.
func (mem *CListMempool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs {
	mem.updateMtx.Lock()
	defer mem.updateMtx.Unlock()

	var (
		totalBytes int64
		totalGas   int64
		txs        types.Txs
	)

	// Iterate through transactions in the mempool
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		tx := e.Value.(types.Tx)

		// Check if adding this transaction would exceed the limits
		txBytes := int64(len(tx))
		if totalBytes+txBytes > maxBytes {
			return txs
		}

		// For now, assume gas is 0 for all transactions
		txGas := int64(0)
		// TODO: Implement proper gas calculation
		// if isStaker, _ := mem.fluentumKeeper.IsStaker(tx.Sender()); !isStaker {
		// 	txGas = tx.Gas()
		// }

		if totalGas+txGas > maxGas {
			return txs
		}

		txs = append(txs, tx)
		totalBytes += txBytes
		totalGas += txGas
	}

	return txs
}

// Size returns the number of transactions in the mempool.
func (mem *CListMempool) Size() int {
	return mem.txs.Len()
}

// SizeBytes returns the total size of all transactions in the mempool.
func (mem *CListMempool) SizeBytes() int64 {
	return mem.txsBytes
}

// Flush removes all transactions from the mempool.
func (mem *CListMempool) Flush() {
	mem.updateMtx.Lock()
	defer mem.updateMtx.Unlock()

	mem.txs = clist.New()
	mem.txsBytes = 0
	mem.notifiedTxsAvailable = false
}

// TxsAvailable returns a channel that fires when transactions are available.
func (mem *CListMempool) TxsAvailable() <-chan struct{} {
	return mem.txsAvailable
}

// EnableTxsAvailable enables the txsAvailable channel.
func (mem *CListMempool) EnableTxsAvailable() {
	mem.updateMtx.Lock()
	defer mem.updateMtx.Unlock()

	mem.notifiedTxsAvailable = false
}
