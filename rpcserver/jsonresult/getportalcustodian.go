package jsonresult

import "github.com/levietcuong2602/incognito-chain/metadata"

type PortalCustodianWithdrawRequest struct {
	CustodianWithdrawRequest metadata.CustodianWithdrawRequestStatus `json:"CustodianWithdraw"`
}
