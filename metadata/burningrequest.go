package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// whoever can send this type of tx
type BurningRequest struct {
	BurnerAddress privacy.PaymentAddress
	BurningAmount uint64 // must be equal to vout value
	TokenID       common.Hash
	TokenName     string
	RemoteAddress string
	MetadataBase
}

func NewBurningRequest(
	burnerAddress privacy.PaymentAddress,
	burningAmount uint64,
	tokenID common.Hash,
	tokenName string,
	remoteAddress string,
	metaType int,
) (*BurningRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	burningReq := &BurningRequest{
		BurnerAddress: burnerAddress,
		BurningAmount: burningAmount,
		TokenID:       tokenID,
		TokenName:     tokenName,
		RemoteAddress: remoteAddress,
	}
	burningReq.MetadataBase = metadataBase
	return burningReq, nil
}

func (bReq BurningRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	//bridgeTokenExisted, err := statedb.IsBridgeTokenExistedByType(beaconViewRetriever.GetBeaconFeatureStateDB(), bReq.TokenID, false)
	//if err != nil {
	//	return false, err
	//}
	//if !bridgeTokenExisted {
	//	return false, errors.New("the burning token is not existed in bridge tokens")
	//}
	return true, nil
}

func (bReq BurningRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	// if reflect.TypeOf(tx).String() == "*transaction.Tx" {
	// 	return true, true, nil
	// }

	if _, err := hex.DecodeString(bReq.RemoteAddress); err != nil {
		return false, false, err
	}
	if bReq.BurningAmount == 0 {
		return false, false, errors.New("wrong request info's burned amount")
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, fmt.Errorf("it is not transaction burn. Error %v", err)
	}
	if !bytes.Equal(burnedTokenID[:], bReq.TokenID[:]) {
		return false, false, fmt.Errorf("wrong request info's token id and token burned")
	}
	burnAmount := burnCoin.GetValue()
	if burnAmount != bReq.BurningAmount || burnAmount == 0 {
		return false, false, fmt.Errorf("burn amount is incorrect %v", burnAmount)
	}

	if shardViewRetriever.GetEpoch() >= config.Param().ETHRemoveBridgeSigEpoch && (bReq.Type == BurningRequestMeta || bReq.Type == BurningForDepositToSCRequestMeta) {
		return false, false, fmt.Errorf("metadata type %d is deprecated", bReq.Type)
	}
	if shardViewRetriever.GetEpoch() < config.Param().ETHRemoveBridgeSigEpoch &&
		(bReq.Type == BurningRequestMetaV2 || bReq.Type == BurningForDepositToSCRequestMetaV2 ||
			bReq.Type == BurningPBSCRequestMeta || bReq.Type == metadataCommon.BurningPRVERC20RequestMeta ||
			bReq.Type == metadataCommon.BurningPRVBEP20RequestMeta || bReq.Type == metadataCommon.BurningPBSCForDepositToSCRequestMeta) {
		return false, false, fmt.Errorf("metadata type %d is not supported", bReq.Type)
	}

	if (bReq.Type == metadataCommon.BurningPRVERC20RequestMeta || bReq.Type == metadataCommon.BurningPRVBEP20RequestMeta) && bReq.TokenID.String() != common.PRVIDStr {
		return false, false, fmt.Errorf("metadata type %d does not support for incTokenID %v", bReq.Type, bReq.TokenID.String())
	} else if (bReq.Type != metadataCommon.BurningPRVERC20RequestMeta && bReq.Type != metadataCommon.BurningPRVBEP20RequestMeta) && bReq.TokenID.String() == common.PRVIDStr {
		return false, false, fmt.Errorf("metadata type %d does not support for incTokenID %v", bReq.Type, bReq.TokenID.String())
	}

	return true, true, nil
}

func (bReq BurningRequest) ValidateMetadataByItself() bool {
	return bReq.Type == BurningRequestMeta || bReq.Type == BurningForDepositToSCRequestMeta || bReq.Type == BurningRequestMetaV2 ||
		bReq.Type == BurningForDepositToSCRequestMetaV2 || bReq.Type == BurningPBSCRequestMeta ||
		bReq.Type == metadataCommon.BurningPRVERC20RequestMeta || bReq.Type == metadataCommon.BurningPRVBEP20RequestMeta ||
		bReq.Type == metadataCommon.BurningPBSCForDepositToSCRequestMeta
}

func (bReq BurningRequest) Hash() *common.Hash {
	record := bReq.MetadataBase.Hash().String()
	record += bReq.BurnerAddress.String()
	record += bReq.TokenID.String()
	record += strconv.FormatUint(bReq.BurningAmount, 10)
	record += bReq.TokenName
	record += bReq.RemoteAddress
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (bReq BurningRequest) HashWithoutSig() *common.Hash {
	record := bReq.MetadataBase.Hash().String()
	record += bReq.BurnerAddress.String()
	record += bReq.TokenID.String()
	record += strconv.FormatUint(bReq.BurningAmount, 10)
	record += bReq.TokenName
	record += bReq.RemoteAddress

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (bReq *BurningRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":          *bReq,
		"RequestedTxID": tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(bReq.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bReq *BurningRequest) CalculateSize() uint64 {
	return calculateSize(bReq)
}
