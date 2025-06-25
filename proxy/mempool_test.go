package proxy_test

import (
	"context"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/fluentum-chain/fluentum/proxy/mocks"
	"github.com/stretchr/testify/require"
)

func TestMempoolCheckTx(t *testing.T) {
	t.Parallel()

	t.Run("successful check tx", func(t *testing.T) {
		mock := mocks.NewMockMempool()
		mock.CheckTxFn = func(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
			require.Equal(t, []byte("test_tx"), req.Tx)
			return &types.ResponseCheckTx{Code: 0}, nil
		}

		res, err := mock.CheckTx(context.Background(), &types.RequestCheckTx{Tx: []byte("test_tx")})
		require.NoError(t, err)
		require.Equal(t, uint32(0), res.Code)
	})

	t.Run("timeout handling", func(t *testing.T) {
		mock := mocks.NewMockMempool()
		mock.CheckTxFn = func(ctx context.Context, req *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
			<-ctx.Done() // Wait for cancellation
			return nil, ctx.Err()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := mock.CheckTx(ctx, &types.RequestCheckTx{Tx: []byte("test")})
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestConsensusFinalizeBlock(t *testing.T) {
	t.Parallel()

	mock := &mocks.MockAppConnConsensus{}
	mock.FinalizeBlockFn = func(ctx context.Context, req *types.RequestFinalizeBlock) (*types.ResponseFinalizeBlock, error) {
		require.Equal(t, int64(1), req.Height)
		return &types.ResponseFinalizeBlock{
			TxResults: []*types.ExecTxResult{{Code: 0}},
		}, nil
	}

	res, err := mock.FinalizeBlock(context.Background(), &types.RequestFinalizeBlock{Height: 1})
	require.NoError(t, err)
	require.Len(t, res.TxResults, 1)
}
