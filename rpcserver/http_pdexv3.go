package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (httpServer *HttpServer) handleGetPdexv3State(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Beacon height is invalid"))
	}
	filter, ok := data["Filter"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Filter is invalid"))
	}
	result, err := httpServer.blockService.GetPdexv3State(filter, uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	return result, nil
}

/*
	Params Modifying
*/

func (httpServer *HttpServer) handleCreateAndSendTxWithPdexv3ModifyParams(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPdexv3ModifyParams(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithPdexv3ModifyParams(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	newParams, ok := tokenParamsRaw["NewParams"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NewParams is invalid"))
	}

	defaultFeeRateBPS, err := common.AssertAndConvertStrToNumber(newParams["DefaultFeeRateBPS"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("DefaultFeeRateBPS is invalid"))
	}

	feeRateBPSTemp, ok := newParams["FeeRateBPS"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("FeeRateBPS is invalid"))
	}
	feeRateBPS := map[string]uint{}
	for key, feeRatePool := range feeRateBPSTemp {
		value, err := common.AssertAndConvertStrToNumber(feeRatePool)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("FeeRateBPS is invalid"))
		}
		feeRateBPS[key] = uint(value)
	}

	prvDiscountPercent, err := common.AssertAndConvertStrToNumber(newParams["PRVDiscountPercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PRVDiscountPercent is invalid"))
	}

	tradingProtocolFeePercent, err := common.AssertAndConvertStrToNumber(newParams["TradingProtocolFeePercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("TradingProtocolFeePercent is invalid"))
	}

	tradingStakingPoolRewardPercent, err := common.AssertAndConvertStrToNumber(newParams["TradingStakingPoolRewardPercent"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("TradingStakingPoolRewardPercent is invalid"))
	}

	pdexRewardPoolPairsShareTemp, ok := newParams["PDEXRewardPoolPairsShare"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PDEXRewardPoolPairsShare is invalid"))
	}
	pdexRewardPoolPairsShare := map[string]uint{}
	for key, share := range pdexRewardPoolPairsShareTemp {
		value, err := common.AssertAndConvertStrToNumber(share)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PDEXRewardPoolPairsShare is invalid"))
		}
		pdexRewardPoolPairsShare[key] = uint(value)
	}

	stakingPoolsShareTemp, ok := newParams["StakingPoolsShare"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingPoolsShare is invalid"))
	}
	stakingPoolsShare := map[string]uint{}
	for key, share := range stakingPoolsShareTemp {
		value, err := common.AssertAndConvertStrToNumber(share)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingPoolsShare is invalid"))
		}
		stakingPoolsShare[key] = uint(value)
	}

	stakingRewardTokensRaw, ok := newParams["StakingRewardTokens"].([]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingRewardTokens is invalid"))
	}
	stakingRewardTokens := []common.Hash{}
	for _, tokenIDRaw := range stakingRewardTokensRaw {
		tokenIDStr, ok := tokenIDRaw.(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingRewardTokens is invalid"))
		}
		tokenID, err := new(common.Hash).NewHashFromStr(tokenIDStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Token %v of StakingRewardTokens is invalid", tokenIDStr))
		}
		stakingRewardTokens = append(stakingRewardTokens, *tokenID)
	}

	mintNftRequireAmount, err := common.AssertAndConvertStrToNumber(newParams["MintNftRequireAmount"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("MintNftRequireAmount is invalid"))
	}

	maxOrdersPerNft, err := common.AssertAndConvertStrToNumber(newParams["MaxOrdersPerNft"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("MaxOrdersPerNft is invalid"))
	}
	autoWithdrawOrderLimitAmount, err := common.AssertAndConvertStrToNumber(newParams["AutoWithdrawOrderLimitAmount"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("AutoWithdrawOrderLimitAmount is invalid"))
	}

	minPRVReserveTradingRate, err := common.AssertAndConvertStrToNumber(newParams["MinPRVReserveTradingRate"])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("MinPRVReserveTradingRate is invalid"))
	}

	orderMiningRewardRatioBPSTmp, ok := newParams["OrderMiningRewardRatioBPS"].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("OrderMiningRewardRatioBPS is invalid"))
	}
	orderMiningRewardRatioBPS := map[string]uint{}
	for key, share := range orderMiningRewardRatioBPSTmp {
		value, err := common.AssertAndConvertStrToNumber(share)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("OrderMiningRewardRatioBPS is invalid"))
		}
		orderMiningRewardRatioBPS[key] = uint(value)
	}

	meta, err := metadataPdexv3.NewPdexv3ParamsModifyingRequest(
		metadataCommon.Pdexv3ModifyParamsMeta,
		metadataPdexv3.Pdexv3Params{
			DefaultFeeRateBPS:               uint(defaultFeeRateBPS),
			FeeRateBPS:                      feeRateBPS,
			PRVDiscountPercent:              uint(prvDiscountPercent),
			TradingProtocolFeePercent:       uint(tradingProtocolFeePercent),
			TradingStakingPoolRewardPercent: uint(tradingStakingPoolRewardPercent),
			PDEXRewardPoolPairsShare:        pdexRewardPoolPairsShare,
			StakingPoolsShare:               stakingPoolsShare,
			StakingRewardTokens:             stakingRewardTokens,
			MintNftRequireAmount:            mintNftRequireAmount,
			MaxOrdersPerNft:                 uint(maxOrdersPerNft),
			AutoWithdrawOrderLimitAmount:    uint(autoWithdrawOrderLimitAmount),
			MinPRVReserveTradingRate:        minPRVReserveTradingRate,
			OrderMiningRewardRatioBPS:       orderMiningRewardRatioBPS,
		},
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleGetPdexv3ParamsModifyingRequestStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPdexv3ParamsModifyingRequestStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3ParamsModyfingStatusError, err)
	}
	return status, nil
}

/*
	Fee Management
*/
func (httpServer *HttpServer) handleGetPdexv3EstimatedLPValue(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	pairID, ok := data["PoolPairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PairID is invalid"))
	}
	nftIDStr, ok := data["NftID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID is invalid"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(nftIDStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	beaconBestView := httpServer.config.BlockChain.GetBeaconBestState()
	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok || beaconHeight == 0 {
		beaconHeight = float64(beaconBestView.BeaconHeight)
	}

	beaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, err)
	}

	if uint64(beaconHeight) < config.Param().PDexParams.Pdexv3BreakPointHeight {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, errors.New("pDEX v3 is not available"))
	}

	pDexv3State, err := pdex.InitStateFromDB(beaconFeatureStateDB, uint64(beaconHeight), pdex.AmplifierVersion)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, err)
	}

	poolPairs := make(map[string]*pdex.PoolPairState)
	err = json.Unmarshal(pDexv3State.Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	if _, ok := poolPairs[pairID]; !ok {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, errors.New("PairID is not existed"))
	}

	pair := poolPairs[pairID]
	pairState := pair.State()

	result := jsonresult.Pdexv3LPValue{
		PoolValue:  map[string]uint64{},
		TradingFee: map[string]uint64{},
	}

	uncollectedLPReward := map[common.Hash]uint64{}
	uncollectedOrderReward := map[common.Hash]uint64{}

	share, ok := pair.Shares()[nftIDStr]
	if ok {
		shareAmount := share.Amount()
		if shareAmount != 0 {
			poolAmount0 := new(big.Int).Mul(
				new(big.Int).SetUint64(pairState.Token0RealAmount()),
				new(big.Int).SetUint64(shareAmount),
			)
			poolAmount0 = new(big.Int).Div(
				poolAmount0,
				new(big.Int).SetUint64(pairState.ShareAmount()),
			)
			if !poolAmount0.IsUint64() {
				return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, errors.New("Could not get pool amount"))
			}

			poolAmount1 := new(big.Int).Mul(
				new(big.Int).SetUint64(pairState.Token1RealAmount()),
				new(big.Int).SetUint64(shareAmount),
			)
			poolAmount1 = new(big.Int).Div(
				poolAmount1,
				new(big.Int).SetUint64(pairState.ShareAmount()),
			)
			if !poolAmount0.IsUint64() {
				return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, errors.New("Could not get pool amount"))
			}

			result.PoolValue[pairState.Token0ID().String()] = poolAmount0.Uint64()
			result.PoolValue[pairState.Token1ID().String()] = poolAmount1.Uint64()
		}

		uncollectedLPReward, err = pair.RecomputeLPFee(*nftID)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, err)
		}
	}

	order, ok := pair.OrderRewards()[nftIDStr]
	if ok {
		// compute amount of received LOP reward
		uncollectedOrderReward = order.UncollectedRewards()
	}

	reward := pdex.CombineReward(uncollectedLPReward, uncollectedOrderReward)

	for tokenID, amount := range reward {
		result.TradingFee[tokenID.String()] = amount
	}

	return result, nil
}

func (httpServer *HttpServer) handleGetPdexv3EstimatedLPPoolReward(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	pairID, ok := data["PoolPairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PairID is invalid"))
	}
	beaconBestView := httpServer.config.BlockChain.GetBeaconBestState()
	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok || beaconHeight == 0 {
		beaconHeight = float64(beaconBestView.BeaconHeight)
	}

	if uint64(beaconHeight) < config.Param().PDexParams.Pdexv3BreakPointHeight {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, errors.New("pDEX v3 is not available"))
	}

	beaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	stateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, err)
	}

	prevBeaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight-1))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight-1, err))
	}
	prevStateDB, err := statedb.NewWithPrefixTrie(prevBeaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, err)
	}

	result, err := httpServer.blockService.GetPdexv3BlockLPReward(
		pairID, uint64(beaconHeight), stateDB, prevStateDB,
	)

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3LPFeeError, err)
	}

	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPdexv3WithdrawLPFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPdexv3WithdrawLPFee(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	// send raw transaction
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithPdexv3WithdrawLPFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) >= 7 {
		hasPrivacyTokenParam, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("HasPrivacyToken is invalid"))
		}
		hasPrivacyToken := int(hasPrivacyTokenParam) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("The privacy mode must be disabled"))
		}
	}
	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	beaconBestView, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	poolPairs := make(map[string]*pdex.PoolPairState)
	err = json.Unmarshal(beaconBestView.PdeState(pdex.AmplifierVersion).Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	pairID, ok := tokenParamsRaw["PoolPairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PairID is invalid"))
	}

	poolPair, found := poolPairs[pairID]
	if !found {
		err = fmt.Errorf("Can't find poolPairID %s", pairID)
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	poolPairState := poolPair.State()

	nftIDStr, ok := tokenParamsRaw["NftID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID is invalid"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(nftIDStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID is invalid"))
	}

	// payment address v2
	feeReceiver, ok := tokenParamsRaw["FeeReceiver"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("FeeReceiver is invalid"))
	}

	tokenIDs := []string{
		poolPairState.Token0ID().String(),
		poolPairState.Token1ID().String(),
		common.PRVIDStr,
		common.PDEXIDStr,
		nftIDStr,
	}

	keyWallet, err := wallet.Base58CheckDeserialize(feeReceiver)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Cannot deserialize payment address: %v", err))
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payment address is invalid"))
	}

	receivers := map[common.Hash]privacy.OTAReceiver{}
	for _, tokenIDStr := range tokenIDs {
		tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("TokenID %v is invalid", tokenIDStr))
		}
		receiver := privacy.OTAReceiver{}
		err = receiver.FromAddress(keyWallet.KeySet.PaymentAddress)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
		}
		receivers[*tokenID] = receiver
	}

	meta, err := metadataPdexv3.NewPdexv3WithdrawalLPFeeRequest(
		metadataCommon.Pdexv3WithdrawLPFeeRequestMeta,
		pairID,
		*nftID,
		receivers,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err2 := json.Marshal(customTokenTx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleGetPdexv3WithdrawalLPFeeStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPdexv3WithdrawalLPFeeStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3WithdrawalLPFeeStatusError, err)
	}
	return status, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPdexv3WithdrawProtocolFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPdexv3WithdrawProtocolFee(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithPdexv3WithdrawProtocolFee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	beaconBestView, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	poolPairs := make(map[string]*pdex.PoolPairState)
	err = json.Unmarshal(beaconBestView.PdeState(pdex.AmplifierVersion).Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	pairID, ok := tokenParamsRaw["PoolPairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("PairID is invalid"))
	}

	_, found := poolPairs[pairID]
	if !found {
		err = fmt.Errorf("Can't find poolPairID %s", pairID)
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	meta, err := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeRequest(
		metadataCommon.Pdexv3WithdrawProtocolFeeRequestMeta,
		pairID,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleGetPdexv3WithdrawalProtocolFeeStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPdexv3WithdrawalProtocolFeeStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3WithdrawalProtocolFeeStatusError, err)
	}
	return status, nil
}

func (httpServer *HttpServer) handleAddLiquidityV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, isPRV, err := httpServer.createPdexv3AddLiquidityTransaction(params)
	if err != nil {
		return nil, err
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, isPRV, closeChan)
}

func (httpServer *HttpServer) createPdexv3AddLiquidityTransaction(params interface{}) (
	*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError,
) {
	arrayParams := common.InterfaceSlice(params)
	isPRV := false
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}

	if len(arrayParams) != 5 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}

	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}

	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverStr, err := otaReceiver.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		NftID             string       `json:"NftID"`
		TokenID           string       `json:"TokenID"`
		PoolPairID        string       `json:"PoolPairID"`
		PairHash          string       `json:"PairHash"`
		ContributedAmount Uint64Reader `json:"ContributedAmount"`
		Amplifier         Uint64Reader `json:"Amplifier"`
	}{}
	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}

	md := metadataPdexv3.NewAddLiquidityRequestWithValue(
		mdReader.PoolPairID, mdReader.PairHash, otaReceiverStr, mdReader.TokenID, mdReader.NftID,
		uint64(mdReader.ContributedAmount), uint(mdReader.Amplifier))
	tokenHash, err := common.Hash{}.NewHashFromStr(mdReader.TokenID)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}
	paramSelect.SetTokenID(*tokenHash)
	isPRV = md.TokenID() == common.PRVIDStr
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for order, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         md.TokenAmount(),
		},
	}
	if isPRV {
		paramSelect.PRV.PaymentInfos = burnPayments
	} else {
		paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
		paramSelect.SetTokenReceivers(burnPayments)
	}

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, isPRV, nil
}

func (httpServer *HttpServer) handleGetPdexv3ContributionStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}

	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3ContributionStatusPrefix(),
		txID.Bytes(),
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handleWithdrawLiquidityV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createPdexv3WithdrawLiquidityTransaction(params)
	if err != nil {
		return nil, err
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err = sendCreatedTransaction(httpServer, newParam, false, closeChan)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (httpServer *HttpServer) createPdexv3WithdrawLiquidityTransaction(
	params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		PoolPairID  string       `json:"PoolPairID"`
		NftID       string       `json:"NftID"`
		ShareAmount Uint64Reader `json:"ShareAmount"`
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters"))
	}

	otaReceivers := make(map[string]string)
	otaReceiverNft := privacy.OTAReceiver{}
	err = otaReceiverNft.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverNftStr, err := otaReceiverNft.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceivers[mdReader.NftID] = otaReceiverNftStr
	otaReceiverToken0 := privacy.OTAReceiver{}
	err = otaReceiverToken0.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverToken0Str, err := otaReceiverToken0.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverToken1 := privacy.OTAReceiver{}
	err = otaReceiverToken1.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverToken1Str, err := otaReceiverToken1.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	beaconBestView, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	poolPairs := make(map[string]*pdex.PoolPairState)
	err = json.Unmarshal(beaconBestView.PdeState(pdex.AmplifierVersion).Reader().PoolPairs(), &poolPairs)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	poolPair, found := poolPairs[mdReader.PoolPairID]
	if !found {
		err = fmt.Errorf("Can't find poolPairID %s", mdReader.PoolPairID)
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	poolPairState := poolPair.State()
	otaReceivers[poolPairState.Token0ID().String()] = otaReceiverToken0Str
	otaReceivers[poolPairState.Token1ID().String()] = otaReceiverToken1Str

	md := metadataPdexv3.NewWithdrawLiquidityRequestWithValue(
		mdReader.PoolPairID, mdReader.NftID, otaReceivers, uint64(mdReader.ShareAmount),
	)
	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	if md.NftID() == common.PRVIDStr {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("Cannot use PRV for withdrawLiquidity tx"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(md.NftID())
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	paramSelect.SetTokenID(*nftID)
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn 1 governance-NFT to withdraw order
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         1,
		},
	}
	paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
	paramSelect.SetTokenReceivers(burnPayments)

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}

func (httpServer *HttpServer) handlePdexv3MintNft(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createPdexv3MintNftTransaction(params)
	if err != nil {
		return nil, err
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)

	res, err = sendCreatedTransaction(httpServer, newParam, true, closeChan)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (httpServer *HttpServer) createPdexv3MintNftTransaction(
	params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 5, len(arrayParams)))
	}

	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiveStr, err := otaReceiver.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	// metadata object format to read from RPC parameters
	mdReader := &struct {
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}
	paramSelect.SetTokenID(common.PRVCoinID)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress
	amount := bc.GetBeaconBestState().PdeState(pdex.AmplifierVersion).Reader().Params().MintNftRequireAmount

	md := metadataPdexv3.NewUserMintNftRequestWithValue(otaReceiveStr, amount)
	paramSelect.SetMetadata(md)

	// burn selling amount for order, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         md.Amount(),
		},
	}
	paramSelect.PRV.PaymentInfos = burnPayments

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetPdexv3WithdrawLiquidityStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawLiquidityStatusPrefix(),
		txID.Bytes(),
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetPdexv3MintNftStatus(
	params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3UserMintNftStatusPrefix(),
		txID.Bytes(),
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

// --- Trade - Order ---

func (httpServer *HttpServer) handlePdexv3TxTradeRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// create tx
	data, isPRV, err := createPdexv3TradeRequestTransaction(httpServer, params)
	if err != nil {
		return nil, err
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, isPRV, closeChan)
}

func (httpServer *HttpServer) handlePdexv3TxAddOrderRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// create tx
	data, isPRV, err := createPdexv3AddOrderRequestTransaction(httpServer, params)
	if err != nil {
		return nil, err
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, isPRV, closeChan)
}

func (httpServer *HttpServer) handlePdexv3TxWithdrawOrderRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// create tx
	data, err := createPdexv3WithdrawOrderRequestTransaction(httpServer, params)
	if err != nil {
		return nil, err
	}
	createTxResult := []interface{}{data.Base58CheckData}
	// send tx
	return sendCreatedTransaction(httpServer, createTxResult, false, closeChan)
}

func (httpServer *HttpServer) handlePdexv3GetTradeStatus(params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}

	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3TradeStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get TradeStatus data"))
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handlePdexv3GetAddOrderStatus(params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}

	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3AddOrderStatusPrefix(),
		txID[:],
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get AddOrderStatus data"))
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handlePdexv3GetWithdrawOrderStatus(params interface{}, closeChan <-chan struct{},
) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	var statusSuffix []byte
	for _, item := range arrayParams {
		s, ok := item.(string)
		h, err := common.Hash{}.NewHashFromStr(s)
		if !ok || err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
				errors.New("Invalid hash from parameters"))
		}
		statusSuffix = append(statusSuffix, h[:]...)
	}

	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3WithdrawOrderStatusPrefix(),
		statusSuffix,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Cannot get WithdrawOrderStatus data"))
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

// --- Helpers ---

func createPdexv3TradeRequestTransaction(
	httpServer *HttpServer, params interface{},
) (*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError) {
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		TradePath           []string
		TokenToSell         common.Hash
		TokenToBuy          common.Hash
		SellAmount          Uint64Reader
		MinAcceptableAmount Uint64Reader
		TradingFee          Uint64Reader
		FeeInPRV            bool
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}
	err = httpServer.pdexTxService.ValidateTokenIDs(&mdReader.TokenToSell, &mdReader.TokenToBuy)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	md, _ := metadataPdexv3.NewTradeRequest(
		mdReader.TradePath, mdReader.TokenToSell, uint64(mdReader.SellAmount),
		uint64(mdReader.MinAcceptableAmount), uint64(mdReader.TradingFee), nil,
		metadataCommon.Pdexv3TradeRequestMeta,
	)

	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	paramSelect.SetTokenID(md.TokenToSell)
	isPRV := md.TokenToSell == common.PRVCoinID
	tokenList := []common.Hash{md.TokenToSell, mdReader.TokenToBuy}
	if mdReader.FeeInPRV && !isPRV && mdReader.TokenToBuy != common.PRVCoinID {
		tokenList = append(tokenList, common.PRVCoinID)
	}
	md.Receiver, err = httpServer.pdexTxService.GenerateOTAReceivers(
		tokenList, paramSelect.PRV.SenderKeySet.PaymentAddress)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for trade, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         md.SellAmount + md.TradingFee,
		},
	}
	if isPRV {
		paramSelect.PRV.PaymentInfos = burnPayments
	} else {
		if mdReader.FeeInPRV {
			// sell amount in token
			paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{
				&privacy.PaymentInfo{
					PaymentAddress: burnAddr,
					Amount:         md.TradingFee,
				},
			}
			// trading fee in PRV
			paramSelect.SetTokenReceivers([]*privacy.PaymentInfo{
				&privacy.PaymentInfo{
					PaymentAddress: burnAddr,
					Amount:         md.SellAmount,
				},
			})
		} else {
			paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
			paramSelect.SetTokenReceivers(burnPayments)
		}
	}

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, isPRV, nil
}

func createPdexv3AddOrderRequestTransaction(
	httpServer *HttpServer, params interface{},
) (*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError) {
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		TokenToSell         common.Hash
		TokenToBuy          common.Hash
		PoolPairID          string
		SellAmount          Uint64Reader
		MinAcceptableAmount Uint64Reader
		NftID               common.Hash
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters"))
	}
	err = httpServer.pdexTxService.ValidateTokenIDs(&mdReader.TokenToSell, &mdReader.TokenToBuy)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	md, _ := metadataPdexv3.NewAddOrderRequest(
		mdReader.TokenToSell, mdReader.PoolPairID, uint64(mdReader.SellAmount),
		uint64(mdReader.MinAcceptableAmount), nil,
		mdReader.NftID, metadataCommon.Pdexv3AddOrderRequestMeta,
	)

	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	paramSelect.SetTokenID(md.TokenToSell)
	isPRV := md.TokenToSell == common.PRVCoinID
	tokenList := []common.Hash{md.TokenToSell, mdReader.TokenToBuy}
	md.Receiver, err = httpServer.pdexTxService.GenerateOTAReceivers(
		tokenList, paramSelect.PRV.SenderKeySet.PaymentAddress)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for order, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         md.SellAmount,
		},
	}
	if isPRV {
		paramSelect.PRV.PaymentInfos = burnPayments
	} else {
		paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
		paramSelect.SetTokenReceivers(burnPayments)
	}

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, isPRV, nil
}

func createPdexv3WithdrawOrderRequestTransaction(
	httpServer *HttpServer, params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	// metadata object format to read from RPC parameters
	mdReader := &struct {
		PoolPairID       string
		OrderID          string
		WithdrawTokenIDs []common.Hash
		Amount           Uint64Reader
		NftID            common.Hash
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters"))
	}
	// sanity check for withdrawing token IDs
	for _, tokenID := range mdReader.WithdrawTokenIDs {
		if tokenID.IsZeroValue() {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
				fmt.Errorf("Invalid WithdrawTokenID %v", tokenID))
		}
	}
	switch len(mdReader.WithdrawTokenIDs) {
	case 1:
		// withdraw single token: proceed
	case 2:
		// withdraw both tokens from order: set withdraw amount to 0
		mdReader.Amount = 0
	default:
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("Invalid WithdrawTokenIDs count %d, expect 1 or 2", len(mdReader.WithdrawTokenIDs)))
	}

	md, _ := metadataPdexv3.NewWithdrawOrderRequest(
		mdReader.PoolPairID, mdReader.OrderID, uint64(mdReader.Amount),
		nil, mdReader.NftID, metadataCommon.Pdexv3WithdrawOrderRequestMeta)

	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	if md.NftID == common.PRVCoinID {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("Cannot use PRV for withdrawOrder TX"))
	}
	paramSelect.SetTokenID(md.NftID)
	tokenList := append(mdReader.WithdrawTokenIDs, md.NftID)
	recv, err := httpServer.pdexTxService.GenerateOTAReceivers(
		tokenList, paramSelect.PRV.SenderKeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	md.Receiver = recv
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn 1 governance-NFT to withdraw order
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         1,
		},
	}
	paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
	paramSelect.SetTokenReceivers(burnPayments)

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}

func sendCreatedTransaction(httpServer *HttpServer, params interface{}, isPRV bool, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var sendTxResult interface{}
	var err *rpcservice.RPCError
	if isPRV {
		sendTxResult, err = httpServer.handleSendRawTransaction(params, closeChan)
	} else {
		sendTxResult, err = httpServer.handleSendRawPrivacyCustomTokenTransaction(params, closeChan)
	}
	if err != nil {
		return nil, err
	}
	return sendTxResult, nil
}

func (httpServer *HttpServer) handlePdexv3Staking(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, isPRV, err := httpServer.createPdexv3StakingRequestTransaction(params)
	if err != nil {
		return nil, err
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err = sendCreatedTransaction(httpServer, newParam, isPRV, closeChan)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (httpServer *HttpServer) createPdexv3StakingRequestTransaction(
	params interface{},
) (*jsonresult.CreateTransactionResult, bool, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	isPRV := false
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	if len(arrayParams) != 5 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}
	otaReceiver := privacy.OTAReceiver{}
	err = otaReceiver.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverStr, err := otaReceiver.String()
	if err != nil {
		return nil, isPRV, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		NftID         string       `json:"NftID"`
		StakingPoolID string       `json:"StakingPoolID"`
		Amount        Uint64Reader `json:"Amount"`
	}{}
	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}

	md := metadataPdexv3.NewStakingRequestWithValue(
		mdReader.StakingPoolID, mdReader.NftID, otaReceiverStr, uint64(mdReader.Amount),
	)
	tokenHash, err := common.Hash{}.NewHashFromStr(mdReader.StakingPoolID)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters %v", err))
	}

	paramSelect.SetTokenID(*tokenHash)
	isPRV = md.TokenID() == common.PRVIDStr
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn selling amount for order, plus fee
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         md.TokenAmount(),
		},
	}
	if isPRV {
		paramSelect.PRV.PaymentInfos = burnPayments
	} else {
		paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
		paramSelect.SetTokenReceivers(burnPayments)
	}

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, false, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, isPRV, nil

}

func (httpServer *HttpServer) handleGetPdexv3StakingStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3StakingStatusPrefix(),
		txID.Bytes(),
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handlePdexv3Unstaking(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var res interface{}
	data, err := httpServer.createPdexv3UnstakingRequestTransaction(params)
	if err != nil {
		return nil, err
	}
	base58CheckData := data.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	res, err = sendCreatedTransaction(httpServer, newParam, false, closeChan)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (httpServer *HttpServer) createPdexv3UnstakingRequestTransaction(
	params interface{},
) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	privateKey, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	privacyDetect, ok := arrayParams[3].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("privacy detection param need to be int"))
	}
	if int(privacyDetect) <= 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Tx has to be a privacy tx"))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize private"))
	}
	if len(keyWallet.KeySet.PrivateKey) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid private key"))
	}
	if len(arrayParams) != 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid length of rpc expect %v but get %v", 4, len(arrayParams)))
	}

	// metadata object format to read from RPC parameters
	mdReader := &struct {
		StakingPoolID string       `json:"StakingPoolID"`
		NftID         string       `json:"NftID"`
		Amount        Uint64Reader `json:"Amount"`
	}{}

	// parse params & metadata
	paramSelect, err := httpServer.pdexTxService.ReadParamsFrom(params, mdReader)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot deserialize parameters"))
	}
	otaReceivers := make(map[string]string)
	otaReceiverNft := privacy.OTAReceiver{}
	err = otaReceiverNft.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverNftStr, err := otaReceiverNft.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceivers[mdReader.NftID] = otaReceiverNftStr
	otaReceiverUnstakingToken := privacy.OTAReceiver{}
	err = otaReceiverUnstakingToken.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceiverUnstakingTokenStr, err := otaReceiverUnstakingToken.String()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	otaReceivers[mdReader.StakingPoolID] = otaReceiverUnstakingTokenStr
	md := metadataPdexv3.NewUnstakingRequestWithValue(
		mdReader.StakingPoolID, mdReader.NftID, otaReceivers, uint64(mdReader.Amount),
	)
	// set token ID & metadata to paramSelect struct. Generate new OTAReceivers from private key
	if md.NftID() == common.PRVIDStr {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			fmt.Errorf("Cannot use PRV for withdrawLiquidity tx"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(md.NftID())
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}
	paramSelect.SetTokenID(*nftID)
	paramSelect.SetMetadata(md)

	// get burning address
	bc := httpServer.pdexTxService.BlockChain
	bestState, err := bc.GetClonedBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
	}
	temp := bc.GetBurningAddress(bestState.BeaconHeight)
	w, _ := wallet.Base58CheckDeserialize(temp)
	burnAddr := w.KeySet.PaymentAddress

	// burn 1 governance-NFT to withdraw order
	burnPayments := []*privacy.PaymentInfo{
		&privacy.PaymentInfo{
			PaymentAddress: burnAddr,
			Amount:         1,
		},
	}
	paramSelect.Token.PaymentInfos = []*privacy.PaymentInfo{}
	paramSelect.SetTokenReceivers(burnPayments)

	// create transaction
	tx, err1 := httpServer.pdexTxService.BuildTransaction(paramSelect, md)
	// error must be of type *RPCError for equality
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err1)
	}

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.CreateTxDataError, err)
	}
	res := &jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(marshaledTx, 0x00),
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetPdexv3UnstakingStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// read txID
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Incorrect parameter length"))
	}
	s, ok := arrayParams[0].(string)
	txID, err := common.Hash{}.NewHashFromStr(s)
	if !ok || err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,
			errors.New("Invalid TxID from parameters"))
	}
	stateDB := httpServer.blockService.BlockChain.GetBeaconBestState().GetBeaconFeatureStateDB()
	data, err := statedb.GetPdexv3Status(
		stateDB,
		statedb.Pdexv3UnstakingStatusPrefix(),
		txID.Bytes(),
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	var res json.RawMessage
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetPdexv3EstimatedStakingReward(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	stakingPoolID, ok := data["StakingPoolID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingPoolID is invalid"))
	}
	nftIDStr, ok := data["NftID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID is invalid"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(nftIDStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	beaconBestView := httpServer.config.BlockChain.GetBeaconBestState()
	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok || beaconHeight == 0 {
		beaconHeight = float64(beaconBestView.BeaconHeight)
	}

	beaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	beaconFeatureStateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, err)
	}

	if uint64(beaconHeight) < config.Param().PDexParams.Pdexv3BreakPointHeight {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, errors.New("pDEX v3 is not available"))
	}

	pDexv3State, err := pdex.InitStateFromDB(beaconFeatureStateDB, uint64(beaconHeight), pdex.AmplifierVersion)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, err)
	}

	stakingPools := pDexv3State.Reader().StakingPools()

	if _, ok := stakingPools[stakingPoolID]; !ok {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, errors.New("TokenID is not existed"))
	}

	pool := stakingPools[stakingPoolID].Clone()

	if _, ok := pool.Stakers()[nftIDStr]; !ok {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, errors.New("NftID is not existed"))
	}

	uncollectedStakingRewards, err := pool.RecomputeStakingRewards(*nftID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, err)
	}
	result := map[string]uint64{}
	for tokenID := range uncollectedStakingRewards {
		result[tokenID.String()] = uncollectedStakingRewards[tokenID]
	}

	return result, nil
}

func (httpServer *HttpServer) handleGetPdexv3EstimatedStakingPoolReward(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	stakingPoolID, ok := data["StakingPoolID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingPoolID is invalid"))
	}

	beaconBestView := httpServer.config.BlockChain.GetBeaconBestState()
	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok || beaconHeight == 0 {
		beaconHeight = float64(beaconBestView.BeaconHeight)
	}

	if uint64(beaconHeight) < config.Param().PDexParams.Pdexv3BreakPointHeight {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, errors.New("pDEX v3 is not available"))
	}

	beaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err))
	}
	stateDB, err := statedb.NewWithPrefixTrie(beaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, err)
	}

	prevBeaconFeatureStateRootHash, err := httpServer.config.BlockChain.GetBeaconFeatureRootHash(beaconBestView, uint64(beaconHeight-1))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, fmt.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight-1, err))
	}
	prevStateDB, err := statedb.NewWithPrefixTrie(prevBeaconFeatureStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.GetBeaconChainDatabase()))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, err)
	}

	result, err := httpServer.blockService.GetPdexv3BlockStakingReward(
		stakingPoolID, uint64(beaconHeight), stateDB, prevStateDB,
	)

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StakingRewardError, err)
	}

	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPdexv3WithdrawStakingReward(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPdexv3WithdrawStakingReward(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	// send raw transaction
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithPdexv3WithdrawStakingReward(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// parse params
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) >= 7 {
		hasPrivacyTokenParam, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("HasPrivacyToken is invalid"))
		}
		hasPrivacyToken := int(hasPrivacyTokenParam) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("The privacy mode must be disabled"))
		}
	}
	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param metadata is invalid"))
	}

	beaconBestView, err := httpServer.blockService.GetBeaconBestState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
	}
	stakingPools := beaconBestView.PdeState(pdex.AmplifierVersion).Reader().StakingPools()

	stakingToken, ok := tokenParamsRaw["StakingPoolID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("StakingPoolID is invalid"))
	}

	pool, found := stakingPools[stakingToken]
	if !found {
		err = fmt.Errorf("Can't find staking token %s", stakingToken)
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3StateError, err)
	}

	nftIDStr, ok := tokenParamsRaw["NftID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID is invalid"))
	}
	nftID, err := common.Hash{}.NewHashFromStr(nftIDStr)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("NftID is invalid"))
	}

	// payment address v2
	feeReceiver, ok := tokenParamsRaw["FeeReceiver"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("FeeReceiver is invalid"))
	}

	tokenIDStrs := []string{
		nftIDStr,
	}

	for tokenID := range pool.RewardsPerShare() {
		tokenIDStrs = append(tokenIDStrs, tokenID.String())
	}

	keyWallet, err := wallet.Base58CheckDeserialize(feeReceiver)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Cannot deserialize payment address: %v", err))
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payment address is invalid"))
	}

	receivers := map[common.Hash]privacy.OTAReceiver{}
	for _, tokenIDStr := range tokenIDStrs {
		tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("TokenID %v is invalid", tokenIDStr))
		}
		receiver := privacy.OTAReceiver{}
		err = receiver.FromAddress(keyWallet.KeySet.PaymentAddress)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GenerateOTAFailError, err)
		}
		receivers[*tokenID] = receiver
	}

	meta, err := metadataPdexv3.NewPdexv3WithdrawalStakingRewardRequest(
		metadataCommon.Pdexv3WithdrawStakingRewardRequestMeta,
		stakingToken,
		*nftID,
		receivers,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err2 := json.Marshal(customTokenTx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleGetPdexv3WithdrawalStakingRewardStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param array must be at least one"))
	}
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	reqTxID, ok := data["ReqTxID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Param ReqTxID is invalid"))
	}
	status, err := httpServer.blockService.GetPdexv3WithdrawalStakingRewardStatus(reqTxID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPdexv3WithdrawalStakingRewardStatusError, err)
	}
	return status, nil
}
