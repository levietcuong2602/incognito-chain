package pdex

import "github.com/levietcuong2602/incognito-chain/common"

type PDEXLogger struct {
	log common.Logger
}

func (pDEXLogger *PDEXLogger) Init(logger common.Logger) {
	pDEXLogger.log = logger
}

// Global instant to use
var Logger = PDEXLogger{}
