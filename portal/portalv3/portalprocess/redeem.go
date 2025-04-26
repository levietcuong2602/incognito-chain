package portalprocess

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	"github.com/incognitochain/incognito-chain/wallet"
	"sort"
	"strconv"
)

/* =======
Portal Redeem Request Processor
======= */

type PortalRedeemRequestProcessor struct {
	*PortalInstProcessorV3
}

func (p *PortalRedeemRequestProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalRedeemRequestProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalRedeemRequestProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal redeem request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal redeem request action: %+v", err)
	}
	var actionData metadata.PortalRedeemRequestActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal redeem request action: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal redeem request action: %+v", err)
	}

	optionalData := make(map[string]interface{})

	// Get redeem request with uniqueID from stateDB
	redeemRequestBytes, err := statedb.GetPortalRedeemRequestStatus(stateDB, actionData.Meta.UniqueRedeemID)
	if err != nil {
		Logger.log.Errorf("Redeem request: an error occurred while get redeem request Id from DB: %+v", err)
		return nil, fmt.Errorf("Redeem request: an error occurred while get redeem request Id from DB: %+v", err)
	}

	optionalData["isExistRedeemID"] = len(redeemRequestBytes) > 0
	return optionalData, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildRedeemRequestInst(
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	remoteAddress string,
	redeemFee uint64,
	matchingCustodianDetail []*statedb.MatchingRedeemCustodianDetail,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
	shardHeight uint64,
	redeemerExternalAddress string,
) []string {
	redeemRequestContent := metadata.PortalRedeemRequestContent{
		UniqueRedeemID:          uniqueRedeemID,
		TokenID:                 tokenID,
		RedeemAmount:            redeemAmount,
		RedeemerIncAddressStr:   incAddressStr,
		RemoteAddress:           remoteAddress,
		MatchingCustodianDetail: matchingCustodianDetail,
		RedeemFee:               redeemFee,
		TxReqID:                 txReqID,
		ShardID:                 shardID,
		ShardHeight:             shardHeight,
		RedeemerExternalAddress: redeemerExternalAddress,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

func (p *PortalRedeemRequestProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal redeem request action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRedeemRequestActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal redeem request action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildRedeemRequestInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemFee,
		nil,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
		actionData.ShardHeight,
		meta.RedeemerExternalAddress,
	)

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForRedeemRequest]: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	redeemID := meta.UniqueRedeemID
	// check uniqueRedeemID is existed waitingRedeem list or not
	keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(redeemID)
	keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
	waitingRedeemRequest := currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr]
	if waitingRedeemRequest != nil {
		Logger.log.Errorf("RedeemID is existed in waiting redeem requests list %v\n", redeemID)
		return [][]string{rejectInst}, nil
	}

	// check uniqueRedeemID is existed matched Redeem request list or not
	keyMatchedRedeemRequest := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID).String()
	matchedRedeemRequest := currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequest]
	if matchedRedeemRequest != nil {
		Logger.log.Errorf("RedeemID is existed in matched redeem requests list %v\n", redeemID)
		return [][]string{rejectInst}, nil
	}

	// check uniqueRedeemID is existed in db or not
	if optionalData == nil {
		Logger.log.Errorf("Redeem request: optionalData is null")
		return [][]string{rejectInst}, nil
	}
	isExist, ok := optionalData["isExistRedeemID"].(bool)
	if !ok {
		Logger.log.Errorf("Redeem request: optionalData isExistPortingID is invalid")
		return [][]string{rejectInst}, nil
	}
	if isExist {
		Logger.log.Errorf("Redeem request: Porting request id exist in db %v", redeemID)
		return [][]string{rejectInst}, nil
	}

	// get tokenID from redeemTokenID
	tokenID := meta.TokenID

	// check redeem fee
	if currentPortalState.FinalExchangeRatesState == nil {
		Logger.log.Errorf("Redeem request: Can not get exchange rate at beaconHeight %v\n", beaconHeight)
		return [][]string{rejectInst}, nil
	}
	minRedeemFee, err := CalMinRedeemFee(meta.RedeemAmount, tokenID, currentPortalState.FinalExchangeRatesState, portalParams)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v\n", err)
		return [][]string{rejectInst}, nil
	}

	if meta.RedeemFee < minRedeemFee {
		Logger.log.Errorf("Redeem fee is invalid, minRedeemFee %v, but get %v\n", minRedeemFee, meta.RedeemFee)
		return [][]string{rejectInst}, nil
	}

	// add to waiting Redeem list
	redeemRequest := statedb.NewRedeemRequestWithValue(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemAmount,
		[]*statedb.MatchingRedeemCustodianDetail{},
		meta.RedeemFee,
		beaconHeight+1,
		actionData.TxReqID,
		actionData.ShardID,
		actionData.ShardHeight,
		meta.RedeemerExternalAddress,
	)
	currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr] = redeemRequest

	Logger.log.Infof("[Portal] Build accepted instruction for redeem request")
	inst := buildRedeemRequestInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemFee,
		[]*statedb.MatchingRedeemCustodianDetail{},
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
		actionData.ShardHeight,
		meta.RedeemerExternalAddress,
	)
	return [][]string{inst}, nil
}

func (p *PortalRedeemRequestProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalRedeemRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == pCommon.PortalRequestAcceptedChainStatus {
		// add waiting redeem request into waiting redeems list
		keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(actionData.UniqueRedeemID)
		keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
		redeemRequest := statedb.NewRedeemRequestWithValue(
			actionData.UniqueRedeemID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RemoteAddress,
			actionData.RedeemAmount,
			actionData.MatchingCustodianDetail,
			actionData.RedeemFee,
			beaconHeight+1,
			actionData.TxReqID,
			actionData.ShardID,
			actionData.ShardHeight,
			actionData.RedeemerExternalAddress,
		)
		currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr] = redeemRequest

		// track status of redeem request by redeemID
		redeemRequestStatus := metadata.PortalRedeemRequestStatus{
			Status:                  pCommon.PortalRedeemReqWaitingStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 actionData.TxReqID,
			ShardID:                 actionData.ShardID,
			ShardHeight:             actionData.ShardHeight,
			BeaconHeight:            beaconHeight + 1,
			RedeemerExternalAddress: actionData.RedeemerExternalAddress,
		}
		redeemRequestStatusBytes, _ := json.Marshal(redeemRequestStatus)
		err := statedb.StorePortalRedeemRequestStatus(
			stateDB,
			actionData.UniqueRedeemID,
			redeemRequestStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when storing status of redeem request by redeemID: %v\n", err)
			return nil
		}

		// track status of redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalRedeemRequestStatus{
			Status:                  pCommon.PortalRequestAcceptedStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 actionData.TxReqID,
			ShardID:                 actionData.ShardID,
			ShardHeight:             actionData.ShardHeight,
			BeaconHeight:            beaconHeight + 1,
			RedeemerExternalAddress: actionData.RedeemerExternalAddress,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalRedeemRequestByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}

		// update bridge/portal token info
		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
			return nil
		}
		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
		if found {
			updatingInfo.DeductAmt += actionData.RedeemAmount
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:      0,
				DeductAmt:       actionData.RedeemAmount,
				TokenID:         *incTokenID,
				ExternalTokenID: nil,
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo

	} else if reqStatus == pCommon.PortalRequestRejectedChainStatus {
		// track status of redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalRedeemRequestStatus{
			Status:                  pCommon.PortalRequestRejectedStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 actionData.TxReqID,
			ShardID:                 actionData.ShardID,
			ShardHeight:             actionData.ShardHeight,
			BeaconHeight:            beaconHeight + 1,
			RedeemerExternalAddress: actionData.RedeemerExternalAddress,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalRedeemRequestByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}
	} else if reqStatus == pCommon.PortalRedeemReqCancelledByLiquidationChainStatus {
		keyWaitingRedeemRequest := statedb.GenerateWaitingRedeemRequestObjectKey(actionData.UniqueRedeemID)
		keyWaitingRedeemRequestStr := keyWaitingRedeemRequest.String()
		redeemReq := currentPortalState.WaitingRedeemRequests[keyWaitingRedeemRequestStr]
		if redeemReq == nil {
			Logger.log.Errorf("[processPortalRedeemRequest] redeemReq with ID %v not found: %v\n", actionData.UniqueRedeemID)
			return nil
		}

		// reject waiting redeem request, return ptoken and redeem fee for users
		// update custodian state (return holding public token amount)
		err := UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(currentPortalState, redeemReq, beaconHeight)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when updating custodian state %v - RedeemID %v\n: ",
				err, redeemReq.GetUniqueRedeemID())
			return nil
		}

		// remove redeem request from waiting redeem requests list
		deleteWaitingRedeemRequest(currentPortalState, keyWaitingRedeemRequestStr)
		statedb.DeleteWaitingRedeemRequest(stateDB, redeemReq.GetUniqueRedeemID())

		// update status of redeem request by redeemID to rejected by liquidation
		redeemRequestStatus := metadata.PortalRedeemRequestStatus{
			Status:                  pCommon.PortalRedeemReqCancelledByLiquidationStatus,
			UniqueRedeemID:          actionData.UniqueRedeemID,
			TokenID:                 actionData.TokenID,
			RedeemAmount:            actionData.RedeemAmount,
			RedeemerIncAddressStr:   actionData.RedeemerIncAddressStr,
			RemoteAddress:           actionData.RemoteAddress,
			RedeemFee:               actionData.RedeemFee,
			MatchingCustodianDetail: actionData.MatchingCustodianDetail,
			TxReqID:                 redeemReq.GetTxReqID(),
			ShardID:                 actionData.ShardID,
			ShardHeight:             actionData.ShardHeight,
			BeaconHeight:            redeemReq.GetBeaconHeight(),
			RedeemerExternalAddress: actionData.RedeemerExternalAddress,
		}
		redeemRequestStatusBytes, _ := json.Marshal(redeemRequestStatus)
		err = statedb.StorePortalRedeemRequestStatus(
			stateDB,
			actionData.UniqueRedeemID,
			redeemRequestStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalRedeemRequest] Error when storing status of redeem request by redeemID: %v\n", err)
			return nil
		}

		// update bridge/portal token info
		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
			return nil
		}
		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
		if found {
			updatingInfo.CountUpAmt += actionData.RedeemAmount
		} else {
			updatingInfo = metadata.UpdatingInfo{
				CountUpAmt:      actionData.RedeemAmount,
				DeductAmt:       0,
				TokenID:         *incTokenID,
				ExternalTokenID: nil,
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo
	}

	return nil
}

/* =======
Portal Request Mactching Redeem Processor
======= */

type PortalRequestMatchingRedeemProcessor struct {
	*PortalInstProcessorV3
}

func (p *PortalRequestMatchingRedeemProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalRequestMatchingRedeemProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalRequestMatchingRedeemProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildReqMatchingRedeemInst(
	redeemID string,
	incAddressStr string,
	matchingAmount uint64,
	isFullCustodian bool,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	reqMatchingRedeemContent := metadata.PortalReqMatchingRedeemContent{
		CustodianAddressStr: incAddressStr,
		RedeemID:            redeemID,
		MatchingAmount:      matchingAmount,
		IsFullCustodian:     isFullCustodian,
		TxReqID:             txReqID,
		ShardID:             shardID,
	}
	reqMatchingRedeemContentBytes, _ := json.Marshal(reqMatchingRedeemContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(reqMatchingRedeemContentBytes),
	}
}

func (p *PortalRequestMatchingRedeemProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal req matching redeem action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalReqMatchingRedeemAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal req matching redeem action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildReqMatchingRedeemInst(
		meta.RedeemID,
		meta.CustodianAddressStr,
		0,
		false,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForRedeemRequest]: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	redeemID := meta.RedeemID
	amountMatching, isEnoughCustodians, err := MatchCustodianToWaitingRedeemReq(meta.CustodianAddressStr, redeemID, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when processing matching custodian and redeemID %v\n", err)
		return [][]string{rejectInst}, nil
	}

	// if custodian is valid to matching, append custodian in matchingCustodians of redeem request
	// update custodian state
	_, err = UpdatePortalStateAfterCustodianReqMatchingRedeem(meta.CustodianAddressStr, redeemID, amountMatching, isEnoughCustodians, currentPortalState)
	if err != nil {
		Logger.log.Errorf("Error when updating portal state %v\n", err)
		return [][]string{rejectInst}, nil
	}

	inst := buildReqMatchingRedeemInst(
		meta.RedeemID,
		meta.CustodianAddressStr,
		amountMatching,
		isEnoughCustodians,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
	)
	return [][]string{inst}, nil
}

func (p *PortalRequestMatchingRedeemProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData metadata.PortalReqMatchingRedeemContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == pCommon.PortalRequestAcceptedChainStatus {
		updatedRedeemRequest, err := UpdatePortalStateAfterCustodianReqMatchingRedeem(
			actionData.CustodianAddressStr,
			actionData.RedeemID,
			actionData.MatchingAmount,
			actionData.IsFullCustodian,
			currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating portal state of request matching redeem request %v", err)
			return nil
		}

		newStatus := pCommon.PortalRedeemReqWaitingStatus
		if actionData.IsFullCustodian {
			statedb.DeleteWaitingRedeemRequest(stateDB, actionData.RedeemID)
			newStatus = pCommon.PortalRedeemReqMatchedStatus
		}

		// update status of redeem ID by redeemID and matching custodians
		redeemRequest := metadata.PortalRedeemRequestStatus{
			Status:                  byte(newStatus),
			UniqueRedeemID:          updatedRedeemRequest.GetUniqueRedeemID(),
			TokenID:                 updatedRedeemRequest.GetTokenID(),
			RedeemAmount:            updatedRedeemRequest.GetRedeemAmount(),
			RedeemerIncAddressStr:   updatedRedeemRequest.GetRedeemerAddress(),
			RemoteAddress:           updatedRedeemRequest.GetRedeemerRemoteAddress(),
			RedeemFee:               updatedRedeemRequest.GetRedeemFee(),
			MatchingCustodianDetail: updatedRedeemRequest.GetCustodians(),
			TxReqID:                 updatedRedeemRequest.GetTxReqID(),
			ShardID:                 updatedRedeemRequest.ShardID(),
			ShardHeight:             updatedRedeemRequest.ShardHeight(),
			BeaconHeight:            updatedRedeemRequest.GetBeaconHeight(),
			RedeemerExternalAddress: updatedRedeemRequest.GetRedeemerExternalAddress(),
		}
		newRedeemRequest, err := json.Marshal(redeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when marshaling status of redeem request %v", err)
			return nil
		}
		err = statedb.StorePortalRedeemRequestStatus(stateDB, actionData.RedeemID, newRedeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when storing status of redeem request %v", err)
			return err
		}

		// track status of req matching redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalReqMatchingRedeemStatus{
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemID:            actionData.RedeemID,
			MatchingAmount:      actionData.MatchingAmount,
			Status:              pCommon.PortalRequestAcceptedStatus,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalReqMatchingRedeemByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalReqMatchingRedeem] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}

	} else if reqStatus == pCommon.PortalRequestRejectedChainStatus {
		// track status of req matching redeem request by txReqID
		redeemRequestByTxIDStatus := metadata.PortalReqMatchingRedeemStatus{
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemID:            actionData.RedeemID,
			MatchingAmount:      actionData.MatchingAmount,
			Status:              pCommon.PortalRequestRejectedStatus,
		}
		redeemRequestByTxIDStatusBytes, _ := json.Marshal(redeemRequestByTxIDStatus)
		err = statedb.StorePortalReqMatchingRedeemByTxIDStatus(
			stateDB, actionData.TxReqID.String(), redeemRequestByTxIDStatusBytes)
		if err != nil {
			Logger.log.Errorf("[processPortalReqMatchingRedeem] Error when tracking status of redeem request by txReqID: %v\n", err)
			return nil
		}

	}
	return nil
}

//todo:

// PortalPickMoreCustodiansForRedeemReqContent - Beacon builds a new instruction with this content after timeout of redeem request
// It will be appended to beaconBlock
type PortalPickMoreCustodiansForRedeemReqContent struct {
	RedeemID   string
	Custodians []*statedb.MatchingRedeemCustodianDetail
}

func buildInstPickingCustodiansForTimeoutWaitingRedeem(
	redeemID string,
	custodians []*statedb.MatchingRedeemCustodianDetail,
	metaType int,
	status string) []string {
	pickMoreCustodians := PortalPickMoreCustodiansForRedeemReqContent{
		RedeemID:   redeemID,
		Custodians: custodians,
	}
	pickMoreCustodiansBytes, _ := json.Marshal(pickMoreCustodians)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(-1),
		status,
		string(pickMoreCustodiansBytes),
	}
}

// checkAndPickMoreCustodianForWaitingRedeemRequest check waiting redeem requests get timeout or not
// if the waiting redeem request gets timeout, but not enough matching custodians, auto pick up more custodians
func CheckAndPickMoreCustodianForWaitingRedeemRequest(
	bc metadata.ChainRetriever,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams) ([][]string, error) {
	insts := [][]string{}
	waitingRedeemKeys := []string{}
	for key := range currentPortalState.WaitingRedeemRequests {
		waitingRedeemKeys = append(waitingRedeemKeys, key)
	}
	sort.Strings(waitingRedeemKeys)
	for _, waitingRedeemKey := range waitingRedeemKeys {
		waitingRedeem := currentPortalState.WaitingRedeemRequests[waitingRedeemKey]
		if !bc.CheckBlockTimeIsReached(beaconHeight, waitingRedeem.GetBeaconHeight(), shardHeights[waitingRedeem.ShardID()], waitingRedeem.ShardHeight(), portalParams.TimeOutWaitingRedeemRequest) {
			continue
		}

		// calculate amount need to be matched
		totalMatchedAmount := uint64(0)
		for _, cus := range waitingRedeem.GetCustodians() {
			totalMatchedAmount += cus.GetAmount()
		}
		neededMatchingAmountInPToken := waitingRedeem.GetRedeemAmount() - totalMatchedAmount
		if neededMatchingAmountInPToken > waitingRedeem.GetRedeemAmount() || neededMatchingAmountInPToken <= 0 {
			Logger.log.Errorf("Amount need to matching in redeem request %v is less than zero", neededMatchingAmountInPToken)
			continue
		}

		// pick up more custodians
		moreCustodians, err := pickupCustodianForRedeem(neededMatchingAmountInPToken, waitingRedeem.GetTokenID(), currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when pick up more custodians for timeout redeem request %v\n", err)
			inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
				waitingRedeem.GetUniqueRedeemID(),
				moreCustodians,
				metadata.PortalPickMoreCustodianForRedeemMeta,
				pCommon.PortalProducerInstFailedChainStatus,
			)
			insts = append(insts, inst)

			// build instruction reject redeem request
			err := UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(currentPortalState, waitingRedeem, beaconHeight)
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when updating custodian state %v - RedeemID %v\n: ",
					err, waitingRedeem.GetUniqueRedeemID())
				break
			}

			// remove redeem request from waiting redeem requests list
			deleteWaitingRedeemRequest(currentPortalState, waitingRedeemKey)

			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(waitingRedeem.GetRedeemerAddress())
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					waitingRedeem.GetUniqueRedeemID(), err)
				break
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// build instruction
			inst2 := buildRedeemRequestInst(
				waitingRedeem.GetUniqueRedeemID(),
				waitingRedeem.GetTokenID(),
				waitingRedeem.GetRedeemAmount(),
				waitingRedeem.GetRedeemerAddress(),
				waitingRedeem.GetRedeemerRemoteAddress(),
				waitingRedeem.GetRedeemFee(),
				waitingRedeem.GetCustodians(),
				metadata.PortalRedeemRequestMetaV3,
				shardID,
				common.Hash{},
				pCommon.PortalRedeemReqCancelledByLiquidationChainStatus,
				waitingRedeem.ShardHeight(),
				waitingRedeem.GetRedeemerExternalAddress(),
			)
			insts = append(insts, inst2)
			continue
		}

		// update custodian state (holding public tokens)
		_, err = UpdatePortalStateAfterPickMoreCustodiansForWaitingRedeemReq(
			moreCustodians, waitingRedeem, currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating portal state %v\n", err)
			inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
				waitingRedeem.GetUniqueRedeemID(),
				moreCustodians,
				metadata.PortalPickMoreCustodianForRedeemMeta,
				pCommon.PortalProducerInstFailedChainStatus,
			)
			insts = append(insts, inst)
			continue
		}

		inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
			waitingRedeem.GetUniqueRedeemID(),
			moreCustodians,
			metadata.PortalPickMoreCustodianForRedeemMeta,
			pCommon.PortalProducerInstSuccessChainStatus,
		)
		insts = append(insts, inst)
	}

	return insts, nil
}

/* =======
Portal Pickup more Custodian For Waiting Redeem Requests Processor
======= */

type PortalPickMoreCustodianForRedeemProcessor struct {
	*PortalInstProcessorV3
}

func (p *PortalPickMoreCustodianForRedeemProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalPickMoreCustodianForRedeemProcessor) PutAction(action []string, shardID byte) {
	// @NOTE: do nothing, because beacon auto check and pick custodians, have no action from shard blocks
	//_, found := p.actions[shardID]
	//if !found {
	//	p.actions[shardID] = [][]string{action}
	//} else {
	//	p.actions[shardID] = append(p.actions[shardID], action)
	//}
}

func (p *PortalPickMoreCustodianForRedeemProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildPortalPickMoreCustodianForRedeemInst(
	redeemID string,
	custodians []*statedb.MatchingRedeemCustodianDetail,
	metaType int,
	status string,
) []string {
	pickMoreCustodians := PortalPickMoreCustodiansForRedeemReqContent{
		RedeemID:   redeemID,
		Custodians: custodians,
	}
	pickMoreCustodiansBytes, _ := json.Marshal(pickMoreCustodians)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(-1),
		status,
		string(pickMoreCustodiansBytes),
	}
}

func (p *PortalPickMoreCustodianForRedeemProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	insts := [][]string{}
	waitingRedeemKeys := []string{}
	for key := range currentPortalState.WaitingRedeemRequests {
		waitingRedeemKeys = append(waitingRedeemKeys, key)
	}
	sort.Strings(waitingRedeemKeys)
	for _, waitingRedeemKey := range waitingRedeemKeys {
		waitingRedeem := currentPortalState.WaitingRedeemRequests[waitingRedeemKey]
		if !bc.CheckBlockTimeIsReached(beaconHeight, waitingRedeem.GetBeaconHeight(), shardHeights[waitingRedeem.ShardID()], waitingRedeem.ShardHeight(), portalParams.TimeOutWaitingRedeemRequest) {
			continue
		}

		// calculate amount need to be matched
		totalMatchedAmount := uint64(0)
		for _, cus := range waitingRedeem.GetCustodians() {
			totalMatchedAmount += cus.GetAmount()
		}
		neededMatchingAmountInPToken := waitingRedeem.GetRedeemAmount() - totalMatchedAmount
		if neededMatchingAmountInPToken > waitingRedeem.GetRedeemAmount() || neededMatchingAmountInPToken <= 0 {
			Logger.log.Errorf("Amount need to matching in redeem request %v is less than zero", neededMatchingAmountInPToken)
			continue
		}

		// pick up more custodians
		moreCustodians, err := pickupCustodianForRedeem(neededMatchingAmountInPToken, waitingRedeem.GetTokenID(), currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when pick up more custodians for timeout redeem request %v\n", err)
			inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
				waitingRedeem.GetUniqueRedeemID(),
				moreCustodians,
				metadata.PortalPickMoreCustodianForRedeemMeta,
				pCommon.PortalProducerInstFailedChainStatus,
			)
			insts = append(insts, inst)

			// build instruction reject redeem request
			err := UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(currentPortalState, waitingRedeem, beaconHeight)
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when updating custodian state %v - RedeemID %v\n: ",
					err, waitingRedeem.GetUniqueRedeemID())
				break
			}

			// remove redeem request from waiting redeem requests list
			deleteWaitingRedeemRequest(currentPortalState, waitingRedeemKey)

			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(waitingRedeem.GetRedeemerAddress())
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					waitingRedeem.GetUniqueRedeemID(), err)
				break
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// build instruction
			inst2 := buildRedeemRequestInst(
				waitingRedeem.GetUniqueRedeemID(),
				waitingRedeem.GetTokenID(),
				waitingRedeem.GetRedeemAmount(),
				waitingRedeem.GetRedeemerAddress(),
				waitingRedeem.GetRedeemerRemoteAddress(),
				waitingRedeem.GetRedeemFee(),
				waitingRedeem.GetCustodians(),
				metadata.PortalRedeemRequestMetaV3,
				shardID,
				common.Hash{},
				pCommon.PortalRedeemReqCancelledByLiquidationChainStatus,
				waitingRedeem.ShardHeight(),
				waitingRedeem.GetRedeemerExternalAddress(),
			)
			insts = append(insts, inst2)
			continue
		}

		// update custodian state (holding public tokens)
		_, err = UpdatePortalStateAfterPickMoreCustodiansForWaitingRedeemReq(
			moreCustodians, waitingRedeem, currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating portal state %v\n", err)
			inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
				waitingRedeem.GetUniqueRedeemID(),
				moreCustodians,
				metadata.PortalPickMoreCustodianForRedeemMeta,
				pCommon.PortalProducerInstFailedChainStatus,
			)
			insts = append(insts, inst)
			continue
		}

		inst := buildInstPickingCustodiansForTimeoutWaitingRedeem(
			waitingRedeem.GetUniqueRedeemID(),
			moreCustodians,
			metadata.PortalPickMoreCustodianForRedeemMeta,
			pCommon.PortalProducerInstSuccessChainStatus,
		)
		insts = append(insts, inst)
	}

	return insts, nil
}

func (p *PortalPickMoreCustodianForRedeemProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData PortalPickMoreCustodiansForRedeemReqContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == pCommon.PortalProducerInstSuccessChainStatus {
		waitingRedeemKey := statedb.GenerateWaitingRedeemRequestObjectKey(actionData.RedeemID).String()
		waitingRedeem := currentPortalState.WaitingRedeemRequests[waitingRedeemKey]
		updatedRedeemRequest, err := UpdatePortalStateAfterPickMoreCustodiansForWaitingRedeemReq(
			actionData.Custodians,
			waitingRedeem,
			currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating portal state of request matching redeem request %v", err)
			return nil
		}
		// delete waiting redeem request
		statedb.DeleteWaitingRedeemRequest(stateDB, actionData.RedeemID)

		// update status of redeem ID by redeemID and matching custodians
		newStatus := pCommon.PortalRedeemReqMatchedStatus
		redeemRequest := metadata.PortalRedeemRequestStatus{
			Status:                  byte(newStatus),
			UniqueRedeemID:          updatedRedeemRequest.GetUniqueRedeemID(),
			TokenID:                 updatedRedeemRequest.GetTokenID(),
			RedeemAmount:            updatedRedeemRequest.GetRedeemAmount(),
			RedeemerIncAddressStr:   updatedRedeemRequest.GetRedeemerAddress(),
			RemoteAddress:           updatedRedeemRequest.GetRedeemerRemoteAddress(),
			RedeemFee:               updatedRedeemRequest.GetRedeemFee(),
			MatchingCustodianDetail: updatedRedeemRequest.GetCustodians(),
			TxReqID:                 updatedRedeemRequest.GetTxReqID(),
			ShardID:                 updatedRedeemRequest.ShardID(),
			ShardHeight:             updatedRedeemRequest.ShardHeight(),
			BeaconHeight:            updatedRedeemRequest.GetBeaconHeight(),
			RedeemerExternalAddress: updatedRedeemRequest.GetRedeemerExternalAddress(),
		}
		newRedeemRequest, err := json.Marshal(redeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when marshaling status of redeem request %v", err)
			return nil
		}
		err = statedb.StorePortalRedeemRequestStatus(stateDB, actionData.RedeemID, newRedeemRequest)
		if err != nil {
			Logger.log.Errorf("Error when storing status of redeem request %v", err)
			return err
		}
	}
	return nil
}

/* =======
Portal Request Unlock Collateral Processor
======= */

type PortalRequestUnlockCollateralProcessor struct {
	*PortalInstProcessorV3
}

func (p *PortalRequestUnlockCollateralProcessor) GetActions() map[byte][][]string {
	return p.Actions
}

func (p *PortalRequestUnlockCollateralProcessor) PutAction(action []string, shardID byte) {
	_, found := p.Actions[shardID]
	if !found {
		p.Actions[shardID] = [][]string{action}
	} else {
		p.Actions[shardID] = append(p.Actions[shardID], action)
	}
}

func (p *PortalRequestUnlockCollateralProcessor) PrepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildReqUnlockCollateralInst(
	uniqueRedeemID string,
	tokenID string,
	custodianAddressStr string,
	redeemAmount uint64,
	unlockAmount uint64,
	redeemProof string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	reqUnlockCollateralContent := metadata.PortalRequestUnlockCollateralContent{
		UniqueRedeemID:      uniqueRedeemID,
		TokenID:             tokenID,
		CustodianAddressStr: custodianAddressStr,
		RedeemAmount:        redeemAmount,
		UnlockAmount:        unlockAmount,
		RedeemProof:         redeemProof,
		TxReqID:             txReqID,
		ShardID:             shardID,
	}
	reqUnlockCollateralContentBytes, _ := json.Marshal(reqUnlockCollateralContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(reqUnlockCollateralContentBytes),
	}
}

func (p *PortalRequestUnlockCollateralProcessor) BuildNewInsts(
	bc metadata.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portalv3.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal request unlock collateral action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRequestUnlockCollateralAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal request unlock collateral action: %+v", err)
		return [][]string{}, nil
	}
	meta := actionData.Meta

	rejectInst := buildReqUnlockCollateralInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.CustodianAddressStr,
		meta.RedeemAmount,
		0,
		meta.RedeemProof,
		meta.Type,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("WARN - [buildInstructionsForReqUnlockCollateral]: Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	// check meta.UniqueRedeemID is in matched RedeemRequests list in portal state or not
	redeemID := meta.UniqueRedeemID
	keyMatchedRedeemRequestStr := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID).String()
	matchedRedeemRequest := currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr]
	if matchedRedeemRequest == nil {
		Logger.log.Errorf("redeemID is not existed in matched redeem requests list")
		return [][]string{rejectInst}, nil
	}

	// check tokenID
	if meta.TokenID != matchedRedeemRequest.GetTokenID() {
		Logger.log.Errorf("TokenID is not correct in redeemID req")
		return [][]string{rejectInst}, nil
	}

	// check redeem amount of matching custodian
	matchedCustodian := new(statedb.MatchingRedeemCustodianDetail)
	for _, cus := range matchedRedeemRequest.GetCustodians() {
		if cus.GetIncognitoAddress() == meta.CustodianAddressStr {
			matchedCustodian = cus
			break
		}
	}
	if matchedCustodian.GetIncognitoAddress() == "" {
		Logger.log.Errorf("Custodian address %v is not in redeemID req %v", meta.CustodianAddressStr, meta.UniqueRedeemID)
		return [][]string{rejectInst}, nil
	}

	if meta.RedeemAmount != matchedCustodian.GetAmount() {
		Logger.log.Errorf("RedeemAmount is not correct in redeemID req")
		return [][]string{rejectInst}, nil
	}

	portalTokenProcessor := portalParams.PortalTokens[meta.TokenID]
	if portalTokenProcessor == nil {
		Logger.log.Errorf("TokenID %v is not supported currently on Portal", meta.TokenID)
		return [][]string{rejectInst}, nil
	}

	expectedMemo := portalTokenProcessor.GetExpectedMemoForRedeem(meta.UniqueRedeemID, meta.CustodianAddressStr)
	expectedPaymentInfos := map[string]uint64{
		matchedRedeemRequest.GetRedeemerRemoteAddress(): matchedCustodian.GetAmount(),
	}
	isValid, err := portalTokenProcessor.ParseAndVerifyProof(meta.RedeemProof, bc, expectedMemo, expectedPaymentInfos)
	if !isValid || err != nil {
		Logger.log.Error("Parse and verify redeem proof failed: %v", err)
		return [][]string{rejectInst}, nil
	}

	// calculate unlock amount
	custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.CustodianAddressStr)
	var unlockAmount uint64
	custodianStateKeyStr := custodianStateKey.String()
	if meta.Type == metadata.PortalRequestUnlockCollateralMeta {
		unlockAmount, err = CalUnlockCollateralAmount(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian %v", err)
			return [][]string{rejectInst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		// unlock amount in prv
		err = updateCustodianStateAfterReqUnlockCollateral(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID)
		if err != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err)
			return [][]string{rejectInst}, nil
		}
	} else {
		unlockAmount, _, err = CalUnlockCollateralAmountV3(currentPortalState, custodianStateKeyStr, meta.RedeemAmount, meta.TokenID, portalParams)
		if err != nil {
			Logger.log.Errorf("Error calculating unlock amount for custodian V3 %v", err)
			return [][]string{rejectInst}, nil
		}

		// update custodian state (FreeCollateral, LockedAmountCollateral)
		// unlock amount in usdt
		_, err = updateCustodianStateAfterReqUnlockCollateralV3(
			currentPortalState.CustodianPoolState[custodianStateKeyStr],
			unlockAmount, meta.TokenID, portalParams, currentPortalState)
		if err != nil {
			Logger.log.Errorf("Error when updating custodian state after unlocking collateral %v", err)
			return [][]string{rejectInst}, nil
		}
	}

	// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
	updatedCustodians, err := removeCustodianFromMatchingRedeemCustodians(
		currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians(), meta.CustodianAddressStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while removing custodian %v from matching custodians", meta.CustodianAddressStr)
		return [][]string{rejectInst}, nil
	}
	currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].SetCustodians(updatedCustodians)

	// remove redeem request from WaitingRedeemRequest list when all matching custodians return public token to user
	// when list matchingCustodianDetail is empty
	if len(currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians()) == 0 {
		deleteMatchedRedeemRequest(currentPortalState, keyMatchedRedeemRequestStr)
	}

	inst := buildReqUnlockCollateralInst(
		meta.UniqueRedeemID,
		meta.TokenID,
		meta.CustodianAddressStr,
		meta.RedeemAmount,
		unlockAmount,
		meta.RedeemProof,
		meta.Type,
		shardID,
		actionData.TxReqID,
		pCommon.PortalRequestAcceptedChainStatus,
	)

	return [][]string{inst}, nil
}

func (p *PortalRequestUnlockCollateralProcessor) ProcessInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portalv3.PortalParams,
	updatingInfoByTokenID map[common.Hash]metadata.UpdatingInfo,
) error {

	// unmarshal instructions content
	var actionData metadata.PortalRequestUnlockCollateralContent
	var err error
	err = json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	// get tokenID from redeemTokenID
	tokenID := actionData.TokenID
	reqStatus := instructions[2]
	if reqStatus == pCommon.PortalRequestAcceptedChainStatus {
		// update custodian state (FreeCollateral, LockedAmountCollateral)
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		// portal unlock collateral v2 and v3
		if instructions[0] == strconv.Itoa(metadata.PortalRequestUnlockCollateralMeta) {
			err = updateCustodianStateAfterReqUnlockCollateral(
				currentPortalState.CustodianPoolState[custodianStateKeyStr],
				actionData.UnlockAmount, tokenID)
		} else {
			_, err = updateCustodianStateAfterReqUnlockCollateralV3(
				currentPortalState.CustodianPoolState[custodianStateKeyStr],
				actionData.UnlockAmount, tokenID, portalParams, currentPortalState)
		}
		if err != nil {
			Logger.log.Errorf("Error when update custodian state", err)
			return nil
		}

		redeemID := actionData.UniqueRedeemID
		keyMatchedRedeemRequest := statedb.GenerateMatchedRedeemRequestObjectKey(redeemID)
		keyMatchedRedeemRequestStr := keyMatchedRedeemRequest.String()

		// update redeem request state in WaitingRedeemRequest (remove custodian from matchingCustodianDetail)
		newCustodians, err := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians(), actionData.CustodianAddressStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while removing custodian %v from matching custodians", actionData.CustodianAddressStr)
			return nil
		}
		currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].SetCustodians(newCustodians)

		// remove redeem request from WaitingRedeemRequest list when all matching custodians return public token to user
		// when list matchingCustodianDetail is empty
		if len(currentPortalState.MatchedRedeemRequests[keyMatchedRedeemRequestStr].GetCustodians()) == 0 {
			deleteMatchedRedeemRequest(currentPortalState, keyMatchedRedeemRequestStr)
			statedb.DeleteMatchedRedeemRequest(stateDB, actionData.UniqueRedeemID)

			// update status of redeem request with redeemID
			err = updateRedeemRequestStatusByRedeemId(redeemID, pCommon.PortalRedeemReqSuccessStatus, stateDB)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
				return nil
			}
		}

		// track reqUnlockCollateral status by txID into DB
		reqUnlockCollateralTrackData := metadata.PortalRequestUnlockCollateralStatus{
			Status:              pCommon.PortalRequestAcceptedStatus,
			UniqueRedeemID:      actionData.UniqueRedeemID,
			TokenID:             actionData.TokenID,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemAmount:        actionData.RedeemAmount,
			UnlockAmount:        actionData.UnlockAmount,
			RedeemProof:         actionData.RedeemProof,
			TxReqID:             actionData.TxReqID,
		}
		reqUnlockCollateralTrackDataBytes, _ := json.Marshal(reqUnlockCollateralTrackData)
		err = statedb.StorePortalRequestUnlockCollateralStatus(
			stateDB,
			actionData.TxReqID.String(),
			reqUnlockCollateralTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking request unlock collateral tx: %+v", err)
			return nil
		}

	} else if reqStatus == pCommon.PortalRequestRejectedChainStatus {
		// track reqUnlockCollateral status by txID into DB
		reqUnlockCollateralTrackData := metadata.PortalRequestUnlockCollateralStatus{
			Status:              pCommon.PortalRequestRejectedStatus,
			UniqueRedeemID:      actionData.UniqueRedeemID,
			TokenID:             actionData.TokenID,
			CustodianAddressStr: actionData.CustodianAddressStr,
			RedeemAmount:        actionData.RedeemAmount,
			UnlockAmount:        actionData.UnlockAmount,
			RedeemProof:         actionData.RedeemProof,
			TxReqID:             actionData.TxReqID,
		}
		reqUnlockCollateralTrackDataBytes, _ := json.Marshal(reqUnlockCollateralTrackData)
		err = statedb.StorePortalRequestUnlockCollateralStatus(
			stateDB,
			actionData.TxReqID.String(),
			reqUnlockCollateralTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking request unlock collateral tx: %+v", err)
			return nil
		}
	}

	return nil
}