// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	incognito_chaincommon "github.com/levietcuong2602/incognito-chain/common"
	mock "github.com/stretchr/testify/mock"
)

// ValidationEnviroment is an autogenerated mock type for the ValidationEnviroment type
type ValidationEnviroment struct {
	mock.Mock
}

// BeaconHeight provides a mock function with given fields:
func (_m *ValidationEnviroment) BeaconHeight() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// ConfirmedTime provides a mock function with given fields:
func (_m *ValidationEnviroment) ConfirmedTime() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// DBData provides a mock function with given fields:
func (_m *ValidationEnviroment) DBData() [][]byte {
	ret := _m.Called()

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func() [][]byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	return r0
}

// HasCA provides a mock function with given fields:
func (_m *ValidationEnviroment) HasCA() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsConfimed provides a mock function with given fields:
func (_m *ValidationEnviroment) IsConfimed() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsPrivacy provides a mock function with given fields:
func (_m *ValidationEnviroment) IsPrivacy() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ShardHeight provides a mock function with given fields:
func (_m *ValidationEnviroment) ShardHeight() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// ShardID provides a mock function with given fields:
func (_m *ValidationEnviroment) ShardID() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// SigPubKey provides a mock function with given fields:
func (_m *ValidationEnviroment) SigPubKey() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// TokenID provides a mock function with given fields:
func (_m *ValidationEnviroment) TokenID() incognito_chaincommon.Hash {
	ret := _m.Called()

	var r0 incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func() incognito_chaincommon.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(incognito_chaincommon.Hash)
		}
	}

	return r0
}

// TxAction provides a mock function with given fields:
func (_m *ValidationEnviroment) TxAction() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// TxType provides a mock function with given fields:
func (_m *ValidationEnviroment) TxType() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Version provides a mock function with given fields:
func (_m *ValidationEnviroment) Version() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}
