package peerv2

import "github.com/levietcuong2602/incognito-chain/common"

type Peerv2Logger struct {
	common.Logger
}

func (self *Peerv2Logger) Init(inst common.Logger) {
	self.Logger = inst
}

// Global instant to use
var Logger = Peerv2Logger{}
