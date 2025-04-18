package wire

import (
	"encoding/json"
	"github.com/levietcuong2602/incognito-chain/blockchain/types"

	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/incognitokey"
	"github.com/libp2p/go-libp2p-peer"
)

type MessageBlockShard struct {
	Block                  *types.ShardBlock
	PreviousValidationData string
}

func (msg *MessageBlockShard) Hash() string {
	rawBytes, err := msg.JsonSerialize()
	if err != nil {
		return ""
	}
	return common.HashH(rawBytes).String()
}

func (msg *MessageBlockShard) MessageType() string {
	return CmdBlockShard
}

func (msg *MessageBlockShard) MaxPayloadLength(pver int) int {
	return MaxBlockPayload
}

func (msg *MessageBlockShard) JsonSerialize() ([]byte, error) {
	jsonBytes, err := json.Marshal(msg)
	return jsonBytes, err
}

func (msg *MessageBlockShard) JsonDeserialize(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), msg)
	return err
}

func (msg *MessageBlockShard) SetSenderID(senderID peer.ID) error {
	return nil
}

func (msg *MessageBlockShard) SignMsg(_ *incognitokey.KeySet) error {
	return nil
}

func (msg *MessageBlockShard) VerifyMsgSanity() error {
	return nil
}
