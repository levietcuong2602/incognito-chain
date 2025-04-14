package finishsync

import "github.com/levietcuong2602/incognito-chain/common"

type FinishSyncLogger struct {
	common.Logger
}

func (self *FinishSyncLogger) Init(inst common.Logger) {
	self.Logger = inst
}

// Global instant to use
var Logger = FinishSyncLogger{}
