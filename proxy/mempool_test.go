package proxy_test

import (
	"context"
	"testing"
	"time"

	abci "github.com/fluentum-chain/fluentum/abci/types"
	"github.com/fluentum-chain/fluentum/proxy/mocks"
	"github.com/stretchr/testify/require"
)

func TestMempoolCheckTx(t *testing.T) {
	t.Parallel()

	t.Run("successful check tx", func(t *testing.T) {
		mock := mocks.NewMockMempool()
		mock.CheckTxFn = func(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
			require.Equal(t, []byte("test_tx"), req.Tx)
			return &abci.CheckTxResponse{Code: 0}, nil
		}

		res, err := mock.CheckTx(context.Background(), &abci.CheckTxRequest{Tx: []byte("test_tx")})
		require.NoError(t, err)
		require.Equal(t, uint32(0), res.Code)
	})

	t.Run("timeout handling", func(t *testing.T) {
		mock := mocks.NewMockMempool()
		mock.CheckTxFn = func(ctx context.Context, req *abci.CheckTxRequest) (*abci.CheckTxResponse, error) {
			<-ctx.Done() // Wait for cancellation
			return nil, ctx.Err()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := mock.CheckTx(ctx, &abci.CheckTxRequest{Tx: []byte("test")})
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestConsensusFinalizeBlock(t *testing.T) {
	t.Parallel()

	mock := &mocks.MockAppConnConsensus{}
	mock.FinalizeBlockFn = func(ctx context.Context, req *abci.FinalizeBlockRequest) (*abci.FinalizeBlockResponse, error) {
		require.Equal(t, int64(1), req.Height)
		return &abci.FinalizeBlockResponse{
			TxResults: []*abci.ExecTxResult{{Code: 0}},
		}, nil
	}

	res, err := mock.FinalizeBlock(context.Background(), &abci.FinalizeBlockRequest{Height: 1})
	require.NoError(t, err)
	require.Len(t, res.TxResults, 1)
}
