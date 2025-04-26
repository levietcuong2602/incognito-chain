package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"

	eCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/portal/portalrelaying"
	"github.com/incognitochain/incognito-chain/portal/portalv3"
	pCommon "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	portaltokensv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portaltokens"
	"github.com/incognitochain/incognito-chain/relaying/bnb"
	"github.com/stretchr/testify/suite"
)

var _ = func() (_ struct{}) {
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	portalprocessv3.Logger.Init(common.NewBackend(nil).Logger("test", true))
	portaltokensv3.Logger.Init(common.NewBackend(nil).Logger("test", true))
	metadata.Logger.Init(common.NewBackend(nil).Logger("test", true))
	pCommon.Logger.Init(common.NewBackend(nil).Logger("test", true))
	Logger.log.Info("This runs before init()!")
	return
}()

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PortalTestSuiteV3 struct {
	suite.Suite
	currentPortalStateForProducer portalprocessv3.CurrentPortalState
	currentPortalStateForProcess  portalprocessv3.CurrentPortalState
	sdb                           *statedb.StateDB
	portalParams                  portalv3.PortalParams
	blockChain                    *BlockChain
}

const USER_INC_ADDRESS_1 = "12S5pBBRDf1GqfRHouvCV86sWaHzNfvakAWpVMvNnWu2k299xWCgQzLLc9wqPYUHfMYGDprPvQ794dbi6UU1hfRN4tPiU61txWWenhC"
const USER_INC_ADDRESS_2 = "12S1a8VnkwhDTQWZ5PhdpySwiFZj7p8sKdG7oAQFZ3dLsWaV6fhDWk5aSFHpt1jcPBjY4sYgwqAqRzx3oTYDZCvCei1LSCdJARXWiyK"
const USER_INC_ADDRESS_3 = "12Rs38qgRQZXdeKgy9QS2ebg6pHA875eJcxQRof3QAkhBFaBDKovGbxvPqV2wa2k128vQY7ahuXLSNcUvs2g52QsFFbdiaqQUWKLdc4"
const USER_INC_ADDRESS_4 = "12Rs38qgRQZXdeKgy9QS2ebg6pHA875eJcxQRof3QAkhBFaBDKovGbxvPqV2wa2k128vQY7ahuXLSNcUvs2g52QsFFbdiaqQUWKLdc5"
const USER_INC_ADDRESS_5 = "12Rs38qgRQZXdeKgy9QS2ebg6pHA875eJcxQRof3QAkhBFaBDKovGbxvPqV2wa2k128vQY7ahuXLSNcUvs2g52QsFFbdiaqQUWKLdc6"
const USER_INC_ADDRESS_6 = "12Rs38qgRQZXdeKgy9QS2ebg6pHA875eJcxQRof3QAkhBFaBDKovGbxvPqV2wa2k128vQY7ahuXLSNcUvs2g52QsFFbdiaqQUWKLdc7"
const USER_INC_ADDRESS_7 = "12Rs38qgRQZXdeKgy9QS2ebg6pHA875eJcxQRof3QAkhBFaBDKovGbxvPqV2wa2k128vQY7ahuXLSNcUvs2g52QsFFbdiaqQUWKLdc8"
const USER_BNB_ADDRESS_1 = "tbnb1d90lad6rg5ldh8vxgtuwzxd8n6rhhx7mfqek38"
const USER_BNB_ADDRESS_2 = "tbnb172pnrmd0409237jwlq5qjhw2s2r7lq6ukmaeke"
const USER_ETH_ADDRESS_1 = "user-eth-address-1"
const USER_ETH_ADDRESS_2 = "user-eth-address-2"

const CUS_INC_ADDRESS_1 = "12RuEdPjq4yxivzm8xPxRVHmkL74t4eAdUKPdKKhMEnpxPH3k8GEyULbwq4hjwHWmHQr7MmGBJsMpdCHsYAqNE18jipWQwciBf9yqvQ"
const CUS_INC_ADDRESS_2 = "12Rwz4HXkVABgRnSb5Gfu1FaJ7auo3fLNXVGFhxx1dSytxHpWhbkimT1Mv5Z2oCMsssSXTVsapY8QGBZd2J4mPiCTzJAtMyCzb4dDcy"
const CUS_INC_ADDRESS_3 = "12S4NL3DZ1KoprFRy1k5DdYSXUq81NtxFKdvUTP3PLqQypWzceL5fBBwXooAsX5s23j7cpb1Za37ddmfSaMpEJDPsnJGZuyWTXJSZZ5"
const CUS_INC_ADDRESS_4 = "12S4NL3DZ1KoprFRy1k5DdYSXUq81NtxFKdvUTP3PLqQypWzceL5fBBwXooAsX5s23j7cpb1Za37ddmfSaMpEJDPsnJGZuyWTXJSZZ6"
const CUS_BNB_ADDRESS_1 = "tbnb19cmxazhx5ujlhhlvj9qz0wv8a4vvsx8vuy9cyc"
const CUS_BNB_ADDRESS_2 = "tbnb1zyqrky9zcumc2e4smh3xwh2u8kudpdc56gafuk"
const CUS_BNB_ADDRESS_3 = "tbnb1n5lrzass9l28djvv7drst53dcw7y9yj4pyvksf"
const CUS_BTC_ADDRESS_1 = "btc-address-1"
const CUS_BTC_ADDRESS_2 = "btc-address-2"
const CUS_BTC_ADDRESS_3 = "btc-address-3"
const CUS_ETH_ADDRESS_1 = "cus-eth-address-1"
const CUS_ETH_ADDRESS_2 = "cus-eth-address-2"
const CUS_ETH_ADDRESS_3 = "cus-eth-address-3"

const USDT_ID = "64fbdbc6bf5b228814b58706d91ed03777f0edf6"
const DAI_ID = "4f96fe3b7a6cf9725f59d353f723c1bdb64ca6aa"
const ETH_ID = "0000000000000000000000000000000000000000"

const BNB_NODE_URL = "https://data-seed-pre-0-s3.binance.org:443"

var supportedCollaterals = []portalv3.PortalCollateral{
	{"0000000000000000000000000000000000000000", 9},  // eth
	{"64fbdbc6bf5b228814b58706d91ed03777f0edf6", 6},  // usdt, kovan testnet
	{"7079f3762805cff9c979a5bdc6f5648bcfee76c8", 6},  // usdc, kovan testnet
	{"4f96fe3b7a6cf9725f59d353f723c1bdb64ca6aa", 18}, // dai, kovan testnet
}

func (s *PortalTestSuiteV3) SetupTest() {
	dbPath, err := ioutil.TempDir(os.TempDir(), "portal_test_statedb_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBStatedbTest := statedb.NewDatabaseAccessWarper(diskBD)
	emptyRoot := common.HexToHash(common.HexEmptyRoot)
	stateDB, _ := statedb.NewWithPrefixTrie(emptyRoot, warperDBStatedbTest)

	s.sdb = stateDB

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:        {Amount: 1000000000},
			pCommon.PortalBNBIDStr: {Amount: 20000000000},
			pCommon.PortalBTCIDStr: {Amount: 10000000000000},
			ETH_ID:                 {Amount: 400000000000},
			USDT_ID:                {Amount: 1000000000},
			DAI_ID:                 {Amount: 1000000000},
		})
	s.currentPortalStateForProducer = portalprocessv3.CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    finalExchangeRate,
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.currentPortalStateForProcess = portalprocessv3.CurrentPortalState{
		CustodianPoolState:         map[string]*statedb.CustodianState{},
		WaitingPortingRequests:     map[string]*statedb.WaitingPortingRequest{},
		WaitingRedeemRequests:      map[string]*statedb.RedeemRequest{},
		MatchedRedeemRequests:      map[string]*statedb.RedeemRequest{},
		FinalExchangeRatesState:    finalExchangeRate,
		LiquidationPool:            map[string]*statedb.LiquidationPool{},
		LockedCollateralForRewards: new(statedb.LockedCollateralState),
		ExchangeRatesRequests:      map[string]*metadata.ExchangeRatesRequestStatus{},
	}
	s.portalParams = portalv3.PortalParams{
		TimeOutCustodianReturnPubToken:       24 * time.Hour,
		TimeOutWaitingPortingRequest:         24 * time.Hour,
		TimeOutWaitingRedeemRequest:          10 * time.Minute,
		MaxPercentLiquidatedCollateralAmount: 120,
		MaxPercentCustodianRewards:           10,
		MinPercentCustodianRewards:           1,
		MinLockCollateralAmountInEpoch:       10000 * 1e6, // 10000 usd
		MinPercentLockedCollateral:           200,
		TP120:                                120,
		TP130:                                130,
		MinPercentPortingFee:                 0.01,
		MinPercentRedeemFee:                  0.01,
		SupportedCollateralTokens:            supportedCollaterals,
		MinPortalFee:                         100,
		MinUnlockOverRateCollaterals:         25,
		PortalTokens: map[string]portaltokensv3.PortalTokenProcessorV3{
			pCommon.PortalBTCIDStr: &portaltokensv3.PortalBTCTokenProcessor{
				&portaltokensv3.PortalTokenV3{
					ChainID:        "Bitcoin-Testnet",
					MinTokenAmount: 10,
				},
			},
			pCommon.PortalBNBIDStr: &portaltokensv3.PortalBNBTokenProcessor{
				&portaltokensv3.PortalTokenV3{
					ChainID:        "Binance-Chain-Ganges",
					MinTokenAmount: 10,
				},
			},
		},

		PortalETHContractAddressStr: "0xDdFe62F1022a62bF8Dc007cb4663228C71F5235b",
	}
	tempPortalParam := &portal.PortalParams{
		PortalParamsV3: map[uint64]portalv3.PortalParams{
			0: s.portalParams,
		},
		RelayingParam: portalrelaying.RelayingParams{
			BNBRelayingHeaderChainID: "Binance-Chain-Ganges",
			BTCRelayingHeaderChainID: "Bitcoin-Testnet",
			BTCDataFolderName:        "btcrelayingv14",
			BNBFullNodeProtocol:      "https",
			BNBFullNodeHost:          "data-seed-pre-0-s3.binance.org",
			BNBFullNodePort:          "443",
		},
	}
	config.AbortParam()
	config.Param().BlockTime.MinBeaconBlockInterval = 40 * time.Second
	config.Param().BlockTime.MinShardBlockInterval = 40 * time.Second
	config.Param().EpochParam.NumberOfBlockInEpoch = 100
	portal.SetupPortalParam(tempPortalParam)
}

/*
 Utility functions
*/

//// external tokenID there is no 0x prefix, in lower case
//// @@Note: need to update before deploying
//func getSupportedPortalCollateralsTestnet() []portal.PortalCollateral {
//	return []portal.PortalCollateral{
//		{"0000000000000000000000000000000000000000", 9}, // eth
//		{"3a829f4b97660d970428cd370c4e41cbad62092b", 6}, // usdt, kovan testnet
//		{"75b0622cec14130172eae9cf166b92e5c112faff", 6}, // usdc, kovan testnet
//	}
//}

func exchangeRates(amount uint64, tokenIDFrom string, tokenIDTo string, finalExchangeRate *statedb.FinalExchangeRatesState, portalParams portalv3.PortalParams) uint64 {
	convertTool := portalprocessv3.NewPortalExchangeRateTool(finalExchangeRate, portalParams)
	res, _ := convertTool.Convert(tokenIDFrom, tokenIDTo, amount)
	return res
}

func getLockedCollateralAmount(
	portingAmount uint64, tokenID string, collateralTokenID string, finalExchangeRate *statedb.FinalExchangeRatesState, percent uint64, portalParams portalv3.PortalParams) uint64 {
	amount := portalprocessv3.UpPercent(portingAmount, percent)
	return exchangeRates(amount, tokenID, collateralTokenID, finalExchangeRate, portalParams)
}

func getMinFee(amount uint64, tokenID string, finalExchangeRate *statedb.FinalExchangeRatesState, percent float64, portalParams portalv3.PortalParams) uint64 {
	amountInPRV := exchangeRates(amount, tokenID, common.PRVIDStr, finalExchangeRate, portalParams)
	fee := float64(amountInPRV) * percent / float64(100)
	return uint64(math.Round(fee))
}

func getUnlockAmount(totalLockedAmount uint64, totalPTokenAmount uint64, pTokenAmount uint64) uint64 {
	amount := new(big.Int).Mul(new(big.Int).SetUint64(pTokenAmount), new(big.Int).SetUint64(totalLockedAmount))
	amount = amount.Div(amount, new(big.Int).SetUint64(totalPTokenAmount))
	return amount.Uint64()
}

func (s *PortalTestSuiteV3) TestGetLockedCollateralAmount() {
	portingAmount := uint64(30 * 1e9)
	tokenID := pCommon.PortalBNBIDStr
	collateralTokenID := USDT_ID

	percent := s.portalParams.MinPercentLockedCollateral
	amount := getLockedCollateralAmount(portingAmount, tokenID, collateralTokenID, s.currentPortalStateForProducer.FinalExchangeRatesState, percent, s.portalParams)
	fmt.Println("Result from TestGetLockedCollateralAmount: ", amount)
}

func (s *PortalTestSuiteV3) TestGetMinFee() {
	amount := uint64(140 * 1e9)
	tokenID := pCommon.PortalBNBIDStr
	percent := s.portalParams.MinPercentPortingFee

	fee := getMinFee(amount, tokenID, s.currentPortalStateForProducer.FinalExchangeRatesState, percent, s.portalParams)
	fmt.Println("Result from TestGetMinFee: ", fee)
}

func (s *PortalTestSuiteV3) TestGetUnlockAmount() {
	totalLockedAmount := uint64(40000000000)
	totalPTokenAmount := uint64(1 * 1e9)
	pTokenAmount := uint64(0.3 * 1e9)

	unlockAmount := getUnlockAmount(totalLockedAmount, totalPTokenAmount, pTokenAmount)
	fmt.Println("Result from TestGetUnlockAmount: ", unlockAmount)
}

func (s *PortalTestSuiteV3) TestExchangeRate() {
	// uncomment this code to update final exchange rate for converting
	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:        {Amount: 1000000000},
			pCommon.PortalBNBIDStr: {Amount: 35000000000}, // x1.75
			pCommon.PortalBTCIDStr: {Amount: 10000000000000},
			ETH_ID:                 {Amount: 400000000000},
			USDT_ID:                {Amount: 1000000000},
			DAI_ID:                 {Amount: 1000000000},
		})
	s.currentPortalStateForProducer.FinalExchangeRatesState = finalExchangeRate

	amount := uint64(13500000000)
	tokenIDFrom := pCommon.PortalBNBIDStr
	tokenIDTo := USDT_ID
	convertAmount := exchangeRates(amount, tokenIDFrom, tokenIDTo, s.currentPortalStateForProducer.FinalExchangeRatesState, s.portalParams)
	fmt.Println("Result from TestExchangeRate: ", convertAmount)
}

type instructionForProducer struct {
	inst         []string
	optionalData map[string]interface{}
}

func producerPortalInstructions(
	blockchain metadata.ChainRetriever,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	insts []instructionForProducer,
	currentPortalState *portalprocessv3.CurrentPortalState,
	portalParams portalv3.PortalParams,
	shardID byte,
	pm map[int]portalprocessv3.PortalInstructionProcessorV3,
) ([][]string, error) {
	var newInsts [][]string

	for _, item := range insts {
		inst := item.inst
		optionalData := item.optionalData

		metaType, _ := strconv.Atoi(inst[0])
		contentStr := inst[1]
		portalProcessor := portalprocessv3.GetPortalInstProcessorByMetaType(pm, metaType)
		newInst, err := portalProcessor.BuildNewInsts(
			blockchain,
			contentStr,
			shardID,
			currentPortalState,
			beaconHeight,
			shardHeights,
			portalParams,
			optionalData,
		)
		if err != nil {
			Logger.log.Error(err)
			return newInsts, err
		}

		newInsts = append(newInsts, newInst...)
	}

	return newInsts, nil
}

func processPortalInstructions(
	blockchain metadata.ChainRetriever,
	beaconHeight uint64,
	insts [][]string,
	portalStateDB *statedb.StateDB,
	currentPortalState *portalprocessv3.CurrentPortalState,
	portalParams portalv3.PortalParams,
	pm map[int]portalprocessv3.PortalInstructionProcessorV3,
) error {
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
	//todo
	epoch := uint64(100)
	for _, inst := range insts {
		if len(inst) < 4 {
			continue // Not error, just not Portal instruction
		}

		var err error
		metaType, _ := strconv.Atoi(inst[0])
		processor := portalprocessv3.GetPortalInstProcessorByMetaType(pm, metaType)
		if processor != nil {
			err = processor.ProcessInsts(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, updatingInfoByTokenID)
			if err != nil {
				Logger.log.Errorf("Process portal instruction err: %v, inst %+v", err, inst)
			}
			continue
		}

		switch inst[0] {
		// ============ Reward ============
		// portal reward
		case strconv.Itoa(metadata.PortalRewardMeta), strconv.Itoa(metadata.PortalRewardMetaV3):
			err = portalprocessv3.ProcessPortalReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, epoch)
		// total custodian reward instruction
		case strconv.Itoa(metadata.PortalTotalRewardCustodianMeta):
			err = portalprocessv3.ProcessPortalTotalCustodianReward(portalStateDB, beaconHeight, inst, currentPortalState, portalParams, epoch)

		// ============ Portal smart contract ============
		// todo: add more metadata need to unlock token from sc
		case strconv.Itoa(metadata.PortalCustodianWithdrawConfirmMetaV3),
			strconv.Itoa(metadata.PortalRedeemFromLiquidationPoolConfirmMetaV3),
			strconv.Itoa(metadata.PortalLiquidateRunAwayCustodianConfirmMetaV3):
			err = portalprocessv3.ProcessPortalConfirmWithdrawInstV3(portalStateDB, beaconHeight, inst, currentPortalState, portalv3.PortalParams{})
		}

		if err != nil {
			Logger.log.Errorf("Process portal instruction err: %v, inst %+v", err, inst)
		}

	}

	// pick the final exchangeRates
	portalprocessv3.PickExchangesRatesFinal(currentPortalState)

	// update info of bridge portal token
	for _, updatingInfo := range updatingInfoByTokenID {
		var updatingAmt uint64
		var updatingType string
		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
			updatingType = "+"
		}
		if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
			updatingAmt = updatingInfo.DeductAmt - updatingInfo.CountUpAmt
			updatingType = "-"
		}
		err := statedb.UpdateBridgeTokenInfo(
			portalStateDB,
			updatingInfo.TokenID,
			updatingInfo.ExternalTokenID,
			updatingInfo.IsCentralized,
			updatingAmt,
			updatingType,
		)
		if err != nil {
			return err
		}
	}

	// store updated currentPortalState to leveldb with new beacon height
	err := portalprocessv3.StorePortalStateToDB(portalStateDB, currentPortalState)
	if err != nil {
		Logger.log.Error(err)
	}

	return nil
}

func buildPortalRelayExchangeRateAction(
	incAddressStr string,
	rates []*metadata.ExchangeRateInfo,
	shardID byte,
) []string {
	data := metadata.PortalExchangeRates{
		MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
			MetadataBase: metadata.MetadataBase{
				Type: metadata.PortalExchangeRatesMeta,
			},
		},
		SenderAddress: incAddressStr,
		Rates:         rates,
	}

	actionContent := metadata.PortalExchangeRatesAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalExchangeRatesMeta), actionContentBase64Str}
}

func buildPortalCustodianDepositAction(
	incAddressStr string,
	remoteAddress map[string]string,
	depositAmount uint64,
	shardID byte,
) []string {
	data := metadata.PortalCustodianDeposit{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianDepositMeta,
		},
		IncogAddressStr: incAddressStr,
		RemoteAddresses: remoteAddress,
		DepositedAmount: depositAmount,
	}

	actionContent := metadata.PortalCustodianDepositAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianDepositMeta), actionContentBase64Str}
}

func buildPortalCustodianDepositActionV3(
	remoteAddress map[string]string,
	blockHash eCommon.Hash,
	txIndex uint,
	proofStrs []string,
	shardID byte,
) []string {
	data := metadata.PortalCustodianDepositV3{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianDepositMetaV3,
		},
		RemoteAddresses: remoteAddress,
		BlockHash:       blockHash,
		TxIndex:         txIndex,
		ProofStrs:       proofStrs,
	}

	actionContent := metadata.PortalCustodianDepositActionV3{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianDepositMetaV3), actionContentBase64Str}
}

func buildPortalCustodianWithdrawActionV3(
	custodianIncAddress string,
	custodianExternalAddress string,
	externalTokenID string,
	amount uint64,
	shardID byte,
) []string {
	data := metadata.PortalCustodianWithdrawRequestV3{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianWithdrawRequestMetaV3,
		},
		CustodianIncAddress:      custodianIncAddress,
		CustodianExternalAddress: custodianExternalAddress,
		ExternalTokenID:          externalTokenID,
		Amount:                   amount,
	}

	actionContent := metadata.PortalCustodianWithdrawRequestActionV3{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianWithdrawRequestMetaV3), actionContentBase64Str}
}

func buildPortalUserRegisterAction(
	portingID string,
	incAddressStr string,
	pTokenID string,
	portingAmount uint64,
	portingFee uint64,
	shardID byte,
	shardHeight uint64,
) []string {
	data := metadata.PortalUserRegister{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalRequestPortingMetaV3,
		},
		UniqueRegisterId: portingID,
		IncogAddressStr:  incAddressStr,
		PTokenId:         pTokenID,
		RegisterAmount:   portingAmount,
		PortingFee:       portingFee,
	}

	actionContent := metadata.PortalUserRegisterActionV3{
		Meta:        data,
		TxReqID:     common.Hash{},
		ShardID:     shardID,
		ShardHeight: shardHeight,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalRequestPortingMetaV3), actionContentBase64Str}
}

func buildPortalUserReqPTokenAction(
	portingID string,
	incAddressStr string,
	pTokenID string,
	portingAmount uint64,
	portingProof string,
	shardID byte,
) []string {
	data := metadata.PortalRequestPTokens{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalUserRequestPTokenMeta,
		},
		UniquePortingID: portingID,
		TokenID:         pTokenID,
		IncogAddressStr: incAddressStr,
		PortingAmount:   portingAmount,
		PortingProof:    portingProof,
	}

	actionContent := metadata.PortalRequestPTokensAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalUserRequestPTokenMeta), actionContentBase64Str}
}

func buildTopupWaitingPortingAction(
	incAddressStr string,
	portingID string,
	ptokenID string,
	depositAmount uint64,
	shardID byte,
	freeCollateralAmount uint64,
) []string {
	data := metadata.PortalTopUpWaitingPortingRequest{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalTopUpWaitingPortingRequestMeta,
		},
		IncogAddressStr:      incAddressStr,
		PortingID:            portingID,
		PTokenID:             ptokenID,
		DepositedAmount:      depositAmount,
		FreeCollateralAmount: freeCollateralAmount,
	}

	actionContent := metadata.PortalTopUpWaitingPortingRequestAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalTopUpWaitingPortingRequestMeta), actionContentBase64Str}
}

func buildPortalTopupCustodianAction(
	incAddressStr string,
	ptokenID string,
	depositAmount uint64,
	shardID byte,
	freeCollateralAmount uint64,
) []string {
	data := metadata.PortalLiquidationCustodianDepositV2{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalCustodianTopupMetaV2,
		},
		IncogAddressStr:      incAddressStr,
		PTokenId:             ptokenID,
		DepositedAmount:      depositAmount,
		FreeCollateralAmount: freeCollateralAmount,
	}

	actionContent := metadata.PortalLiquidationCustodianDepositActionV2{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalCustodianTopupMetaV2), actionContentBase64Str}
}

func buildPortalRequestRedeemActionV3(
	uniqueRedeemID string,
	tokenID string,
	redeemAmount uint64,
	redeemerIncAddressStr string,
	remoteAddress string,
	redeemFee uint64,
	redeemerExternalAddress string,
	shardID byte,
	shardHeight uint64,
) []string {
	data := metadata.PortalRedeemRequestV3{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalRedeemRequestMetaV3,
		},
		UniqueRedeemID:          uniqueRedeemID,
		TokenID:                 tokenID,
		RedeemAmount:            redeemAmount,
		RedeemerIncAddressStr:   redeemerIncAddressStr,
		RemoteAddress:           remoteAddress,
		RedeemFee:               redeemFee,
		RedeemerExternalAddress: redeemerExternalAddress,
	}

	actionContent := metadata.PortalRedeemRequestActionV3{
		Meta:        data,
		TxReqID:     common.Hash{},
		ShardID:     shardID,
		ShardHeight: shardHeight,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalRedeemRequestMetaV3), actionContentBase64Str}
}

func buildPortalRequestMatchingWRedeemActionV3(
	uniqueRedeemID string,
	custodianIncAddress string,
	shardID byte,
) []string {
	data := metadata.PortalReqMatchingRedeem{
		MetadataBaseWithSignature: metadata.MetadataBaseWithSignature{
			MetadataBase: metadata.MetadataBase{
				Type: metadata.PortalReqMatchingRedeemMeta,
			},
		},
		CustodianAddressStr: custodianIncAddress,
		RedeemID:            uniqueRedeemID,
	}

	actionContent := metadata.PortalReqMatchingRedeemAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalReqMatchingRedeemMeta), actionContentBase64Str}
}

func buildPortalRequestUnlockCollateralsActionV3(
	uniqueRedeemID string,
	portalTokenID string,
	custodianIncAddress string,
	redeemAmount uint64,
	redeemProof string,
	shardID byte,
) []string {
	data := metadata.PortalRequestUnlockCollateral{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalRequestUnlockCollateralMetaV3,
		},
		UniqueRedeemID:      uniqueRedeemID,
		TokenID:             portalTokenID,
		CustodianAddressStr: custodianIncAddress,
		RedeemAmount:        redeemAmount,
		RedeemProof:         redeemProof,
	}

	actionContent := metadata.PortalRequestUnlockCollateralAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalRequestUnlockCollateralMetaV3), actionContentBase64Str}
}

/*
	Feature 0: Relay exchange rate
*/
type TestCaseRelayExchangeRate struct {
	senderAddressStr string
	rates            []*metadata.ExchangeRateInfo
}

func buildPortalExchangeRateActionsFromTcs(tcs []TestCaseRelayExchangeRate, shardID byte) [][]string {
	insts := [][]string{}

	for _, tc := range tcs {
		inst := buildPortalRelayExchangeRateAction(tc.senderAddressStr, tc.rates, shardID)
		insts = append(insts, inst)
	}

	return insts
}

//func (s *PortalTestSuiteV3) TestRelayExchangeRate() {
//	fmt.Println("Running TestRelayExchangeRate - beacon height 999 ...")
//	bc := s.blockChain
//	pm := NewPortalManager()
//	beaconHeight := uint64(999)
//	shardID := byte(0)
//	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
//
//	// build test cases
//	testcases := []TestCaseRelayExchangeRate{
//		// valid
//		{
//			senderAddressStr: "feeder1",
//			rates: []*metadata.ExchangeRateInfo{
//				{
//					PTokenID: common.PRVIDStr,
//					Rate:     1000000,
//				},
//				{
//					PTokenID: pCommon.PortalBNBIDStr,
//					Rate:     20000000,
//				},
//				{
//					PTokenID: pCommon.PortalBTCIDStr,
//					Rate:     10000000000,
//				},
//			},
//		},
//	}
//
//	// build actions from testcases
//	insts := buildPortalExchangeRateActionsFromTcs(testcases, shardID)
//
//	// producer instructions
//	newInsts, err := producerPortalInstructions(
//		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.portalParams, shardID, pm)
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)
//
//	// check results
//	s.Equal(1, len(newInsts))
//	s.Equal(nil, err)
//
//	//exchangeRateKey := statedb.GeneratePortalFinalExchangeRatesStateObjectKey().String()
//
//	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
//		map[string]statedb.FinalExchangeRatesDetail{
//			common.PRVIDStr:       {Amount: 1000000},
//			pCommon.PortalBNBIDStr: {Amount: 20000000},
//			pCommon.PortalBTCIDStr: {Amount: 10000000000},
//		})
//
//	s.Equal(finalExchangeRate, s.currentPortalStateForProcess.FinalExchangeRatesState)
//}

/*
	Feature 1: Custodians deposit collateral (PRV)
*/

type TestCaseCustodianDeposit struct {
	custodianIncAddress string
	remoteAddress       map[string]string
	depositAmount       uint64
}

type ExpectedResultCustodianDeposit struct {
	custodianPool  map[string]*statedb.CustodianState
	numBeaconInsts uint
}

func buildTestCaseAndExpectedResultCustodianDeposit() ([]TestCaseCustodianDeposit, *ExpectedResultCustodianDeposit) {
	testcases := []TestCaseCustodianDeposit{
		// valid
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
				pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
			},
			depositAmount: 5000 * 1e9,
		},
		// custodian deposit more with new remote addresses
		// expect don't change to new remote addresses,
		// custodian is able to update new remote addresses when total collaterals is empty
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
				pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
			},
			depositAmount: 2000 * 1e9,
		},
		// new custodian supply only bnb address
		{
			custodianIncAddress: CUS_INC_ADDRESS_2,
			remoteAddress: map[string]string{
				pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
			},
			depositAmount: 1000 * 1e9,
		},
		// new custodian supply only btc address
		{
			custodianIncAddress: CUS_INC_ADDRESS_3,
			remoteAddress: map[string]string{
				pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
			},
			depositAmount: 10000 * 1e9,
		},
	}

	// build expected results
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(7000 * 1e9)
	custodian1.SetFreeCollateral(7000 * 1e9)

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalCollateral(1000 * 1e9)
	custodian2.SetFreeCollateral(1000 * 1e9)

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
		})
	custodian3.SetTotalCollateral(10000 * 1e9)
	custodian3.SetFreeCollateral(10000 * 1e9)

	expectedRes := &ExpectedResultCustodianDeposit{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		numBeaconInsts: 4,
	}

	return testcases, expectedRes
}

func buildCustodianDepositActionsFromTcs(tcs []TestCaseCustodianDeposit, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalCustodianDepositAction(tc.custodianIncAddress, tc.remoteAddress, tc.depositAmount, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestCustodianDepositCollateral() {
	fmt.Println("Running TestCustodianDepositCollateral - beacon height 1000 ...")
	bc := s.blockChain
	pm := portal.NewPortalManager()
	beaconHeight := uint64(1000)
	shardHeights := map[byte]uint64{
		0: uint64(1000),
	}
	shardID := byte(0)

	// build test cases
	testcases, expectedRes := buildTestCaseAndExpectedResultCustodianDeposit()

	// build actions from testcases
	instsForProducer := buildCustodianDepositActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

	// check results
	s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)
	s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 2: Custodians deposit collateral (ETH/ERC20)
*/

type TestCaseCustodianDepositV3 struct {
	custodianIncAddress string
	remoteAddress       map[string]string
	depositAmount       uint64
	collateralTokenID   string
	blockHash           eCommon.Hash
	txIndex             uint
	proofStrs           []string

	isSubmitted      bool
	uniqExternalTxID []byte
}

type ExpectedResultCustodianDepositV3 struct {
	custodianPool  map[string]*statedb.CustodianState
	numBeaconInsts uint
	statusInsts    []string
}

func buildTestCaseAndExpectedResultCustodianDepositV3() ([]TestCaseCustodianDepositV3, *ExpectedResultCustodianDepositV3) {
	// build test cases
	// kovan env and inc local
	// custodian addr: 12S5Lrs1XeQLbqN4ySyKtjAjd2d7sBP2tjFijzmp6avrrkQCNFMpkXm3FPzj2Wcu2ZNqJEmh9JriVuRErVwhuQnLmWSaggobEWsBEci
	testcases := []TestCaseCustodianDepositV3{
		// valid
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
				pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
			},
			depositAmount:     1 * 1e8, // 0.1 ETH
			collateralTokenID: ETH_ID,
			blockHash:         eCommon.HexToHash("0xb9501b5c90fda7c22b29f6268eb20fbf6e13f599149a3af4654fdcc041cd42ba"),
			txIndex:           0,
			proofStrs:         []string{"+FGgk3WNTlqE/WmqFuZw5BT+8jXcjqkx87k/ylg4lhs+k8aAgICAgICAoJwqPBAQ6mJR6Y5apWOx7jnMO5oeAmuIwp00iK/w59DSgICAgICAgIA=", "+QJOMLkCSvkCRwGChZu5AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA+QE9+QE6lN3+YvECKmK/jcAHy0ZjIoxx9SNb4aAtS1l5NfPNZ/su6/HbTevJNM7lx7qnFT+YD9vrLnQITrkBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABY0V4XYoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABnMTJSdUVkUGpxNHl4aXZ6bTh4UHhSVkhta0w3NHQ0ZUFkVUtQZEtLaE1FbnB4UEgzazhHRXlVTGJ3cTRoandIV21IUXI3TW1HQkpzTXBkQ0hzWUFxTkUxOGppcFdRd2NpQmY5eXF2UQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},
			uniqExternalTxID:  []byte{},
			isSubmitted:       false,
		},
		// the existed custodian deposits other token more with new remote addresses
		// expect don't change to new remote addresses,
		// custodian is able to update new remote addresses when total collaterals is empty
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
				pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
			},
			depositAmount:     10 * 1e9, // 10 DAI
			collateralTokenID: DAI_ID,
			blockHash:         eCommon.HexToHash("0x3bc93973b60f4f1d7b078229f80651aa24835e18b75781f2ef7478a34771150a"),
			txIndex:           0,
			proofStrs:         []string{"+QLtgiCAuQLn+QLkAYLZILkBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACACAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAgAAAAAAAAABAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAQAAEAAAAAAAAAAAAAAEAAIAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAIAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD5Adr4m5RPlv47emz5cl9Z01P3I8G9tkymqvhjoN3yUq0b4sibacKwaPw3jaqVK6fxY8ShFij1Wk31I7PvoAAAAAAAAAAAAAAAAP2UxGq43PCSjVETpv6qkleTUE4WoAAAAAAAAAAAAAAAAN3+YvECKmK/jcAHy0ZjIoxx9SNboAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIrHIwSJ6AAA+QE6lN3+YvECKmK/jcAHy0ZjIoxx9SNb4aAtS1l5NfPNZ/su6/HbTevJNM7lx7qnFT+YD9vrLnQITrkBAAAAAAAAAAAAAAAAAE+W/jt6bPlyX1nTU/cjwb22TKaqAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACVAvkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABnMTJSdUVkUGpxNHl4aXZ6bTh4UHhSVkhta0w3NHQ0ZUFkVUtQZEtLaE1FbnB4UEgzazhHRXlVTGJ3cTRoandIV21IUXI3TW1HQkpzTXBkQ0hzWUFxTkUxOGppcFdRd2NpQmY5eXF2UQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},
			uniqExternalTxID:  []byte{},
			isSubmitted:       false,
		},
		// invalid: submit the used proof
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			remoteAddress: map[string]string{
				pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
				pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
			},
			depositAmount:     1 * 1e8, // 0.1 ETH
			collateralTokenID: ETH_ID,
			blockHash:         eCommon.HexToHash("0xb9501b5c90fda7c22b29f6268eb20fbf6e13f599149a3af4654fdcc041cd42ba"),
			txIndex:           0,
			proofStrs:         []string{"+FGgk3WNTlqE/WmqFuZw5BT+8jXcjqkx87k/ylg4lhs+k8aAgICAgICAoJwqPBAQ6mJR6Y5apWOx7jnMO5oeAmuIwp00iK/w59DSgICAgICAgIA=", "+QJOMLkCSvkCRwGChZu5AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA+QE9+QE6lN3+YvECKmK/jcAHy0ZjIoxx9SNb4aAtS1l5NfPNZ/su6/HbTevJNM7lx7qnFT+YD9vrLnQITrkBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABY0V4XYoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABnMTJSdUVkUGpxNHl4aXZ6bTh4UHhSVkhta0w3NHQ0ZUFkVUtQZEtLaE1FbnB4UEgzazhHRXlVTGJ3cTRoandIV21IUXI3TW1HQkpzTXBkQ0hzWUFxTkUxOGppcFdRd2NpQmY5eXF2UQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},
			uniqExternalTxID:  []byte{},
			isSubmitted:       true,
		},
		// new custodian deposit ERC20 (DAI)
		{
			custodianIncAddress: CUS_INC_ADDRESS_2,
			remoteAddress: map[string]string{
				pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
			},
			depositAmount:     10 * 1e9, // 10 DAI
			collateralTokenID: DAI_ID,
			blockHash:         eCommon.HexToHash("0xf87360c12ccc3c673417c56f1d3ff02b603d61a3294553d4b24d840d5283a875"),
			txIndex:           0,
			proofStrs:         []string{"+FGgLol8ZRst35UPhCuvVQQG/4UtrK8Bhl7e9Imro+uJjpyAgICAgICAoCN6C36Lr1Z6/ZfQzhQKe47dWHE8tD87xuNWQ0NaQ2uqgICAgICAgIA=", "+QLrMLkC5/kC5AGC2SC5AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAgAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAIAAAAAAAAAAQAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAEAABAAAAAAAAAAAAAABAACAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAACAAAAAAUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA+QHa+JuUT5b+O3ps+XJfWdNT9yPBvbZMpqr4Y6Dd8lKtG+LIm2nCsGj8N42qlSun8WPEoRYo9VpN9SOz76AAAAAAAAAAAAAAAAD9lMRquNzwko1RE6b+qpJXk1BOFqAAAAAAAAAAAAAAAADd/mLxAipiv43AB8tGYyKMcfUjW6AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACKxyMEiegAAPkBOpTd/mLxAipiv43AB8tGYyKMcfUjW+GgLUtZeTXzzWf7Luvx203ryTTO5ce6pxU/mA/b6y50CE65AQAAAAAAAAAAAAAAAABPlv47emz5cl9Z01P3I8G9tkymqgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAlQL5AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAZzEyUnd6NEhYa1ZBQmdSblNiNUdmdTFGYUo3YXVvM2ZMTlhWR0ZoeHgxZFN5dHhIcFdoYmtpbVQxTXY1WjJvQ01zc3NTWFRWc2FwWThRR0JaZDJKNG1QaUNUekpBdE15Q3piNGREY3kAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
			uniqExternalTxID:  []byte{},
			isSubmitted:       false,
		},
		// invalid: collateral tokenID is not supported
		{
			custodianIncAddress: CUS_INC_ADDRESS_3,
			remoteAddress: map[string]string{
				pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
			},
			depositAmount:     10 * 1e9, // 10 KNC
			collateralTokenID: "ad67cb4d63c9da94aca37fdf2761aadf780ff4a2",
			blockHash:         eCommon.HexToHash("0xfff082e249d4921bd95b96596761683de764c34316db34860b40d8f7a067d4db"),
			txIndex:           0,
			proofStrs:         []string{"+FGgGPhZ7YB/7bFsx8Cz9d0tREoH6Dw2HOVVFVZWYZchx6GAgICAgICAoGn26Vtqf8FBFOyVjEvDAtrNK5YL91jmWIVrM/MI5e34gICAgICAgIA=", "+QLsMLkC6PkC5QGDAQx9uQEAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAACAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAkAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAQAAAAAAAAAAAAAAAAAAABAAAQAAAAAAAAAAAAAAQAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAgAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAPkB2viblK1ny01jydqUrKN/3ydhqt94D/Si+GOg3fJSrRviyJtpwrBo/DeNqpUrp/FjxKEWKPVaTfUjs++gAAAAAAAAAAAAAAAA/ZTEarjc8JKNUROm/qqSV5NQThagAAAAAAAAAAAAAAAA3f5i8QIqYr+NwAfLRmMijHH1I1ugAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAiscjBInoAAD5ATqU3f5i8QIqYr+NwAfLRmMijHH1I1vhoC1LWXk1881n+y7r8dtN68k0zuXHuqcVP5gP2+sudAhOuQEAAAAAAAAAAAAAAAAArWfLTWPJ2pSso3/fJ2Gq33gP9KIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJUC+QAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGcxMlJ3ejRIWGtWQUJnUm5TYjVHZnUxRmFKN2F1bzNmTE5YVkdGaHh4MWRTeXR4SHBXaGJraW1UMU12NVoyb0NNc3NzU1hUVnNhcFk4UUdCWmQySjRtUGlDVHpKQXRNeUN6YjRkRGN5AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="},
			uniqExternalTxID:  []byte{},
			isSubmitted:       false,
		},
	}

	// build expected results
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID: 1 * 1e8,
			DAI_ID: 10 * 1e9,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID: 1 * 1e8,
			DAI_ID: 10 * 1e9,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			DAI_ID: 10 * 1e9,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			DAI_ID: 10 * 1e9,
		})

	expectedRes := &ExpectedResultCustodianDepositV3{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
		},
		numBeaconInsts: 5,
		statusInsts: []string{
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildCustodianDepositActionsV3FromTcs(tcs []TestCaseCustodianDepositV3, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalCustodianDepositActionV3(tc.remoteAddress, tc.blockHash, tc.txIndex, tc.proofStrs, shardID)
		uniqExternalTxID := portalprocessv3.GetUniqExternalTxID(pCommon.ETHChainName, tc.blockHash, tc.txIndex)
		insts = append(insts, instructionForProducer{
			inst: inst,
			optionalData: map[string]interface{}{
				"isSubmitted":      tc.isSubmitted,
				"uniqExternalTxID": uniqExternalTxID,
			},
		})
	}

	return insts
}

/*func (s *PortalTestSuiteV3) TestCustodianDepositCollateralV3() {*/
//fmt.Println("Running TestCustodianDepositCollateralV3 - beacon height 1000 ...")
//bc := s.blockChain
//pm := portal.NewPortalManager()
//beaconHeight := uint64(1000)
//shardHeights := map[byte]uint64{
//0: uint64(1000),
//}
//shardID := byte(0)
//// build test cases and expected results
//testcases, expectedRes := buildTestCaseAndExpectedResultCustodianDepositV3()

//// build actions from testcases
//instsForProducer := buildCustodianDepositActionsV3FromTcs(testcases, shardID)

//// producer instructions
//newInsts, err := producerPortalInstructions(
//bc, beaconHeight, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)

//// process new instructions
//err = processPortalInstructions(
//bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

//for i, inst := range newInsts {
//s.Equal(expectedRes.statusInsts[i], inst[2])
//}

//// check results
//s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
//s.Equal(nil, err)
//s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)

//s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
/*}*/

/*
	Feature 3: Custodians withdraw free collaterals (ETH/ERC20)
*/

type TestCaseCustodianWithdrawV3 struct {
	custodianIncAddress      string
	custodianExternalAddress string
	externalTokenID          string
	amount                   uint64
}

type ExpectedResultCustodianWithdrawV3 struct {
	custodianPool  map[string]*statedb.CustodianState
	numBeaconInsts uint
}

func (s *PortalTestSuiteV3) SetupTestCustodianWithdrawV3() {
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	custodianPool := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodianPool
	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodianPool)
}

func buildTestCaseAndExpectedResultCustodianWithdrawV3() ([]TestCaseCustodianWithdrawV3, *ExpectedResultCustodianWithdrawV3) {
	// build test cases
	//todo: build the eth proof
	testcases := []TestCaseCustodianWithdrawV3{
		// valid: withdrawal amount less than free collateral amount (ETH)
		{
			custodianIncAddress:      CUS_INC_ADDRESS_1,
			custodianExternalAddress: CUS_ETH_ADDRESS_1,
			externalTokenID:          ETH_ID,
			amount:                   1 * 1e9,
		},
		// valid: withdrawal amount equal to free collateral amount (ERC20)
		{
			custodianIncAddress:      CUS_INC_ADDRESS_1,
			custodianExternalAddress: CUS_ETH_ADDRESS_1,
			externalTokenID:          USDT_ID,
			amount:                   500 * 1e6,
		},
		// invalid: withdrawal amount greater than free collateral amount
		{
			custodianIncAddress:      CUS_INC_ADDRESS_2,
			custodianExternalAddress: CUS_ETH_ADDRESS_2,
			externalTokenID:          USDT_ID,
			amount:                   3000 * 1e6,
		},
		// invalid: invalid collateral tokenID
		{
			custodianIncAddress:      CUS_INC_ADDRESS_2,
			custodianExternalAddress: CUS_ETH_ADDRESS_2,
			externalTokenID:          ETH_ID,
			amount:                   1 * 1e9,
		},
	}

	// build expected results
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  9 * 1e9,
			USDT_ID: 0 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  9 * 1e9,
			USDT_ID: 0 * 1e6,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	expectedRes := &ExpectedResultCustodianWithdrawV3{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		numBeaconInsts: 6,
	}
	// todo: check confirm withdraw instructions

	return testcases, expectedRes
}

func buildCustodianWithdrawActionsV3FromTcs(tcs []TestCaseCustodianWithdrawV3, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalCustodianWithdrawActionV3(tc.custodianIncAddress, tc.custodianExternalAddress, tc.externalTokenID, tc.amount, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestCustodianWithdrawCollateralV3() {
	fmt.Println("Running TestCustodianWithdrawCollateralV3 - beacon height 1000 ...")
	bc := s.blockChain
	pm := portal.NewPortalManager()
	beaconHeight := uint64(1000)
	shardHeights := map[byte]uint64{
		0: uint64(1000),
	}
	shardID := byte(0)

	s.SetupTestCustodianWithdrawV3()

	// build test cases and expected results
	testcases, expectedRes := buildTestCaseAndExpectedResultCustodianWithdrawV3()

	// build actions from testcases
	instsForProducer := buildCustodianWithdrawActionsV3FromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

	// check results
	s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)
	s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 4: Users create porting request
*/
type TestCaseRequestPorting struct {
	portingID     string
	incAddressStr string
	pTokenID      string
	portingAmount uint64
	portingFee    uint64
	isExisted     bool
}

type ExpectedResultPortingRequest struct {
	waitingPortingRes map[string]*statedb.WaitingPortingRequest
	custodianPool     map[string]*statedb.CustodianState
	numBeaconInsts    uint
}

func (s *PortalTestSuiteV3) SetupTestPortingRequest() {
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	custodianPool := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodianPool
	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodianPool)
}

func buildTestCaseAndExpectedResultPortingRequest() ([]TestCaseRequestPorting, *ExpectedResultPortingRequest) {
	beaconHeight := uint64(1001)
	shardHeight := uint64(1001)
	shardID := byte(0)
	// build test cases
	testcases := []TestCaseRequestPorting{
		// valid porting request: match to one custodian, 0.01% porting fee
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    2000000,
			isExisted:     false,
		},
		// valid porting request: match to many custodians
		{
			portingID:     "porting-bnb-2",
			incAddressStr: USER_INC_ADDRESS_2,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 150 * 1e9,
			portingFee:    300000000,
			isExisted:     false,
		},
		// invalid porting request with duplicate porting ID
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    2000000,
			isExisted:     true,
		},
		// invalid porting request with invalid porting fee
		{
			portingID:     "porting-bnb-3",
			incAddressStr: USER_INC_ADDRESS_3,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			portingFee:    99900,
			isExisted:     false,
		},
		// invalid porting request: total collaterals of the custodians are not enough for the porting amount
		{
			portingID:     "porting-btc-4",
			incAddressStr: USER_INC_ADDRESS_3,
			pTokenID:      pCommon.PortalBTCIDStr,
			portingAmount: 10 * 1e9,
			portingFee:    1000000000,
			isExisted:     false,
		},
	}

	// build expected results
	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1460 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(map[string]map[string]uint64{
		pCommon.PortalBNBIDStr: {
			USDT_ID: 540 * 1e6,
		},
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-1", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 1*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 1 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 40000000,
				},
			},
		}, 2000000, beaconHeight, shardHeight, shardID)

	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-2").String()
	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-2", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_2, 150*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian1.GetIncognitoAddress(),
				RemoteAddress:          custodian1.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 137500000000,
				LockedAmountCollateral: 1000000000000,
				LockedTokenCollaterals: map[string]uint64{
					ETH_ID:  10000000000,
					USDT_ID: 500000000,
				},
			},
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 12500000000,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 500000000,
				},
			},
		}, 300000000, beaconHeight, shardHeight, shardID)

	expectedRes := &ExpectedResultPortingRequest{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{
			wPortingReqKey1: wPortingRequest1,
			wPortingReqKey2: wPortingRequest2,
		},
		numBeaconInsts: 5,
	}

	return testcases, expectedRes
}

func buildRequestPortingActionsFromTcs(tcs []TestCaseRequestPorting, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	//todo: shardHeight
	for _, tc := range tcs {
		inst := buildPortalUserRegisterAction(
			tc.portingID, tc.incAddressStr, tc.pTokenID, tc.portingAmount, tc.portingFee, shardID, 1001)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: map[string]interface{}{"isExistPortingID": tc.isExisted},
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestPortingRequest() {
	fmt.Println("Running TestPortingRequest - beacon height 1001 ...")
	bc := s.blockChain
	pm := portal.NewPortalManager()
	beaconHeight := uint64(1001)
	shardHeights := map[byte]uint64{
		0: uint64(1001),
	}
	shardID := byte(0)

	s.SetupTestPortingRequest()

	// build test cases
	testcases, expectedRes := buildTestCaseAndExpectedResultPortingRequest()

	// build actions from testcases
	instsForProducer := buildRequestPortingActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight-1, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

	// check results
	s.Equal(expectedRes.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)
	s.Equal(expectedRes.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedRes.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 5: Users submit proof to request pTokens after sending public tokens to custodians
*/
type TestCaseRequestPtokens struct {
	portingID     string
	incAddressStr string
	pTokenID      string
	portingAmount uint64

	txID         string
	portingProof string
}

type ExpectedResultRequestPTokens struct {
	waitingPortingRes map[string]*statedb.WaitingPortingRequest
	custodianPool     map[string]*statedb.CustodianState
	numBeaconInsts    uint
	statusInsts       []string
}

func (s *PortalTestSuiteV3) SetupTestRequestPtokens() {
	beaconHeight := uint64(1002)
	shardHeight := uint64(1002)
	shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1460 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(map[string]map[string]uint64{
		pCommon.PortalBNBIDStr: {
			USDT_ID: 540 * 1e6,
		},
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-1", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 1*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 1 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 40000000,
				},
			},
		}, 2000000, beaconHeight, shardHeight, shardID)

	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-2").String()
	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-2", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_2, 150*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian1.GetIncognitoAddress(),
				RemoteAddress:          custodian1.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 137500000000,
				LockedAmountCollateral: 1000000000000,
				LockedTokenCollaterals: map[string]uint64{
					ETH_ID:  10000000000,
					USDT_ID: 500000000,
				},
			},
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 12500000000,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 500000000,
				},
			},
		}, 300000000, beaconHeight, shardHeight, shardID)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey1: wPortingRequest1,
		wPortingReqKey2: wPortingRequest2,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests

	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = portalprocessv3.CloneWPortingRequests(wPortingRequests)
}

func buildTestCaseAndExpectedResultRequestPTokens() ([]TestCaseRequestPtokens, *ExpectedResultRequestPTokens) {
	// build test cases
	testcases := []TestCaseRequestPtokens{
		// valid request ptokens: porting request matching one custodian
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			txID:          "A71E509490912AA82E138361D63253601C4CD20F9508F6880738FEF495CE42C6",
			portingProof:  "",
		},
		// valid request ptokens: porting request matching many custodians
		{
			portingID:     "porting-bnb-2",
			incAddressStr: USER_INC_ADDRESS_2,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 150 * 1e9,
			txID:          "37B26ADEFB57534C9F38C07D948B8241099E02A0D3CAFA60B4400BD200197B52",
			portingProof:  "",
		},
		// invalid request ptokens: resubmit porting proof
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			txID:          "A71E509490912AA82E138361D63253601C4CD20F9508F6880738FEF495CE42C6",
			portingProof:  "",
		},
		// invalid request ptokens: invalid porting proof
		{
			portingID:     "porting-bnb-1",
			incAddressStr: USER_INC_ADDRESS_1,
			pTokenID:      pCommon.PortalBNBIDStr,
			portingAmount: 1 * 1e9,
			txID:          "B186FDFBB781D118E25FB37E85BEB83865C3C575485DC43C28220DD7F91BD70A",
			portingProof:  "",
		},
	}

	// build porting proof for testcases
	for i, tc := range testcases {
		proof, err := bnb.BuildProofFromTxID(tc.txID, BNB_NODE_URL)
		if err != nil {
			fmt.Errorf("err build proof: %v", err)
		}
		testcases[i].portingProof = proof
	}

	// build expected results
	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 137500000000,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1460 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 540 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 13500000000,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	expectedRes := &ExpectedResultRequestPTokens{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{},
		numBeaconInsts:    4,
		statusInsts: []string{
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildRequestPtokensActionsFromTcs(tcs []TestCaseRequestPtokens, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalUserReqPTokenAction(
			tc.portingID, tc.incAddressStr, tc.pTokenID, tc.portingAmount, tc.portingProof, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestRequestPtokens() {
	// TODO: update new proof for new memo
	//fmt.Println("Running TestRequestPtokens - beacon height 1002 ...")
	//bc := s.blockChain
	//pm := portal.NewPortalManager()
	//beaconHeight := uint64(1002)
	//shardHeights := map[byte]uint64{
	//	0: uint64(1002),
	//}
	//shardID := byte(0)
	//
	//s.SetupTestRequestPtokens()
	//
	//// build test cases
	//testcases, expectedResult := buildTestCaseAndExpectedResultRequestPTokens()
	//
	//// build actions from testcases
	//instsForProducer := buildRequestPtokensActionsFromTcs(testcases, shardID)
	//
	//// producer instructions
	//newInsts, err := producerPortalInstructions(
	//	bc, beaconHeight-1, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)
	//s.Equal(nil, err)
	//
	//// process new instructions
	//err = processPortalInstructions(
	//	bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)
	//
	//// check results
	//s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	//s.Equal(nil, err)
	//
	//for i, inst := range newInsts {
	//	s.Equal(expectedResult.statusInsts[i], inst[2], i)
	//}
	//
	//s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	//s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)
	//
	//s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)

	return
}

/*
	Feature 6: Users redeem request
*/
type TestCaseRequestRedeemV3 struct {
	uniqueRedeemID          string
	tokenID                 string
	redeemAmount            uint64
	redeemerIncAddressStr   string
	remoteAddress           string
	redeemFee               uint64
	redeemerExternalAddress string
	isExisted               bool
}

type ExpectedResultRequestRedeemV3 struct {
	waitingPortingRes map[string]*statedb.WaitingPortingRequest
	custodianPool     map[string]*statedb.CustodianState
	waitingRedeemReq  map[string]*statedb.RedeemRequest
	numBeaconInsts    uint
	statusInsts       []string
}

func (s *PortalTestSuiteV3) SetupTestRequestRedeemV3() {
	beaconHeight := uint64(1003)
	shardHeight := uint64(1003)
	shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 137500000000,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1740 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 13500000000,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey3: wPortingRequest3,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests

	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = portalprocessv3.CloneWPortingRequests(wPortingRequests)
}

func buildTestCaseAndExpectedResultRequestRedeemV3() ([]TestCaseRequestRedeemV3, *ExpectedResultRequestRedeemV3) {
	beaconHeight := uint64(1003)
	shardHeight := uint64(1003)
	shardID := byte(0)
	// build test cases
	testcases := []TestCaseRequestRedeemV3{
		// valid request redeem: matching one custodian
		{
			uniqueRedeemID:          "redeem-bnb-1",
			tokenID:                 pCommon.PortalBNBIDStr,
			redeemAmount:            1 * 1e9,
			redeemerIncAddressStr:   USER_INC_ADDRESS_1,
			remoteAddress:           USER_BNB_ADDRESS_1,
			redeemFee:               2000000,
			redeemerExternalAddress: USER_ETH_ADDRESS_1,
			isExisted:               false,
		},
		// valid request redeem: matching many custodians
		{
			uniqueRedeemID:          "redeem-bnb-2",
			tokenID:                 pCommon.PortalBNBIDStr,
			redeemAmount:            140 * 1e9,
			redeemerIncAddressStr:   USER_INC_ADDRESS_1,
			remoteAddress:           USER_BNB_ADDRESS_1,
			redeemFee:               280000000,
			redeemerExternalAddress: USER_ETH_ADDRESS_1,
			isExisted:               false,
		},
		// valid request redeem: redeem amount exceeds total available public token amount
		{
			uniqueRedeemID:          "redeem-bnb-3",
			tokenID:                 pCommon.PortalBNBIDStr,
			redeemAmount:            130 * 1e9,
			redeemerIncAddressStr:   USER_INC_ADDRESS_1,
			remoteAddress:           USER_BNB_ADDRESS_1,
			redeemFee:               280000000,
			redeemerExternalAddress: USER_ETH_ADDRESS_1,
			isExisted:               false,
		},
		// invalid request redeem: duplicate redeem ID
		{
			uniqueRedeemID:          "redeem-bnb-1",
			tokenID:                 pCommon.PortalBNBIDStr,
			redeemAmount:            1 * 1e9,
			redeemerIncAddressStr:   USER_INC_ADDRESS_1,
			remoteAddress:           USER_BNB_ADDRESS_1,
			redeemFee:               2000000,
			redeemerExternalAddress: USER_ETH_ADDRESS_1,
			isExisted:               true,
		},
		// invalid request redeem: redeem fee is less than min redeem fee
		{
			uniqueRedeemID:          "redeem-bnb-4",
			tokenID:                 pCommon.PortalBNBIDStr,
			redeemAmount:            1 * 1e9,
			redeemerIncAddressStr:   USER_INC_ADDRESS_1,
			remoteAddress:           USER_BNB_ADDRESS_1,
			redeemFee:               1900000,
			redeemerExternalAddress: USER_ETH_ADDRESS_1,
			isExisted:               false,
		},
	}

	// build expected results
	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 137500000000,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1740 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 13500000000,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	// waiting redeem requests
	wRedeemReqKey1 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-1").String()
	wRedeemReq1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	wRedeemReqKey2 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-1").String()
	wRedeemReq2 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-2", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		140*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	wRedeemReqKey3 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-1").String()
	wRedeemReq3 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		130*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	expectedRes := &ExpectedResultRequestRedeemV3{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{
			wPortingReqKey3: wPortingRequest3,
		},
		waitingRedeemReq: map[string]*statedb.RedeemRequest{
			wRedeemReqKey1: wRedeemReq1,
			wRedeemReqKey2: wRedeemReq2,
			wRedeemReqKey3: wRedeemReq3,
		},
		numBeaconInsts: 5,
		statusInsts: []string{
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildRequestRedeemActionsFromTcs(tcs []TestCaseRequestRedeemV3, shardID byte, shardHeight uint64) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalRequestRedeemActionV3(
			tc.uniqueRedeemID, tc.tokenID, tc.redeemAmount, tc.redeemerIncAddressStr, tc.remoteAddress, tc.redeemFee, tc.redeemerExternalAddress, shardID, shardHeight)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: map[string]interface{}{"isExistRedeemID": tc.isExisted},
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestRequestRedeemV3() {
	fmt.Println("Running TestRequestRedeemV3 - beacon height 1003 ...")
	bc := s.blockChain
	pm := portal.NewPortalManager()
	beaconHeight := uint64(1003)
	shardHeight := uint64(1003)
	shardHeights := map[byte]uint64{
		0: uint64(1002),
	}
	shardID := byte(0)

	s.SetupTestRequestRedeemV3()

	// build test cases
	testcases, expectedResult := buildTestCaseAndExpectedResultRequestRedeemV3()

	// build actions from testcases
	instsForProducer := buildRequestRedeemActionsFromTcs(testcases, shardID, shardHeight)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight-1, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)

	for i, inst := range newInsts {
		s.Equal(expectedResult.statusInsts[i], inst[2], "Instruction index %v", i)
	}

	s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 7: Custodians request matching waiting redeem requests
*/
type TestCaseRequestMatchingWRedeemV3 struct {
	uniqueRedeemID   string
	cusIncAddressStr string
}

type ExpectedResultRequestMatchingWRedeemV3 struct {
	waitingPortingRes map[string]*statedb.WaitingPortingRequest
	custodianPool     map[string]*statedb.CustodianState
	waitingRedeemReq  map[string]*statedb.RedeemRequest
	numBeaconInsts    uint
	statusInsts       []string
}

func (s *PortalTestSuiteV3) SetupTestRequestMatchingWRedeemV3() {
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 137500000000, // 137.5 BNB
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1740 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 13500000000, // 13.5 BNB
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	// waiting redeem requests
	wRedeemReqKey1 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-1").String()
	wRedeemReq1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	wRedeemReqKey2 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-2").String()
	wRedeemReq2 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-2", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		140*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	wRedeemReqKey3 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-3").String()
	wRedeemReq3 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-3", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		130*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}
	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey3: wPortingRequest3,
	}
	wRedeemRequests := map[string]*statedb.RedeemRequest{
		wRedeemReqKey1: wRedeemReq1,
		wRedeemReqKey2: wRedeemReq2,
		wRedeemReqKey3: wRedeemReq3,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
	s.currentPortalStateForProducer.WaitingRedeemRequests = wRedeemRequests

	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = portalprocessv3.CloneWPortingRequests(wPortingRequests)
	s.currentPortalStateForProcess.WaitingRedeemRequests = portalprocessv3.CloneRedeemRequests(wRedeemRequests)
}

func buildTestCaseAndExpectedResultRequestMatchingWRedeemV3() ([]TestCaseRequestMatchingWRedeemV3, *ExpectedResultRequestMatchingWRedeemV3) {
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardID := byte(0)
	// build test cases
	testcases := []TestCaseRequestMatchingWRedeemV3{
		// valid request matching waiting redeem: matching full redeem amount
		{
			uniqueRedeemID:   "redeem-bnb-1",
			cusIncAddressStr: CUS_INC_ADDRESS_2,
		},
		// valid request matching waiting redeem: matching a part of redeem amount
		{
			uniqueRedeemID:   "redeem-bnb-2",
			cusIncAddressStr: CUS_INC_ADDRESS_1,
		},
		// invalid request matching waiting redeem: duplicate waiting redeem request ID
		{
			uniqueRedeemID:   "redeem-bnb-1",
			cusIncAddressStr: CUS_INC_ADDRESS_2,
		},
		// invalid request matching waiting redeem: invalid custodian
		{
			uniqueRedeemID:   "redeem-bnb-3",
			cusIncAddressStr: CUS_INC_ADDRESS_3,
		},
	}

	// build expected results
	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1740 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 12500000000,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	// waiting redeem requests
	wRedeemReqKey1 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-1").String()
	wRedeemReq1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 1*1e9),
		},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	wRedeemReqKey2 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-2").String()
	wRedeemReq2 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-2", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		140*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 137500000000),
		},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	wRedeemReqKey3 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-3").String()
	wRedeemReq3 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-3", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		130*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	expectedRes := &ExpectedResultRequestMatchingWRedeemV3{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{
			wPortingReqKey3: wPortingRequest3,
		},
		waitingRedeemReq: map[string]*statedb.RedeemRequest{
			wRedeemReqKey1: wRedeemReq1,
			wRedeemReqKey2: wRedeemReq2,
			wRedeemReqKey3: wRedeemReq3,
		},
		numBeaconInsts: 4,
		statusInsts: []string{
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildRequestMatchingWRedeemV3ActionsFromTcs(tcs []TestCaseRequestMatchingWRedeemV3, shardID byte, shardHeight uint64) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalRequestMatchingWRedeemActionV3(
			tc.uniqueRedeemID, tc.cusIncAddressStr, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestRequestMatchingWRedeemV3() {
	fmt.Println("Running TestRequestMatchingWRedeemV3 - beacon height 1004 ...")
	bc := s.blockChain
	pm := portal.NewPortalManager()
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardHeights := map[byte]uint64{
		0: uint64(1004),
	}
	shardID := byte(0)

	s.SetupTestRequestMatchingWRedeemV3()

	// build test cases
	testcases, expectedResult := buildTestCaseAndExpectedResultRequestMatchingWRedeemV3()

	// build actions from testcases
	instsForProducer := buildRequestMatchingWRedeemV3ActionsFromTcs(testcases, shardID, shardHeight)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight-1, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)

	for i, inst := range newInsts {
		s.Equal(expectedResult.statusInsts[i], inst[2], "Instruction index %v", i)
	}

	s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 8: auto pick up custodians for waiting redeem requests
*/
type ExpectedResultPickMoreCustodianForWRequestRedeem struct {
	custodianPool        map[string]*statedb.CustodianState
	waitingPortingRes    map[string]*statedb.WaitingPortingRequest
	waitingRedeemRequest map[string]*statedb.RedeemRequest
	matchedRedeemRequest map[string]*statedb.RedeemRequest
	numBeaconInsts       uint
	statusInsts          []string
}

func buildExpectedResultPickMoreCustodianForWRequestRedeem() *ExpectedResultPickMoreCustodianForWRequestRedeem {
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardID := byte(0)

	// build expected results
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 7500000000, // 7.5 BNB
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1740 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 12500000000, // 12.5 BNB
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	// matched redeem requests: 3 => 1 => 2
	matchedRedeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
	matchedRedeemReq1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 1*1e9),
		},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	matchedRedeemReqKey3 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-3").String()
	matchedRedeemReq3 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-3", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		130*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 130*1e9),
		},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	expectedRes := &ExpectedResultPickMoreCustodianForWRequestRedeem{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{
			wPortingReqKey3: wPortingRequest3,
		},
		waitingRedeemRequest: map[string]*statedb.RedeemRequest{},
		matchedRedeemRequest: map[string]*statedb.RedeemRequest{
			matchedRedeemReqKey1: matchedRedeemReq1,
			matchedRedeemReqKey3: matchedRedeemReq3,
		},
		numBeaconInsts: 4,
		statusInsts: []string{
			pCommon.PortalProducerInstSuccessChainStatus,
			pCommon.PortalProducerInstSuccessChainStatus,
			pCommon.PortalProducerInstFailedChainStatus,
			pCommon.PortalRedeemReqCancelledByLiquidationChainStatus,
		},
	}

	return expectedRes
}

/*func (s *PortalTestSuiteV3) TestAutoPickMoreCustodiansForWRedeemRequest() {*/
//fmt.Println("Running TestRequestPtokens - beacon height 1020 ...")
//bc := s.blockChain
//beaconHeight := uint64(1019)
//shardHeights := map[byte]uint64{
//0: 1019,
//}
//pm := portal.NewPortalManager()

//s.SetupTestRequestMatchingWRedeemV3()

//wRedeem := s.currentPortalStateForProducer.WaitingRedeemRequests

//for key, req := range wRedeem {
//fmt.Printf("key %v - redeemID %v\n", key, req.GetUniqueRedeemID())
//}

//// build test cases
//expectedResult := buildExpectedResultPickMoreCustodianForWRequestRedeem()

//processor := portalprocessv3.GetPortalInstProcessorByMetaType(pm.PortalInstProcessorsV3, metadata.PortalPickMoreCustodianForRedeemMeta)
//newInsts, err := processor.BuildNewInsts(bc, "", 0, &s.currentPortalStateForProducer, beaconHeight, shardHeights, s.portalParams, nil)
//s.Equal(nil, err)

//// process new instructions
//err = processPortalInstructions(
//bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

//// check results
//s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
//s.Equal(nil, err)

//for i, inst := range newInsts {
//s.Equal(expectedResult.statusInsts[i], inst[2], "Instruction index %v", i)
//}

//s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
//s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)
//s.Equal(expectedResult.waitingRedeemRequest, s.currentPortalStateForProducer.WaitingRedeemRequests)
//s.Equal(expectedResult.matchedRedeemRequest, s.currentPortalStateForProducer.MatchedRedeemRequests)

//s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
/*}*/

/*
	Feature 9: Custodians request unlock collaterals
*/
type TestCaseRequestUnlockCollateralsV3 struct {
	uniqueRedeemID      string
	portalTokenID       string
	custodianIncAddress string
	redeemAmount        uint64
	redeemProof         string
	externalTxID        string
}

type ExpectedResultRequestUnlockCollateralsV3 struct {
	waitingPortingRes    map[string]*statedb.WaitingPortingRequest
	custodianPool        map[string]*statedb.CustodianState
	matchedRedeemRequest map[string]*statedb.RedeemRequest
	numBeaconInsts       uint
	statusInsts          []string
}

func (s *PortalTestSuiteV3) SetupTestRequestUnlockCollateralsV3() {
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0, // 137.5 BNB
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1740 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 10000000000, // 13.5 BNB
		})
	custodian2.SetTotalCollateral(100 * 1e9)
	custodian2.SetFreeCollateral(0 * 1e9)
	custodian2.SetLockedAmountCollateral(map[string]uint64{
		pCommon.PortalBNBIDStr: 100 * 1e9,
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 50 * 1e9,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	// matched redeem requests
	matchedRedeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
	matchedRedeemReq1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 1*1e9),
		},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	matchedRedeemReqKey2 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-2").String()
	matchedRedeemReq2 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-2", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		140*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 136500000000),
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 3500000000),
		},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}
	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey3: wPortingRequest3,
	}
	matchedRedeemRequests := map[string]*statedb.RedeemRequest{
		matchedRedeemReqKey1: matchedRedeemReq1,
		matchedRedeemReqKey2: matchedRedeemReq2,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
	s.currentPortalStateForProducer.MatchedRedeemRequests = matchedRedeemRequests

	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = portalprocessv3.CloneWPortingRequests(wPortingRequests)
	s.currentPortalStateForProcess.MatchedRedeemRequests = portalprocessv3.CloneRedeemRequests(matchedRedeemRequests)
}

func buildTestCaseAndExpectedResultRequestUnlockCollateralsV3() ([]TestCaseRequestUnlockCollateralsV3, *ExpectedResultRequestUnlockCollateralsV3) {
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardID := byte(0)
	// build test cases
	testcases := []TestCaseRequestUnlockCollateralsV3{
		// valid request unlock collaterals: from redeem request match to one custodian
		{
			uniqueRedeemID:      "redeem-bnb-1",
			portalTokenID:       pCommon.PortalBNBIDStr,
			custodianIncAddress: CUS_INC_ADDRESS_1,
			redeemAmount:        1 * 1e9,
			redeemProof:         "",
			externalTxID:        "A0D740A9C619D6F6E6A9983736CC0E0811CECAE52493708C99F9447C66D04F17",
		},
		// valid request unlock collaterals: from redeem request match to many custodians
		{
			uniqueRedeemID:      "redeem-bnb-2",
			portalTokenID:       pCommon.PortalBNBIDStr,
			custodianIncAddress: CUS_INC_ADDRESS_1,
			redeemAmount:        136500000000,
			redeemProof:         "",
			externalTxID:        "E1663D62152C6A5527081963441C17137ED99FE8C4B0CF549CF54CE9B4B49746",
		},
		// invalid request unlock collaterals: resubmit the proof for the same redeem request
		{
			uniqueRedeemID:      "redeem-bnb-1",
			portalTokenID:       pCommon.PortalBNBIDStr,
			custodianIncAddress: CUS_INC_ADDRESS_1,
			redeemAmount:        1 * 1e9,
			redeemProof:         "",
			externalTxID:        "A0D740A9C619D6F6E6A9983736CC0E0811CECAE52493708C99F9447C66D04F17",
		},
		// invalid request unlock collaterals: resubmit the proof for the other redeem request
		{
			uniqueRedeemID:      "redeem-bnb-2",
			portalTokenID:       pCommon.PortalBNBIDStr,
			custodianIncAddress: CUS_INC_ADDRESS_1,
			redeemAmount:        1 * 1e9,
			redeemProof:         "",
			externalTxID:        "A0D740A9C619D6F6E6A9983736CC0E0811CECAE52493708C99F9447C66D04F17",
		},
		// invalid request unlock collaterals: invalid redeem memo
		{
			uniqueRedeemID:      "redeem-bnb-2",
			portalTokenID:       pCommon.PortalBNBIDStr,
			custodianIncAddress: CUS_INC_ADDRESS_2,
			redeemAmount:        3500000000,
			redeemProof:         "",
			externalTxID:        "8DE6AC94DDE88658E60F6AEBF2FD2AC79E61185EA58772B8BAE00B4102CA0EDE",
		},
		// valid request unlock collaterals: custodian that matching to one waiting porting
		{
			uniqueRedeemID:      "redeem-bnb-2",
			portalTokenID:       pCommon.PortalBNBIDStr,
			custodianIncAddress: CUS_INC_ADDRESS_2,
			redeemAmount:        3500000000,
			redeemProof:         "",
			externalTxID:        "82D3A930C73266F134A42A0EECD496777D876473837983C975C2963FE5147C47",
		},
	}

	// build expected results
	// custodian state after matching porting requests
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  0,
				USDT_ID: 0,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 362962962,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1637037038,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 10000000000,
		})
	custodian2.SetTotalCollateral(100 * 1e9)
	custodian2.SetFreeCollateral(50 * 1e9)
	custodian2.SetLockedAmountCollateral(map[string]uint64{
		pCommon.PortalBNBIDStr: 50 * 1e9,
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 50 * 1e9,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	expectedRes := &ExpectedResultRequestUnlockCollateralsV3{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{
			wPortingReqKey3: wPortingRequest3,
		},
		matchedRedeemRequest: map[string]*statedb.RedeemRequest{
			//matchedRedeemReqKey2: matchedRedeemReq2,
		},
		numBeaconInsts: 6,
		statusInsts: []string{
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
		},
	}

	return testcases, expectedRes
}

func buildRequestUnlockCollateralsV3ActionsFromTcs(tcs []TestCaseRequestUnlockCollateralsV3, shardID byte, shardHeight uint64) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		tc.redeemProof, _ = bnb.BuildProofFromTxID(tc.externalTxID, BNB_NODE_URL)
		inst := buildPortalRequestUnlockCollateralsActionV3(
			tc.uniqueRedeemID, tc.portalTokenID, tc.custodianIncAddress, tc.redeemAmount, tc.redeemProof, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func (s *PortalTestSuiteV3) TestRequestUnlockCollateralsV3() {
	fmt.Println("Running TestRequestUnlockCollateralsV3 - beacon height 1004 ...")
	bc := s.blockChain
	pm := portal.NewPortalManager()
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardHeights := map[byte]uint64{
		0: 1004,
	}
	shardID := byte(0)

	s.SetupTestRequestUnlockCollateralsV3()

	// build test cases
	testcases, expectedResult := buildTestCaseAndExpectedResultRequestUnlockCollateralsV3()

	// build actions from testcases
	instsForProducer := buildRequestUnlockCollateralsV3ActionsFromTcs(testcases, shardID, shardHeight)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight-1, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)
	s.Equal(nil, err)

	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)

	for i, inst := range newInsts {
		s.Equal(expectedResult.statusInsts[i], inst[2], "Instruction index %v", i)
	}

	s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)
	s.Equal(expectedResult.matchedRedeemRequest, s.currentPortalStateForProducer.MatchedRedeemRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 10:
	+ auto-unlockCollaterals: when porting request expired
*/
type ExpectedResultExpiredPortingRequest struct {
	waitingPortingRes map[string]*statedb.WaitingPortingRequest
	custodianPool     map[string]*statedb.CustodianState
	numBeaconInsts    uint
	statusInsts       []string
}

func (s *PortalTestSuiteV3) SetupTestExpiredPortingRequest() {
	beaconHeight := uint64(1002)
	shardHeight := uint64(1002)
	shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1460 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(map[string]map[string]uint64{
		pCommon.PortalBNBIDStr: {
			USDT_ID: 540 * 1e6,
		},
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-1", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 1*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 1 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 40000000,
				},
			},
		}, 2000000, beaconHeight, shardHeight, shardID)

	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-2").String()
	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-2", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_2, 150*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian1.GetIncognitoAddress(),
				RemoteAddress:          custodian1.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 137500000000,
				LockedAmountCollateral: 1000000000000,
				LockedTokenCollaterals: map[string]uint64{
					ETH_ID:  10000000000,
					USDT_ID: 500000000,
				},
			},
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 12500000000,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 500000000,
				},
			},
		}, 300000000, beaconHeight, shardHeight, shardID)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey1: wPortingRequest1,
		wPortingReqKey2: wPortingRequest2,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests

	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = portalprocessv3.CloneWPortingRequests(wPortingRequests)
}

func buildExpectedResultExpiredPortingRequest() *ExpectedResultExpiredPortingRequest {
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  0,
				USDT_ID: 0,
			},
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(map[string]map[string]uint64{
		pCommon.PortalBNBIDStr: {
			USDT_ID: 0,
		},
	})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}

	res := &ExpectedResultExpiredPortingRequest{
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{},
		custodianPool:     custodians,
		numBeaconInsts:    2,
		statusInsts:       nil,
	}
	return res
}

func (s *PortalTestSuiteV3) TestExpiredPortingRequest() {
	fmt.Println("Running TestAutoLiquidationCustodian - beacon height 31602 ...")
	bc := s.blockChain
	beaconHeight := uint64(3162) // ~ after 24 hours from porting request
	shardHeights := map[byte]uint64{
		0: 3162,
	}
	pm := portal.NewPortalManager()

	s.SetupTestExpiredPortingRequest()
	expectedResult := buildExpectedResultExpiredPortingRequest()

	processor := portalprocessv3.GetPortalInstProcessorByMetaType(pm.PortalInstProcessorsV3, metadata.PortalExpiredWaitingPortingReqMeta)
	newInsts, err := processor.BuildNewInsts(bc, "", 0, &s.currentPortalStateForProducer, beaconHeight, shardHeights, s.portalParams, nil)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)
	s.Equal(nil, err)

	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}

/*
	Feature 11:
	+ auto-liquidation: the custodians don't send back public token to the users
*/
/*
func (s *PortalTestSuiteV3) SetupTestAutoLiquidation() {
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
	custodianKey4 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_4).String()
	custodianKey5 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress5").String()
	custodianKey6 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress6").String()
	custodianKey7 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress7").String()
	custodian1 := statedb.NewCustodianStateWithValue(
		CUS_INC_ADDRESS_1, 7000*1e9, 6680*1e9,
		map[string]uint64{
			pCommon.PortalBNBIDStr: 8 * 1e9,
		},
		map[string]uint64{
			pCommon.PortalBNBIDStr: 320 * 1e9,
		},
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		},
		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
	custodian2 := statedb.NewCustodianStateWithValue(
		CUS_INC_ADDRESS_2, 1000*1e9, 976000000000,
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0.6 * 1e9,
		},
		map[string]uint64{
			pCommon.PortalBNBIDStr: 24 * 1e9,
		},
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		},
		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
	custodian3 := statedb.NewCustodianStateWithValue(
		CUS_INC_ADDRESS_3, 200*1e9, 0,
		map[string]uint64{
			pCommon.PortalBTCIDStr: 0.1 * 1e9,
		},
		map[string]uint64{
			pCommon.PortalBTCIDStr: 200 * 1e9,
		},
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
		},
		map[string]uint64{}, map[string]uint64{
			ETH_ID: 5 * 1e9,
		}, map[string]uint64{
			ETH_ID: 3 * 1e9,
		}, map[string]map[string]uint64{
			pCommon.PortalBTCIDStr: {
				ETH_ID: 2 * 1e9,
			},
		})
	custodian4 := statedb.NewCustodianStateWithValue(
		CUS_INC_ADDRESS_4, 0, 0,
		map[string]uint64{
			pCommon.PortalBNBIDStr: 20 * 1e9,
		},
		map[string]uint64{},
		map[string]string{
			pCommon.PortalBNBIDStr: "bnbAddress4",
		},
		map[string]uint64{}, map[string]uint64{
			ETH_ID: 5 * 1e9,
		}, map[string]uint64{
			ETH_ID: 4 * 1e9,
		}, map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID: 1e9,
			},
		})
	custodian5 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress5", 20*1e9, 20*1e9,
		map[string]uint64{},
		map[string]uint64{},
		map[string]string{
			pCommon.PortalBNBIDStr: "bnbAddress5",
		},
		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
	custodian6 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress6", 10*1e9, 10*1e9,
		map[string]uint64{},
		map[string]uint64{},
		map[string]string{
			pCommon.PortalBNBIDStr: "bnbAddress6",
		},
		map[string]uint64{}, map[string]uint64{
			ETH_ID: 1 * 1e9,
		}, map[string]uint64{
			ETH_ID: 1 * 1e9,
		}, map[string]map[string]uint64{})
	custodian7 := statedb.NewCustodianStateWithValue(
		"custodianIncAddress7", 0, 0,
		map[string]uint64{},
		map[string]uint64{},
		map[string]string{
			pCommon.PortalBNBIDStr: "bnbAddress7",
		},
		map[string]uint64{}, map[string]uint64{
			ETH_ID: 2 * 1e9,
		}, map[string]uint64{
			ETH_ID: 2 * 1e9,
		}, map[string]map[string]uint64{})
	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
		custodianKey4: custodian4,
		custodianKey5: custodian5,
		custodianKey6: custodian6,
		custodianKey7: custodian7,
	}
	// redeem match prv only
	redeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
	redeemRequest1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, "userCUS_BNB_ADDRESS_1", 8.3*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 8*1e9),
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 0.3*1e9),
		}, 4600000, 1000, common.Hash{}, 0, 1000, "f7E20F75782279547ad1eD99d37f020dF1028d07")
	// redeem match prv and eth
	redeemReqKey2 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-btc-2").String()
	redeemRequest2 := statedb.NewRedeemRequestWithValue(
		"redeem-btc-2", pCommon.PortalBTCIDStr,
		USER_INC_ADDRESS_2, "userCUS_BTC_ADDRESS_2", 0.05*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_3, CUS_BTC_ADDRESS_3, 0.05*1e9),
		}, 30000000, 1500, common.Hash{}, 0, 1000, "f7E20F75782279547ad1eD99d37f020dF1028d07")
	// redeem match eth only
	redeemReqKey3 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-2").String()
	redeemRequest3 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-2", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_4, "userCUS_BNB_ADDRESS_2", 10*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_4, CUS_BNB_ADDRESS_3, 10*1e9),
		}, 30000000, 1500, common.Hash{}, 0, 1000, "f7E20F75782279547ad1eD99d37f020dF1028d07")
	matchedRedeemRequest := map[string]*statedb.RedeemRequest{
		redeemReqKey1: redeemRequest1,
		redeemReqKey2: redeemRequest2,
		redeemReqKey3: redeemRequest3,
	}
	wRedeemReqKey3 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-3").String()
	wRedeemRequest3 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-3", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, "userCUS_BNB_ADDRESS_1", 0.1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 0.1*1e9),
		}, 4600000, 1500, common.Hash{}, 0, 1000, "f7E20F75782279547ad1eD99d37f020dF1028d07")
	wRedeemRequests := map[string]*statedb.RedeemRequest{
		wRedeemReqKey3: wRedeemRequest3,
	}
	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-1", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_5, 0.5*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             "custodianIncAddress5",
				RemoteAddress:          "bnbAddress5",
				Amount:                 0.5 * 1e9,
				LockedAmountCollateral: 20 * 1e9,
			},
		}, 2000000, 1500, 1500, 0)
	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-2").String()
	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-2", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_6, 10*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             "custodianIncAddress6",
				RemoteAddress:          "bnbAddress6",
				Amount:                 10 * 1e9,
				LockedAmountCollateral: 10 * 1e9,
				LockedTokenCollaterals: map[string]uint64{
					ETH_ID: 0.975 * 1e9,
				},
			},
		}, 2000000, 1500, 1500, 0)
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 25*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:    "custodianIncAddress7",
				RemoteAddress: "bnbAddress7",
				Amount:        25 * 1e9,
				LockedTokenCollaterals: map[string]uint64{
					ETH_ID: 1.25 * 1e9,
				},
			},
		}, 2000000, 1500, 1500, 0)
	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey1: wPortingRequest1,
		wPortingReqKey2: wPortingRequest2,
		wPortingReqKey3: wPortingRequest3,
	}
	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.MatchedRedeemRequests = matchedRedeemRequest
	s.currentPortalStateForProducer.WaitingRedeemRequests = wRedeemRequests
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
	s.currentPortalStateForProcess.CustodianPoolState = cloneCustodians(custodians)
	s.currentPortalStateForProcess.MatchedRedeemRequests = cloneRedeemRequests(matchedRedeemRequest)
	s.currentPortalStateForProcess.WaitingRedeemRequests = cloneRedeemRequests(wRedeemRequests)
	s.currentPortalStateForProcess.WaitingPortingRequests = cloneWPortingRequests(wPortingRequests)
}
func (s *PortalTestSuiteV3) TestAutoLiquidationCustodian() {
	fmt.Println("Running TestAutoLiquidationCustodian - beacon height 3161 ...")
	bc := s.blockChain
	beaconHeight := uint64(3161) // ~ after 24 hours from redeem request and porting request
	//shardHeight := uint64(3161)
	shardHeights := map[byte]uint64{
		0: 3161,
	}
	//shardID := byte(0)
	//pm := NewPortalManager()
	//newMatchedRedeemReqIDs := []string{}
	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
	s.SetupTestAutoLiquidation()
	newInsts, err := s.blockChain.CheckAndBuildInstForCustodianLiquidation(beaconHeight, shardHeights, &s.currentPortalStateForProducer, s.portalParams)
	s.Equal(nil, err)
	// process new instructions redeem
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)
	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)
	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}
*/

/*
	Feature 12: auto-liquidation: the proportion between the collateral and public token is drop down below 120%
*/
type ExpectedResultAutoLiquidationByRates struct {
	custodianPool         map[string]*statedb.CustodianState
	waitingPortingRes     map[string]*statedb.WaitingPortingRequest
	waitingRedeemRequests map[string]*statedb.RedeemRequest
	matchedRedeemRequest  map[string]*statedb.RedeemRequest
	liquidationPool       map[string]*statedb.LiquidationPool
	numBeaconInsts        uint
	statusInsts           []string
}

func (s *PortalTestSuiteV3) SetupTestAutoLiquidationByRates() {
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	// process 3 => 2 => 1
	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 2000 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1740 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0, // 10 BNB
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	// waiting redeem requests
	wRedeemReqKey4 := statedb.GenerateWaitingRedeemRequestObjectKey("redeem-bnb-4").String()
	wRedeemReq4 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-4", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		12*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 10*1e9),
		},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	// matched redeem requests
	matchedRedeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
	matchedRedeemReq1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 1*1e9),
		},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	matchedRedeemReqKey2 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-2").String()
	matchedRedeemReq2 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-2", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		140*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 136500000000),
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 3500000000),
		},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}
	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
		wPortingReqKey3: wPortingRequest3,
	}
	matchedRedeemRequests := map[string]*statedb.RedeemRequest{
		matchedRedeemReqKey1: matchedRedeemReq1,
		matchedRedeemReqKey2: matchedRedeemReq2,
	}
	waitingRedeemRequests := map[string]*statedb.RedeemRequest{
		wRedeemReqKey4: wRedeemReq4,
	}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
	s.currentPortalStateForProducer.WaitingRedeemRequests = waitingRedeemRequests
	s.currentPortalStateForProducer.MatchedRedeemRequests = matchedRedeemRequests

	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = portalprocessv3.CloneWPortingRequests(wPortingRequests)
	s.currentPortalStateForProcess.WaitingRedeemRequests = portalprocessv3.CloneRedeemRequests(waitingRedeemRequests)
	s.currentPortalStateForProcess.MatchedRedeemRequests = portalprocessv3.CloneRedeemRequests(matchedRedeemRequests)

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:        {Amount: 1000000000},
			pCommon.PortalBNBIDStr: {Amount: 40000000000}, // x2
			pCommon.PortalBTCIDStr: {Amount: 10000000000000},
			ETH_ID:                 {Amount: 400000000000},
			USDT_ID:                {Amount: 1000000000},
			DAI_ID:                 {Amount: 1000000000},
		})
	s.currentPortalStateForProducer.FinalExchangeRatesState = finalExchangeRate
	s.currentPortalStateForProcess.FinalExchangeRatesState = finalExchangeRate
}

func buildExpectedResultAutoLiquidationByRates() *ExpectedResultAutoLiquidationByRates {
	beaconHeight := uint64(1004)
	shardHeight := uint64(1004)
	shardID := byte(0)

	// build expected results
	// custodian state after liquidated
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			USDT_ID: 1600 * 1e6,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			USDT_ID: 260 * 1e6,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				USDT_ID: 1340 * 1e6,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0, // 10 BNB
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(1000 * 1e9)

	// waiting porting requests
	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, 30*1e9,
		[]*statedb.MatchingPortingCustodianDetail{
			{
				IncAddress:             custodian2.GetIncognitoAddress(),
				RemoteAddress:          custodian2.GetRemoteAddresses()[pCommon.PortalBNBIDStr],
				Amount:                 30 * 1e9,
				LockedAmountCollateral: 0,
				LockedTokenCollaterals: map[string]uint64{
					USDT_ID: 1200 * 1e6,
				},
			},
		}, 60000000, beaconHeight, shardHeight, shardID)

	// waiting redeem requests
	//wRedeemReqKey4 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-4").String()
	//wRedeemReq4 := statedb.NewRedeemRequestWithValue(
	//	"redeem-bnb-4", pCommon.PortalBNBIDStr,
	//	USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
	//	12*1e9,
	//	[]*statedb.MatchingRedeemCustodianDetail{
	//		statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 10*1e9),
	//	},
	//	2000000,
	//	beaconHeight, common.Hash{},
	//	shardID, shardHeight,
	//	USER_ETH_ADDRESS_1)

	// matched redeem requests
	matchedRedeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
	matchedRedeemReq1 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-1", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		1*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 1*1e9),
		},
		2000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	matchedRedeemReqKey2 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-2").String()
	matchedRedeemReq2 := statedb.NewRedeemRequestWithValue(
		"redeem-bnb-2", pCommon.PortalBNBIDStr,
		USER_INC_ADDRESS_1, USER_BNB_ADDRESS_1,
		140*1e9,
		[]*statedb.MatchingRedeemCustodianDetail{
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 136500000000),
			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 3500000000),
		},
		280000000,
		beaconHeight, common.Hash{},
		shardID, shardHeight,
		USER_ETH_ADDRESS_1)

	liquidationPoolKey := statedb.GeneratePortalLiquidationPoolObjectKey().String()
	liquidationPool := map[string]*statedb.LiquidationPool{}
	liquidationPool[liquidationPoolKey] = statedb.NewLiquidationPoolWithValue(
		map[string]statedb.LiquidationPoolDetail{
			pCommon.PortalBNBIDStr: {
				PubTokenAmount:   10 * 1e9,
				CollateralAmount: 0,
				TokensCollateralAmount: map[string]uint64{
					USDT_ID: 400 * 1e6,
				},
			},
		})

	expectedRes := &ExpectedResultAutoLiquidationByRates{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes: map[string]*statedb.WaitingPortingRequest{
			wPortingReqKey3: wPortingRequest3,
		},
		waitingRedeemRequests: map[string]*statedb.RedeemRequest{},
		matchedRedeemRequest: map[string]*statedb.RedeemRequest{
			matchedRedeemReqKey1: matchedRedeemReq1,
			matchedRedeemReqKey2: matchedRedeemReq2,
		},
		liquidationPool: liquidationPool,
		numBeaconInsts:  2,
		statusInsts: []string{
			pCommon.PortalRedeemReqCancelledByLiquidationChainStatus,
			pCommon.PortalProducerInstSuccessChainStatus,
		},
	}

	return expectedRes
}

/*func (s *PortalTestSuiteV3) TestAutoLiquidationByExchangeRate() {*/
//fmt.Println("Running TestAutoLiquidationByExchangeRate - beacon height 1501 ...")
//bc := s.blockChain
//beaconHeight := uint64(1501)
//shardHeights := map[byte]uint64{
//0: 1501,
//}
//pm := portal.NewPortalManager()

//s.SetupTestAutoLiquidationByRates()

//custodianPool := s.currentPortalStateForProducer.CustodianPoolState
//for cusKey, cus := range custodianPool {
//fmt.Printf("cusKey %v - custodian address: %v\n", cusKey, cus.GetIncognitoAddress())
//}

//expectedResult := buildExpectedResultAutoLiquidationByRates()

//// producer instructions
//processor := portalprocessv3.GetPortalInstProcessorByMetaType(pm.PortalInstProcessorsV3, metadata.PortalLiquidateByRatesMetaV3)
//newInsts, err := processor.BuildNewInsts(bc, "", 0, &s.currentPortalStateForProducer, beaconHeight, shardHeights, s.portalParams, nil)
//s.Equal(nil, err)

//// process new instructions
//err = processPortalInstructions(
//bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

//// check results
//s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
//s.Equal(nil, err)

//for i, inst := range newInsts {
//s.Equal(expectedResult.statusInsts[i], inst[2], "Instruction index %v", i)
//}

//s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
//s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)
//s.Equal(expectedResult.waitingRedeemRequests, s.currentPortalStateForProducer.WaitingRedeemRequests)
//s.Equal(expectedResult.matchedRedeemRequest, s.currentPortalStateForProducer.MatchedRedeemRequests)
//s.Equal(expectedResult.liquidationPool, s.currentPortalStateForProducer.LiquidationPool)

//s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
/*}*/

///*
//	Feature 13: auto-liquidation: the custodian top up the collaterals
//*/
//func (s *PortalTestSuiteV3) SetupTestTopupCustodian() {
//	s.SetupTestAutoLiquidationByRates()
//}
//
//type TestCaseTopupCustodian struct {
//	incAddressStr        string
//	ptokenID             string
//	depositAmount        uint64
//	freeCollateralAmount uint64
//}
//
//func buildTopupCustodianActionsFromTcs(tcs []TestCaseTopupCustodian, shardID byte) [][]string {
//	insts := [][]string{}
//
//	for _, tc := range tcs {
//		inst := buildPortalTopupCustodianAction(tc.incAddressStr, tc.ptokenID, tc.depositAmount, shardID, tc.freeCollateralAmount)
//		insts = append(insts, inst)
//	}
//
//	return insts
//}
//
//type TestCaseTopupWaitingPorting struct {
//	incAddressStr        string
//	portingID            string
//	ptokenID             string
//	depositAmount        uint64
//	freeCollateralAmount uint64
//}
//
//func buildTopupWaitingPortingActionsFromTcs(tcs []TestCaseTopupWaitingPorting, shardID byte) [][]string {
//	insts := [][]string{}
//
//	for _, tc := range tcs {
//		inst := buildTopupWaitingPortingAction(tc.incAddressStr, tc.portingID, tc.ptokenID, tc.depositAmount, shardID, tc.freeCollateralAmount)
//		insts = append(insts, inst)
//	}
//
//	return insts
//}
//func (s *PortalTestSuiteV3) TestTopupCustodian() {
//	fmt.Println("Running TestAutoLiquidationCustodian - beacon height 1501 ...")
//	bc := s.blockChain
//	beaconHeight := uint64(1501)
//	pm := NewPortalManager()
//	shardID := byte(0)
//	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
//
//	s.SetupTestAutoLiquidationByRates()
//
//	// build test cases for topup custodian
//	testcases := []TestCaseTopupCustodian{
//		// topup by burning more collaterals
//		{
//			incAddressStr:        CUS_INC_ADDRESS_2,
//			ptokenID:             pCommon.PortalBNBIDStr,
//			depositAmount:        500 * 1e9,
//			freeCollateralAmount: 0,
//		},
//		// topup by using free collaterals
//		{
//			incAddressStr:        CUS_INC_ADDRESS_2,
//			ptokenID:             pCommon.PortalBNBIDStr,
//			depositAmount:        0,
//			freeCollateralAmount: 500 * 1e9,
//		},
//	}
//
//	// build actions from testcases
//	insts := buildTopupCustodianActionsFromTcs(testcases, shardID)
//
//	// build test cases for topup waiting porting
//	testcases2 := []TestCaseTopupWaitingPorting{
//		// topup by burning more collaterals
//		{
//			incAddressStr:        "custodianIncAddress4",
//			portingID:            "porting-bnb-1",
//			ptokenID:             pCommon.PortalBNBIDStr,
//			depositAmount:        20 * 1e9,
//			freeCollateralAmount: 0,
//		},
//		// topup by using free collaterals
//		{
//			incAddressStr:        "custodianIncAddress4",
//			portingID:            "porting-bnb-1",
//			ptokenID:             pCommon.PortalBNBIDStr,
//			depositAmount:        0,
//			freeCollateralAmount: 50 * 1e9,
//		},
//	}
//
//	// build actions from testcases2
//	insts2 := buildTopupWaitingPortingActionsFromTcs(testcases2, shardID)
//
//	insts = append(insts, insts2...)
//
//	// producer instructions
//	newInsts, err := producerPortalInstructions(
//		bc, beaconHeight, insts, s.sdb, &s.currentPortalStateForProducer, s.portalParams, shardID, pm)
//
//	// check liquidation by exchange rates
//	newInstsForLiquidationByExchangeRate, err := buildInstForLiquidationTopPercentileExchangeRates(
//		beaconHeight-1, &s.currentPortalStateForProducer, s.portalParams)
//
//	s.Equal(0, len(newInstsForLiquidationByExchangeRate))
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)
//
//	// check results
//	s.Equal(4, len(newInsts))
//	s.Equal(nil, err)
//
//	// remain waiting redeem requests and matched redeem requests
//	s.Equal(2, len(s.currentPortalStateForProducer.MatchedRedeemRequests))
//	s.Equal(1, len(s.currentPortalStateForProducer.WaitingRedeemRequests))
//	s.Equal(1, len(s.currentPortalStateForProducer.WaitingPortingRequests))
//
//	//custodian state after auto liquidation
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//	custodianKey4 := statedb.GenerateCustodianStateObjectKey("custodianIncAddress4").String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 7000*1e9, 6920000000000,
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 80000000000,
//		},
//		map[string]string{
//			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 1500*1e9, 460000000000,
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 0.6 * 1e9,
//		},
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 1040000000000,
//		},
//		map[string]string{
//			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000*1e9, 8000000000000,
//		map[string]uint64{
//			pCommon.PortalBTCIDStr: 0.1 * 1e9,
//		},
//		map[string]uint64{
//			pCommon.PortalBTCIDStr: 2000000000000,
//		},
//		map[string]string{
//			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian4 := statedb.NewCustodianStateWithValue(
//		"custodianIncAddress4", 5020*1e9, 4910000000000,
//		map[string]uint64{
//		},
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 110000000000,
//		},
//		map[string]string{
//			pCommon.PortalBNBIDStr: "bnbAddress4",
//		},
//		map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
//	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-bnb-1", common.Hash{}, pCommon.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, 1*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             "custodianIncAddress4",
//				RemoteAddress:          "bnbAddress4",
//				Amount:                 1 * 1e9,
//				LockedAmountCollateral: 110000000000,
//			},
//		}, 2000000, 1500, 1500, 0)
//
//	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
//		wPortingReqKey1: wPortingRequest1,
//	}
//
//	s.Equal(custodian1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1])
//	s.Equal(custodian2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2])
//	s.Equal(custodian3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3])
//	s.Equal(custodian4, s.currentPortalStateForProducer.CustodianPoolState[custodianKey4])
//	s.Equal(0, len(s.currentPortalStateForProducer.LiquidationPool))
//	s.Equal(wPortingRequests, s.currentPortalStateForProducer.WaitingPortingRequests)
//
//	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
//}

/**
	Feature 14: Custodian rewards from DAO funds and porting/redeem fee
**/
//func (s *PortalTestSuiteV3) SetupTestCustodianRewards() {
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//
//	custodian1 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_1, 7000000000000, 6708000000000,
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 292000000000,
//		},
//		map[string]string{
//			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
//			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian2 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_2, 1000000000000, 972000000000,
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 0,
//		},
//		map[string]uint64{
//			pCommon.PortalBNBIDStr: 28000000000,
//		},
//		map[string]string{
//			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodian3 := statedb.NewCustodianStateWithValue(
//		CUS_INC_ADDRESS_3, 10000000000000, 7988000000000,
//		nil,
//		map[string]uint64{
//			pCommon.PortalBTCIDStr: 2012000000000,
//		},
//		map[string]string{
//			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_3,
//		},  map[string]uint64{}, map[string]uint64{}, map[string]uint64{}, map[string]map[string]uint64{})
//
//	custodians := map[string]*statedb.CustodianState{
//		custodianKey1: custodian1,
//		custodianKey2: custodian2,
//		custodianKey3: custodian3,
//	}
//
//	// waiting porting requests
//	wPortingReqKey1 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-1").String()
//	wPortingRequest1 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-bnb-1", common.Hash{}, pCommon.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, 1*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             CUS_INC_ADDRESS_1,
//				RemoteAddress:          CUS_BNB_ADDRESS_1,
//				Amount:                 0.3 * 1e9,
//				LockedAmountCollateral: 12000000000,
//			},
//			{
//				IncAddress:             CUS_INC_ADDRESS_2,
//				RemoteAddress:          CUS_BNB_ADDRESS_1,
//				Amount:                 0.7 * 1e9,
//				LockedAmountCollateral: 28000000000,
//			},
//		}, 2000000, 1000, 1000, 0)
//
//	wPortingReqKey2 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-btc-2").String()
//	wPortingRequest2 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-btc-2", common.Hash{}, pCommon.PortalBTCIDStr,
//		USER_INC_ADDRESS_2, 0.1*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             CUS_INC_ADDRESS_3,
//				RemoteAddress:          CUS_BTC_ADDRESS_3,
//				Amount:                 0.1 * 1e9,
//				LockedAmountCollateral: 2000000000000,
//			},
//		}, 100000001, 1000, 1000, 0)
//
//	wPortingReqKey3 := statedb.GeneratePortalWaitingPortingRequestObjectKey("porting-bnb-3").String()
//	wPortingRequest3 := statedb.NewWaitingPortingRequestWithValue(
//		"porting-bnb-3", common.Hash{}, pCommon.PortalBNBIDStr,
//		USER_INC_ADDRESS_2, 5*1e9,
//		[]*statedb.MatchingPortingCustodianDetail{
//			{
//				IncAddress:             CUS_INC_ADDRESS_1,
//				RemoteAddress:          CUS_BNB_ADDRESS_1,
//				Amount:                 5 * 1e9,
//				LockedAmountCollateral: 200000000000,
//			},
//		}, 2000000, 900, 900, 0)
//
//	wPortingRequests := map[string]*statedb.WaitingPortingRequest{
//		wPortingReqKey1: wPortingRequest1,
//		wPortingReqKey2: wPortingRequest2,
//		wPortingReqKey3: wPortingRequest3,
//	}
//
//	// matched redeem requests
//	redeemReqKey1 := statedb.GenerateMatchedRedeemRequestObjectKey("redeem-bnb-1").String()
//	redeemRequest1 := statedb.NewRedeemRequestWithValue(
//		"redeem-bnb-1", pCommon.PortalBNBIDStr,
//		USER_INC_ADDRESS_1, "userCUS_BNB_ADDRESS_1", 2.3*1e9,
//		[]*statedb.MatchingRedeemCustodianDetail{
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_1, CUS_BNB_ADDRESS_1, 2*1e9),
//			statedb.NewMatchingRedeemCustodianDetailWithValue(CUS_INC_ADDRESS_2, CUS_BNB_ADDRESS_2, 0.3*1e9),
//		}, 4600000, 990, common.Hash{}, 0, 990, "")
//
//	matchedRedeemRequest := map[string]*statedb.RedeemRequest{
//		redeemReqKey1: redeemRequest1,
//	}
//
//	// locked collaterals
//	lockedCollateralDetail := map[string]uint64{
//		CUS_INC_ADDRESS_1: 292000000000,
//		CUS_INC_ADDRESS_2: 28000000000,
//		CUS_INC_ADDRESS_3: 2012000000000,
//	}
//	totalLockedCollateralInEpoch := uint64(2332000000000)
//	s.currentPortalStateForProducer.LockedCollateralForRewards = statedb.NewLockedCollateralStateWithValue(
//		totalLockedCollateralInEpoch, lockedCollateralDetail)
//	s.currentPortalStateForProducer.CustodianPoolState = custodians
//	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
//	s.currentPortalStateForProducer.MatchedRedeemRequests = matchedRedeemRequest
//
//	s.currentPortalStateForProcess.CustodianPoolState = cloneCustodians(custodians)
//	s.currentPortalStateForProcess.WaitingPortingRequests = cloneWPortingRequests(wPortingRequests)
//	s.currentPortalStateForProcess.MatchedRedeemRequests = cloneRedeemRequests(matchedRedeemRequest)
//	s.currentPortalStateForProcess.LockedCollateralForRewards = statedb.NewLockedCollateralStateWithValue(
//		totalLockedCollateralInEpoch, cloneMap(lockedCollateralDetail))
//}
//
//func (s *PortalTestSuiteV3) TestCustodianRewards() {
//	fmt.Println("Running TestCustodianRewards - beacon height 1000 ...")
//	bc := s.blockChain
//	beaconHeight := uint64(1000) // after 24 hours from requesting porting (bch = 100)
//	//shardID := byte(0)
//	newMatchedRedeemReqIDs := []string{"redeem-bnb-1"}
//	rewardForCustodianByEpoch := map[common.Hash]uint64{
//		common.PRVCoinID: 100000000000, // 100 prv
//		common.Hash{1}:   200000,
//	}
//	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
//
//	s.SetupTestCustodianRewards()
//
//	// producer instructions
//	newInsts, err := bc.buildPortalRewardsInsts(
//		beaconHeight-1, &s.currentPortalStateForProducer, rewardForCustodianByEpoch, newMatchedRedeemReqIDs)
//	s.Equal(nil, err)
//
//	// process new instructions
//	err = processPortalInstructions(
//		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, updatingInfoByTokenID)
//
//	// check results
//	s.Equal(2, len(newInsts))
//	s.Equal(nil, err)
//
//	//custodian state after auto liquidation
//	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
//	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
//	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()
//
//	reward1 := map[string]uint64{
//		common.PRVIDStr:         12526040824,
//		common.Hash{1}.String(): 25044,
//	}
//
//	reward2 := map[string]uint64{
//		common.PRVIDStr:         1202686106,
//		common.Hash{1}.String(): 2401,
//	}
//	reward3 := map[string]uint64{
//		common.PRVIDStr:         86377873071,
//		common.Hash{1}.String(): 172555,
//	}
//
//	s.Equal(reward1, s.currentPortalStateForProducer.CustodianPoolState[custodianKey1].GetRewardAmount())
//	s.Equal(reward2, s.currentPortalStateForProducer.CustodianPoolState[custodianKey2].GetRewardAmount())
//	s.Equal(reward3, s.currentPortalStateForProducer.CustodianPoolState[custodianKey3].GetRewardAmount())
//
//	s.Equal(reward1, s.currentPortalStateForProcess.CustodianPoolState[custodianKey1].GetRewardAmount())
//	s.Equal(reward2, s.currentPortalStateForProcess.CustodianPoolState[custodianKey2].GetRewardAmount())
//	s.Equal(reward3, s.currentPortalStateForProcess.CustodianPoolState[custodianKey3].GetRewardAmount())
//}

/**
	Feature 15: Custodian unlock over rate collaterals
**/

type TestCaseUnlockOverRateCollaterals struct {
	custodianIncAddress string
	tokenID             string
}

type ExpectedResultUnlockOverRateCollaterals struct {
	custodianPool         map[string]*statedb.CustodianState
	waitingPortingRes     map[string]*statedb.WaitingPortingRequest
	waitingRedeemRequests map[string]*statedb.RedeemRequest
	matchedRedeemRequest  map[string]*statedb.RedeemRequest
	liquidationPool       map[string]*statedb.LiquidationPool
	numBeaconInsts        uint
	statusInsts           []string
}

func (s *PortalTestSuiteV3) SetupTestUnlockOverRateCollaterals() {
	//beaconHeight := uint64(1004)
	//shardHeight := uint64(1004)
	//shardID := byte(0)

	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	// process 3 => 2 => 1
	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(0)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  0,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 1000 * 1e9,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  10 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 137.5 * 1e9,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID: 10 * 1e9,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID: 5 * 1e9,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID: 5 * 1e9,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 100 * 1e9,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetFreeCollateral(500 * 1e9)
	custodian3.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 12.5 * 1e9,
		})
	custodian3.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 500 * 1e9,
		})

	custodians := map[string]*statedb.CustodianState{
		custodianKey1: custodian1,
		custodianKey2: custodian2,
		custodianKey3: custodian3,
	}
	wPortingRequests := map[string]*statedb.WaitingPortingRequest{}
	matchedRedeemRequests := map[string]*statedb.RedeemRequest{}
	waitingRedeemRequests := map[string]*statedb.RedeemRequest{}

	s.currentPortalStateForProducer.CustodianPoolState = custodians
	s.currentPortalStateForProducer.WaitingPortingRequests = wPortingRequests
	s.currentPortalStateForProducer.WaitingRedeemRequests = waitingRedeemRequests
	s.currentPortalStateForProducer.MatchedRedeemRequests = matchedRedeemRequests

	s.currentPortalStateForProcess.CustodianPoolState = portalprocessv3.CloneCustodians(custodians)
	s.currentPortalStateForProcess.WaitingPortingRequests = portalprocessv3.CloneWPortingRequests(wPortingRequests)
	s.currentPortalStateForProcess.WaitingRedeemRequests = portalprocessv3.CloneRedeemRequests(waitingRedeemRequests)
	s.currentPortalStateForProcess.MatchedRedeemRequests = portalprocessv3.CloneRedeemRequests(matchedRedeemRequests)

	finalExchangeRate := statedb.NewFinalExchangeRatesStateWithValue(
		map[string]statedb.FinalExchangeRatesDetail{
			common.PRVIDStr:        {Amount: 3000000000}, // x3
			pCommon.PortalBNBIDStr: {Amount: 20000000000},
			pCommon.PortalBTCIDStr: {Amount: 10000000000000},
			ETH_ID:                 {Amount: 1200000000000}, // x3
			USDT_ID:                {Amount: 1000000000},
			DAI_ID:                 {Amount: 1000000000},
		})
	s.currentPortalStateForProducer.FinalExchangeRatesState = finalExchangeRate
	s.currentPortalStateForProcess.FinalExchangeRatesState = finalExchangeRate
}

func buildUnlockOverRateCollateralsActionsFromTcs(tcs []TestCaseUnlockOverRateCollaterals, shardID byte) []instructionForProducer {
	insts := []instructionForProducer{}

	for _, tc := range tcs {
		inst := buildPortalCustodianUnlockOverRateCollateralsAction(tc.custodianIncAddress, tc.tokenID, shardID)
		insts = append(insts, instructionForProducer{
			inst:         inst,
			optionalData: nil,
		})
	}

	return insts
}

func buildPortalCustodianUnlockOverRateCollateralsAction(
	incAddressStr string,
	tokenID string,
	shardID byte,
) []string {
	data := metadata.PortalUnlockOverRateCollaterals{
		MetadataBase: metadata.MetadataBase{
			Type: metadata.PortalUnlockOverRateCollateralsMeta,
		},
		CustodianAddressStr: incAddressStr,
		TokenID:             tokenID,
	}

	actionContent := metadata.PortalUnlockOverRateCollateralsAction{
		Meta:    data,
		TxReqID: common.Hash{},
		ShardID: shardID,
	}
	actionContentBytes, _ := json.Marshal(actionContent)
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	return []string{strconv.Itoa(metadata.PortalUnlockOverRateCollateralsMeta), actionContentBase64Str}
}

func buildExpectedResultUnlockOverRateCollaterals() ([]TestCaseUnlockOverRateCollaterals, *ExpectedResultUnlockOverRateCollaterals) {

	testcases := []TestCaseUnlockOverRateCollaterals{
		// unlocked both prv and eth
		{
			custodianIncAddress: CUS_INC_ADDRESS_1,
			tokenID:             pCommon.PortalBNBIDStr,
		},
		// only eth unlocked
		{
			custodianIncAddress: CUS_INC_ADDRESS_2,
			tokenID:             pCommon.PortalBNBIDStr,
		},
		// only prv unlocked
		{
			custodianIncAddress: CUS_INC_ADDRESS_3,
			tokenID:             pCommon.PortalBNBIDStr,
		},
		// can not unlock cause rate not changed
		{
			custodianIncAddress: CUS_INC_ADDRESS_2,
			tokenID:             pCommon.PortalBNBIDStr,
		},
	}

	// build expected results
	// custodian state after liquidated
	custodianKey1 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_1).String()
	custodianKey2 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_2).String()
	custodianKey3 := statedb.GenerateCustodianStateObjectKey(CUS_INC_ADDRESS_3).String()

	custodian1 := statedb.NewCustodianState()
	custodian1.SetIncognitoAddress(CUS_INC_ADDRESS_1)
	custodian1.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_1,
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_1,
		})
	custodian1.SetTotalCollateral(1000 * 1e9)
	custodian1.SetFreeCollateral(1000 * 1e9)
	custodian1.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID:  10 * 1e9,
			USDT_ID: 500 * 1e6,
		})
	custodian1.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID:  4.6875 * 1e9,
			USDT_ID: 0,
		})
	custodian1.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 0,
		})
	custodian1.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID:  5.3125 * 1e9,
				USDT_ID: 500 * 1e6,
			},
		})
	custodian1.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 137.5 * 1e9,
		})

	custodian2 := statedb.NewCustodianState()
	custodian2.SetIncognitoAddress(CUS_INC_ADDRESS_2)
	custodian2.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBNBIDStr: CUS_BNB_ADDRESS_2,
		})
	custodian2.SetTotalTokenCollaterals(
		map[string]uint64{
			ETH_ID: 10 * 1e9,
		})
	custodian2.SetFreeTokenCollaterals(
		map[string]uint64{
			ETH_ID: 5.833333333 * 1e9,
		})
	custodian2.SetLockedTokenCollaterals(
		map[string]map[string]uint64{
			pCommon.PortalBNBIDStr: {
				ETH_ID: 4166666667,
			}})
	custodian2.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 100 * 1e9,
		})

	custodian3 := statedb.NewCustodianState()
	custodian3.SetIncognitoAddress(CUS_INC_ADDRESS_3)
	custodian3.SetRemoteAddresses(
		map[string]string{
			pCommon.PortalBTCIDStr: CUS_BTC_ADDRESS_2,
		})
	custodian3.SetTotalCollateral(1000 * 1e9)
	custodian3.SetLockedAmountCollateral(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 208333333334,
		})
	custodian3.SetFreeCollateral(791666666666)
	custodian3.SetHoldingPublicTokens(
		map[string]uint64{
			pCommon.PortalBNBIDStr: 12.5 * 1e9,
		})

	expectedRes := &ExpectedResultUnlockOverRateCollaterals{
		custodianPool: map[string]*statedb.CustodianState{
			custodianKey1: custodian1,
			custodianKey2: custodian2,
			custodianKey3: custodian3,
		},
		waitingPortingRes:     map[string]*statedb.WaitingPortingRequest{},
		waitingRedeemRequests: map[string]*statedb.RedeemRequest{},
		matchedRedeemRequest:  map[string]*statedb.RedeemRequest{},
		numBeaconInsts:        4,
		liquidationPool:       map[string]*statedb.LiquidationPool{},
		statusInsts: []string{
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestAcceptedChainStatus,
			pCommon.PortalRequestRejectedChainStatus,
		},
	}

	return testcases, expectedRes
}

func (s *PortalTestSuiteV3) TestUnlockOverRateCollaterals() {
	fmt.Println("Running TestUnlockOverRateCollaterals - beacon height 1501 ...")
	bc := s.blockChain
	beaconHeight := uint64(1501)
	shardHeights := map[byte]uint64{
		0: uint64(1501),
	}
	shardID := byte(0)
	//newMatchedRedeemReqIDs := []string{}
	pm := portal.NewPortalManager()
	//updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}

	s.SetupTestUnlockOverRateCollaterals()

	custodianPool := s.currentPortalStateForProducer.CustodianPoolState
	for cusKey, cus := range custodianPool {
		fmt.Printf("cusKey %v - custodian address: %v\n", cusKey, cus.GetIncognitoAddress())
	}

	testcases, expectedResult := buildExpectedResultUnlockOverRateCollaterals()

	// build actions from testcases
	instsForProducer := buildUnlockOverRateCollateralsActionsFromTcs(testcases, shardID)

	// producer instructions
	newInsts, err := producerPortalInstructions(
		bc, beaconHeight, shardHeights, instsForProducer, &s.currentPortalStateForProducer, s.portalParams, shardID, pm.PortalInstProcessorsV3)
	s.Equal(nil, err)

	// process new instructions
	err = processPortalInstructions(
		bc, beaconHeight-1, newInsts, s.sdb, &s.currentPortalStateForProcess, s.portalParams, pm.PortalInstProcessorsV3)

	// check results
	s.Equal(expectedResult.numBeaconInsts, uint(len(newInsts)))
	s.Equal(nil, err)

	for i, inst := range newInsts {
		s.Equal(expectedResult.statusInsts[i], inst[2], "Instruction index %v", i)
	}

	s.Equal(expectedResult.custodianPool, s.currentPortalStateForProducer.CustodianPoolState)
	s.Equal(expectedResult.waitingPortingRes, s.currentPortalStateForProducer.WaitingPortingRequests)
	s.Equal(expectedResult.waitingRedeemRequests, s.currentPortalStateForProducer.WaitingRedeemRequests)
	s.Equal(expectedResult.matchedRedeemRequest, s.currentPortalStateForProducer.MatchedRedeemRequests)
	s.Equal(expectedResult.liquidationPool, s.currentPortalStateForProducer.LiquidationPool)

	s.Equal(s.currentPortalStateForProcess, s.currentPortalStateForProducer)
}
