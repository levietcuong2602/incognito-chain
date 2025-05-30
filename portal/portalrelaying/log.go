package portalrelaying

import "github.com/levietcuong2602/incognito-chain/common"

type PortalRelayingLogger struct {
	log common.Logger
}

func (p *PortalRelayingLogger) Init(inst common.Logger) {
	p.log = inst
}

// Global instant to use
var Logger = PortalRelayingLogger{}
