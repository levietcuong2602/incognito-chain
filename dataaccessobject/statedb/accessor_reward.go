package statedb

import (
	"encoding/json"
	"fmt"
	"github.com/levietcuong2602/incognito-chain/trie"

	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/common/base58"
	"github.com/levietcuong2602/incognito-chain/privacy/key"
)

// AddShardRewardRequestMultiset
func AddShardRewardRequestMultiset(
	stateDB *StateDB,
	epoch uint64,
	shardID, subsetID byte,
	tokenID common.Hash,
	rewardAmount uint64,
) error {
	key := generateRewardRequestMultisetObjectKey(epoch, shardID, subsetID, tokenID)
	r, has, err := stateDB.getRewardRequestStateV3(key)
	if err != nil {
		return NewStatedbError(StoreRewardRequestError, err)
	}
	if has {
		rewardAmount += r.Amount()
	}
	value := NewRewardRequestStateV3WithValue(epoch, shardID, subsetID, tokenID, rewardAmount)
	err = stateDB.SetStateObject(RewardRequestV3ObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreRewardRequestError, err)
	}
	return nil
}

func GetRewardOfShardByEpochMultiset(
	stateDB *StateDB,
	epoch uint64,
	shardID, subsetID byte,
	tokenID common.Hash,
) (uint64, error) {
	key := generateRewardRequestMultisetObjectKey(epoch, shardID, subsetID, tokenID)
	amount, has, err := stateDB.getRewardRequestAmountV3(key)
	if err != nil {
		return 0, NewStatedbError(GetRewardRequestError, err)
	}
	if !has {
		return 0, nil
	}
	return amount, nil
}

func GetAllTokenIDForRewardMultiset(stateDB *StateDB, epoch uint64) []common.Hash {
	_, rewardRequestStates := stateDB.getAllRewardRequestStateV3(epoch)
	m := make(map[common.Hash]struct{})
	tokenIDs := []common.Hash{}
	for _, rewardRequestState := range rewardRequestStates {
		m[rewardRequestState.TokenID()] = struct{}{}
	}
	for k, _ := range m {
		tokenIDs = append(tokenIDs, k)
	}
	return tokenIDs
}

// Reward in Beacon
func AddShardRewardRequest(stateDB *StateDB, epoch uint64, shardID byte, tokenID common.Hash, rewardAmount uint64) error {
	key := GenerateRewardRequestObjectKey(epoch, shardID, tokenID)
	r, has, err := stateDB.getRewardRequestState(key)
	if err != nil {
		return NewStatedbError(StoreRewardRequestError, err)
	}
	if has {
		rewardAmount += r.Amount()
	}
	value := NewRewardRequestStateWithValue(epoch, shardID, tokenID, rewardAmount)
	err = stateDB.SetStateObject(RewardRequestObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreRewardRequestError, err)
	}
	return nil
}

func GetRewardOfShardByEpoch(stateDB *StateDB, epoch uint64, shardID byte, tokenID common.Hash) (uint64, error) {
	key := GenerateRewardRequestObjectKey(epoch, shardID, tokenID)
	amount, has, err := stateDB.getRewardRequestAmount(key)
	if err != nil {
		return 0, NewStatedbError(GetRewardRequestError, err)
	}
	if !has {
		return 0, nil
	}
	return amount, nil
}

func GetAllTokenIDForReward(stateDB *StateDB, epoch uint64) []common.Hash {
	_, rewardRequestStates := stateDB.getAllRewardRequestState(epoch)
	tokenIDs := []common.Hash{}
	for _, rewardRequestState := range rewardRequestStates {
		tokenIDs = append(tokenIDs, rewardRequestState.TokenID())
	}
	return tokenIDs
}

func RemoveRewardOfShardByEpoch(stateDB *StateDB, epoch uint64) {
	rewardRequestKeys, _ := stateDB.getAllRewardRequestState(epoch)
	for _, k := range rewardRequestKeys {
		stateDB.MarkDeleteStateObject(RewardRequestObjectType, k)
	}
}

// Reward in Shard
func AddCommitteeReward(stateDB *StateDB, incognitoPublicKey string, committeeReward uint64, tokenID common.Hash) error {
	key, err := GenerateCommitteeRewardObjectKey(incognitoPublicKey)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	c, has, err := stateDB.getCommitteeRewardState(key)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	committeeRewardM := make(map[common.Hash]uint64)
	if has {
		committeeRewardM = c.Reward()
	}
	amount, ok := committeeRewardM[tokenID]
	if ok {
		committeeReward += amount
	}
	committeeRewardM[tokenID] = committeeReward
	value := NewCommitteeRewardStateWithValue(committeeRewardM, incognitoPublicKey)
	err = stateDB.SetStateObject(CommitteeRewardObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	return nil
}

func GetCommitteeReward(stateDB *StateDB, incognitoPublicKey string, tokenID common.Hash) (uint64, error) {
	key, err := GenerateCommitteeRewardObjectKey(incognitoPublicKey)
	if err != nil {
		return 0, NewStatedbError(GetCommitteeRewardError, err)
	}
	r, has, err := stateDB.getCommitteeRewardAmount(key)
	if err != nil {
		return 0, NewStatedbError(GetCommitteeRewardError, err)
	}
	if !has {
		return 0, nil
	}
	if amount, ok := r[tokenID]; !ok {
		return 0, nil
	} else {
		return amount, nil
	}
}

func ListCommitteeReward(stateDB *StateDB) map[string]map[common.Hash]uint64 {
	return stateDB.getAllCommitteeReward()
}

func RemoveCommitteeReward(stateDB *StateDB, incognitoPublicKeyBytes []byte, withdrawAmount uint64, tokenID common.Hash) error {
	incognitoPublicKey := base58.Base58Check{}.Encode(incognitoPublicKeyBytes, common.Base58Version)
	key, err := GenerateCommitteeRewardObjectKey(incognitoPublicKey)
	if err != nil {
		return NewStatedbError(RemoveCommitteeRewardError, err)
	}
	c, has, err := stateDB.getCommitteeRewardState(key)
	if err != nil {
		return NewStatedbError(RemoveCommitteeRewardError, err)
	}
	if !has {
		return nil
	}
	committeeRewardM := c.Reward()
	currentReward := committeeRewardM[tokenID]
	if withdrawAmount > currentReward {
		return NewStatedbError(RemoveCommitteeRewardError, fmt.Errorf("Current Reward %+v but got withdraw %+v", currentReward, withdrawAmount))
	}
	remain := currentReward - withdrawAmount
	if remain == 0 {
		delete(committeeRewardM, tokenID)
	} else {
		committeeRewardM[tokenID] = remain
	}
	if len(committeeRewardM) == 0 {
		stateDB.MarkDeleteStateObject(CommitteeRewardObjectType, key)
		return nil
	}
	value := NewCommitteeRewardStateWithValue(committeeRewardM, incognitoPublicKey)
	err = stateDB.SetStateObject(CommitteeRewardObjectType, key, value)
	if err != nil {
		return NewStatedbError(StoreCommitteeRewardError, err)
	}
	return nil
}

// ================================= Testing ======================================
func GetRewardRequestInfoByEpoch(stateDB *StateDB, epoch uint64) []*RewardRequestState {
	_, rewardRequestStates := stateDB.getAllRewardRequestState(epoch)
	return rewardRequestStates
}

// ================================= Delegation Reward =============================
func StoreDelegationReward(stateDB *StateDB, incognitoPublicKeyBytes []byte, incognitoPaymentAddress key.PaymentAddress, shardCPK string, epoch int, beaconUID string, amount uint64) error {
	incognitoPublicKey := base58.Base58Check{}.Encode(incognitoPublicKeyBytes, common.Base58Version)
	key, err := GenerateDelegateRewardObjectKey(incognitoPublicKey)
	reward, _, err := GetDelegationReward(stateDB, incognitoPublicKeyBytes)
	if err != nil {
		return err
	}
	reward.incognitoPublicKey = incognitoPublicKey
	reward.incognitoPaymentAddress = incognitoPaymentAddress
	if reward.Reward[shardCPK] == nil {
		reward.Reward[shardCPK] = map[int]DelegateInfo{}
	}
	reward.Reward[shardCPK][epoch] = DelegateInfo{
		beaconUID, amount,
	}
	err = stateDB.SetStateObject(DelegationRewardObjectType, key, reward)
	if err != nil {
		return NewStatedbError(StoreDelegationRewardError, err)
	}
	return nil
}

func ResetDelegationReward(stateDB *StateDB, incognitoPublicKeyBytes []byte, incognitoPaymentAddress key.PaymentAddress, epoch int) error {
	incognitoPublicKey := base58.Base58Check{}.Encode(incognitoPublicKeyBytes, common.Base58Version)
	key, err := GenerateDelegateRewardObjectKey(incognitoPublicKey)
	reward, _, err := GetDelegationReward(stateDB, incognitoPublicKeyBytes)
	if err != nil {
		return err
	}
	reward.incognitoPublicKey = incognitoPublicKey
	reward.incognitoPaymentAddress = incognitoPaymentAddress
	for shardCPK, dInfoByEpoch := range reward.Reward {
		latestDelegateEpoch := 0
		for k, _ := range dInfoByEpoch {
			if latestDelegateEpoch < k {
				latestDelegateEpoch = k
			}
		}
		latestInfo := dInfoByEpoch[latestDelegateEpoch]
		dInfoByEpoch = map[int]DelegateInfo{}
		if latestInfo.BeaconUID != "" {
			dInfoByEpoch = map[int]DelegateInfo{
				epoch: latestInfo,
			}
		}
		reward.Reward[shardCPK] = dInfoByEpoch
	}
	err = stateDB.SetStateObject(DelegationRewardObjectType, key, reward)
	if err != nil {
		return NewStatedbError(StoreDelegationRewardError, err)
	}
	return nil
}

func GetDelegationReward(stateDB *StateDB, incognitoPublicKeyBytes []byte) (*DelegationRewardState, bool, error) {
	incognitoPublicKey := base58.Base58Check{}.Encode(incognitoPublicKeyBytes, common.Base58Version)
	key, err := GenerateDelegateRewardObjectKey(incognitoPublicKey)
	if err != nil {
		return nil, false, err
	}
	return stateDB.getDelegationRewardState(key)
}

func ListDelegationReward(stateDB *StateDB) (map[string]*DelegationRewardState, error) {
	m := map[string]*DelegationRewardState{}
	keys := []common.Hash{}
	prefixHash := GetDelegationRewardPrefix()
	temp := stateDB.trie.NodeIterator(prefixHash)
	it := trie.NewIterator(temp)
	for it.Next(true, false, true) {
		key := it.Key
		newKey := make([]byte, len(key))
		copy(newKey, key)
		keys = append(keys, common.BytesToHash(newKey))
		value := it.Value
		newValue := make([]byte, len(value))
		copy(newValue, value)
		delegateRewardState := NewDelegationRewardState()
		err := json.Unmarshal(newValue, delegateRewardState)
		if err != nil {
			panic("wrong value type")
		}
		m[delegateRewardState.incognitoPublicKey] = delegateRewardState
	}
	return m, nil
}
