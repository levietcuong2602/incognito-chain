package addrmanager

import "github.com/incognitochain/incognito-chain/common"

type AddrManagerLogger struct {
	log common.Logger
}

func (addrManagerLogger *AddrManagerLogger) Init(inst common.Logger) {
	addrManagerLogger.log = inst
}

// Global instant to use
var Logger = AddrManagerLogger{}
