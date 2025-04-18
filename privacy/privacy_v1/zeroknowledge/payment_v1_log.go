package zkp

import (
	"github.com/levietcuong2602/incognito-chain/common"
	agg "github.com/levietcuong2602/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	oom "github.com/levietcuong2602/incognito-chain/privacy/privacy_v1/zeroknowledge/oneoutofmany"
	snn "github.com/levietcuong2602/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumbernoprivacy"
	snp "github.com/levietcuong2602/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumberprivacy"
	utils "github.com/levietcuong2602/incognito-chain/privacy/privacy_util"
)

type PaymentV1Logger struct {
	Log common.Logger
}

func (logger *PaymentV1Logger) Init(inst common.Logger) {
	logger.Log = inst
	agg.Logger.Init(inst)
	oom.Logger.Init(inst)
	snn.Logger.Init(inst)
	snp.Logger.Init(inst)
	utils.Logger.Init(inst)
}

// Global instant to use
var Logger = PaymentV1Logger{}
