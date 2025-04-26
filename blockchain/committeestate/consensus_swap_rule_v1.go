package committeestate

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

// createSwapInstruction creates swap inst and return new validator list
// Return param:
// #1: swap inst
// #2: new pending validator list after swapped
// #3: new committees after swapped
// #4: error
func createSwapInstruction(
	pendingValidator []string,
	commitees []string,
	maxCommitteeSize int,
	minCommitteeSize int,
	shardID byte,
	offset int,
	swapOffset int,
) (*instruction.SwapInstruction, []string, []string, error) {
	newPendingValidator, newShardCommittees, shardSwapedCommittees, shardNewCommittees, err :=
		SwapValidator(pendingValidator, commitees, maxCommitteeSize, minCommitteeSize, offset, swapOffset)
	if err != nil {
		return nil, nil, nil, err
	}
	swapInstruction := instruction.NewSwapInstructionWithValue(shardNewCommittees, shardSwapedCommittees, int(shardID))
	return swapInstruction, newPendingValidator, newShardCommittees, nil
}

// assignShardCandidate Assign Candidates Into Shard Pending Validator List
// Each Shard Pending Validator List has a limit
// If a candidate is assigned into shard which Pending Validator List has reach its limit then candidate will get back into candidate list
// Otherwise, candidate will be converted to shard pending validator
// - return param #1: remain shard candidate (not assign yet)
// - return param #2: assigned candidate
func assignShardCandidate(candidates []string, numberOfPendingValidator map[byte]int, rand int64, testnetAssignOffset int, activeShards int) ([]string, map[byte][]string) {
	assignedCandidates := make(map[byte][]string)
	remainShardCandidates := []string{}
	shuffledCandidate := shuffleShardCandidate(candidates, rand)
	for _, candidate := range shuffledCandidate {
		shardID := calculateCandidateShardID(candidate, rand, activeShards)
		if numberOfPendingValidator[shardID]+1 > testnetAssignOffset {
			remainShardCandidates = append(remainShardCandidates, candidate)
			continue
		} else {
			assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			numberOfPendingValidator[shardID] += 1
		}
	}
	return remainShardCandidates, assignedCandidates
}

// shuffleShardCandidate Shuffle Position Of Shard Candidates in List with Random Number
func shuffleShardCandidate(candidates []string, rand int64) []string {
	m := make(map[string]string)
	temp := []string{}
	shuffledCandidates := []string{}
	for _, candidate := range candidates {
		seed := strconv.Itoa(int(rand)) + candidate
		hash := common.HashH([]byte(seed)).String()
		m[hash] = candidate
		temp = append(temp, hash)
	}
	if len(m) != len(temp) {
		panic("Failed To Shuffle Shard Candidate Before Assign to Shard")
	}
	sort.Strings(temp)
	for _, key := range temp {
		shuffledCandidates = append(shuffledCandidates, m[key])
	}
	if len(shuffledCandidates) != len(candidates) {
		panic("Failed To Shuffle Shard Candidate Before Assign to Shard")
	}
	return shuffledCandidates
}

// Formula ShardID: LSB[hash(candidatePubKey+randomNumber)]
// Last byte of hash(candidatePubKey+randomNumber)
func calculateCandidateShardID(candidate string, rand int64, activeShards int) (shardID byte) {
	seed := candidate + strconv.Itoa(int(rand))
	hash := common.HashB([]byte(seed))
	shardID = byte(int(hash[len(hash)-1]) % activeShards)
	Logger.log.Critical("calculateCandidateShardID/shardID", shardID)
	return shardID
}

func filterValidators(
	validators []string,
	producersBlackList map[string]uint8,
	isExistenceIncluded bool,
) []string {
	resultingValidators := []string{}
	for _, pv := range validators {
		_, found := producersBlackList[pv]
		if (found && isExistenceIncluded) || (!found && !isExistenceIncluded) {
			resultingValidators = append(resultingValidators, pv)
		}
	}
	return resultingValidators
}

func isBadProducer(badProducers []string, producer string) bool {
	for _, badProducer := range badProducers {
		if badProducer == producer {
			return true
		}
	}
	return false
}

func swap(
	badPendingValidators []string,
	goodPendingValidators []string,
	currentGoodProducers []string,
	currentBadProducers []string,
	maxCommittee int,
	offset int,
) ([]string, []string, []string, []string, error) {
	// if swap offset = 0 then do nothing
	if offset == 0 {
		// return pendingValidators, currentGoodProducers, currentBadProducers, []string{}, errors.New("no pending validator for swapping")
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, currentBadProducers, []string{}, nil
	}
	if offset > maxCommittee {
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, currentBadProducers, []string{}, errors.New("try to swap too many validators")
	}
	tempValidators := []string{}
	swapValidator := currentBadProducers
	diff := maxCommittee - len(currentGoodProducers)
	if diff >= offset {
		tempValidators = append(tempValidators, goodPendingValidators[:offset]...)
		currentGoodProducers = append(currentGoodProducers, tempValidators...)
		goodPendingValidators = goodPendingValidators[offset:]
		return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, swapValidator, tempValidators, nil
	}
	offset -= diff
	tempValidators = append(tempValidators, goodPendingValidators[:diff]...)
	goodPendingValidators = goodPendingValidators[diff:]
	currentGoodProducers = append(currentGoodProducers, tempValidators...)

	// out pubkey: swapped out validator
	swapValidator = append(swapValidator, currentGoodProducers[:offset]...)
	// unqueue validator with index from 0 to offset-1 from currentValidators list
	currentGoodProducers = currentGoodProducers[offset:]
	// in pubkey: unqueue validator with index from 0 to offset-1 from pendingValidators list
	tempValidators = append(tempValidators, goodPendingValidators[:offset]...)
	// enqueue new validator to the remaning of current validators list
	currentGoodProducers = append(currentGoodProducers, goodPendingValidators[:offset]...)
	// save new pending validators list
	goodPendingValidators = goodPendingValidators[offset:]
	return append(goodPendingValidators, badPendingValidators...), currentGoodProducers, swapValidator, tempValidators, nil
}

// SwapValidator consider these list as queue structure
// unqueue a number of validator out of currentValidators list
// enqueue a number of validator into currentValidators list <=> unqueue a number of validator out of pendingValidators list
// return value: #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator #5 error
func SwapValidator(
	pendingValidators []string,
	currentValidators []string,
	maxCommittee int,
	minCommittee int,
	offset int,
	swapOffset int,
) ([]string, []string, []string, []string, error) {
	producersBlackList := make(map[string]uint8)
	goodPendingValidators := filterValidators(pendingValidators, producersBlackList, false)
	badPendingValidators := filterValidators(pendingValidators, producersBlackList, true)
	currentBadProducers := filterValidators(currentValidators, producersBlackList, true)
	currentGoodProducers := filterValidators(currentValidators, producersBlackList, false)
	goodPendingValidatorsLen := len(goodPendingValidators)
	currentGoodProducersLen := len(currentGoodProducers)

	if currentGoodProducersLen >= minCommittee {
		if currentGoodProducersLen == maxCommittee {
			offset = swapOffset
		}
		if offset > goodPendingValidatorsLen {
			offset = goodPendingValidatorsLen
		}
		return swap(badPendingValidators, goodPendingValidators, currentGoodProducers, currentBadProducers, maxCommittee, offset)
	}

	minProducersNeeded := minCommittee - currentGoodProducersLen
	if len(pendingValidators) >= minProducersNeeded {
		if offset < minProducersNeeded {
			offset = minProducersNeeded
		} else if offset > goodPendingValidatorsLen {
			offset = goodPendingValidatorsLen
		}
		return swap(badPendingValidators, goodPendingValidators, currentGoodProducers, currentBadProducers, maxCommittee, offset)
	}

	producersNumCouldBeSwapped := len(goodPendingValidators) + len(currentValidators) - minCommittee
	swappedProducers := []string{}
	remainingProducers := []string{}
	for _, producer := range currentValidators {
		if isBadProducer(currentBadProducers, producer) && len(swappedProducers) < producersNumCouldBeSwapped {
			swappedProducers = append(swappedProducers, producer)
			continue
		}
		remainingProducers = append(remainingProducers, producer)
	}
	newProducers := append(remainingProducers, goodPendingValidators...)
	return badPendingValidators, newProducers, swappedProducers, goodPendingValidators, nil
}

// removeValidatorV1 remove validator and return removed list
// return: #param1: validator list after remove
// in parameter: #param1: list of full validator
// in parameter: #param2: list of removed validator
// removed validators list must be a subset of full validator list and it must be first in the list
func removeValidatorV1(validators []string, removedValidators []string) ([]string, error) {
	// if number of pending validator is less or equal than offset, set offset equal to number of pending validator
	if len(removedValidators) > len(validators) {
		return validators, errors.New("trying to remove too many validators")
	}
	remainingValidators := []string{}
	for _, validator := range validators {
		isRemoved := false
		for _, removedValidator := range removedValidators {
			if strings.Compare(validator, removedValidator) == 0 {
				isRemoved = true
			}
		}
		if !isRemoved {
			remainingValidators = append(remainingValidators, validator)
		}
	}
	return remainingValidators, nil
}

// Shuffle Candidate: suffer unassignedCommonPool with random number and return suffered list
// Candidate Value Concatenate with Random Number
// then Hash and Obtain Hash Value
// Sort Hash Value Then Re-arrange Candidate corresponding to Hash Value
func ShuffleCandidate(candidates []incognitokey.CommitteePublicKey, rand int64) ([]incognitokey.CommitteePublicKey, error) {
	Logger.log.Debug("Beacon Process/Shuffle Candidate: Candidate Before Sort ", candidates)
	hashes := []string{}
	m := make(map[string]incognitokey.CommitteePublicKey)
	sortedCandidate := []incognitokey.CommitteePublicKey{}
	for _, candidate := range candidates {
		candidateStr, _ := candidate.ToBase58()
		seed := candidateStr + strconv.Itoa(int(rand))
		hash := common.HashB([]byte(seed))
		hashes = append(hashes, string(hash[:32]))
		m[string(hash[:32])] = candidate
	}
	sort.Strings(hashes)
	for _, hash := range hashes {
		sortedCandidate = append(sortedCandidate, m[hash])
	}
	Logger.log.Debug("Beacon Process/Shuffle Candidate: Candidate After Sort ", sortedCandidate)
	return sortedCandidate, nil
}
