package committeestate

import "github.com/levietcuong2602/incognito-chain/common"

type committeeStateLogger struct {
	log common.Logger
}

func (i *committeeStateLogger) Init(inst common.Logger) {
	i.log = inst
}

// Global instant to use
var Logger = committeeStateLogger{}
