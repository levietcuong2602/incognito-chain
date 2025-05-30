package pubsub

import (
	"github.com/levietcuong2602/incognito-chain/common"
)

type PubSubLogger struct {
	log common.Logger
}

func (metricLogger *PubSubLogger) Init(inst common.Logger) {
	metricLogger.log = inst
}

// Global instant to use
var Logger = PubSubLogger{}
