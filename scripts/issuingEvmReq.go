package main

import (
	"fmt"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/incognitochain/incognito-chain/utils"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.

	utils.InitTxLogger("logs", "tx_log.log")
	utils.LogPrintf("Log file: %s", cfg.LogFileName)
	defer func() {
		utils.CloseTxLogger()
	}()
	config.LoadParam()
	evmcaller.InitCacher()

	blockHash := "0x5b3e3f0c1e3482031ee04d84b0f27817da5cec7dc6f67ea491fdef1348e0328b"
	incTokenID := "0000000000000000000000000000000000000000000000000000000000000000"

	BlockHash := ethCommon.HexToHash(blockHash)
	TxIndex := uint(0)
	ProofStrs := []string{"+QJwgiCAuQJq+QJnAYJ+VbkBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD5AV35AVqUi5F2t7ABwlT0UzDh+yS4T+S4DWvhoC1LWXk1881n+y7r8dtN68k0zuXHuqcVP5gP2+sudAhOuQEgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA3gtrOnZAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJQxMnNlbmRLUkY1SlhUU3pGb0hOUzRKZHRXWUxjeFZzaGFTd2lSVGVMUkJLdjRIRFJwM0V3ckhZNVpXUHhjZGI5dmU1ajFwUDE2N004Z0hiTUJpMkFMVnplTHluQmU1c3JjVWZDNDhHc2o3b3NLdkxjQjJ0VE5lZEpxYzZ2b3BNa241a280eUI2anV4ZDk3cEY3Q0x3AAAAAAAAAAAAAAAA"}
	IncTokenID := common.HexToHash(incTokenID)
	NetworkID := uint(0)
	MetadataType := metadataCommon.IssuingETHRequestMeta

	request, err := bridge.NewIssuingEVMRequest(BlockHash, TxIndex, ProofStrs, IncTokenID, NetworkID, MetadataType)
	if err != nil {
		fmt.Println("Failed to create IssuingEVMRequest: %v", err)
	}

	result := request.ValidateMetadataByItself()
	fmt.Println("ValidateMetadataByItself: %v", result)
}
