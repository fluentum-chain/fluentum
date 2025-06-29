package txindex

import (
	"context"

	"github.com/fluentum-chain/fluentum/libs/service"
	"github.com/fluentum-chain/fluentum/state/indexer"
	"github.com/fluentum-chain/fluentum/types"
)

// XXX/TODO: These types should be moved to the indexer package.

const (
	subscriber = "IndexerService"
)

// IndexerService connects event bus, transaction and block indexers together in
// order to index transactions and blocks coming from the event bus.
type IndexerService struct {
	service.BaseService

	txIdxr           TxIndexer
	blockIdxr        indexer.BlockIndexer
	eventBus         *types.EventBus
	terminateOnError bool
}

// NewIndexerService returns a new service instance.
func NewIndexerService(
	txIdxr TxIndexer,
	blockIdxr indexer.BlockIndexer,
	eventBus *types.EventBus,
	terminateOnError bool,
) *IndexerService {

	is := &IndexerService{txIdxr: txIdxr, blockIdxr: blockIdxr, eventBus: eventBus, terminateOnError: terminateOnError}
	is.BaseService = *service.NewBaseService(nil, "IndexerService", is)
	return is
}

// OnStart implements service.Service by subscribing for all transactions
// and indexing them by events.
func (is *IndexerService) OnStart() error {
	// Use SubscribeUnbuffered here to ensure both subscriptions does not get
	// canceled due to not pulling messages fast enough. Cause this might
	// sometimes happen when there are no other subscribers.
	blockHeadersSub, err := is.eventBus.SubscribeUnbuffered(
		context.Background(),
		subscriber,
		types.EventQueryNewBlockHeader)
	if err != nil {
		return err
	}

	txsSub, err := is.eventBus.SubscribeUnbuffered(context.Background(), subscriber, types.EventQueryTx)
	if err != nil {
		return err
	}

	go func() {
		for {
			msg := <-blockHeadersSub.Out()
			eventDataHeader := msg.Data().(types.EventDataNewBlockHeader)
			height := eventDataHeader.Header.Height
			batch := NewBatch(eventDataHeader.NumTxs)

			for i := int64(0); i < eventDataHeader.NumTxs; i++ {
				msg2 := <-txsSub.Out()
				txResult := msg2.Data().(types.EventDataTx).TxResult

				if err = batch.Add(&txResult); err != nil {
					is.Logger.Error(
						"failed to add tx to batch",
						"height", height,
						"index", txResult.Index,
						"err", err,
					)

					if is.terminateOnError {
						if err := is.Stop(); err != nil {
							is.Logger.Error("failed to stop", "err", err)
						}
						return
					}
				}
			}

			if err := is.blockIdxr.Index(eventDataHeader); err != nil {
				is.Logger.Error("failed to index block", "height", height, "err", err)
				if is.terminateOnError {
					if err := is.Stop(); err != nil {
						is.Logger.Error("failed to stop", "err", err)
					}
					return
				}
			} else {
				is.Logger.Info("indexed block exents", "height", height)
			}

			if err = is.txIdxr.AddBatch(batch); err != nil {
				is.Logger.Error("failed to index block txs", "height", height, "err", err)
				if is.terminateOnError {
					if err := is.Stop(); err != nil {
						is.Logger.Error("failed to stop", "err", err)
					}
					return
				}
			} else {
				is.Logger.Debug("indexed transactions", "height", height, "num_txs", eventDataHeader.NumTxs)
			}
		}
	}()
	return nil
}

// OnStop implements service.Service by unsubscribing from all transactions.
func (is *IndexerService) OnStop() {
	if is.eventBus.IsRunning() {
		_ = is.eventBus.UnsubscribeAll(context.Background(), subscriber)
	}
}
