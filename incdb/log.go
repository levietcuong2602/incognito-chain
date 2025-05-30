package incdb

import (
	"github.com/levietcuong2602/incognito-chain/common"
)

type DbLogger struct {
	Log common.Logger
}

func (dbLogger *DbLogger) Init(inst common.Logger) {
	dbLogger.Log = inst
}

// Global instant to use
var Logger = DbLogger{}
