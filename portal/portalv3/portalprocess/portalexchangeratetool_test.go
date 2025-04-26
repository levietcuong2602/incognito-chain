package portalprocess

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	"testing"
)

func TestPortalExchangeRateTool(t *testing.T) {
	finalExchangeRates := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:        {Amount: 1000000},
			pCommon.PortalBNBIDStr: {Amount: 40000000},
			pCommon.PortalBTCIDStr: {Amount: 10000000000},
			"USDT":                 {Amount: 1000000},
			common.EthAddrStr:      {Amount: 400000000},
			"Rose":                 {Amount: 500000},
		})

	portalParams := portalv3.PortalParams{
		SupportedCollateralTokens: []portalv3.PortalCollateral{
			{common.EthAddrStr, 9},
			{"USDT", 6},
			{"Rose", 7},
		},
	}
	tool := NewPortalExchangeRateTool(finalExchangeRates, portalParams)

	res, _ := tool.Convert(common.EthAddrStr, "USDT", 1)
	fmt.Println("Res: ", res)
	res2, _ := tool.Convert(common.EthAddrStr, "Rose", 500000000)
	fmt.Println("Res2: ", res2)

	res3, _ := tool.ConvertToUSD(common.EthAddrStr, 10)
	fmt.Println("Res3: ", res3)

	res4, _ := tool.ConvertFromUSD(common.EthAddrStr, 4)
	fmt.Println("res4: ", res4)

}
