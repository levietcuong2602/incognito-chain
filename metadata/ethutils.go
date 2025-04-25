package metadata

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/pkg/errors"
)

func VerifyProofAndParseReceipt(blockHash eCommon.Hash, txIndex uint, proofStrs []string) (*types.Receipt, error) {
	gethParam := config.Config().GethParam
	ethHeader, err := GetEVMHeader(blockHash, gethParam.Protocol, gethParam.Host, gethParam.Port)
	if err != nil {
		return nil, NewMetadataTxError(VerifyProofAndParseReceiptError, err)
	}
	if ethHeader == nil {
		Logger.log.Info("WARNING: Could not find out the EVM block header with the hash: ", blockHash)
		return nil, NewMetadataTxError(VerifyProofAndParseReceiptError, errors.Errorf("WARNING: Could not find out the EVM block header with the hash: %s", blockHash.String()))
	}

	mostRecentBlkNum, err := GetMostRecentEVMBlockHeight(gethParam.Protocol, gethParam.Host, gethParam.Port)
	if err != nil {
		Logger.log.Info("WARNING: Could not find the most recent block height on Ethereum")
		return nil, NewMetadataTxError(VerifyProofAndParseReceiptError, err)
	}

	if mostRecentBlkNum.Cmp(big.NewInt(0).Add(ethHeader.Number, big.NewInt(EVMConfirmationBlocks))) == -1 {
		errMsg := fmt.Sprintf("WARNING: It needs 15 confirmation blocks for the process, the requested block (%s) but the latest block (%s)", ethHeader.Number.String(), mostRecentBlkNum.String())
		Logger.log.Info(errMsg)
		return nil, NewMetadataTxError(VerifyProofAndParseReceiptError, errors.New(errMsg))
	}

	keybuf := new(bytes.Buffer)
	keybuf.Reset()
	rlp.Encode(keybuf, txIndex)

	nodeList := new(light.NodeList)
	for _, proofStr := range proofStrs {
		proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
		if err != nil {
			return nil, err
		}
		nodeList.Put([]byte{}, proofBytes)
	}
	proof := nodeList.NodeSet()
	val, _, err := trie.VerifyProof(ethHeader.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		fmt.Printf("WARNING: ETH proof verification failed: %v", err)
		return nil, NewMetadataTxError(VerifyProofAndParseReceiptError, err)
	}
	// Decode value from VerifyProof into Receipt
	constructedReceipt := new(types.Receipt)
	err = rlp.DecodeBytes(val, constructedReceipt)
	if err != nil {
		return nil, NewMetadataTxError(VerifyProofAndParseReceiptError, err)
	}

	if constructedReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, NewMetadataTxError(VerifyProofAndParseReceiptError, errors.New("The constructedReceipt's status is not success"))
	}

	return constructedReceipt, nil
}

func PickAndParseLogMapFromReceiptByContractAddr(
	constructedReceipt *types.Receipt,
	ethContractAddressStr string,
	eventName string) (map[string]interface{}, error) {
	logData := []byte{}
	logLen := len(constructedReceipt.Logs)
	if logLen == 0 {
		Logger.log.Errorf("WARNING: LOG data is invalid.")
		return nil, nil
	}
	for _, log := range constructedReceipt.Logs {
		if bytes.Equal(eCommon.HexToAddress(ethContractAddressStr).Bytes(), log.Address.Bytes()) {
			logData = log.Data
			break
		}
	}
	if len(logData) == 0 {
		Logger.log.Errorf("WARNING: logData is empty.")
		return nil, nil
	}
	return ParseEVMLogDataByEventName(logData, eventName)
}

func ParseEVMLogDataByEventName(data []byte, name string) (map[string]interface{}, error) {
	abiIns, err := abi.JSON(strings.NewReader(common.AbiJson))
	if err != nil {
		return nil, err
	}
	dataMap := map[string]interface{}{}
	if err = abiIns.UnpackIntoMap(dataMap, name, data); err != nil {
		return nil, NewMetadataTxError(UnexpectedError, err)
	}
	return dataMap, nil
}
