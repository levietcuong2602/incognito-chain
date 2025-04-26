package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// whoever can send this type of tx
type ContractingRequest struct {
	BurnerAddress privacy.PaymentAddress
	BurnedAmount  uint64 // must be equal to vout value
	TokenID       common.Hash
	MetadataBase
}

type ContractingReqAction struct {
	Meta    ContractingRequest `json:"meta"`
	TxReqID common.Hash        `json:"txReqId"`
}

func NewContractingRequest(
	burnerAddress privacy.PaymentAddress,
	burnedAmount uint64,
	tokenID common.Hash,
	metaType int,
) (*ContractingRequest, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	contractingReq := &ContractingRequest{
		TokenID:       tokenID,
		BurnedAmount:  burnedAmount,
		BurnerAddress: burnerAddress,
	}
	contractingReq.MetadataBase = metadataBase
	return contractingReq, nil
}

func (cReq ContractingRequest) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	//bridgeTokenExisted, err := statedb.IsBridgeTokenExistedByType(beaconViewRetriever.GetBeaconFeatureStateDB(), cReq.TokenID, true)
	//if err != nil {
	//	return false, err
	//}
	//if !bridgeTokenExisted {
	//	return false, errors.New("the burning token is not existed in bridge tokens")
	//}
	return true, nil
}

func (cReq ContractingRequest) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {

	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	// if reflect.TypeOf(tx).String() == "*transaction.Tx" {
	// 	return true, true, nil
	// }

	if cReq.Type != ContractingRequestMeta {
		return false, false, errors.New("Wrong request info's meta type")
	}
	if _, err := AssertPaymentAddressAndTxVersion(cReq.BurnerAddress, tx.GetVersion()); err != nil {
		return false, false, err
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, errors.New("Error This is not Tx Burn")
	}

	if cReq.BurnedAmount == 0 || cReq.BurnedAmount != burnCoin.GetValue() {
		return false, false, errors.New("Wrong request info's burned amount")
	}

	if !bytes.Equal(burnedTokenID[:], cReq.TokenID[:]) {
		return false, false, errors.New("Wrong request info's token id, it should be equal to tx's token id.")
	}

	return true, true, nil
}

func (cReq ContractingRequest) ValidateMetadataByItself() bool {
	return cReq.Type == ContractingRequestMeta
}

func (cReq ContractingRequest) Hash() *common.Hash {
	record := cReq.MetadataBase.Hash().String()
	record += cReq.BurnerAddress.String()
	record += cReq.TokenID.String()
	// TODO: @hung change to record += fmt.Sprint(cReq.BurnedAmount)
	record += string(cReq.BurnedAmount)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (cReq *ContractingRequest) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":          *cReq,
		"RequestedTxID": tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(ContractingRequestMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (cReq *ContractingRequest) CalculateSize() uint64 {
	return calculateSize(cReq)
}
