// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	"github.com/cometbft/cometbft/abci/types"
	mock "github.com/stretchr/testify/mock"
)

// AppConnQuery is an autogenerated mock type for the AppConnQuery type
type AppConnQuery struct {
	mock.Mock
}

// EchoSync provides a mock function with given fields: _a0
func (_m *AppConnQuery) EchoSync(_a0 string) (*types.ResponseEcho, error) {
	ret := _m.Called(_a0)

	var r0 *types.ResponseEcho
	if rf, ok := ret.Get(0).(func(string) *types.ResponseEcho); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseEcho)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Error provides a mock function with given fields:
func (_m *AppConnQuery) Error() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InfoSync provides a mock function with given fields: _a0
func (_m *AppConnQuery) InfoSync(_a0 types.RequestInfo) (*types.ResponseInfo, error) {
	ret := _m.Called(_a0)

	var r0 *types.ResponseInfo
	if rf, ok := ret.Get(0).(func(types.RequestInfo) *types.ResponseInfo); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.RequestInfo) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QuerySync provides a mock function with given fields: _a0
func (_m *AppConnQuery) QuerySync(_a0 types.RequestQuery) (*types.ResponseQuery, error) {
	ret := _m.Called(_a0)

	var r0 *types.ResponseQuery
	if rf, ok := ret.Get(0).(func(types.RequestQuery) *types.ResponseQuery); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.ResponseQuery)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.RequestQuery) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewAppConnQueryT interface {
	mock.TestingT
	Cleanup(func())
}

// NewAppConnQuery creates a new instance of AppConnQuery. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAppConnQuery(t NewAppConnQueryT) *AppConnQuery {
	mock := &AppConnQuery{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
