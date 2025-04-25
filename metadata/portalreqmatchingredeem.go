package metadata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// PortalReqMatchingRedeem - portal custodian request matching redeem requests
// metadata - request matching redeem requests - create normal tx with this metadata
type PortalReqMatchingRedeem struct {
	MetadataBaseWithSignature
	CustodianAddressStr string
	RedeemID            string
}

// PortalReqMatchingRedeemAction - shard validator creates instruction that contain this action content
// it will be append to ShardToBeaconBlock
type PortalReqMatchingRedeemAction struct {
	Meta    PortalReqMatchingRedeem
	TxReqID common.Hash
	ShardID byte
}

// PortalReqMatchingRedeemContent - Beacon builds a new instruction with this content after receiving a instruction from shard
// It will be appended to beaconBlock
// both accepted and rejected status
type PortalReqMatchingRedeemContent struct {
	CustodianAddressStr string
	RedeemID            string
	MatchingAmount      uint64
	IsFullCustodian     bool
	TxReqID             common.Hash
	ShardID             byte
}

// PortalReqMatchingRedeemStatus - Beacon tracks status of request matching redeem tx into db
type PortalReqMatchingRedeemStatus struct {
	CustodianAddressStr string
	RedeemID            string
	MatchingAmount      uint64
	Status              byte
}

func NewPortalReqMatchingRedeem(metaType int, custodianAddrStr string, redeemID string) (*PortalReqMatchingRedeem, error) {
	metadataBase := NewMetadataBaseWithSignature(metaType)
	custodianDepositMeta := &PortalReqMatchingRedeem{
		CustodianAddressStr: custodianAddrStr,
		RedeemID:            redeemID,
	}
	custodianDepositMeta.MetadataBaseWithSignature = *metadataBase
	return custodianDepositMeta, nil
}

func (req PortalReqMatchingRedeem) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (req PortalReqMatchingRedeem) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	// validate IncogAddressStr
	keyWallet, err := wallet.Base58CheckDeserialize(req.CustodianAddressStr)
	if err != nil {
		return false, false, errors.New("CustodianAddressStr of custodian incorrect")
	}
	incogAddr := keyWallet.KeySet.PaymentAddress
	if len(incogAddr.Pk) == 0 {
		return false, false, errors.New("wrong custodian incognito address")
	}

	if ok, err := req.MetadataBaseWithSignature.VerifyMetadataSignature(incogAddr.Pk, txr); err != nil || !ok {
		return false, false, errors.New("Request matching redeem is unauthorized")
	}

	// check tx type
	if txr.GetType() != common.TxNormalType {
		return false, false, errors.New("tx custodian deposit must be TxNormalType")
	}

	// validate RedeemID
	if req.RedeemID == "" {
		return false, false, errors.New("RedeemID is incorrect")
	}
	return true, true, nil
}

func (req PortalReqMatchingRedeem) ValidateMetadataByItself() bool {
	return req.Type == PortalReqMatchingRedeemMeta
}

func (req PortalReqMatchingRedeem) Hash() *common.Hash {
	record := req.MetadataBaseWithSignature.Hash().String()
	record += req.CustodianAddressStr
	record += req.RedeemID
	if req.Sig != nil && len(req.Sig) != 0 {
		record += string(req.Sig)
	}
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (req PortalReqMatchingRedeem) HashWithoutSig() *common.Hash {
	record := req.MetadataBaseWithSignature.Hash().String()
	record += req.CustodianAddressStr
	record += req.RedeemID

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (req *PortalReqMatchingRedeem) BuildReqActions(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := PortalReqMatchingRedeemAction{
		Meta:    *req,
		TxReqID: *tx.Hash(),
		ShardID: shardID,
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(PortalReqMatchingRedeemMeta), actionContentBase64Str}
	return [][]string{action}, nil
}

func (req *PortalReqMatchingRedeem) CalculateSize() uint64 {
	return calculateSize(req)
}
