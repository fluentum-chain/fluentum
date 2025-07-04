// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	state "github.com/fluentum-chain/fluentum/state"

	types "github.com/fluentum-chain/fluentum/types"
)

// EvidencePool is an autogenerated mock type for the EvidencePool type
type EvidencePool struct {
	mock.Mock
}

// AddEvidence provides a mock function with given fields: _a0
func (_m *EvidencePool) AddEvidence(_a0 types.Evidence) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.Evidence) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CheckEvidence provides a mock function with given fields: _a0
func (_m *EvidencePool) CheckEvidence(_a0 types.EvidenceList) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.EvidenceList) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PendingEvidence provides a mock function with given fields: maxBytes
func (_m *EvidencePool) PendingEvidence(maxBytes int64) ([]types.Evidence, int64) {
	ret := _m.Called(maxBytes)

	var r0 []types.Evidence
	if rf, ok := ret.Get(0).(func(int64) []types.Evidence); ok {
		r0 = rf(maxBytes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.Evidence)
		}
	}

	var r1 int64
	if rf, ok := ret.Get(1).(func(int64) int64); ok {
		r1 = rf(maxBytes)
	} else {
		r1 = ret.Get(1).(int64)
	}

	return r0, r1
}

// Update provides a mock function with given fields: _a0, _a1
func (_m *EvidencePool) Update(_a0 state.State, _a1 types.EvidenceList) {
	_m.Called(_a0, _a1)
}

type NewEvidencePoolT interface {
	mock.TestingT
	Cleanup(func())
}

// NewEvidencePool creates a new instance of EvidencePool. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEvidencePool(t NewEvidencePoolT) *EvidencePool {
	mock := &EvidencePool{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
