package metadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalLiquidationCustodianDeposit struct {
	MetadataBase
	IncogAddressStr        string
	PTokenId               string
	DepositedAmount        uint64
	FreeCollateralSelected bool
}

type PortalLiquidationCustodianDepositAction struct {
	Meta    PortalLiquidationCustodianDeposit
	TxReqID common.Hash
	ShardID byte
}

type PortalLiquidationCustodianDepositContent struct {
	IncogAddressStr        string
	PTokenId               string
	DepositedAmount        uint64
	FreeCollateralSelected bool
	TxReqID                common.Hash
	ShardID                byte
}

type LiquidationCustodianDepositStatus struct {
	TxReqID                common.Hash
	IncogAddressStr        string
	PTokenId               string
	DepositAmount          uint64
	FreeCollateralSelected bool
	Status                 byte
}

func NewLiquidationCustodianDepositStatus(txReqID common.Hash, incogAddressStr string, PTokenId string, depositAmount uint64, freeCollateralSelected bool, status byte) *LiquidationCustodianDepositStatus {
	return &LiquidationCustodianDepositStatus{TxReqID: txReqID, IncogAddressStr: incogAddressStr, PTokenId: PTokenId, DepositAmount: depositAmount, FreeCollateralSelected: freeCollateralSelected, Status: status}
}

func NewPortalLiquidationCustodianDeposit(metaType int, incognitoAddrStr string, pToken string, amount uint64, freeCollateralSelected bool) (*PortalLiquidationCustodianDeposit, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	custodianDepositMeta := &PortalLiquidationCustodianDeposit{
		IncogAddressStr:        incognitoAddrStr,
		PTokenId:               pToken,
		DepositedAmount:        amount,
		FreeCollateralSelected: freeCollateralSelected,
	}
	custodianDepositMeta.MetadataBase = metadataBase
	return custodianDepositMeta, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	// Note: the metadata was already verified with *transaction.TxCustomToken level so no need to verify with *transaction.Tx level again as *transaction.Tx is embedding property of *transaction.TxCustomToken
	// if txr.GetType() == common.TxCustomTokenPrivacyType && reflect.TypeOf(txr).String() == "*transaction.Tx" {
	// 	return true, true, nil
	// }

	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(custodianDeposit.IncogAddressStr)
	if err != nil {
		return false, false, errors.New("IncogAddressStr of custodian incorrect")
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return false, false, errors.New("wrong custodian incognito address")
	}

	// check burning tx
	isBurned, burnCoin, burnedTokenID, err := txr.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, errors.New("Error This is not Tx Burn")
	}
	// validate amount deposit
	if custodianDeposit.DepositedAmount == 0 || custodianDeposit.DepositedAmount != burnCoin.GetValue() {
		return false, false, errors.New("deposit amount should be equal to the tx value")
	}
	// check tx type
	if txr.GetType() != common.TxNormalType || !bytes.Equal(burnedTokenID.Bytes(), common.PRVCoinID[:]) {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	isPortalToken, err := chainRetriever.IsPortalToken(beaconHeight, custodianDeposit.PTokenId, common.PortalVersion3)
	if !isPortalToken || err != nil {
		return false, false, errors.New("TokenID in remote address is invalid")
	}
	return true, true, nil
}

func (custodianDeposit PortalLiquidationCustodianDeposit) ValidateMetadataByItself() bool {
	return custodianDeposit.Type == PortalCustodianTopupMeta
}

func (custodianDeposit PortalLiquidationCustodianDeposit) Hash() *common.Hash {
	record := custodianDeposit.MetadataBase.Hash().String()
	record += custodianDeposit.IncogAddressStr
	record += custodianDeposit.PTokenId
	record += strconv.FormatUint(custodianDeposit.DepositedAmount, 10)
	record += strconv.FormatBool(custodianDeposit.FreeCollateralSelected)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (custodianDeposit *PortalLiquidationCustodianDeposit) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalLiquidationCustodianDepositAction{
		Meta:    *custodianDeposit,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalCustodianTopupMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (custodianDeposit *PortalLiquidationCustodianDeposit) CalculateSize() uint64 {
	return calculateSize(custodianDeposit)
}
