package statedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggVaultState struct {
	// vault's volume amount - available for unshield
	amount uint64
	// vault's amount was locked with current waiting unshield reqs
	lockedAmount uint64
	// total shortage unshield amount of current waiting unshield reqs
	waitingUnshieldAmount uint64
	// total unshield fee corresponding to waitingUnshieldAmount
	waitingUnshieldFee uint64

	extDecimal uint8
	networkID  uint8
	incTokenID common.Hash
}

func (state *BridgeAggVaultState) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount                uint64      `json:"Amount"`
		LockedAmount          uint64      `json:"LockedAmount"`
		WaitingUnshieldAmount uint64      `json:"WaitingUnshieldAmount"`
		WaitingUnshieldFee    uint64      `json:"WaitingUnshieldFee"`
		ExtDecimal            uint8       `json:"ExtDecimal"`
		NetworkID             uint8       `json:"NetworkID"`
		IncTokenID            common.Hash `json:"IncTokenID"`
	}{
		Amount:                state.amount,
		LockedAmount:          state.lockedAmount,
		WaitingUnshieldAmount: state.waitingUnshieldAmount,
		WaitingUnshieldFee:    state.waitingUnshieldFee,
		ExtDecimal:            state.extDecimal,
		NetworkID:             state.networkID,
		IncTokenID:            state.incTokenID,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (state *BridgeAggVaultState) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount                uint64      `json:"Amount"`
		LockedAmount          uint64      `json:"LockedAmount"`
		WaitingUnshieldAmount uint64      `json:"WaitingUnshieldAmount"`
		WaitingUnshieldFee    uint64      `json:"WaitingUnshieldFee"`
		ExtDecimal            uint8       `json:"ExtDecimal"`
		NetworkID             uint8       `json:"NetworkID"`
		IncTokenID            common.Hash `json:"IncTokenID"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	state.amount = temp.Amount
	state.lockedAmount = temp.LockedAmount
	state.waitingUnshieldAmount = temp.WaitingUnshieldAmount
	state.waitingUnshieldFee = temp.WaitingUnshieldFee
	state.extDecimal = temp.ExtDecimal
	state.networkID = temp.NetworkID
	state.incTokenID = temp.IncTokenID
	return nil
}

func NewBridgeAggVaultState() *BridgeAggVaultState {
	return &BridgeAggVaultState{}
}

func NewBridgeAggVaultStateWithValue(
	amount, lockedAmount, waitingUnshieldAmount, waitingUnshieldFee uint64, extDecimal uint8, networkID uint8, tokenID common.Hash,
) *BridgeAggVaultState {
	return &BridgeAggVaultState{
		amount:                amount,
		lockedAmount:          lockedAmount,
		waitingUnshieldAmount: waitingUnshieldAmount,
		waitingUnshieldFee:    waitingUnshieldFee,
		extDecimal:            extDecimal,
		networkID:             networkID,
		incTokenID:            tokenID,
	}
}

func (b *BridgeAggVaultState) Amount() uint64 {
	return b.amount
}

func (b *BridgeAggVaultState) LockedAmount() uint64 {
	return b.lockedAmount
}

func (b *BridgeAggVaultState) WaitingUnshieldAmount() uint64 {
	return b.waitingUnshieldAmount
}

func (b *BridgeAggVaultState) WaitingUnshieldFee() uint64 {
	return b.waitingUnshieldFee
}

func (b *BridgeAggVaultState) ExtDecimal() uint8 {
	return b.extDecimal
}

func (b *BridgeAggVaultState) NetworkID() uint8 {
	return b.networkID
}

func (b *BridgeAggVaultState) IncTokenID() common.Hash {
	return b.incTokenID
}

func (b *BridgeAggVaultState) SetAmount(amount uint64) {
	b.amount = amount
}

func (b *BridgeAggVaultState) SetLockedAmount(amount uint64) {
	b.lockedAmount = amount
}

func (b *BridgeAggVaultState) SetWaitingUnshieldAmount(amount uint64) {
	b.waitingUnshieldAmount = amount
}

func (b *BridgeAggVaultState) SetWaitingUnshieldFee(amount uint64) {
	b.waitingUnshieldFee = amount
}

func (b *BridgeAggVaultState) SetExtDecimal(extDecimal uint8) {
	b.extDecimal = extDecimal
}

func (b *BridgeAggVaultState) SetNetworkID(networkID uint8) {
	b.networkID = networkID
}

func (b *BridgeAggVaultState) SetIncTokenID(tokenID common.Hash) {
	b.incTokenID = tokenID
}

func (b *BridgeAggVaultState) UpdateAmount(amount uint64, operator int) error {
	tmpAmt := uint64(0)
	switch operator {
	case common.SubOperator:
		{
			tmpAmt = b.amount - amount
			if tmpAmt > b.amount {
				return errors.New("decrease vault amount out of range uint64")
			}
		}
	case common.AddOperator:
		{
			tmpAmt = b.amount + amount
			if tmpAmt < b.amount {
				return errors.New("increase vault amount out of range uint64")
			}
		}
	default:
		return errors.New("invalid operator")
	}

	b.amount = tmpAmt
	return nil
}

func (b *BridgeAggVaultState) UpdateLockedAmount(amount uint64, operator int) error {
	tmpAmt := uint64(0)
	switch operator {
	case common.SubOperator:
		{
			tmpAmt = b.lockedAmount - amount
			if tmpAmt > b.lockedAmount {
				return errors.New("decrease vault locked amount out of range uint64")
			}
		}
	case common.AddOperator:
		{
			tmpAmt = b.lockedAmount + amount
			if tmpAmt < b.lockedAmount {
				return errors.New("increase vault locked amount out of range uint64")
			}
		}
	default:
		return errors.New("invalid operator")
	}

	b.lockedAmount = tmpAmt
	return nil
}

func (b *BridgeAggVaultState) UpdateWaitingUnshieldAmount(amount uint64, operator int) error {
	tmpAmt := uint64(0)
	switch operator {
	case common.SubOperator:
		{
			tmpAmt = b.waitingUnshieldAmount - amount
			if tmpAmt > b.waitingUnshieldAmount {
				return errors.New("decrease vault waiting unshield amount out of range uint64")
			}
		}
	case common.AddOperator:
		{
			tmpAmt = b.waitingUnshieldAmount + amount
			if tmpAmt < b.waitingUnshieldAmount {
				return errors.New("increase vault waiting unshield amount out of range uint64")
			}
		}
	default:
		return errors.New("invalid operator")
	}

	b.waitingUnshieldAmount = tmpAmt
	return nil
}

func (b *BridgeAggVaultState) UpdateWaitingUnshieldFee(amount uint64, operator int) error {
	tmpAmt := uint64(0)
	switch operator {
	case common.SubOperator:
		{
			tmpAmt = b.waitingUnshieldFee - amount
			if tmpAmt > b.waitingUnshieldFee {
				return errors.New("decrease vault waiting unshield fee out of range uint64")
			}
		}
	case common.AddOperator:
		{
			tmpAmt = b.waitingUnshieldFee + amount
			if tmpAmt < b.waitingUnshieldFee {
				return errors.New("increase vault waiting unshield fee out of range uint64")
			}
		}
	default:
		return errors.New("invalid operator")
	}

	b.waitingUnshieldFee = tmpAmt
	return nil
}

func (b *BridgeAggVaultState) Clone() *BridgeAggVaultState {
	return &BridgeAggVaultState{
		amount:                b.amount,
		lockedAmount:          b.lockedAmount,
		waitingUnshieldAmount: b.waitingUnshieldAmount,
		waitingUnshieldFee:    b.waitingUnshieldFee,
		extDecimal:            b.extDecimal,
		networkID:             b.networkID,
		incTokenID:            b.incTokenID,
	}
}

func (b *BridgeAggVaultState) GetDiff(compareState *BridgeAggVaultState) (*BridgeAggVaultState, error) {
	if compareState == nil {
		return nil, errors.New("compareState is nil")
	}
	if b.amount != compareState.amount || b.lockedAmount != compareState.lockedAmount ||
		b.waitingUnshieldAmount != compareState.waitingUnshieldAmount || b.waitingUnshieldFee != compareState.waitingUnshieldFee ||
		b.extDecimal != compareState.extDecimal ||
		b.networkID != compareState.networkID || b.incTokenID != compareState.incTokenID {
		return b.Clone(), nil
	}
	return nil, nil
}

func (b *BridgeAggVaultState) IsDiff(compareState *BridgeAggVaultState) (bool, error) {
	if compareState == nil {
		return false, errors.New("compareState is nil")
	}
	if b.amount != compareState.amount || b.lockedAmount != compareState.lockedAmount ||
		b.waitingUnshieldAmount != compareState.waitingUnshieldAmount || b.waitingUnshieldFee != compareState.waitingUnshieldFee ||
		b.extDecimal != compareState.extDecimal ||
		b.networkID != compareState.networkID || b.incTokenID != compareState.incTokenID {
		return true, nil
	}
	return false, nil
}

func (b *BridgeAggVaultState) IsEmpty() bool {
	return b.amount == 0 && b.lockedAmount == 0 &&
		b.waitingUnshieldAmount == 0 && b.waitingUnshieldFee == 0 &&
		b.extDecimal == 0 && b.incTokenID == common.Hash{}
}

type BridgeAggVaulltObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	hash       common.Hash
	state      *BridgeAggVaultState
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newBridgeAggVaultObject(db *StateDB, hash common.Hash) *BridgeAggVaulltObject {
	return &BridgeAggVaulltObject{
		version:    defaultVersion,
		db:         db,
		hash:       hash,
		state:      NewBridgeAggVaultState(),
		objectType: BridgeAggVaultObjectType,
		deleted:    false,
	}
}

func newBridgeAggVaultObjectWithValue(db *StateDB, key common.Hash, data interface{}) (
	*BridgeAggVaulltObject, error,
) {
	var newBridgeAggVaultState = NewBridgeAggVaultState()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, newBridgeAggVaultState)
		if err != nil {
			return nil, err
		}
	} else {
		newBridgeAggVaultState, ok = data.(*BridgeAggVaultState)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggVaultStateType, reflect.TypeOf(data))
		}
	}
	return &BridgeAggVaulltObject{
		version:    defaultVersion,
		hash:       key,
		state:      newBridgeAggVaultState,
		db:         db,
		objectType: BridgeAggVaultObjectType,
		deleted:    false,
	}, nil
}

func generateBridgeAggVaultObjectPrefix(unifiedTokenID common.Hash) []byte {
	b := append(GetBridgeAggVaultPrefix(), unifiedTokenID.Bytes()...)
	h := common.HashH(b)
	return h[:prefixHashKeyLength]
}

func GenerateBridgeAggVaultObjectKey(unifiedTokenID, tokenID common.Hash) common.Hash {
	prefixHash := generateBridgeAggVaultObjectPrefix(unifiedTokenID)
	valueHash := common.HashH(tokenID.Bytes())
	return common.BytesToHash(append(prefixHash, valueHash[:prefixKeyLength]...))
}

func (object *BridgeAggVaulltObject) GetVersion() int {
	return object.version
}

// setError remembers the first non-nil error it is called with.
func (object *BridgeAggVaulltObject) SetError(err error) {
	if object.dbErr == nil {
		object.dbErr = err
	}
}

func (object *BridgeAggVaulltObject) GetTrie(db DatabaseAccessWarper) Trie {
	return object.trie
}

func (object *BridgeAggVaulltObject) SetValue(data interface{}) error {
	newBridgeAggVaultState, ok := data.(*BridgeAggVaultState)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidBridgeAggVaultStateType, reflect.TypeOf(data))
	}
	object.state = newBridgeAggVaultState
	return nil
}

func (object *BridgeAggVaulltObject) GetValue() interface{} {
	return object.state
}

func (object *BridgeAggVaulltObject) GetValueBytes() []byte {
	state, ok := object.GetValue().(*BridgeAggVaultState)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(state)
	if err != nil {
		panic("failed to marshal bridge agg vault state")
	}
	return value
}

func (object *BridgeAggVaulltObject) GetHash() common.Hash {
	return object.hash
}

func (object *BridgeAggVaulltObject) GetType() int {
	return object.objectType
}

// MarkDelete will delete an object in trie
func (object *BridgeAggVaulltObject) MarkDelete() {
	object.deleted = true
}

// reset all shard committee value into default value
func (object *BridgeAggVaulltObject) Reset() bool {
	object.state = NewBridgeAggVaultState()
	return true
}

func (object *BridgeAggVaulltObject) IsDeleted() bool {
	return object.deleted
}

// value is either default or nil
func (object *BridgeAggVaulltObject) IsEmpty() bool {
	temp := NewBridgeAggVaultState()
	return reflect.DeepEqual(temp, object.state) || object.state == nil
}
