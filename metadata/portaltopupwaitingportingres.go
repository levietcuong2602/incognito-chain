package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

type PortalTopUpWaitingPortingResponse struct {
	MetadataBase
	DepositStatus string
	ReqTxID       common.Hash
	SharedRandom  []byte `json:"SharedRandom,omitempty"`
}

func NewPortalTopUpWaitingPortingResponse(
	depositStatus string,
	reqTxID common.Hash,
	metaType int,
) *PortalTopUpWaitingPortingResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}

	return &PortalTopUpWaitingPortingResponse{
		DepositStatus: depositStatus,
		ReqTxID:       reqTxID,
		MetadataBase:  metadataBase,
	}
}

func (iRes PortalTopUpWaitingPortingResponse) CheckTransactionFee(tr Transaction, minFeePerKb uint64, minFeePerTx uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (iRes PortalTopUpWaitingPortingResponse) ValidateTxWithBlockChain(txr Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, db *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requested tx (via RequestedTxID)
	return false, nil
}

func (iRes PortalTopUpWaitingPortingResponse) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	return false, true, nil
}

func (iRes PortalTopUpWaitingPortingResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return iRes.Type == PortalTopUpWaitingPortingResponseMeta
}

func (iRes PortalTopUpWaitingPortingResponse) Hash() *common.Hash {
	record := iRes.DepositStatus
	record += iRes.ReqTxID.String()
	record += iRes.MetadataBase.Hash().String()

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (iRes *PortalTopUpWaitingPortingResponse) CalculateSize() uint64 {
	return calculateSize(iRes)
}

func (iRes PortalTopUpWaitingPortingResponse) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *MintData, shardID byte, tx Transaction, chainRetriever ChainRetriever, ac *AccumulatedValues, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever) (bool, error) {
	idx := -1

	for i, inst := range mintData.Insts {
		if len(inst) < 4 { // this is not PortalTopUpWaitingPorting response instruction
			continue
		}
		instMetaType := inst[0]
		if mintData.InstsUsed[i] > 0 ||
			instMetaType != strconv.Itoa(PortalTopUpWaitingPortingRequestMeta) {
			continue
		}
		instDepositStatus := inst[2]
		if instDepositStatus != iRes.DepositStatus ||
			(instDepositStatus != pCommon.PortalRequestRejectedChainStatus) {
			continue
		}

		var shardIDFromInst byte
		var txReqIDFromInst common.Hash
		var depositorAddrStrFromInst string
		var depositedAmountFromInst uint64

		contentBytes := []byte(inst[3])
		var topUpWaitingPortingReqContent PortalTopUpWaitingPortingRequestContent
		err := json.Unmarshal(contentBytes, &topUpWaitingPortingReqContent)
		if err != nil {
			Logger.log.Error("WARNING - VALIDATION: an error occured while parsing portal top up waiting porting request content: ", err)
			continue
		}
		shardIDFromInst = topUpWaitingPortingReqContent.ShardID
		txReqIDFromInst = topUpWaitingPortingReqContent.TxReqID
		depositorAddrStrFromInst = topUpWaitingPortingReqContent.IncogAddressStr
		depositedAmountFromInst = topUpWaitingPortingReqContent.DepositedAmount

		if !bytes.Equal(iRes.ReqTxID[:], txReqIDFromInst[:]) ||
			shardID != shardIDFromInst {
			continue
		}
		key, err := wallet.Base58CheckDeserialize(depositorAddrStrFromInst)
		if err != nil {
			Logger.log.Info("WARNING - VALIDATION: an error occurred while deserializing custodian address string: ", err)
			continue
		}

		isMinted, mintCoin, coinID, err := tx.GetTxMintData()
		if err != nil || !isMinted {
			Logger.log.Info("WARNING - VALIDATION: Error occured while validate tx mint.  ", err)
			continue
		}
		if coinID.String() != common.PRVCoinID.String() {
			Logger.log.Info("WARNING - VALIDATION: Receive Token ID in tx mint maybe not correct. Must be PRV")
			continue
		}
		if ok := mintCoin.CheckCoinValid(key.KeySet.PaymentAddress, iRes.SharedRandom, depositedAmountFromInst); !ok {
			Logger.log.Info("WARNING - VALIDATION: Error occured while check receiver and amount. CheckCoinValid return false ")
			continue
		}

		idx = i
		break
	}

	if idx == -1 { // not found the issuance request tx for this response
		return false, fmt.Errorf(fmt.Sprintf("no PortalTopUpWaitingPortingRequestMeta instruction found for PortalTopUpWaitingPortingResponse tx %s", tx.Hash().String()))
	}
	mintData.InstsUsed[idx] = 1
	return true, nil
}

func (iRes *PortalTopUpWaitingPortingResponse) SetSharedRandom(r []byte) {
	iRes.SharedRandom = r
}
