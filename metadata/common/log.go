package common

import "github.com/levietcuong2602/incognito-chain/common"

type MetaDataLogger struct {
	Log common.Logger
}

func (metricLogger *MetaDataLogger) Init(inst common.Logger) {
	metricLogger.Log = inst
}

// Global instant to use
var Logger = MetaDataLogger{}
