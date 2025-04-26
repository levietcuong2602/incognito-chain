package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/syncker/finishsync"
)

/*
handleGetMempoolInfo - RPC returns information about the node's current txs memory pool
*/
func (httpServer *HttpServer) handleGetMempoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if httpServer.config.BlockChain.UsingNewPool() {
		pM := httpServer.config.BlockChain.GetPoolManager()
		if pM != nil {
			return jsonresult.NewGetMempoolInfoV2(pM.GetMempoolInfo()), nil
		} else {
			return nil, rpcservice.NewRPCError(rpcservice.GeTxFromPoolError, errors.New("PoolManager is nil"))
		}
	} else {
		return jsonresult.NewGetMempoolInfo(httpServer.config.TxMemPool), nil
	}
}

/*
handleGetMempoolInfoDetails - RPC to return all txs along with their detailed info from mempool
*/
func (httpServer *HttpServer) handleGetMempoolInfoDetails(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.GetTxDetailsFromMempool(httpServer.config.TxMemPool)
	return result, nil
}

/*
handleGetRawMempool - RPC returns all transaction ids in memory pool as a json array of string transaction ids
Hint: use getmempoolentry to fetch a specific transaction from the mempool.
*/
func (httpServer *HttpServer) handleGetRawMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetRawMempoolResult(httpServer.config.TxMemPool)
	return result, nil
}

/*
handleGetPendingTxsInBlockgen - RPC returns all transaction ids in blockgen
*/
func (httpServer *HttpServer) handleGetPendingTxsInBlockgen(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetPendingTxsInBlockgenResult(httpServer.config.Blockgen.GetPendingTxsV2(255))
	return result, nil
}

func (httpServer *HttpServer) handleGetNumberOfTxsInMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := httpServer.txMemPoolService.GetNumberOfTxsInMempool()
	return result, nil
}

/*
handleMempoolEntry - RPC fetch a specific transaction from the mempool
*/
func (httpServer *HttpServer) handleMempoolEntry(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// Param #1: hash string of tx(tx id)
	txIDParam, ok := params.(string)
	if !ok || txIDParam == "" {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("transaction id is invalid"))
	}

	var txInPool metadata.Transaction
	var shardID byte
	if httpServer.txService.BlockChain.UsingNewPool() {
		var err error
		pM := httpServer.txService.BlockChain.GetPoolManager()
		if pM != nil {
			txInPool, err = pM.GetTransactionByHash(txIDParam)
			if err == nil {
				shardID = common.GetShardIDFromLastByte(txInPool.GetSenderAddrLastByte())
			}
		} else {
			err = errors.New("PoolManager is nil")
		}

		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	} else {
		tmpTxInPool, tmpShardID, err := httpServer.txMemPoolService.MempoolEntry(txIDParam)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		}

		txInPool = tmpTxInPool
		shardID = tmpShardID
	}

	tx, errM := jsonresult.NewTransactionDetail(txInPool, nil, 0, 0, shardID)
	if errM != nil {
		Logger.log.Error(errM)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errM)
	}
	tx.IsInMempool = true
	return tx, nil
}

// handleRemoveTxInMempool - try to remove tx from tx mempool
func (httpServer *HttpServer) handleRemoveTxInMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrays := common.InterfaceSlice(params)
	if arrays == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param is invalid"))
	}

	result := []bool{}
	for _, txHashString := range arrays {
		txHash, ok := txHashString.(string)
		if ok {
			t := false
			if httpServer.config.BlockChain.UsingNewPool() {
				pM := httpServer.config.BlockChain.GetPoolManager()
				if pM != nil {
					pM.RemoveTransactionInPool(txHash)
					t = true
				} else {
					return nil, rpcservice.NewRPCError(rpcservice.GeTxFromPoolError, errors.New("PoolManager is nil"))
				}
			} else {
				t, _ = httpServer.txMemPoolService.RemoveTxInMempool(txHash)
			}
			result = append(result, t)
		} else {
			result = append(result, false)
		}
	}
	return result, nil
}

// handleHasSerialNumbersInMempool - check list serial numbers existed in mempool or not
func (httpServer *HttpServer) handleHasSerialNumbersInMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	//#0: list serialnumbers in base58check encode string
	serialNumbersStr, ok := arrayParams[0].([]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("serialNumbers must be an array of string"))
	}
	return httpServer.txMemPoolService.CheckListSerialNumbersExistedInMempool(serialNumbersStr)
}

// handleHasSerialNumbersInMempool - check list serial numbers existed in mempool or not
func (httpServer *HttpServer) handleGetSyncPoolValidator(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return finishsync.DefaultFinishSyncMsgPool.GetFinishedSyncValidators(), nil
}

// handleHasSerialNumbersInMempool - check list serial numbers existed in mempool or not
func (httpServer *HttpServer) handleGetSyncPoolValidatorDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	validators := finishsync.DefaultFinishSyncMsgPool.GetFinishedSyncValidators()

	validatorsDetail := make(map[byte][]incognitokey.CommitteeKeyString)
	for k, v := range validators {
		temp, _ := incognitokey.CommitteeBase58KeyListToStruct(v)
		validatorsDetail[k] = incognitokey.CommitteeKeyListToStringList(temp)
	}

	return validatorsDetail, nil
}
