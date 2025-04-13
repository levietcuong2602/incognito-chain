package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/syncker"
	"sort"
)

type PoolInfo struct {
	Info map[int][]BlockInfo `json:"Info"`
}

type SyncStats struct {
	Beacon syncker.SyncInfo
	Shard  map[int]*syncker.SyncInfo
}

type BlockInfo struct {
	Height  uint64 `json:"BlockHeight"`
	Hash    string `json:"BlockHash"`
	PreHash string `json:"PreHash"`
}

func NewPoolInfo(blks []types.BlockPoolInterface) *PoolInfo {
	res := &PoolInfo{}
	res.Info = map[int][]BlockInfo{}
	for _, blk := range blks {
		res.Info[blk.GetShardID()] = append(res.Info[blk.GetShardID()], BlockInfo{
			Height:  blk.GetHeight(),
			Hash:    blk.Hash().String(),
			PreHash: blk.GetPrevHash().String(),
		})
	}
	for k, v := range res.Info {
		sort.Slice(v, func(i, j int) bool {
			return v[i].Height < v[j].Height
		})
		res.Info[k] = v
	}
	return res
}

func NewSyncStats(stats syncker.SynckerStats) *SyncStats {
	return &SyncStats{
		Beacon: stats.Beacon,
		Shard:  stats.Shard,
	}
}
