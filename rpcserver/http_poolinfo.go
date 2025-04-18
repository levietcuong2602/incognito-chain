package rpcserver

import (
	"errors"
	"github.com/levietcuong2602/incognito-chain/blockchain"
	"github.com/levietcuong2602/incognito-chain/blockchain/types"

	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/multiview"
	"github.com/levietcuong2602/incognito-chain/rpcserver/jsonresult"
	"github.com/levietcuong2602/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) hanldeGetFeatureStats(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := blockchain.DefaultFeatureStat.Report(httpServer.config.BlockChain.GetBeaconBestState())
	return result, nil
}

func (httpServer *HttpServer) hanldeGetSyncStats(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	stats := httpServer.synkerService.GetSyncStats()
	result := jsonresult.NewSyncStats(stats)
	return result, nil
}

func (httpServer *HttpServer) hanldeGetBeaconPoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("hanldeGetBeaconPoolInfo params: %+v", params)
	blks := httpServer.synkerService.GetBeaconPoolInfo()
	result := jsonresult.NewPoolInfo(blks)
	Logger.log.Debugf("hanldeGetBeaconPoolInfo result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) hanldeGetShardPoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}
	Logger.log.Debugf("hanldeGetShardPoolInfo params: %+v", params)
	blks := httpServer.synkerService.GetShardPoolInfo(int(shardID))
	result := jsonresult.NewPoolInfo(blks)
	Logger.log.Debugf("handleGetShardPoolInfo result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) hanldeGetCrossShardPoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID invalid"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}
	Logger.log.Debugf("hanldeGetCrossShardPoolInfo params: %+v", params)
	blks := httpServer.synkerService.GetCrossShardPoolInfo(int(shardID))
	result := jsonresult.NewPoolInfo(blks)
	Logger.log.Debugf("hanldeGetCrossShardPoolInfo result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) hanldeGetAllView(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Invalid param, param 0 must be shardid, 1 is number of blk estimate"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}
	numOfBlks, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Block height component invalid"))
	}
	Logger.log.Debugf("hanldeGetCrossShardPoolInfo params: %+v", params)
	blkOnChain, err := httpServer.blockService.GetBlocks(int(shardID), int(numOfBlks))
	if err != nil {
		return nil, err
	}
	res := []jsonresult.GetViewResult{}
	blksPool := []types.BlockPoolInterface{}
	if shardID == -1 {
		blks := blkOnChain.([]jsonresult.GetBeaconBlockResult)
		if len(blks) == 0 {
			return nil, nil
		}
		blksPool = httpServer.synkerService.GetAllViewBeaconByHash(blks[len(blks)-1].Hash)
		for _, blk := range blks {
			res = append(res, jsonresult.GetViewResult{
				Hash:              blk.Hash,
				PreviousBlockHash: blk.PreviousBlockHash,
				Height:            blk.Height,
				Round:             uint64(blk.Round),
			})
		}
	} else {
		blks := blkOnChain.([]jsonresult.GetShardBlockResult)
		if len(blks) == 0 {
			return nil, nil
		}
		blksPool = httpServer.synkerService.GetAllViewShardByHash(blks[len(blks)-1].Hash, int(shardID))
		for _, blk := range blks {
			res = append(res, jsonresult.GetViewResult{
				Hash:              blk.Hash,
				PreviousBlockHash: blk.PreviousBlockHash,
				Height:            blk.Height,
				Round:             uint64(blk.Round),
			})
		}
	}
	for _, blk := range blksPool {
		res = append(res, jsonresult.GetViewResult{
			Hash:              blk.Hash().String(),
			PreviousBlockHash: blk.GetPrevHash().String(),
			Height:            blk.GetHeight(),
			Round:             uint64(blk.GetRound()),
		})
	}
	return res, nil
}

func (httpServer *HttpServer) hanldeGetAllViewDetail(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Invalid param, param 0 must be shardid"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}
	res := []jsonresult.GetViewResult{}
	var views []multiview.View
	if shardID == -1 {
		views = httpServer.config.BlockChain.BeaconChain.GetAllView()
	} else {
		sChain := httpServer.config.BlockChain.ShardChain[int(shardID)]
		if sChain != nil {
			views = sChain.GetAllView()
		}

	}

	for _, view := range views {
		res = append(res, jsonresult.GetViewResult{
			Hash:              view.GetHash().String(),
			PreviousBlockHash: view.GetPreviousHash().String(),
			Height:            view.GetHeight(),
			Round:             uint64(view.GetBlock().GetRound()),
			ProposeTime:       view.CalculateTimeSlot(view.GetBlock().GetProposeTime()),
			ProduceTime:       view.CalculateTimeSlot(view.GetBlock().GetProduceTime()),
		})
	}
	return res, nil
}

func (httpServer *HttpServer) hanldeIsInstantFinality(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Invalid param, param 0 must be shardid"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}
	var isInstantFinality bool
	if shardID == -1 {
		isInstantFinality = httpServer.config.BlockChain.BeaconChain.GetMultiView().IsInstantFinality()
	} else {
		sChain := httpServer.config.BlockChain.ShardChain[int(shardID)]
		if sChain != nil {
			isInstantFinality = sChain.GetMultiView().IsInstantFinality()
		}

	}
	return isInstantFinality, nil
}
