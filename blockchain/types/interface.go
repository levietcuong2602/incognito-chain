package types

import "github.com/incognitochain/incognito-chain/common"

type BlockPoolInterface interface {
	GetPrevHash() common.Hash
	Hash() *common.Hash
	GetHeight() uint64
	GetShardID() int
	GetRound() int
	FullHashString() string //hash-proposehash
}

type BlockInterface interface {
	GetVersion() int
	GetHeight() uint64
	GetBeaconHeight() uint64
	Hash() *common.Hash
	ProposeHash() *common.Hash
	FullHashString() string //hash-proposehash
	//AddValidationField(validateData string) error
	GetProducer() string
	GetProduceTime() int64
	GetProposeTime() int64
	GetProposer() string
	GetValidationField() string
	AddValidationField(validationData string)
	GetRound() int
	GetRoundKey() string
	GetInstructions() [][]string
	GetConsensusType() string
	GetCurrentEpoch() uint64
	GetPrevHash() common.Hash
	Type() string
	CommitteeFromBlock() common.Hash
	BodyHash() common.Hash
	GetAggregateRootHash() common.Hash
	GetFinalityHeight() uint64
	GetShardID() int
	ToBytes() ([]byte, error)
}

var CommitteeProvider CommitteeProvide

type CommitteeProvide interface {
	GetValidatorFromIndex(
		valdatorIdx int,
		cID byte,
		beaconHash common.Hash,
		prevHash common.Hash,
		epoch uint64,
	) (string, error)
	GetValidatorIndex(
		valdatorPk string,
		cID byte,
		beaconHash common.Hash,
		prevHash common.Hash,
		epoch uint64,
	) (int, error)
}
