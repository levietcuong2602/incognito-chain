package syncker

import (
	"context"
	"github.com/levietcuong2602/incognito-chain/multiview"

	"github.com/levietcuong2602/incognito-chain/blockchain/types"
	"github.com/levietcuong2602/incognito-chain/wire"

	"github.com/levietcuong2602/incognito-chain/incdb"

	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/incognitokey"
)

type Network interface {
	//network
	RequestBeaconBlocksViaStream(ctx context.Context, peerID string, from uint64, to uint64) (blockCh chan types.BlockInterface, err error)
	RequestShardBlocksViaStream(ctx context.Context, peerID string, fromSID int, from uint64, to uint64) (blockCh chan types.BlockInterface, err error)
	RequestCrossShardBlocksViaStream(ctx context.Context, peerID string, fromSID int, toSID int, heights []uint64) (blockCh chan types.BlockInterface, err error)
	RequestCrossShardBlocksByHashViaStream(ctx context.Context, peerID string, fromSID int, toSID int, hashes [][]byte) (blockCh chan types.BlockInterface, err error)
	RequestBeaconBlocksByHashViaStream(ctx context.Context, peerID string, hashes [][]byte) (blockCh chan types.BlockInterface, err error)
	RequestShardBlocksByHashViaStream(ctx context.Context, peerID string, fromSID int, hashes [][]byte) (blockCh chan types.BlockInterface, err error)
	PublishMessageToShard(msg wire.Message, shardID byte) error
	SetSyncMode(string)
}

type BeaconChainInterface interface {
	Chain
	GetShardBestViewHash() map[byte]common.Hash
	GetShardBestViewHeight() map[byte]uint64
}

type ShardChainInterface interface {
	Chain
	GetCrossShardState() map[byte]uint64
}

type Chain interface {
	GetBestView() multiview.View
	GetViewByHash(common.Hash) multiview.View
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
	ReplacePreviousValidationData(previousBlockHash common.Hash, previousProposeHash common.Hash, previousCommittees []incognitokey.CommitteePublicKey, newValidationData string) error
	CheckExistedBlk(block types.BlockInterface) bool
	GetCommitteeForSync(types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) // Using only for stream blocks by gRPC
	CommitteeStateVersion() int
}

const (
	BeaconPoolType = iota
	ShardPoolType
	CrossShardPoolType
)
