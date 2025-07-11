// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	state "github.com/fluentum-chain/fluentum/state"

	tendermintstate "github.com/fluentum-chain/fluentum/proto/tendermint/state"

	tenderminttypes "github.com/fluentum-chain/fluentum/types"

	types "github.com/fluentum-chain/fluentum/proto/tendermint/types"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// Bootstrap provides a mock function with given fields: _a0
func (_m *Store) Bootstrap(_a0 state.State) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(state.State) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *Store) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Load provides a mock function with given fields:
func (_m *Store) Load() (state.State, error) {
	ret := _m.Called()

	var r0 state.State
	if rf, ok := ret.Get(0).(func() state.State); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(state.State)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadABCIResponses provides a mock function with given fields: _a0
func (_m *Store) LoadABCIResponses(_a0 int64) (*tendermintstate.ABCIResponses, error) {
	ret := _m.Called(_a0)

	var r0 *tendermintstate.ABCIResponses
	if rf, ok := ret.Get(0).(func(int64) *tendermintstate.ABCIResponses); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tendermintstate.ABCIResponses)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadConsensusParams provides a mock function with given fields: _a0
func (_m *Store) LoadConsensusParams(_a0 int64) (types.ConsensusParams, error) {
	ret := _m.Called(_a0)

	var r0 types.ConsensusParams
	if rf, ok := ret.Get(0).(func(int64) types.ConsensusParams); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(types.ConsensusParams)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadFromDBOrGenesisDoc provides a mock function with given fields: _a0
func (_m *Store) LoadFromDBOrGenesisDoc(_a0 *tenderminttypes.GenesisDoc) (state.State, error) {
	ret := _m.Called(_a0)

	var r0 state.State
	if rf, ok := ret.Get(0).(func(*tenderminttypes.GenesisDoc) state.State); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(state.State)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*tenderminttypes.GenesisDoc) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadFromDBOrGenesisFile provides a mock function with given fields: _a0
func (_m *Store) LoadFromDBOrGenesisFile(_a0 string) (state.State, error) {
	ret := _m.Called(_a0)

	var r0 state.State
	if rf, ok := ret.Get(0).(func(string) state.State); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(state.State)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadLastABCIResponse provides a mock function with given fields: _a0
func (_m *Store) LoadLastABCIResponse(_a0 int64) (*tendermintstate.ABCIResponses, error) {
	ret := _m.Called(_a0)

	var r0 *tendermintstate.ABCIResponses
	if rf, ok := ret.Get(0).(func(int64) *tendermintstate.ABCIResponses); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tendermintstate.ABCIResponses)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadValidators provides a mock function with given fields: _a0
func (_m *Store) LoadValidators(_a0 int64) (*tenderminttypes.ValidatorSet, error) {
	ret := _m.Called(_a0)

	var r0 *tenderminttypes.ValidatorSet
	if rf, ok := ret.Get(0).(func(int64) *tenderminttypes.ValidatorSet); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tenderminttypes.ValidatorSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PruneStates provides a mock function with given fields: _a0, _a1
func (_m *Store) PruneStates(_a0 int64, _a1 int64) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, int64) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: _a0
func (_m *Store) Save(_a0 state.State) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(state.State) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveABCIResponses provides a mock function with given fields: _a0, _a1
func (_m *Store) SaveABCIResponses(_a0 int64, _a1 *tendermintstate.ABCIResponses) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(int64, *tendermintstate.ABCIResponses) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewStoreT interface {
	mock.TestingT
	Cleanup(func())
}

// NewStore creates a new instance of Store. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStore(t NewStoreT) *Store {
	mock := &Store{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
