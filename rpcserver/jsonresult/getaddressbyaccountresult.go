package jsonresult

import "github.com/levietcuong2602/incognito-chain/wallet"

type GetAddressesByAccount struct {
	Addresses []wallet.KeySerializedData `json:"Addresses"`
}
