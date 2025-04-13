package bridgeagg

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type State struct {
	// unifiedTokenVaults: list of punified tokens and list pTokens are unified in each punified token
	// unifiedTokenID -> tokenID -> vault
	unifiedTokenVaults map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState

	// waitingUnshieldReqs: list of unshielding requests in the queue and it is sorted by beacon height ascending
	// unifiedTokenID -> []waitingUnshieldReq
	waitingUnshieldReqs map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq

	// bridge aggregator param
	param *statedb.BridgeAggParamState

	// temporary state
	// only contains new waiting unshield reqs in processing beacon block
	newWaitingUnshieldReqs map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq
	// only contains deteled (filled) waiting unshield reqs in processing beacon block
	deletedWaitingUnshieldReqKeyHashes []common.Hash
}

// UnifiedTokenVaults read only function do not write to result of function
func (s *State) UnifiedTokenVaults() map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState {
	return s.unifiedTokenVaults
}

// WaitingUnshieldReqs read only function do not write to result of function
func (s *State) WaitingUnshieldReqs() map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq {
	return s.waitingUnshieldReqs
}

func (s *State) DeletedWaitingUnshieldReqKeyHashes() []common.Hash {
	return s.deletedWaitingUnshieldReqKeyHashes
}

func (s *State) NewWaitingUnshieldReqs() map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq {
	return s.newWaitingUnshieldReqs
}

func (s *State) Param() *statedb.BridgeAggParamState {
	return s.param
}

func NewState() *State {
	return &State{
		unifiedTokenVaults:                 make(map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState),
		waitingUnshieldReqs:                make(map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq),
		param:                              nil,
		deletedWaitingUnshieldReqKeyHashes: []common.Hash{},
		newWaitingUnshieldReqs:             make(map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq),
	}
}

func NewStateWithValue(
	unifiedTokenInfos map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState,
	waitingUnshieldReqs map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq,
	param *statedb.BridgeAggParamState,
	newWaitingUnshieldReqs map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq,
	deletedWaitingUnshieldReqKeyHashes []common.Hash,
) *State {
	return &State{
		unifiedTokenVaults:                 unifiedTokenInfos,
		waitingUnshieldReqs:                waitingUnshieldReqs,
		param:                              param,
		newWaitingUnshieldReqs:             newWaitingUnshieldReqs,
		deletedWaitingUnshieldReqKeyHashes: deletedWaitingUnshieldReqKeyHashes,
	}
}

func (s *State) Clone() *State {
	res := NewState()
	res.unifiedTokenVaults = s.CloneUnifiedTokenVaults()
	res.waitingUnshieldReqs = s.CloneWaitingUnshieldReqs()
	if s.param != nil {
		res.param = s.param.Clone()
	}

	// reset temporary state
	res.newWaitingUnshieldReqs = map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq{}
	res.deletedWaitingUnshieldReqKeyHashes = []common.Hash{}

	return res
}

func (s *State) GetDiff(preState *State) (*State, map[common.Hash]bool, error) {
	if preState == nil {
		return nil, nil, errors.New("preState is nil")
	}

	diffState := NewState()
	newUnifiedTokens := map[common.Hash]bool{}

	// get diff unifiedTokenVaults
	for unifiedTokenID, vaults := range s.unifiedTokenVaults {
		if preVaults, found := preState.unifiedTokenVaults[unifiedTokenID]; found {
			for incTokenID, vault := range vaults {
				isUpdate := true
				if preVault, found := preVaults[incTokenID]; found {
					isDiff, err := preVault.IsDiff(vault)
					if err != nil {
						return nil, nil, err
					}
					if !isDiff {
						isUpdate = false
					}
				}
				if !isUpdate {
					continue
				}

				if diffState.unifiedTokenVaults[unifiedTokenID] == nil {
					diffState.unifiedTokenVaults[unifiedTokenID] = map[common.Hash]*statedb.BridgeAggVaultState{}
				}
				diffState.unifiedTokenVaults[unifiedTokenID][incTokenID] = vault
			}
		} else {
			// add new vaults
			newUnifiedTokens[unifiedTokenID] = true
			diffState.unifiedTokenVaults[unifiedTokenID] = vaults
		}
	}

	// get diff param
	if s.param != nil && s.param.IsDiff(preState.param) {
		diffState.param = s.param
	} else {
		diffState.param = nil
	}

	// only store new waiting unshield req in block
	// old waiting unshield reqs don't update state
	diffState.newWaitingUnshieldReqs = s.newWaitingUnshieldReqs
	diffState.deletedWaitingUnshieldReqKeyHashes = s.deletedWaitingUnshieldReqKeyHashes

	return diffState, newUnifiedTokens, nil
}

func (s *State) CloneUnifiedTokenVaults() map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState {
	res := make(map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState)
	for unifiedTokenID, vaults := range s.unifiedTokenVaults {
		res[unifiedTokenID] = make(map[common.Hash]*statedb.BridgeAggVaultState)
		for tokenID, vault := range vaults {
			res[unifiedTokenID][tokenID] = vault.Clone()
		}
	}
	return res
}

func (s *State) CloneWaitingUnshieldReqs() map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq {
	res := make(map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq)
	for unifiedTokenID, reqs := range s.waitingUnshieldReqs {
		res[unifiedTokenID] = []*statedb.BridgeAggWaitingUnshieldReq{}
		for _, req := range reqs {
			res[unifiedTokenID] = append(res[unifiedTokenID], req.Clone())
		}
	}
	return res
}

func (s *State) CloneNewWaitingUnshieldReqs() map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq {
	res := make(map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq)
	for unifiedTokenID, reqs := range s.newWaitingUnshieldReqs {
		res[unifiedTokenID] = []*statedb.BridgeAggWaitingUnshieldReq{}
		for _, req := range reqs {
			res[unifiedTokenID] = append(res[unifiedTokenID], req.Clone())
		}
	}
	return res
}

func (s *State) CloneVaultsByUnifiedTokenID(unifiedTokenID common.Hash) (map[common.Hash]*statedb.BridgeAggVaultState, error) {
	if vaults, found := s.unifiedTokenVaults[unifiedTokenID]; found {
		res := make(map[common.Hash]*statedb.BridgeAggVaultState)
		for networkID, vault := range vaults {
			res[networkID] = vault.Clone()
		}
		return res, nil
	} else {
		return nil, fmt.Errorf("Can't find unifiedTokenID %s", unifiedTokenID.String())
	}
}

func (s *State) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		UnifiedTokenVaults                 map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState `json:"UnifiedTokenVaults"`
		WaitingUnshieldReqs                map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       `json:"WaitingUnshieldReqs"`
		Param                              *statedb.BridgeAggParamState                                 `json:"Param"`
		NewWaitingUnshieldReqs             map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       `json:"NewWaitingUnshieldReqs"`
		DeletedWaitingUnshieldReqKeyHashes []common.Hash                                                `json:"DeletedWaitingUnshieldReqKeyHashes"`
	}{
		UnifiedTokenVaults:                 s.unifiedTokenVaults,
		WaitingUnshieldReqs:                s.waitingUnshieldReqs,
		Param:                              s.param,
		NewWaitingUnshieldReqs:             s.newWaitingUnshieldReqs,
		DeletedWaitingUnshieldReqKeyHashes: s.deletedWaitingUnshieldReqKeyHashes,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (s *State) UnmarshalJSON(data []byte) error {
	temp := struct {
		UnifiedTokenVaults                 map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState `json:"UnifiedTokenVaults"`
		WaitingUnshieldReqs                map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       `json:"WaitingUnshieldReqs"`
		Param                              *statedb.BridgeAggParamState                                 `json:"Param"`
		NewWaitingUnshieldReqs             map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       `json:"NewWaitingUnshieldReqs"`
		DeletedWaitingUnshieldReqKeyHashes []common.Hash                                                `json:"DeletedWaitingUnshieldReqKeyHashes"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	s.unifiedTokenVaults = temp.UnifiedTokenVaults
	s.waitingUnshieldReqs = temp.WaitingUnshieldReqs
	s.param = temp.Param
	s.newWaitingUnshieldReqs = temp.NewWaitingUnshieldReqs
	s.deletedWaitingUnshieldReqKeyHashes = temp.DeletedWaitingUnshieldReqKeyHashes
	return nil
}
