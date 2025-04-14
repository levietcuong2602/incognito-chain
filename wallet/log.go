package wallet

import "github.com/levietcuong2602/incognito-chain/common"

type WalletLogger struct {
	log common.Logger
}

func (walletLogger *WalletLogger) Init(inst common.Logger) {
	walletLogger.log = inst
}

// Global instant to use
var Logger = WalletLogger{}
