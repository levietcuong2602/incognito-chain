package consensus_v2

import (
	"github.com/levietcuong2602/incognito-chain/blockchain"
	"github.com/levietcuong2602/incognito-chain/blockchain/types"
	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/incognitokey"
	"github.com/levietcuong2602/incognito-chain/pubsub"
	"github.com/levietcuong2602/incognito-chain/wire"
	peer "github.com/libp2p/go-libp2p-peer"
)

type EngineConfig struct {
	Node          NodeInterface
	Blockchain    *blockchain.BlockChain
	PubSubManager *pubsub.PubSubManager
}

//Used interfaces
//NodeInterface
type NodeInterface interface {
	PushBlockToAll(block types.BlockInterface, previousValidationData string, isBeacon bool) error
	PushMessageToChain(msg wire.Message, chain common.ChainInterface) error
	IsEnableMining() bool
	GetMiningKeys() string
	GetPrivateKey() string
	GetUserMiningState() (role string, chainID int)
	GetPubkeyMiningState(*incognitokey.CommitteePublicKey) (role string, chainID int)
	IsBeaconFullnode(*incognitokey.CommitteePublicKey) (bool, string)
	RequestMissingViewViaStream(peerID string, hashes [][]byte, fromCID int, chainName string) (err error)
	GetSelfPeerID() peer.ID
}
