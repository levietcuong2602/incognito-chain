package jsonresult

import (
	"github.com/levietcuong2602/incognito-chain/blockchain"
	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/incognitokey"
)

type GetShardBestStateDetail struct {
	BestBlockHash          common.Hash                       `json:"BestBlockHash"` // hash of block.
	BestBeaconHash         common.Hash                       `json:"BestBeaconHash"`
	BeaconHeight           uint64                            `json:"BeaconHeight"`
	ShardID                byte                              `json:"ShardID"`
	Epoch                  uint64                            `json:"Epoch"`
	ShardHeight            uint64                            `json:"ShardHeight"`
	MaxShardCommitteeSize  int                               `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize  int                               `json:"MinShardCommitteeSize"`
	ShardProposerIdx       int                               `json:"ShardProposerIdx"`
	ShardCommittee         []incognitokey.CommitteeKeyString `json:"ShardCommittee"`
	ShardPendingValidator  []incognitokey.CommitteeKeyString `json:"ShardPendingValidator"`
	BestCrossShard         map[byte]uint64                   `json:"BestCrossShard"` // Best cross shard block by heigh
	StakingTx              map[string]string                 `json:"StakingTx"`
	NumTxns                uint64                            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns              uint64                            `json:"TotalTxns"`              // The total number of txns in the chain.
	TotalTxnsExcludeSalary uint64                            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards           int                               `json:"ActiveShards"`
	MetricBlockHeight      uint64                            `json:"MetricBlockHeight"`
}

func NewGetShardBestStateDetail(data *blockchain.ShardBestState) *GetShardBestStateDetail {
	result := &GetShardBestStateDetail{
		Epoch:                  data.Epoch,
		ShardID:                data.ShardID,
		MinShardCommitteeSize:  data.MinShardCommitteeSize,
		ActiveShards:           data.ActiveShards,
		BeaconHeight:           data.BeaconHeight,
		BestBeaconHash:         data.BestBeaconHash,
		BestBlockHash:          data.BestBlockHash,
		MaxShardCommitteeSize:  data.MaxShardCommitteeSize,
		MetricBlockHeight:      data.MetricBlockHeight,
		NumTxns:                data.NumTxns,
		ShardHeight:            data.ShardHeight,
		ShardProposerIdx:       data.ShardProposerIdx,
		TotalTxns:              data.TotalTxns,
		TotalTxnsExcludeSalary: data.TotalTxnsExcludeSalary,
	}

	tempShardCommittee := incognitokey.CommitteeKeyListToStringList(data.ShardCommitteeEngine().GetShardCommittee())
	result.ShardCommittee = make([]incognitokey.CommitteeKeyString, len(data.ShardCommitteeEngine().GetShardCommittee()))
	copy(result.ShardCommittee, tempShardCommittee)

	tempShardPendingValidator := incognitokey.CommitteeKeyListToStringList(data.ShardCommitteeEngine().GetShardSubstitute())
	result.ShardPendingValidator = make([]incognitokey.CommitteeKeyString, len(data.ShardCommitteeEngine().GetShardSubstitute()))
	copy(result.ShardPendingValidator, tempShardPendingValidator)

	result.BestCrossShard = make(map[byte]uint64)
	for k, v := range data.BestCrossShard {
		result.BestCrossShard[k] = v
	}
	result.StakingTx = make(map[string]string)

	return result
}
