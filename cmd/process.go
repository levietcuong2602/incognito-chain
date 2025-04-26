package main

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"log"
)

func parseToJsonString(data interface{}) ([]byte, error) {
	result, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return result, nil
}

func processCmd() {
	switch cfg.Command {
	case getPrivacyTokenID:
		{
			log.Printf("Params %+v", cfg)
			if cfg.PNetwork == "" {
				log.Println("Wrong param")
				return
			}
			if cfg.PToken == "" {
				log.Println("Wrong param")
				return
			}

			//hashPNetWork := common.HashH([]byte(cfg.PNetwork))
			//log.Printf("hashPNetWork: %+v\n", hashPNetWork.String())
			//copy(tokenID[:16], hashPNetWork[:16])
			//log.Printf("tokenID: %+v\n", tokenID.String())
			//
			//hashPToken := common.HashH([]byte(cfg.PToken))
			//log.Printf("hashPToken: %+v\n", hashPToken.String())
			//copy(tokenID[16:], hashPToken[:16])

			point := privacy.HashToPoint([]byte(cfg.PNetwork + "-" + cfg.PToken))
			hash := new(common.Hash)
			err := hash.SetBytes(point.ToBytesS())
			if err != nil {
				log.Println("Wrong param")
				return
			}

			log.Printf("Result tokenID: %+v\n", hash.String())
		}
	case createWalletCmd:
		{
			if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
				log.Println("Wrong param")
				return
			}
			err := createWallet()
			if err != nil {
				log.Println(err)
				return
			}
		}
	case listWalletAccountCmd:
		{
			if cfg.WalletPassphrase == "" || cfg.WalletName == "" {
				log.Println("Wrong param")
				return
			}
			accounts, err := listAccounts()
			if err != nil {
				log.Println(err)
				return
			}
			result, err := parseToJsonString(accounts)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(string(result))
		}
	case getWalletAccountCmd:
		{
			if cfg.WalletPassphrase == "" || cfg.WalletName == "" || cfg.WalletAccountName == "" {
				log.Println("Wrong param")
				return
			}
			account, err := getAccount(cfg.WalletAccountName)
			if err != nil {
				log.Println(err)
				return
			}
			result, err := parseToJsonString(account)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(string(result))
		}
	case createWalletAccountCmd:
		{
			if cfg.WalletPassphrase == "" || cfg.WalletName == "" || cfg.WalletAccountName == "" {
				log.Println("Wrong param")
				return
			}
			var shardID *byte
			if cfg.ShardID > -1 {
				temp := byte(cfg.ShardID)
				shardID = &temp
			}
			account, err := createAccount(cfg.WalletAccountName, shardID)
			if err != nil {
				log.Println(err)
				return
			}
			result, err := parseToJsonString(account)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(string(result))
		}
	}
}
