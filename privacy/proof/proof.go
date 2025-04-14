package proof

import (
	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/privacy/coin"
	"github.com/levietcuong2602/incognito-chain/privacy/env"
	errhandler "github.com/levietcuong2602/incognito-chain/privacy/errorhandler"
	"github.com/levietcuong2602/incognito-chain/privacy/key"
	"github.com/levietcuong2602/incognito-chain/privacy/proof/agg_interface"
)

// Paymentproof
type Proof interface {
	GetVersion() uint8
	Init()
	GetInputCoins() []coin.PlainCoin
	GetOutputCoins() []coin.Coin
	GetAggregatedRangeProof() agg_interface.AggregatedRangeProof

	SetInputCoins([]coin.PlainCoin) error
	SetOutputCoins([]coin.Coin) error

	Bytes() []byte
	SetBytes(proofBytes []byte) *errhandler.PrivacyError

	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error

	IsPrivacy() bool
	ValidateSanity(vEnv env.ValidationEnviroment) (bool, error)

	Verify(boolParams map[string]bool, pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, additionalData interface{}) (bool, error)
	VerifyV2(vEnv env.ValidationEnviroment, fee uint64) (bool, error)
}
