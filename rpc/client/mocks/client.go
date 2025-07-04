// Code generated by mockery 2.7.4. DO NOT EDIT.

package mocks

import (
	bytes "github.com/fluentum-chain/fluentum/libs/bytes"
	client "github.com/fluentum-chain/fluentum/rpc/client"

	context "context"

	coretypes "github.com/fluentum-chain/fluentum/rpc/core/types"

	log "github.com/fluentum-chain/fluentum/libs/log"

	mock "github.com/stretchr/testify/mock"

	types "github.com/fluentum-chain/fluentum/types"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// ABCIInfo provides a mock function with given fields: _a0
func (_m *Client) ABCIInfo(_a0 context.Context) (*coretypes.ResultABCIInfo, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultABCIInfo
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultABCIInfo); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultABCIInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ABCIQuery provides a mock function with given fields: ctx, path, data
func (_m *Client) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error) {
	ret := _m.Called(ctx, path, data)

	var r0 *coretypes.ResultABCIQuery
	if rf, ok := ret.Get(0).(func(context.Context, string, bytes.HexBytes) *coretypes.ResultABCIQuery); ok {
		r0 = rf(ctx, path, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultABCIQuery)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, bytes.HexBytes) error); ok {
		r1 = rf(ctx, path, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ABCIQueryWithOptions provides a mock function with given fields: ctx, path, data, opts
func (_m *Client) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts client.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
	ret := _m.Called(ctx, path, data, opts)

	var r0 *coretypes.ResultABCIQuery
	if rf, ok := ret.Get(0).(func(context.Context, string, bytes.HexBytes, client.ABCIQueryOptions) *coretypes.ResultABCIQuery); ok {
		r0 = rf(ctx, path, data, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultABCIQuery)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, bytes.HexBytes, client.ABCIQueryOptions) error); ok {
		r1 = rf(ctx, path, data, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Block provides a mock function with given fields: ctx, height
func (_m *Client) Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error) {
	ret := _m.Called(ctx, height)

	var r0 *coretypes.ResultBlock
	if rf, ok := ret.Get(0).(func(context.Context, *int64) *coretypes.ResultBlock); ok {
		r0 = rf(ctx, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlock)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *int64) error); ok {
		r1 = rf(ctx, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockByHash provides a mock function with given fields: ctx, hash
func (_m *Client) BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error) {
	ret := _m.Called(ctx, hash)

	var r0 *coretypes.ResultBlock
	if rf, ok := ret.Get(0).(func(context.Context, []byte) *coretypes.ResultBlock); ok {
		r0 = rf(ctx, hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlock)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []byte) error); ok {
		r1 = rf(ctx, hash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockResults provides a mock function with given fields: ctx, height
func (_m *Client) BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error) {
	ret := _m.Called(ctx, height)

	var r0 *coretypes.ResultBlockResults
	if rf, ok := ret.Get(0).(func(context.Context, *int64) *coretypes.ResultBlockResults); ok {
		r0 = rf(ctx, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlockResults)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *int64) error); ok {
		r1 = rf(ctx, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockSearch provides a mock function with given fields: ctx, query, page, perPage, orderBy
func (_m *Client) BlockSearch(ctx context.Context, query string, page *int, perPage *int, orderBy string) (*coretypes.ResultBlockSearch, error) {
	ret := _m.Called(ctx, query, page, perPage, orderBy)

	var r0 *coretypes.ResultBlockSearch
	if rf, ok := ret.Get(0).(func(context.Context, string, *int, *int, string) *coretypes.ResultBlockSearch); ok {
		r0 = rf(ctx, query, page, perPage, orderBy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlockSearch)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *int, *int, string) error); ok {
		r1 = rf(ctx, query, page, perPage, orderBy)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockchainInfo provides a mock function with given fields: ctx, minHeight, maxHeight
func (_m *Client) BlockchainInfo(ctx context.Context, minHeight int64, maxHeight int64) (*coretypes.ResultBlockchainInfo, error) {
	ret := _m.Called(ctx, minHeight, maxHeight)

	var r0 *coretypes.ResultBlockchainInfo
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) *coretypes.ResultBlockchainInfo); ok {
		r0 = rf(ctx, minHeight, maxHeight)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlockchainInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int64, int64) error); ok {
		r1 = rf(ctx, minHeight, maxHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastEvidence provides a mock function with given fields: _a0, _a1
func (_m *Client) BroadcastEvidence(_a0 context.Context, _a1 types.Evidence) (*coretypes.ResultBroadcastEvidence, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *coretypes.ResultBroadcastEvidence
	if rf, ok := ret.Get(0).(func(context.Context, types.Evidence) *coretypes.ResultBroadcastEvidence); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastEvidence)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.Evidence) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastTxAsync provides a mock function with given fields: _a0, _a1
func (_m *Client) BroadcastTxAsync(_a0 context.Context, _a1 types.Tx) (*coretypes.ResultBroadcastTx, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *coretypes.ResultBroadcastTx
	if rf, ok := ret.Get(0).(func(context.Context, types.Tx) *coretypes.ResultBroadcastTx); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.Tx) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastTxCommit provides a mock function with given fields: _a0, _a1
func (_m *Client) BroadcastTxCommit(_a0 context.Context, _a1 types.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *coretypes.ResultBroadcastTxCommit
	if rf, ok := ret.Get(0).(func(context.Context, types.Tx) *coretypes.ResultBroadcastTxCommit); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTxCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.Tx) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastTxSync provides a mock function with given fields: _a0, _a1
func (_m *Client) BroadcastTxSync(_a0 context.Context, _a1 types.Tx) (*coretypes.ResultBroadcastTx, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *coretypes.ResultBroadcastTx
	if rf, ok := ret.Get(0).(func(context.Context, types.Tx) *coretypes.ResultBroadcastTx); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.Tx) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CheckTx provides a mock function with given fields: _a0, _a1
func (_m *Client) CheckTx(_a0 context.Context, _a1 types.Tx) (*coretypes.ResultCheckTx, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *coretypes.ResultCheckTx
	if rf, ok := ret.Get(0).(func(context.Context, types.Tx) *coretypes.ResultCheckTx); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultCheckTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.Tx) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Commit provides a mock function with given fields: ctx, height
func (_m *Client) Commit(ctx context.Context, height *int64) (*coretypes.ResultCommit, error) {
	ret := _m.Called(ctx, height)

	var r0 *coretypes.ResultCommit
	if rf, ok := ret.Get(0).(func(context.Context, *int64) *coretypes.ResultCommit); ok {
		r0 = rf(ctx, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *int64) error); ok {
		r1 = rf(ctx, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConsensusParams provides a mock function with given fields: ctx, height
func (_m *Client) ConsensusParams(ctx context.Context, height *int64) (*coretypes.ResultConsensusParams, error) {
	ret := _m.Called(ctx, height)

	var r0 *coretypes.ResultConsensusParams
	if rf, ok := ret.Get(0).(func(context.Context, *int64) *coretypes.ResultConsensusParams); ok {
		r0 = rf(ctx, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultConsensusParams)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *int64) error); ok {
		r1 = rf(ctx, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConsensusState provides a mock function with given fields: _a0
func (_m *Client) ConsensusState(_a0 context.Context) (*coretypes.ResultConsensusState, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultConsensusState
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultConsensusState); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultConsensusState)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DumpConsensusState provides a mock function with given fields: _a0
func (_m *Client) DumpConsensusState(_a0 context.Context) (*coretypes.ResultDumpConsensusState, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultDumpConsensusState
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultDumpConsensusState); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultDumpConsensusState)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Genesis provides a mock function with given fields: _a0
func (_m *Client) Genesis(_a0 context.Context) (*coretypes.ResultGenesis, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultGenesis
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultGenesis); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultGenesis)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenesisChunked provides a mock function with given fields: _a0, _a1
func (_m *Client) GenesisChunked(_a0 context.Context, _a1 uint) (*coretypes.ResultGenesisChunk, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *coretypes.ResultGenesisChunk
	if rf, ok := ret.Get(0).(func(context.Context, uint) *coretypes.ResultGenesisChunk); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultGenesisChunk)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Health provides a mock function with given fields: _a0
func (_m *Client) Health(_a0 context.Context) (*coretypes.ResultHealth, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultHealth
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultHealth); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultHealth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsRunning provides a mock function with given fields:
func (_m *Client) IsRunning() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NetInfo provides a mock function with given fields: _a0
func (_m *Client) NetInfo(_a0 context.Context) (*coretypes.ResultNetInfo, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultNetInfo
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultNetInfo); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultNetInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NumUnconfirmedTxs provides a mock function with given fields: _a0
func (_m *Client) NumUnconfirmedTxs(_a0 context.Context) (*coretypes.ResultUnconfirmedTxs, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultUnconfirmedTxs
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultUnconfirmedTxs); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultUnconfirmedTxs)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OnReset provides a mock function with given fields:
func (_m *Client) OnReset() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnStart provides a mock function with given fields:
func (_m *Client) OnStart() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnStop provides a mock function with given fields:
func (_m *Client) OnStop() {
	_m.Called()
}

// Quit provides a mock function with given fields:
func (_m *Client) Quit() <-chan struct{} {
	ret := _m.Called()

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// Reset provides a mock function with given fields:
func (_m *Client) Reset() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetLogger provides a mock function with given fields: _a0
func (_m *Client) SetLogger(_a0 log.Logger) {
	_m.Called(_a0)
}

// Start provides a mock function with given fields:
func (_m *Client) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Status provides a mock function with given fields: _a0
func (_m *Client) Status(_a0 context.Context) (*coretypes.ResultStatus, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultStatus
	if rf, ok := ret.Get(0).(func(context.Context) *coretypes.ResultStatus); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultStatus)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Stop provides a mock function with given fields:
func (_m *Client) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// String provides a mock function with given fields:
func (_m *Client) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Subscribe provides a mock function with given fields: ctx, subscriber, query, outCapacity
func (_m *Client) Subscribe(ctx context.Context, subscriber string, query string, outCapacity ...int) (<-chan coretypes.ResultEvent, error) {
	_va := make([]interface{}, len(outCapacity))
	for _i := range outCapacity {
		_va[_i] = outCapacity[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, subscriber, query)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan coretypes.ResultEvent
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...int) <-chan coretypes.ResultEvent); ok {
		r0 = rf(ctx, subscriber, query, outCapacity...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan coretypes.ResultEvent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...int) error); ok {
		r1 = rf(ctx, subscriber, query, outCapacity...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Tx provides a mock function with given fields: ctx, hash, prove
func (_m *Client) Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error) {
	ret := _m.Called(ctx, hash, prove)

	var r0 *coretypes.ResultTx
	if rf, ok := ret.Get(0).(func(context.Context, []byte, bool) *coretypes.ResultTx); ok {
		r0 = rf(ctx, hash, prove)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []byte, bool) error); ok {
		r1 = rf(ctx, hash, prove)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TxSearch provides a mock function with given fields: ctx, query, prove, page, perPage, orderBy
func (_m *Client) TxSearch(ctx context.Context, query string, prove bool, page *int, perPage *int, orderBy string) (*coretypes.ResultTxSearch, error) {
	ret := _m.Called(ctx, query, prove, page, perPage, orderBy)

	var r0 *coretypes.ResultTxSearch
	if rf, ok := ret.Get(0).(func(context.Context, string, bool, *int, *int, string) *coretypes.ResultTxSearch); ok {
		r0 = rf(ctx, query, prove, page, perPage, orderBy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultTxSearch)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, bool, *int, *int, string) error); ok {
		r1 = rf(ctx, query, prove, page, perPage, orderBy)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnconfirmedTxs provides a mock function with given fields: ctx, limit
func (_m *Client) UnconfirmedTxs(ctx context.Context, limit *int) (*coretypes.ResultUnconfirmedTxs, error) {
	ret := _m.Called(ctx, limit)

	var r0 *coretypes.ResultUnconfirmedTxs
	if rf, ok := ret.Get(0).(func(context.Context, *int) *coretypes.ResultUnconfirmedTxs); ok {
		r0 = rf(ctx, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultUnconfirmedTxs)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *int) error); ok {
		r1 = rf(ctx, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Unsubscribe provides a mock function with given fields: ctx, subscriber, query
func (_m *Client) Unsubscribe(ctx context.Context, subscriber string, query string) error {
	ret := _m.Called(ctx, subscriber, query)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, subscriber, query)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnsubscribeAll provides a mock function with given fields: ctx, subscriber
func (_m *Client) UnsubscribeAll(ctx context.Context, subscriber string) error {
	ret := _m.Called(ctx, subscriber)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, subscriber)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Validators provides a mock function with given fields: ctx, height, page, perPage
func (_m *Client) Validators(ctx context.Context, height *int64, page *int, perPage *int) (*coretypes.ResultValidators, error) {
	ret := _m.Called(ctx, height, page, perPage)

	var r0 *coretypes.ResultValidators
	if rf, ok := ret.Get(0).(func(context.Context, *int64, *int, *int) *coretypes.ResultValidators); ok {
		r0 = rf(ctx, height, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultValidators)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *int64, *int, *int) error); ok {
		r1 = rf(ctx, height, page, perPage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
