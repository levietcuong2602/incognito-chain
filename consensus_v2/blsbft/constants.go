package blsbft

import (
	"time"

	"github.com/levietcuong2602/incognito-chain/common"
)

const (
	proposePhase  = "PROPOSE"
	listenPhase   = "LISTEN"
	votePhase     = "VOTE"
	newround      = "NEWROUND"
	consensusName = common.BlsConsensus
)

//
const (
	timeout             = 40 * time.Second       // must be at least twice the time of block interval
	maxNetworkDelayTime = 150 * time.Millisecond // in ms
)

var (
	MAX_FINALITY_PROOF = int64(22)
)
