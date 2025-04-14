package mock

import (
	"github.com/levietcuong2602/incognito-chain/blockchain/types"
)

type Fee struct {
}

func (f *Fee) RegisterBlock(block *types.ShardBlock) error {
	return nil
}
