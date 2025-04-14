package bulletproofs

import "github.com/levietcuong2602/incognito-chain/common"

type BulletproofsLogger struct {
	Log common.Logger
}

func (logger *BulletproofsLogger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = BulletproofsLogger{}
