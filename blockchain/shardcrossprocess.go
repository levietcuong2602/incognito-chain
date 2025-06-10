package blockchain

import (
	"errors"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/utils"
)

// VerifyCrossShardBlock Verify CrossShard Block
// - Agg Signature
// - MerklePath
func VerifyCrossShardBlock(crossShardBlock *types.CrossShardBlock, blockchain *BlockChain, committees []incognitokey.CommitteePublicKey) error {
	utils.LogPrintf("[Cross-Shard-Evidence] ====== START VERIFYING CROSS SHARD BLOCK ======")
	utils.LogPrintf("[Cross-Shard-Evidence] Block Height: %d", crossShardBlock.Header.Height)
	utils.LogPrintf("[Cross-Shard-Evidence] From Shard ID: %d", crossShardBlock.Header.ShardID)
	utils.LogPrintf("[Cross-Shard-Evidence] To Shard ID: %d", crossShardBlock.ToShardID)
	utils.LogPrintf("[Cross-Shard-Evidence] Block Hash: %s", crossShardBlock.Hash().String())

	shardBestState := blockchain.GetBestStateShard(crossShardBlock.Header.ShardID)
	tempShardBlock := types.NewShardBlock()
	tempShardBlock.Header.CommitteeFromBlock = crossShardBlock.Header.CommitteeFromBlock
	tempShardBlock.Header.ProposeTime = crossShardBlock.Header.ProposeTime
	_, committeesForSigning, err := shardBestState.getSigningCommittees(tempShardBlock, blockchain)
	if err != nil {
		return err
	}
	utils.LogPrintf("[Cross-Shard-Evidence] Verifying Committe Signature")
	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(crossShardBlock, committeesForSigning); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	utils.LogPrintf("[Cross-Shard-Evidence] Committe Signature Verification Successful")

	utils.LogPrintf("[Cross-Shard-Evidence] Verifying Merkle Path")
	if ok := types.VerifyCrossShardBlockUTXO(crossShardBlock); !ok {
		return NewBlockChainError(HashError, errors.New("Fail to verify Merkle Path Shard"))
	}
	utils.LogPrintf("[Cross-Shard-Evidence] Merkle Path Verification Successful")
	return nil
}
