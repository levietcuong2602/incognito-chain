package jsonresult

import (
	"github.com/levietcuong2602/incognito-chain/dataaccessobject/statedb"
)

type GetLiquidateExchangeRates struct {
	TokenId     string                        `json:"TokenId"`
	Liquidation statedb.LiquidationPoolDetail `json:"Liquidation"`
}
