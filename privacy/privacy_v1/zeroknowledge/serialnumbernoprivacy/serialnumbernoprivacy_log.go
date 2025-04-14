package serialnumbernoprivacy

import "github.com/levietcuong2602/incognito-chain/common"

type SerialnumbernoprivacyLogger struct {
	Log common.Logger
}

func (logger *SerialnumbernoprivacyLogger) Init(inst common.Logger) {
	logger.Log = inst
}

// Global instant to use
var Logger = SerialnumbernoprivacyLogger{}
