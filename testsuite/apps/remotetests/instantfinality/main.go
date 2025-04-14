package main

import (
	"github.com/levietcuong2602/incognito-chain/testsuite/apps/remotetests"
)

func main() {
	nodeManager := remotetests.NewRemoteNodeManager()
	//NormalScenarioTest(nodeManager)
	InstantFinalityV2(nodeManager)
}
