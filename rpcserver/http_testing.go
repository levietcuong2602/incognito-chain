package rpcserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/dataaccessobject/stats"

	"github.com/incognitochain/incognito-chain/consensus_v2"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wire"
	"github.com/pkg/errors"
)

type txs struct {
	Txs []string `json:"Txs"`
}

func (httpServer *HttpServer) handleTestHttpServer(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return nil, nil
}

/*
For testing and benchmark only
*/
type CountResult struct {
	Success int
	Fail    int
}

func (httpServer *HttpServer) handleUnlockMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	httpServer.config.TxMemPool.SendTransactionToBlockGen()
	return nil, nil
}

func (httpServer *HttpServer) handleGetConsensusInfoV3(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	_, ok := httpServer.config.ConsensusEngine.(*consensus_v2.Engine)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("consensus engine not found, got "+reflect.TypeOf(httpServer.config.ConsensusEngine).String()))
	}

	arr := []interface{}{}
	/*for chainID, bftactor := range engine.BFTProcess() {*/
	//bftactorV2, ok := bftactor.(temp)
	//if !ok {
	//continue
	//}
	//m := map[string]interface{}{
	//"ChainID":              chainID,
	//"VoteHistory":          bftactorV3.GetVoteHistory(),
	//"ReceiveBlockByHash":   bftactorV3.GetReceiveBlockByHash(),
	//"ReceiveBlockByHeight": bftactorV3.GetReceiveBlockByHeight(),
	//}
	//arr = append(arr, m)
	/*}*/

	return arr, nil
}

func (httpServer *HttpServer) handleGetAutoStakingByHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	height := int(arrayParams[0].(float64))
	beaconConsensusStateRootHash, err := httpServer.blockService.BlockChain.GetBeaconConsensusRootHash(httpServer.blockService.BlockChain.GetBeaconBestState(), uint64(height))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	// beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash, statedb.NewDatabaseAccessWarper(httpServer.blockService.BlockChain.GetBeaconChainDatabase()))
	// if err != nil {
	// 	return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	// }
	// _, newAutoStaking := statedb.GetRewardReceiverAndAutoStaking(beaconConsensusStateDB, httpServer.blockService.BlockChain.GetShardIDs())
	newAutoStaking := map[string]bool{}
	return []interface{}{beaconConsensusStateRootHash, newAutoStaking}, nil
}

func (httpServer *HttpServer) handleGetTotalBlockInEpoch(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(-1, errors.New("Invalid number of params, accept only 1 value"))
	}
	epoch := uint64(arrayParams[0].(float64))

	res := make(map[byte]*stats.NumberOfBlockInOneEpochStats)
	for i := 0; i < httpServer.config.BlockChain.BeaconChain.GetActiveShardNumber(); i++ {
		shardID := byte(i)
		numberOfBlockInOneEpochStats, err := stats.GetShardEpochBPV3Stats(httpServer.config.BlockChain.GetShardChainDatabase(shardID), shardID, epoch)
		if err != nil {
			return nil, rpcservice.NewRPCError(-1, err)
		}
		res[shardID] = numberOfBlockInOneEpochStats
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetDetailBlocksOfEpoch(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardID := byte(arrayParams[0].(float64))
	epoch := uint64(arrayParams[1].(float64))
	if len(arrayParams) != 2 {
		return nil, rpcservice.NewRPCError(-1, errors.New("Invalid number of params, accept only 2 value"))
	}
	res, err := stats.GetShardHeightBPV3Stats(httpServer.config.BlockChain.GetShardChainDatabase(shardID), shardID, epoch)
	if err != nil {
		return nil, rpcservice.NewRPCError(-1, err)
	}
	return res, nil
}

func (httpServer *HttpServer) handleGetCommitteeState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	height := uint64(arrayParams[0].(float64))
	tempHash := arrayParams[1].(string)

	var beaconConsensusStateRootHash = &blockchain.BeaconRootHash{}
	var err1 error = nil

	if height == 0 || tempHash != "" {
		hash, err := common.Hash{}.NewHashFromStr(tempHash)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		beaconConsensusStateRootHash, err1 = blockchain.GetBeaconRootsHashByBlockHash(
			httpServer.config.BlockChain.GetBeaconChainDatabase(),
			*hash,
		)
		if err1 != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
		}
	} else {
		beaconConsensusStateRootHash, err1 = httpServer.config.BlockChain.GetBeaconRootsHashFromBlockHeight(
			height,
		)
		if err1 != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
		}
	}
	shardIDs := []int{-1}
	shardIDs = append(shardIDs, httpServer.config.BlockChain.GetShardIDs()...)
	stateDB, err2 := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetBeaconChainDatabase()))
	if err2 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}

	currentValidator, substituteValidator, nextEpochShardCandidate, currentEpochShardCandidate, _, _, syncingValidators, rewardReceivers, autoStaking, stakingTx := statedb.GetAllCandidateSubstituteCommittee(stateDB, shardIDs)
	currentValidatorStr := make(map[int][]string)
	for shardID, v := range currentValidator {
		tempV, _ := incognitokey.CommitteeKeyListToString(v)
		currentValidatorStr[shardID] = tempV
	}
	substituteValidatorStr := make(map[int][]string)
	for shardID, v := range substituteValidator {
		tempV, _ := incognitokey.CommitteeKeyListToString(v)
		substituteValidatorStr[shardID] = tempV
	}
	nextEpochShardCandidateStr, _ := incognitokey.CommitteeKeyListToString(nextEpochShardCandidate)
	currentEpochShardCandidateStr, _ := incognitokey.CommitteeKeyListToString(currentEpochShardCandidate)
	tempStakingTx := make(map[string]string)
	for k, v := range stakingTx {
		tempStakingTx[k] = v.String()
	}
	tempRewardReceiver := make(map[string]string)
	for k, v := range rewardReceivers {
		wl := wallet.KeyWallet{}
		wl.KeySet.PaymentAddress = v
		paymentAddress := wl.Base58CheckSerialize(wallet.PaymentAddressType)
		tempRewardReceiver[k] = paymentAddress
	}
	return map[string]interface{}{
		"root":             beaconConsensusStateRootHash.ConsensusStateDBRootHash,
		"committee":        currentValidatorStr,
		"substitute":       substituteValidatorStr,
		"nextCandidate":    nextEpochShardCandidateStr,
		"currentCandidate": currentEpochShardCandidateStr,
		"rewardReceivers":  tempRewardReceiver,
		"autoStaking":      autoStaking,
		"stakingTx":        tempStakingTx,
		"syncing":          syncingValidators,
	}, nil
}

func (httpServer *HttpServer) handleGetCommitteeStateByShard(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardID := uint64(arrayParams[0].(float64))
	tempHash := arrayParams[1].(string)

	hash, err := common.Hash{}.NewHashFromStr(tempHash)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	shardRootHash, err := blockchain.GetShardRootsHashByBlockHash(
		httpServer.config.BlockChain.GetShardChainDatabase(byte(shardID)),
		byte(shardID),
		*hash,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	stateDB, err := statedb.NewWithPrefixTrie(shardRootHash.ConsensusStateDBRootHash,
		statedb.NewDatabaseAccessWarper(httpServer.config.BlockChain.GetShardChainDatabase(byte(shardID))))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	committees := statedb.GetOneShardCommittee(stateDB, byte(shardID))
	resCommittees := make([]string, len(committees))
	for i := 0; i < len(resCommittees); i++ {
		key, _ := committees[i].ToBase58()
		resCommittees[i] = key
	}
	substitutes := statedb.GetOneShardSubstituteValidator(stateDB, byte(shardID))
	resSubstitutes := make([]string, len(substitutes))
	for i := 0; i < len(resSubstitutes); i++ {
		key, _ := substitutes[i].ToBase58()
		resSubstitutes[i] = key
	}

	return map[string]interface{}{
		"root":       shardRootHash.ConsensusStateDBRootHash,
		"shardID":    shardID,
		"committee":  resCommittees,
		"substitute": resSubstitutes,
	}, nil
}

func (httpServer *HttpServer) handleGetSlashingCommittee(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Number Of Params"))
	}
	epoch := uint64(arrayParams[0].(float64))
	if epoch < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"))
	}
	beaconBestState := httpServer.blockService.BlockChain.GetBeaconBestState()
	if epoch >= beaconBestState.Epoch {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"+
			"expect epoch from %+v to %+v", 1, beaconBestState.Epoch-1))
	}
	slashingCommittee := statedb.GetSlashingCommittee(beaconBestState.GetBeaconSlashStateDB(), epoch)
	return slashingCommittee, nil
}

func (httpServer *HttpServer) handleGetSlashingCommitteeDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Number Of Params"))
	}
	epoch := uint64(arrayParams[0].(float64))
	if epoch < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"))
	}
	beaconBestState := httpServer.blockService.BlockChain.GetBeaconBestState()
	if epoch >= beaconBestState.Epoch {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"+
			"expect epoch from %+v to %+v", 1, beaconBestState.Epoch-1))
	}
	slashingCommittees := statedb.GetSlashingCommittee(beaconBestState.GetBeaconSlashStateDB(), epoch)
	slashingCommitteeDetail := make(map[byte][]incognitokey.CommitteeKeyString)
	for shardID, slashingCommittee := range slashingCommittees {
		res, err := incognitokey.CommitteeBase58KeyListToStruct(slashingCommittee)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}
		slashingCommitteeDetail[shardID] = incognitokey.CommitteeKeyListToStringList(res)
	}
	return slashingCommitteeDetail, nil
}

func (httpServer *HttpServer) handleGetRewardAmountByEpoch(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("want length %+v but got %+v", 2, len(arrayParams)))
	}
	tempShardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid ShardID Value"))
	}
	tempEpoch, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("Invalid Epoch Value"))
	}
	shardID := byte(tempShardID)
	epoch := uint64(tempEpoch)
	rewardStateDB := httpServer.config.BlockChain.GetBeaconBestState().GetBeaconRewardStateDB()
	amount, err := statedb.GetRewardOfShardByEpoch(rewardStateDB, epoch, shardID, common.PRVCoinID)
	return amount, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
}

func (httpServer *HttpServer) handleGetAndSendTxsFromFile(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	shardIDParam := int(arrayParams[0].(float64))
	Logger.log.Critical(arrayParams)
	txType := arrayParams[1].(string)
	isSent := arrayParams[2].(bool)
	interval := int64(arrayParams[3].(float64))
	Logger.log.Criticalf("Interval between transactions %+v \n", interval)
	datadir := "./bin/"
	filename := ""
	success := 0
	fail := 0
	switch txType {
	case "privacy3000_1":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.1.json"
	case "privacy3000_2":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.2.json"
	case "privacy3000_3":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-3000.3.json"
	case "noprivacy9000":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-9000.json"
	case "noprivacy10000_2":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.2.json"
	case "noprivacy10000_3":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.3.json"
	case "noprivacy10000_4":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000.4.json"
	case "noprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-5000.json"
	case "privacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-privacy-5000.json"
	case "cstoken":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-cstoken-5000.json"
	case "cstokenprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-cstokenprivacy-5000.json"
	default:
		return CountResult{}, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Can't find file"))
	}

	Logger.log.Critical("Getting Transactions from file: ", datadir+filename)
	file, err := ioutil.ReadFile(datadir + filename)
	if err != nil {
		Logger.log.Error("Fail to get Transactions from file: ", err)
	}
	data := txs{}
	count := 0
	_ = json.Unmarshal([]byte(file), &data)
	Logger.log.Criticalf("Get %+v Transactions from file \n", len(data.Txs))
	intervalDuration := time.Duration(interval) * time.Millisecond
	for index, txBase58Data := range data.Txs {
		<-time.Tick(intervalDuration)
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			fail++
			continue
		}
		switch txType {
		case "cstokenprivacy":
			{
				tx, err := transaction.NewTransactionTokenFromJsonBytes(rawTxBytes)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxPrivacyToken).Transaction = tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		default:
			tx, err := transaction.NewTransactionFromJsonBytes(rawTxBytes)
			if err != nil {
				fail++
				continue
			}
			if !isSent {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				if err != nil {
					fail++
					continue
				} else {
					success++
					continue
				}
			} else {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
				if err != nil {
					fail++
					continue
				}
				txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
				if err != nil {
					fail++
					continue
				}
				txMsg.(*wire.MessageTx).Transaction = tx
				err = httpServer.config.Server.PushMessageToAll(txMsg)
				if err != nil {
					fail++
					continue
				}
			}
		}
		if err == nil {
			count++
			success++
		}
	}
	return CountResult{Success: success, Fail: fail}, nil
}

func (httpServer *HttpServer) handleGetAndSendTxsFromFileV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	Logger.log.Critical(arrayParams)
	shardIDParam := int(arrayParams[0].(float64))
	txType := arrayParams[1].(string)
	isSent := arrayParams[2].(bool)
	interval := int64(arrayParams[3].(float64))
	Logger.log.Criticalf("Interval between transactions %+v \n", interval)
	datadir := "./utility/"
	Txs := []string{}
	filename := ""
	filenames := []string{}
	success := 0
	fail := 0
	count := 0
	switch txType {
	case "noprivacy":
		filename = "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-9000.json"
		filenames = append(filenames, filename)
		for i := 2; i <= 3; i++ {
			filename := "txs-shard" + fmt.Sprintf("%d", shardIDParam) + "-noprivacy-10000." + fmt.Sprintf("%d", i) + ".json"
			filenames = append(filenames, filename)
		}
	default:
		return CountResult{}, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Can't find file"))
	}
	for _, filename := range filenames {
		Logger.log.Critical("Getting Transactions from file: ", datadir+filename)
		file, err := ioutil.ReadFile(datadir + filename)
		if err != nil {
			Logger.log.Error("Fail to get Transactions from file: ", err)
			continue
		}
		data := txs{}
		_ = json.Unmarshal([]byte(file), &data)
		Logger.log.Criticalf("Get %+v Transactions from file %+v \n", len(data.Txs), filename)
		Txs = append(Txs, data.Txs...)
	}
	intervalDuration := time.Duration(interval) * time.Millisecond
	for index, txBase58Data := range Txs {
		<-time.Tick(intervalDuration)
		Logger.log.Critical("Number of Transaction: ", index)
		//<-time.Tick(50*time.Millisecond)
		rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58Data)
		if err != nil {
			fail++
			continue
		}
		switch txType {
		case "cstokenprivacy":
			{
				tx, err := transaction.NewTransactionTokenFromJsonBytes(rawTxBytes)
				if err != nil {
					fail++
					continue
				}
				if !isSent {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					if err != nil {
						fail++
						continue
					} else {
						success++
						continue
					}
				} else {
					_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
					//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
					if err != nil {
						fail++
						continue
					}
					txMsg, err := wire.MakeEmptyMessage(wire.CmdPrivacyCustomToken)
					if err != nil {
						fail++
						continue
					}
					txMsg.(*wire.MessageTxPrivacyToken).Transaction = tx
					err = httpServer.config.Server.PushMessageToAll(txMsg)
					if err != nil {
						fail++
						continue
					}
				}
				if err == nil {
					count++
					success++
				}
			}
		default:
			tx, err := transaction.NewTransactionFromJsonBytes(rawTxBytes)
			if err != nil {
				fail++
				continue
			}
			if !isSent {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				if err != nil {
					fail++
					continue
				} else {
					success++
					continue
				}
			} else {
				_, _, err = httpServer.config.TxMemPool.MaybeAcceptTransaction(tx, -1)
				//httpServer.config.NetSync.HandleCacheTxHash(*tx.Hash())
				if err != nil {
					fail++
					continue
				}
				txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
				if err != nil {
					fail++
					continue
				}
				txMsg.(*wire.MessageTx).Transaction = tx
				err = httpServer.config.Server.PushMessageToAll(txMsg)
				if err != nil {
					fail++
					continue
				}
			}
		}
		if err == nil {
			count++
			success++
		}
	}
	return CountResult{Success: success, Fail: fail}, nil
}

func (httpServer *HttpServer) handleConvertPaymentAddress(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("a payment address is needed to proceed"))
	}
	address, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("the payment address should be a string"))
	}

	convertedAddress, err := wallet.GetPaymentAddressV1(address, false)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	return convertedAddress, nil
}

// func (httpServer *HttpServer) handleTestBuildDoubleSpendTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildDoubleSpendingTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	// tx := data.(jsonresult.CreateTransactionResult)
// 	// base58CheckData := tx.Base58CheckData
// 	// newParam := make([]interface{}, 0)
// 	// newParam = append(newParam, base58CheckData)
// 	// sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
// 	// if err != nil {
// 	// 	return nil, rpcservice.NewRPCError(rpcservice.SendTxDataError, err)
// 	// }
// 	// result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, tx.ShardID)
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildDuplicateInputTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildDuplicateInputTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildOutGtInTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildOutGtInTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildReceiverExistsTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
// 	if errNewParam != nil {
// 		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	}

// 	txs, err := httpServer.txService.TestBuildReceiverExistsTransaction(createRawTxParam, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildDoubleSpendTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
// 	// createRawTxParam, errNewParam := bean.NewCreateRawPrivacyTokenTxParam(params)
// 	// if errNewParam != nil {
// 	// 	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
// 	// }

// 	txs, err := httpServer.txService.TestBuildDoubleSpendingTokenTransaction(params, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildDuplicateInputTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

// 	txs, err := httpServer.txService.TestBuildDuplicateInputTokenTransaction(params, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }

// func (httpServer *HttpServer) handleTestBuildReceiverExistsTokenTx(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

// 	txs, err := httpServer.txService.TestBuildReceiverExistsTokenTransaction(params, nil)
// 	if err != nil {
// 		// return hex for a new tx
// 		return nil, err
// 	}

// 	var result []jsonresult.CreateTransactionResult
// 	for i,_ := range txs{
// 		jsonBytes, err := json.Marshal(txs[i])
// 		if err != nil {
// 			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
// 		}
// 		result = append(result,jsonresult.NewCreateTransactionResult(txs[i].Hash(), common.EmptyString, jsonBytes, common.GetShardIDFromLastByte(txs[i].GetSenderAddrLastByte())))
// 	}
// 	return result, nil
// }