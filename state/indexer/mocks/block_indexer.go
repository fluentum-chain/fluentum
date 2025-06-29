// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	query "github.com/fluentum-chain/fluentum/libs/pubsub/query"

	types "github.com/fluentum-chain/fluentum/types"
)

// BlockIndexer is an autogenerated mock type for the BlockIndexer type
type BlockIndexer struct {
	mock.Mock
}

// Has provides a mock function with given fields: height
func (_m *BlockIndexer) Has(height int64) (bool, error) {
	ret := _m.Called(height)

	var r0 bool
	if rf, ok := ret.Get(0).(func(int64) bool); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Index provides a mock function with given fields: _a0
func (_m *BlockIndexer) Index(_a0 types.EventDataNewBlockHeader) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.EventDataNewBlockHeader) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Search provides a mock function with given fields: ctx, q
func (_m *BlockIndexer) Search(ctx context.Context, q *query.Query) ([]int64, error) {
	ret := _m.Called(ctx, q)

	var r0 []int64
	if rf, ok := ret.Get(0).(func(context.Context, *query.Query) []int64); ok {
		r0 = rf(ctx, q)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int64)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *query.Query) error); ok {
		r1 = rf(ctx, q)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewBlockIndexer interface {
	mock.TestingT
	Cleanup(func())
}

// NewBlockIndexer creates a new instance of BlockIndexer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBlockIndexer(t mockConstructorTestingTNewBlockIndexer) *BlockIndexer {
	mock := &BlockIndexer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
