package lvdb

import "github.com/levietcuong2602/incognito-chain/common"

// prefix
var (
	txKeyPrefix = []byte("tx-")
)

// splitter
var (
	Splitter = []byte("-[-]-")
)

func getKey(key interface{}) []byte {
	var dbkey []byte
	dbkey = append(txKeyPrefix, key.(*common.Hash)[:]...)
	return dbkey
}
