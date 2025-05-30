package devframework

import (
	"context"

	"github.com/levietcuong2602/incognito-chain/blockchain"
	"github.com/levietcuong2602/incognito-chain/blockchain/types"
	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/common/consensus"
	"github.com/levietcuong2602/incognito-chain/incdb"
	"github.com/levietcuong2602/incognito-chain/incognitokey"
	"github.com/levietcuong2602/incognito-chain/multiview"
	"github.com/levietcuong2602/incognito-chain/syncker"
)

type Chain interface {
	GetBestView() multiview.View
	GetDatabase() incdb.Database
	GetAllViewHash() []common.Hash
	GetBestViewHeight() uint64
	GetFinalViewHeight() uint64
	SetReady(bool)
	IsReady() bool
	GetBestViewHash() string
	GetFinalViewHash() string
	GetEpoch() uint64
	ValidateBlockSignatures(block types.BlockInterface, committees []incognitokey.CommitteePublicKey, numOfFixNode int) error
	GetCommittee() []incognitokey.CommitteePublicKey
	GetLastCommittee() []incognitokey.CommitteePublicKey
	CurrentHeight() uint64
	InsertBlock(block types.BlockInterface, shouldValidate bool) error
	//ReplacePreviousValidationData(blockHash common.Hash, newValidationData string) error
	CheckExistedBlk(block types.BlockInterface) bool
	GetCommitteeByHeight(h uint64) ([]incognitokey.CommitteePublicKey, error)
	CommitteeStateVersion() int
}

type PreView struct {
	View multiview.View
}

type ValidatorIndex []int

type Execute struct {
	sim          *NodeEngine
	appliedChain []int
}

func (exec *Execute) GenerateBlock(args ...interface{}) *Execute {
	args = append(args, exec)
	exec.sim.GenerateBlock(args...)
	return exec
}

func (exec *Execute) NextRound() {
	exec.sim.NextRound()
}

func (sim *NodeEngine) ApplyChain(chain_array ...int) *Execute {
	return &Execute{
		sim,
		chain_array,
	}
}

type Syncker interface {
	GetCrossShardBlocksForShardProducer(state *blockchain.ShardBestState, limit map[byte][]uint64) map[byte][]interface{}
	GetCrossShardBlocksForShardValidator(state *blockchain.ShardBestState, list map[byte][]uint64) (map[byte][]interface{}, error)
	SyncMissingBeaconBlock(ctx context.Context, peerID string, fromHash common.Hash)
	SyncMissingShardBlock(ctx context.Context, peerID string, sid byte, fromHash common.Hash)
	Init(*syncker.SynckerManagerConfig)
	InsertCrossShardBlock(block *types.CrossShardBlock)
}

type Consensus interface {
	GetOneValidator() *consensus.Validator
	GetOneValidatorForEachConsensusProcess() map[int]*consensus.Validator
	ValidateProducerPosition(blk types.BlockInterface, lastProposerIdx int, committee []incognitokey.CommitteePublicKey, minCommitteeSize int) error
	ValidateProducerSig(block types.BlockInterface, consensusType string) error
	ValidateBlockCommitteSig(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error
	IsCommitteeInChain(int) bool
}

const (
	MSG_TX = iota
	MSG_TX_PRIVACYTOKEN

	MSG_BLOCK_SHARD
	MSG_BLOCK_BEACON
	MSG_BLOCK_XSHARD

	MSG_PEER_STATE

	MSG_BFT
)

const (
	BLK_BEACON = iota
	BLK_SHARD
)
