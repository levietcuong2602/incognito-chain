package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

// buildPortalRefundCustodianDepositTx builds refund tx for custodian deposit tx with status "refund"
// mints PRV to return to custodian
func (curView *ShardBestState) buildPortalRefundCustodianDepositTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Info("[Portal refund custodian deposit] Starting...")
	contentBytes := []byte(contentStr)
	var refundDeposit metadata.PortalCustodianDepositContent
	err := json.Unmarshal(contentBytes, &refundDeposit)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal custodian deposit content: %+v", err)
		return nil, nil
	}
	if refundDeposit.ShardID != shardID {
		return nil, nil
	}

	meta := metadata.NewPortalCustodianDepositResponse(
		"refund",
		refundDeposit.TxReqID,
		refundDeposit.IncogAddressStr,
		metadata.PortalCustodianDepositResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(refundDeposit.IncogAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing custodian address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	txParam := transaction.TxSalaryOutputParams{Amount: refundDeposit.DepositedAmount, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

func (curView *ShardBestState) buildPortalRejectedTopUpWaitingPortingTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Info("[buildPortalRejectedTopUpWaitingPortingTx] Starting...")
	contentBytes := []byte(contentStr)
	var topUpInfo metadata.PortalTopUpWaitingPortingRequestContent
	err := json.Unmarshal(contentBytes, &topUpInfo)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling portal top up waiting porting content: %+v", err)
		return nil, nil
	}
	if topUpInfo.ShardID != shardID {
		return nil, nil
	}

	meta := metadata.NewPortalTopUpWaitingPortingResponse(
		pCommon.PortalRequestRejectedChainStatus,
		topUpInfo.TxReqID,
		metadata.PortalTopUpWaitingPortingResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(topUpInfo.IncogAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while deserializing custodian address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	txParam := transaction.TxSalaryOutputParams{Amount: topUpInfo.DepositedAmount, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

func (curView *ShardBestState) buildPortalLiquidationCustodianDepositReject(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Info("[buildPortalLiquidationCustodianDepositReject] Starting...")
	contentBytes := []byte(contentStr)
	var refundDeposit metadata.PortalLiquidationCustodianDepositContent
	err := json.Unmarshal(contentBytes, &refundDeposit)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling portal liquidation custodian deposit content: %+v", err)
		return nil, nil
	}
	if refundDeposit.ShardID != shardID {
		return nil, nil
	}
	meta := metadata.NewPortalLiquidationCustodianDepositResponse(
		pCommon.PortalRequestRejectedChainStatus,
		refundDeposit.TxReqID,
		refundDeposit.IncogAddressStr,
		refundDeposit.DepositedAmount,
		metadata.PortalCustodianTopupResponseMeta,
	)
	keyWallet, err := wallet.Base58CheckDeserialize(refundDeposit.IncogAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while deserializing custodian liquidation address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	txParam := transaction.TxSalaryOutputParams{Amount: refundDeposit.DepositedAmount, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

func (curView *ShardBestState) buildPortalLiquidationCustodianDepositRejectV2(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Info("[buildPortalLiquidationCustodianDepositRejectV2] Starting...")
	contentBytes := []byte(contentStr)
	var refundDeposit metadata.PortalLiquidationCustodianDepositContentV2
	err := json.Unmarshal(contentBytes, &refundDeposit)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling portal liquidation custodian deposit content: %+v", err)
		return nil, nil
	}
	if refundDeposit.ShardID != shardID {
		return nil, nil
	}

	meta := metadata.NewPortalLiquidationCustodianDepositResponseV2(
		pCommon.PortalRequestRejectedChainStatus,
		refundDeposit.TxReqID,
		refundDeposit.IncogAddressStr,
		refundDeposit.DepositedAmount,
		metadata.PortalCustodianTopupResponseMetaV2,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(refundDeposit.IncogAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while deserializing custodian liquidation address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	txParam := transaction.TxSalaryOutputParams{Amount: refundDeposit.DepositedAmount, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalAcceptedRequestPTokensTx builds response tx for user request ptoken tx with status "accepted"
// mints ptoken to return to user
func (curView *ShardBestState) buildPortalAcceptedRequestPTokensTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalAcceptedRequestPTokensTx] Starting...")
	contentBytes := []byte(contentStr)
	var acceptedReqPToken metadata.PortalRequestPTokensContent
	err := json.Unmarshal(contentBytes, &acceptedReqPToken)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal custodian deposit content: %+v", err)
		return nil, nil
	}
	if acceptedReqPToken.ShardID != shardID {
		Logger.log.Errorf("ERROR: ShardID unexpected expect %v, but got %+v", shardID, acceptedReqPToken.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalRequestPTokensResponse(
		"accepted",
		acceptedReqPToken.TxReqID,
		acceptedReqPToken.IncogAddressStr,
		acceptedReqPToken.PortingAmount,
		acceptedReqPToken.TokenID,
		metadata.PortalUserRequestPTokenResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(acceptedReqPToken.IncogAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing custodian address string: %+v", err)
		return nil, nil
	}

	receiverAddr := keyWallet.KeySet.PaymentAddress
	tokenID, err := new(common.Hash).NewHashFromStr(acceptedReqPToken.TokenID)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, nil
	}

	txParam := transaction.TxSalaryOutputParams{Amount: acceptedReqPToken.PortingAmount, ReceiverAddress: &receiverAddr, TokenID: tokenID}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

func (curView *ShardBestState) buildPortalCustodianWithdrawRequest(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Infof("[Shard buildPortalCustodianWithdrawRequest] Starting...")
	contentBytes := []byte(contentStr)
	var custodianWithdrawRequest metadata.PortalCustodianWithdrawRequestContent
	err := json.Unmarshal(contentBytes, &custodianWithdrawRequest)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling portal custodian withdraw request content: %+v", err)
		return nil, nil
	}
	if custodianWithdrawRequest.ShardID != shardID {
		Logger.log.Errorf("ERROR: ShardID unexpected expect %v, but got %+v", shardID, custodianWithdrawRequest.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalCustodianWithdrawResponse(
		pCommon.PortalRequestAcceptedChainStatus,
		custodianWithdrawRequest.TxReqID,
		custodianWithdrawRequest.PaymentAddress,
		custodianWithdrawRequest.Amount,
		metadata.PortalCustodianWithdrawResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(custodianWithdrawRequest.PaymentAddress)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while deserializing custodian address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiveAmt := custodianWithdrawRequest.Amount
	txParam := transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

func (curView *ShardBestState) buildPortalRedeemLiquidateExchangeRatesRequestTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalRedeemLiquidateExchangeRatesRequestTx] Starting...")
	contentBytes := []byte(contentStr)
	var redeemReqContent metadata.PortalRedeemLiquidateExchangeRatesContent
	err := json.Unmarshal(contentBytes, &redeemReqContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling portal redeem liquidate exchange rates content: %+v", err)
		return nil, nil
	}
	if redeemReqContent.ShardID != shardID {
		Logger.log.Errorf("ERROR: ShardID unexpected expect %v, but got %+v", shardID, redeemReqContent.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalRedeemLiquidateExchangeRatesResponse(
		pCommon.PortalProducerInstSuccessChainStatus,
		redeemReqContent.TxReqID,
		redeemReqContent.RedeemerIncAddressStr,
		redeemReqContent.RedeemAmount,
		redeemReqContent.TotalPTokenReceived,
		redeemReqContent.TokenID,
		metadata.PortalRedeemFromLiquidationPoolResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(redeemReqContent.RedeemerIncAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while deserializing custodian address string: %+v", err)
		return nil, nil
	}

	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiveAmt := redeemReqContent.TotalPTokenReceived
	txParam := transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

func (curView *ShardBestState) buildPortalRedeemLiquidateExchangeRatesRequestTxV3(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalRedeemLiquidateExchangeRatesRequestTxV3] Starting...")
	contentBytes := []byte(contentStr)
	var redeemReqContent metadata.PortalRedeemFromLiquidationPoolContentV3
	err := json.Unmarshal(contentBytes, &redeemReqContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling portal redeem liquidate exchange rates content: %+v", err)
		return nil, nil
	}
	if redeemReqContent.ShardID != shardID {
		Logger.log.Errorf("ERROR: ShardID unexpected expect %v, but got %+v", shardID, redeemReqContent.ShardID)
		return nil, nil
	}
	// skip instructions with MintedPRVCollateral = 0
	if redeemReqContent.MintedPRVCollateral == 0 {
		return nil, nil
	}

	meta := metadata.NewPortalRedeemFromLiquidationPoolResponseV3(
		pCommon.PortalProducerInstSuccessChainStatus,
		redeemReqContent.TxReqID,
		redeemReqContent.RedeemerIncAddressStr,
		redeemReqContent.RedeemAmount,
		redeemReqContent.MintedPRVCollateral,
		redeemReqContent.TokenID,
		metadata.PortalRedeemFromLiquidationPoolResponseMetaV3,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(redeemReqContent.RedeemerIncAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while deserializing custodian address string: %+v", err)
		return nil, nil
	}

	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiveAmt := redeemReqContent.MintedPRVCollateral

	//OTA
	txParam := transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalRejectedRedeemRequestTx builds response tx for user request redeem tx with status "rejected"
// mints ptoken to return to user (ptoken that user burned)
func (curView *ShardBestState) buildPortalRejectedRedeemRequestTx(
	beaconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalRejectedRedeemRequestTx] Starting...")
	contentBytes := []byte(contentStr)
	var redeemReqContent metadata.PortalRedeemRequestContent
	err := json.Unmarshal(contentBytes, &redeemReqContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal redeem request content: %+v", err)
		return nil, nil
	}
	if redeemReqContent.ShardID != shardID {
		Logger.log.Errorf("ERROR: unexpected ShardID, expect %v, but got %+v", shardID, redeemReqContent.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalRedeemRequestResponse(
		"rejected",
		redeemReqContent.TxReqID,
		redeemReqContent.RedeemerIncAddressStr,
		redeemReqContent.RedeemAmount,
		redeemReqContent.TokenID,
		metadata.PortalRedeemRequestResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(redeemReqContent.RedeemerIncAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing requester address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiveAmt := redeemReqContent.RedeemAmount
	tokenID, err := new(common.Hash).NewHashFromStr(redeemReqContent.TokenID)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, nil
	}
	txParam := transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: &receiverAddr, TokenID: tokenID}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalRefundCustodianDepositTx builds refund tx for custodian deposit tx with status "refund"
// mints PRV to return to custodian
func (curView *ShardBestState) buildPortalLiquidateCustodianResponseTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Info("[Portal liquidate custodian response] Starting...")
	contentBytes := []byte(contentStr)
	var liqCustodian metadata.PortalLiquidateCustodianContent
	err := json.Unmarshal(contentBytes, &liqCustodian)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal liquidation custodian content: %+v", err)
		return nil, nil
	}
	if liqCustodian.ShardID != shardID {
		Logger.log.Errorf("ERROR: ShardID is invalid: liqCustodian.ShardID %v - shardID %v", liqCustodian.ShardID, shardID)
		return nil, nil
	}

	if liqCustodian.LiquidatedCollateralAmount == 0 {
		return nil, nil
	}

	meta := metadata.NewPortalLiquidateCustodianResponse(
		liqCustodian.UniqueRedeemID,
		liqCustodian.LiquidatedCollateralAmount,
		liqCustodian.RedeemerIncAddressStr,
		liqCustodian.CustodianIncAddressStr,
		metadata.PortalLiquidateCustodianResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(liqCustodian.RedeemerIncAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing redeemer address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiveAmt := liqCustodian.LiquidatedCollateralAmount
	// OTA
	txParam := transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalAcceptedWithdrawRewardTx builds withdraw portal rewards response tx
// mints rewards in PRV for sending to custodian
func (curView *ShardBestState) buildPortalAcceptedWithdrawRewardTx(
	baeconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Info("[buildPortalAcceptedWithdrawRewardTx] Starting...")
	contentBytes := []byte(contentStr)
	var withdrawRewardContent metadata.PortalRequestWithdrawRewardContent
	err := json.Unmarshal(contentBytes, &withdrawRewardContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal withdraw reward content: %+v", err)
		return nil, nil
	}
	if withdrawRewardContent.ShardID != shardID {
		return nil, nil
	}

	meta := metadata.NewPortalWithdrawRewardResponse(
		withdrawRewardContent.TxReqID,
		withdrawRewardContent.CustodianAddressStr,
		withdrawRewardContent.TokenID,
		withdrawRewardContent.RewardAmount,
		metadata.PortalRequestWithdrawRewardResponseMeta,
	)

	tokenID := withdrawRewardContent.TokenID

	keyWallet, err := wallet.Base58CheckDeserialize(withdrawRewardContent.CustodianAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing custodian address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiverAmt := withdrawRewardContent.RewardAmount
	// OTA
	txParam := transaction.TxSalaryOutputParams{Amount: receiverAmt, ReceiverAddress: &receiverAddr, TokenID: &tokenID}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalRefundPortingFeeTx builds portal refund porting fee tx
func (curView *ShardBestState) buildPortalRefundPortingFeeTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Info("[Portal refund porting fee] Starting...")
	contentBytes := []byte(contentStr)
	var portalPortingRequest metadata.PortalPortingRequestContent
	err := json.Unmarshal(contentBytes, &portalPortingRequest)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal porting request content: %+v", err)
		return nil, nil
	}
	if portalPortingRequest.ShardID != shardID {
		return nil, nil
	}

	meta := metadata.NewPortalFeeRefundResponse(
		pCommon.PortalRequestRejectedChainStatus,
		portalPortingRequest.TxReqID,
		metadata.PortalPortingResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(portalPortingRequest.IncogAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing receiver address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiverAmt := portalPortingRequest.PortingFee
	// OTA
	txParam := transaction.TxSalaryOutputParams{Amount: receiverAmt, ReceiverAddress: &receiverAddr, TokenID: nil}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalRefundRedeemFromLiquidationTx builds response tx for user request redeem from liquidation pool tx with status "rejected"
// mints ptoken to return to user (ptoken that user burned)
func (curView *ShardBestState) buildPortalRefundRedeemLiquidateExchangeRatesTx(
	baeconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalRefundRedeemFromLiquidationTx] Starting...")
	contentBytes := []byte(contentStr)
	var redeemReqContent metadata.PortalRedeemLiquidateExchangeRatesContent
	err := json.Unmarshal(contentBytes, &redeemReqContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal redeem request content: %+v", err)
		return nil, nil
	}
	if redeemReqContent.ShardID != shardID {
		Logger.log.Errorf("ERROR: unexpected ShardID, expect %v, but got %+v", shardID, redeemReqContent.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalRedeemLiquidateExchangeRatesResponse(
		pCommon.PortalRequestRejectedChainStatus,
		redeemReqContent.TxReqID,
		redeemReqContent.RedeemerIncAddressStr,
		redeemReqContent.RedeemAmount,
		redeemReqContent.TotalPTokenReceived,
		redeemReqContent.TokenID,
		metadata.PortalRedeemFromLiquidationPoolResponseMeta,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(redeemReqContent.RedeemerIncAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing requester address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiverAmt := redeemReqContent.RedeemAmount
	tokenID, _ := new(common.Hash).NewHashFromStr(redeemReqContent.TokenID)

	// OTA
	txParam := transaction.TxSalaryOutputParams{Amount: receiverAmt, ReceiverAddress: &receiverAddr, TokenID: tokenID}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}

// buildPortalRefundRedeemFromLiquidationTx builds response tx for user request redeem from liquidation pool tx with status "rejected"
// mints ptoken to return to user (ptoken that user burned)
func (curView *ShardBestState) buildPortalRefundRedeemLiquidateExchangeRatesTxV3(
	baeconState *BeaconBestState,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (metadata.Transaction, error) {
	Logger.log.Errorf("[Shard buildPortalRefundRedeemFromLiquidationTx] Starting...")
	contentBytes := []byte(contentStr)
	var redeemReqContent metadata.PortalRedeemFromLiquidationPoolContentV3
	err := json.Unmarshal(contentBytes, &redeemReqContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling portal redeem request content: %+v", err)
		return nil, nil
	}
	if redeemReqContent.ShardID != shardID {
		Logger.log.Errorf("ERROR: unexpected ShardID, expect %v, but got %+v", shardID, redeemReqContent.ShardID)
		return nil, nil
	}

	meta := metadata.NewPortalRedeemFromLiquidationPoolResponseV3(
		pCommon.PortalRequestRejectedChainStatus,
		redeemReqContent.TxReqID,
		redeemReqContent.RedeemerIncAddressStr,
		redeemReqContent.RedeemAmount,
		redeemReqContent.MintedPRVCollateral,
		redeemReqContent.TokenID,
		metadata.PortalRedeemFromLiquidationPoolResponseMetaV3,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(redeemReqContent.RedeemerIncAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while deserializing requester address string: %+v", err)
		return nil, nil
	}
	receiverAddr := keyWallet.KeySet.PaymentAddress
	receiverAmt := redeemReqContent.RedeemAmount
	tokenID, err := new(common.Hash).NewHashFromStr(redeemReqContent.TokenID)

	txParam := transaction.TxSalaryOutputParams{Amount: receiverAmt, ReceiverAddress: &receiverAddr, TokenID: tokenID}
	makeMD := func (c privacy.Coin) metadata.Metadata{
		if c!=nil && c.GetSharedRandom()!=nil{
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
}
