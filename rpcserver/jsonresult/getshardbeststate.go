package jsonresult

import (
	"github.com/levietcuong2602/incognito-chain/blockchain"
	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/incognitokey"
)

type GetShardBestState struct {
	BlockVersion           int               `json:"BlockVersion"`
	BestBlockHash          common.Hash       `json:"BestBlockHash"` // hash of block.
	BestBeaconHash         common.Hash       `json:"BestBeaconHash"`
	BeaconHeight           uint64            `json:"BeaconHeight"`
	ShardID                byte              `json:"ShardID"`
	Epoch                  uint64            `json:"Epoch"`
	ShardHeight            uint64            `json:"ShardHeight"`
	MaxShardCommitteeSize  int               `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize  int               `json:"MinShardCommitteeSize"`
	ShardProposerIdx       int               `json:"ShardProposerIdx"`
	ShardCommittee         []string          `json:"ShardCommittee"`
	ShardPendingValidator  []string          `json:"ShardPendingValidator"`
	BestCrossShard         map[byte]uint64   `json:"BestCrossShard"` // Best cross shard block by heigh
	StakingTx              map[string]string `json:"StakingTx"`
	NumTxns                uint64            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns              uint64            `json:"TotalTxns"`              // The total number of txns in the chain.
	TotalTxnsExcludeSalary uint64            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards           int               `json:"ActiveShards"`
	MetricBlockHeight      uint64            `json:"MetricBlockHeight"`
	CommitteeFromBlock     common.Hash       `json:"CommitteeFromBlock"`
	CommitteeEngineVersion int               `json:"CommitteeStateVersion"`
	TriggeredFeature       map[string]uint64 `json:"TriggeredFeature"`
	MaxTxsReminder         int64             `json:"MaxTxsReminder"`
}

func NewGetShardBestState(data *blockchain.ShardBestState) *GetShardBestState {
	result := &GetShardBestState{
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
		BestCrossShard:         data.BestCrossShard,
		CommitteeFromBlock:     data.CommitteeFromBlock(),
		BlockVersion:           data.BestBlock.GetVersion(),
		MaxTxsReminder:         data.MaxTxsPerBlockRemainder,
	}

	result.TriggeredFeature = make(map[string]uint64)
	for k, v := range data.TriggeredFeature {
		result.TriggeredFeature[k] = v
	}

	result.ShardCommittee = make([]string, len(data.ShardCommitteeEngine().GetShardCommittee()))

	shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(data.ShardCommitteeEngine().GetShardCommittee())
	if err != nil {
		panic(err)
	}
	copy(result.ShardCommittee, shardCommitteeStr)
	result.ShardPendingValidator = make([]string, len(data.ShardCommitteeEngine().GetShardSubstitute()))

	shardPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(data.ShardCommitteeEngine().GetShardSubstitute())
	if err != nil {
		panic(err)
	}
	copy(result.ShardPendingValidator, shardPendingValidatorStr)

	result.StakingTx = make(map[string]string)
	result.CommitteeEngineVersion = data.CommitteeStateVersion()

	return result
}
