package portaltokens

import "github.com/levietcuong2602/incognito-chain/common"

type PortalTokenLoggerV3 struct {
	log common.Logger
}

func (portalTokenLogger *PortalTokenLoggerV3) Init(inst common.Logger) {
	portalTokenLogger.log = inst
}

// Global instant to use
var Logger = PortalTokenLoggerV3{}
