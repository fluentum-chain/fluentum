package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fluentum-chain/fluentum/abci/compat"
	abci "github.com/fluentum-chain/fluentum/abci/types"
	mempl "github.com/fluentum-chain/fluentum/mempool"
	ctypes "github.com/fluentum-chain/fluentum/rpc/core/types"
	rpctypes "github.com/fluentum-chain/fluentum/rpc/jsonrpc/types"
	"github.com/fluentum-chain/fluentum/types"
)

//-----------------------------------------------------------------------------
// NOTE: tx should be signed, but this is only checked at the app level (not by Tendermint!)

// BroadcastTxAsync returns right away, with no response. Does not wait for
// CheckTx nor DeliverTx results.
// More: https://docs.tendermint.com/v0.34/rpc/#/Tx/broadcast_tx_async
func BroadcastTxAsync(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	err := env.Mempool.CheckTx(tx, nil, mempl.TxInfo{})

	if err != nil {
		return nil, err
	}
	return &ctypes.ResultBroadcastTx{Hash: tx.Hash()}, nil
}

// BroadcastTxSync returns with the response from CheckTx. Does not wait for
// DeliverTx result.
// More: https://docs.tendermint.com/v0.34/rpc/#/Tx/broadcast_tx_sync
func BroadcastTxSync(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	resCh := make(chan *abci.Response, 1)
	err := env.Mempool.CheckTx(tx, func(res *abci.Response) {
		select {
		case <-ctx.Context().Done():
		case resCh <- res:
		}

	}, mempl.TxInfo{})
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Context().Done():
		return nil, fmt.Errorf("broadcast confirmation not received: %w", ctx.Context().Err())
	case res := <-resCh:
		r := res.GetCheckTx()
		return &ctypes.ResultBroadcastTx{
			Code:      r.Code,
			Data:      r.Data,
			Log:       r.Log,
			Codespace: r.Codespace,
			Hash:      tx.Hash(),
		}, nil
	}
}

// BroadcastTxCommit returns with the responses from CheckTx and block-level ExecTxResult.
// ABCI 2.0: DeliverTx → ExecTxResult
func BroadcastTxCommit(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	subscriber := ctx.RemoteAddr()

	if env.EventBus.NumClients() >= env.Config.MaxSubscriptionClients {
		return nil, fmt.Errorf("max_subscription_clients %d reached", env.Config.MaxSubscriptionClients)
	} else if env.EventBus.NumClientSubscriptions(subscriber) >= env.Config.MaxSubscriptionsPerClient {
		return nil, fmt.Errorf("max_subscriptions_per_client %d reached", env.Config.MaxSubscriptionsPerClient)
	}

	// Subscribe to tx being committed in block.
	subCtx, cancel := context.WithTimeout(ctx.Context(), SubscribeTimeout)
	defer cancel()
	q := types.EventQueryTxFor(tx)
	deliverTxSub, err := env.EventBus.Subscribe(subCtx, subscriber, q)
	if err != nil {
		err = fmt.Errorf("failed to subscribe to tx: %w", err)
		env.Logger.Error("Error on broadcast_tx_commit", "err", err)
		return nil, err
	}
	defer func() {
		if err := env.EventBus.Unsubscribe(context.Background(), subscriber, q); err != nil {
			env.Logger.Error("Error unsubscribing from eventBus", "err", err)
		}
	}()

	// Broadcast tx and wait for CheckTx result
	checkTxResCh := make(chan *abci.Response, 1)
	err = env.Mempool.CheckTx(tx, func(res *abci.Response) {
		select {
		case <-ctx.Context().Done():
		case checkTxResCh <- res:
		}
	}, mempl.TxInfo{})
	if err != nil {
		env.Logger.Error("Error on broadcastTxCommit", "err", err)
		return nil, fmt.Errorf("error on broadcastTxCommit: %v", err)
	}
	select {
	case <-ctx.Context().Done():
		return nil, fmt.Errorf("broadcast confirmation not received: %w", ctx.Context().Err())
	case checkTxResMsg := <-checkTxResCh:
		checkTxRes := checkTxResMsg.GetCheckTx()
		if checkTxRes.Code != abci.CodeTypeOK {
			return &ctypes.ResultBroadcastTxCommit{
				CheckTx: *compat.CheckTxResponseFromComet(checkTxRes),
				ExecTx:  abci.ExecTxResult{},
				Hash:    tx.Hash(),
			}, nil
		}

		// Wait for the tx to be included in a block or timeout.
		select {
		case msg := <-deliverTxSub.Out(): // The tx was included in a block.
			eventTx := msg.Data().(types.EventDataTx)
			// In ABCI 2.0, eventTx.Result should be of type ExecTxResult
			return &ctypes.ResultBroadcastTxCommit{
				CheckTx: *compat.CheckTxResponseFromComet(checkTxRes),
				ExecTx:  *compat.ExecTxResultFromComet(&eventTx.Result),
				Hash:    tx.Hash(),
				Height:  eventTx.Height,
			}, nil
		case <-deliverTxSub.Cancelled():
			var reason string
			if deliverTxSub.Err() == nil {
				reason = "Tendermint exited"
			} else {
				reason = deliverTxSub.Err().Error()
			}
			err = fmt.Errorf("deliverTxSub was cancelled (reason: %s)", reason)
			env.Logger.Error("Error on broadcastTxCommit", "err", err)
			return &ctypes.ResultBroadcastTxCommit{
				CheckTx: *compat.CheckTxResponseFromComet(checkTxRes),
				ExecTx:  abci.ExecTxResult{},
				Hash:    tx.Hash(),
			}, err
		case <-time.After(env.Config.TimeoutBroadcastTxCommit):
			err = errors.New("timed out waiting for tx to be included in a block")
			env.Logger.Error("Error on broadcastTxCommit", "err", err)
			return &ctypes.ResultBroadcastTxCommit{
				CheckTx: *compat.CheckTxResponseFromComet(checkTxRes),
				ExecTx:  abci.ExecTxResult{},
				Hash:    tx.Hash(),
			}, err
		}
	}
}

// UnconfirmedTxs gets unconfirmed transactions (maximum ?limit entries)
// including their number.
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/unconfirmed_txs
func UnconfirmedTxs(ctx *rpctypes.Context, limitPtr *int) (*ctypes.ResultUnconfirmedTxs, error) {
	// reuse per_page validator
	limit := validatePerPage(limitPtr)

	txs := env.Mempool.ReapMaxTxs(limit)
	return &ctypes.ResultUnconfirmedTxs{
		Count:      len(txs),
		Total:      env.Mempool.Size(),
		TotalBytes: env.Mempool.SizeBytes(),
		Txs:        txs}, nil
}

// NumUnconfirmedTxs gets number of unconfirmed transactions.
// More: https://docs.tendermint.com/v0.34/rpc/#/Info/num_unconfirmed_txs
func NumUnconfirmedTxs(ctx *rpctypes.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	return &ctypes.ResultUnconfirmedTxs{
		Count:      env.Mempool.Size(),
		Total:      env.Mempool.Size(),
		TotalBytes: env.Mempool.SizeBytes()}, nil
}

// CheckTx checks the transaction without executing it. The transaction won't
// be added to the mempool either.
// More: https://docs.tendermint.com/v0.34/rpc/#/Tx/check_tx
func CheckTx(ctx *rpctypes.Context, tx types.Tx) (*ctypes.ResultCheckTx, error) {
	res, err := env.ProxyAppMempool.CheckTx(context.Background(), &abci.CheckTxRequest{Tx: tx})
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultCheckTx{CheckTxResponse: *res}, nil
}

// TODO: Update broadcast and result logic for FinalizeBlock block-level responses
