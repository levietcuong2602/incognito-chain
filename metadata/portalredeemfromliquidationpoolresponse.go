package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalRedeemLiquidateExchangeRatesResponse struct {
	MetadataBase
	RequestStatus    string
	ReqTxID          common.Hash
	RequesterAddrStr string
	RedeemAmount     uint64
	Amount           uint64
	TokenID          string
	SharedRandom     []byte `json:"SharedRandom,omitempty"`
}

func NewPortalRedeemLiquidateExchangeRatesResponse(
	requestStatus string,
	reqTxID common.Hash,
	requesterAddressStr string,
	redeemAmount uint64,
	amount uint64,
	tokenID string,
	metaType int,
) *PortalRedeemLiquidateExchangeRatesResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &PortalRedeemLiquidateExchangeRatesResponse{
		RequestStatus:    requestStatus,
		ReqTxID:          reqTxID,
		MetadataBase:     metadataBase,
		RequesterAddrStr: requesterAddressStr,
		RedeemAmount:     redeemAmount,
		Amount:           amount,
		TokenID:          tokenID,
	}
}

func (iRes PortalRedeemLiquidateExchangeRatesResponse) CheckTransactionFee(tr Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalRedeemLiquidateExchangeRatesResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalRedeemLiquidateExchangeRatesResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalRedeemLiquidateExchangeRatesResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalRedeemFromLiquidationPoolResponseMeta
}

func (iRes PortalRedeemLiquidateExchangeRatesResponse) Hash() *common.Hash {
	record := iRes.MetadataBase.Hash().String()
	record += iRes.RequestStatus
	record += iRes.ReqTxID.String()
	record += iRes.RequesterAddrStr
	record += strconv.FormatUint(iRes.RedeemAmount, 10)
	record += strconv.FormatUint(iRes.Amount, 10)
	record += iRes.TokenID
	if iRes.SharedRandom != nil && len(iRes.SharedRandom) > 0 {
		record += string(iRes.SharedRandom)
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalRedeemLiquidateExchangeRatesResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalRedeemLiquidateExchangeRatesResponse) VerifyMinerCreatedTxBeforeGettingInBlock(
	mintData *MintData,
	shardID byte,
	tx Transaction,
	chainRetriever ChainRetriever,
	ac *AccumulatedValues,
	shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
) (bool, error) {
	idx := -1
	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not PortalRedeemFromLiquidationPoolMeta response instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PortalRedeemFromLiquidationPoolMeta) {
			continue
		}
		instReqStatus := inst[2]
		if instReqStatus != iRes.RequestStatus {
			Logger.log.Errorf("WARNING - VALIDATION: instReqStatus %v is different from iRes.RequestStatus %v", instReqStatus, iRes.RequestStatus)
			continue
		}
		if (instReqStatus != pCommon.PortalProducerInstSuccessChainStatus) &&
			(instReqStatus != pCommon.PortalRequestRejectedChainStatus) {
			Logger.log.Errorf("WARNING - VALIDATION: instReqStatus is not correct %v", instReqStatus)
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var requesterAddrStrFromInst string
		var redeemAmountFromInst uint64
		var totalPTokenReceived uint64
		//var tokenIDStrFromInst string

		contentBytes := []byte(inst[3])
		var redeemReqContent PortalRedeemLiquidateExchangeRatesContent
		err := json.Unmarshal(contentBytes, &redeemReqContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occurred while parsing portal redeem liquidate exchange rates content: ", err)
			continue
		}

		shardIDFromInst = redeemReqContent.ShardID
		txReqIDFromInst = redeemReqContent.TxReqID
		requesterAddrStrFromInst = redeemReqContent.RedeemerIncAddressStr
		redeemAmountFromInst = redeemReqContent.RedeemAmount
		totalPTokenReceived = redeemReqContent.TotalPTokenReceived
		//tokenIDStrFromInst = redeemReqContent.TokenID

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}

		if requesterAddrStrFromInst != iRes.RequesterAddrStr {
			Logger.log.Errorf("Error - VALIDATION: Requester address %v is not matching to Requester address in instruction %v", iRes.RequesterAddrStr, requesterAddrStrFromInst)
			continue
		}

		if totalPTokenReceived != iRes.Amount {
			Logger.log.Errorf("Error - VALIDATION:  totalPTokenReceived %v is not matching to  TotalPTokenReceived in instruction %v", iRes.Amount, redeemAmountFromInst)
			continue
		}

		if redeemAmountFromInst != iRes.RedeemAmount {
			Logger.log.Errorf("Error - VALIDATION: Redeem amount %v is not matching to redeem amount in instruction %v", iRes.RedeemAmount, redeemAmountFromInst)
			continue
		}

		key, err := wallet.Base58CheckDeserialize(requesterAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occurred while deserializing requester address string: ", err)
			continue
		}

		mintedTokenID := common.PRVCoinID.String()
		mintedAmount := totalPTokenReceived
		if instReqStatus == pCommon.PortalRequestRejectedChainStatus {
			mintedTokenID = redeemReqContent.TokenID
			mintedAmount = redeemAmountFromInst
		}

		_, pk, paidAmount, assetID := tx.GetTransferData()
		if !bytes.Equal(key.KeySet.PaymentAddress.Pk[:], pk[:]) ||
			mintedAmount != paidAmount ||
			mintedTokenID != assetID.String() {
			continue
		}
		idx = i
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalRedeemLiquidateExchangeRates instruction found for PortalRedeemLiquidateExchangeRatesResponse tx %s", tx.Hash().String()))
	}

	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *PortalRedeemLiquidateExchangeRatesResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}
