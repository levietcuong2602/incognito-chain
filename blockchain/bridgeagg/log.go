package bridgeagg

import "github.com/levietcuong2602/incognito-chain/common"

type BrideAggLogger struct {
	log common.Logger
}

func (bridgeAggLogger *BrideAggLogger) Init(logger common.Logger) {
	bridgeAggLogger.log = logger
}

// Global instant to use
var Logger = BrideAggLogger{}
