package syncker

import "github.com/levietcuong2602/incognito-chain/common"

type SynckerLogger struct {
	common.Logger
}

func (self *SynckerLogger) Init(inst common.Logger) {
	self.Logger = inst
}

// Global instant to use
var Logger = SynckerLogger{}
