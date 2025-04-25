package databasemp

import (
	"github.com/incognitochain/incognito-chain/common"
)

type DatabaseInterface interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	HasValue(key []byte) (bool, error)

	AddTransaction(txHash *common.Hash, txType string, valueTx []byte, valueDesc []byte) error
	RemoveTransaction(key *common.Hash) error
	GetTransaction(key *common.Hash) ([]byte, error)
	HasTransaction(key *common.Hash) (bool, error)
	Reset() error
	Load() ([][]byte, [][]byte, error)

	Close() error
}
