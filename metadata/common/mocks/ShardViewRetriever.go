// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	incognito_chaincommon "github.com/levietcuong2602/incognito-chain/common"
	mock "github.com/stretchr/testify/mock"

	statedb "github.com/levietcuong2602/incognito-chain/dataaccessobject/statedb"
)

// ShardViewRetriever is an autogenerated mock type for the ShardViewRetriever type
type ShardViewRetriever struct {
	mock.Mock
}

// GetBeaconHeight provides a mock function with given fields:
func (_m *ShardViewRetriever) GetBeaconHeight() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetCopiedFeatureStateDB provides a mock function with given fields:
func (_m *ShardViewRetriever) GetCopiedFeatureStateDB() *statedb.StateDB {
	ret := _m.Called()

	var r0 *statedb.StateDB
	if rf, ok := ret.Get(0).(func() *statedb.StateDB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*statedb.StateDB)
		}
	}

	return r0
}

// GetCopiedTransactionStateDB provides a mock function with given fields:
func (_m *ShardViewRetriever) GetCopiedTransactionStateDB() *statedb.StateDB {
	ret := _m.Called()

	var r0 *statedb.StateDB
	if rf, ok := ret.Get(0).(func() *statedb.StateDB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*statedb.StateDB)
		}
	}

	return r0
}

// GetEpoch provides a mock function with given fields:
func (_m *ShardViewRetriever) GetEpoch() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetHeight provides a mock function with given fields:
func (_m *ShardViewRetriever) GetHeight() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// GetShardID provides a mock function with given fields:
func (_m *ShardViewRetriever) GetShardID() byte {
	ret := _m.Called()

	var r0 byte
	if rf, ok := ret.Get(0).(func() byte); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(byte)
	}

	return r0
}

// GetShardRewardStateDB provides a mock function with given fields:
func (_m *ShardViewRetriever) GetShardRewardStateDB() *statedb.StateDB {
	ret := _m.Called()

	var r0 *statedb.StateDB
	if rf, ok := ret.Get(0).(func() *statedb.StateDB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*statedb.StateDB)
		}
	}

	return r0
}

// GetStakingTx provides a mock function with given fields:
func (_m *ShardViewRetriever) GetStakingTx() map[string]string {
	ret := _m.Called()

	var r0 map[string]string
	if rf, ok := ret.Get(0).(func() map[string]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	return r0
}

// ListShardPrivacyTokenAndPRV provides a mock function with given fields:
func (_m *ShardViewRetriever) ListShardPrivacyTokenAndPRV() []incognito_chaincommon.Hash {
	ret := _m.Called()

	var r0 []incognito_chaincommon.Hash
	if rf, ok := ret.Get(0).(func() []incognito_chaincommon.Hash); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognito_chaincommon.Hash)
		}
	}

	return r0
}
