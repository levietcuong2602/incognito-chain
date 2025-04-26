package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type ExternalFeeInfo struct {
	NetworkFee    uint
	RBFReqIncTxID string
}

type ProcessedUnshieldRequestBatch struct {
	batchID      string
	unshieldsID  []string
	utxos        []*UTXO
	externalFees map[uint64]ExternalFeeInfo // beaconHeight => fee
}

func (us *ProcessedUnshieldRequestBatch) GetUTXOs() []*UTXO {
	return us.utxos
}

func (us *ProcessedUnshieldRequestBatch) SetUTXOs(newUTXOs []*UTXO) {
	us.utxos = newUTXOs
}

func (us *ProcessedUnshieldRequestBatch) GetUnshieldRequests() []string {
	return us.unshieldsID
}

func (us *ProcessedUnshieldRequestBatch) SetUnshieldRequests(usRequests []string) {
	us.unshieldsID = usRequests
}

func (us *ProcessedUnshieldRequestBatch) GetExternalFees() map[uint64]ExternalFeeInfo {
	return us.externalFees
}

func (us *ProcessedUnshieldRequestBatch) SetExternalFees(externalFees map[uint64]ExternalFeeInfo) {
	us.externalFees = externalFees
}

func (us *ProcessedUnshieldRequestBatch) GetBatchID() string {
	return us.batchID
}

func (us *ProcessedUnshieldRequestBatch) SetBatchID(batchID string) {
	us.batchID = batchID
}

func (rq ProcessedUnshieldRequestBatch) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		BatchID      string
		UnshieldsID  []string
		UTXOs        []*UTXO
		ExternalFees map[uint64]ExternalFeeInfo
	}{
		BatchID:      rq.batchID,
		UnshieldsID:  rq.unshieldsID,
		UTXOs:        rq.utxos,
		ExternalFees: rq.externalFees,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (rq *ProcessedUnshieldRequestBatch) UnmarshalJSON(data []byte) error {
	temp := struct {
		BatchID      string
		UnshieldsID  []string
		UTXOs        []*UTXO
		ExternalFees map[uint64]ExternalFeeInfo
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	rq.unshieldsID = temp.UnshieldsID
	rq.utxos = temp.UTXOs
	rq.externalFees = temp.ExternalFees
	rq.batchID = temp.BatchID
	return nil
}

func NewProcessedUnshieldRequestBatchWithValue(
	batchID string,
	unshieldsIDInput []string,
	utxosInput []*UTXO,
	externalFees map[uint64]ExternalFeeInfo) *ProcessedUnshieldRequestBatch {
	return &ProcessedUnshieldRequestBatch{
		batchID:      batchID,
		unshieldsID:  unshieldsIDInput,
		utxos:        utxosInput,
		externalFees: externalFees,
	}
}

func NewProcessedUnshieldRequestBatch() *ProcessedUnshieldRequestBatch {
	return &ProcessedUnshieldRequestBatch{}
}

type ProcessUnshieldRequestBatchObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version                           int
	processedUnshieldRequestBatchHash common.Hash
	processedUnshieldRequestBatch     *ProcessedUnshieldRequestBatch
	objectType                        int
	deleted                           bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newProcessUnshieldRequestBatchObject(db *StateDB, hash common.Hash) *ProcessUnshieldRequestBatchObject {
	return &ProcessUnshieldRequestBatchObject{
		version:                           defaultVersion,
		db:                                db,
		processedUnshieldRequestBatchHash: hash,
		processedUnshieldRequestBatch:     NewProcessedUnshieldRequestBatch(),
		objectType:                        PortalProcessedUnshieldRequestBatchObjectType,
		deleted:                           false,
	}
}

func newProcessUnshieldRequestBatchObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*ProcessUnshieldRequestBatchObject, error) {
	var content = NewProcessedUnshieldRequestBatch()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, content)
		if err != nil {
			return nil, err
		}
	} else {
		content, ok = data.(*ProcessedUnshieldRequestBatch)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidPortalV4BatchUnshieldRequestType, reflect.TypeOf(data))
		}
	}
	return &ProcessUnshieldRequestBatchObject{
		version:                           defaultVersion,
		processedUnshieldRequestBatchHash: key,
		processedUnshieldRequestBatch:     content,
		db:                                db,
		objectType:                        PortalProcessedUnshieldRequestBatchObjectType,
		deleted:                           false,
	}, nil
}

func GenerateProcessedUnshieldRequestBatchObjectKey(tokenID string, batchID string) common.Hash {
	prefixHash := GetProcessedUnshieldRequestBatchPrefix(tokenID)
	valueHash := common.HashH([]byte(batchID))
	return common.BytesToHash(append(prefixHash, valueHash[:][:prefixKeyLength]...))
}

func (t ProcessUnshieldRequestBatchObject) GetVersion() int {
	return t.version
}

// setError remembers the first non-nil error it is called with.
func (t *ProcessUnshieldRequestBatchObject) SetError(err error) {
	if t.dbErr == nil {
		t.dbErr = err
	}
}

func (t ProcessUnshieldRequestBatchObject) GetTrie(db DatabaseAccessWarper) Trie {
	return t.trie
}

func (t *ProcessUnshieldRequestBatchObject) SetValue(data interface{}) error {
	processedUnshieldBatch, ok := data.(*ProcessedUnshieldRequestBatch)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidPortalV4BatchUnshieldRequestType, reflect.TypeOf(data))
	}
	t.processedUnshieldRequestBatch = processedUnshieldBatch
	return nil
}

func (t ProcessUnshieldRequestBatchObject) GetValue() interface{} {
	return t.processedUnshieldRequestBatch
}

func (t ProcessUnshieldRequestBatchObject) GetValueBytes() []byte {
	ProcessUnshield, ok := t.GetValue().(*ProcessedUnshieldRequestBatch)
	if !ok {
		panic("wrong expected value type")
	}
	value, err := json.Marshal(ProcessUnshield)
	if err != nil {
		panic("failed to marshal redeem request")
	}
	return value
}

func (t ProcessUnshieldRequestBatchObject) GetHash() common.Hash {
	return t.processedUnshieldRequestBatchHash
}

func (t ProcessUnshieldRequestBatchObject) GetType() int {
	return t.objectType
}

// MarkDelete will delete an object in trie
func (t *ProcessUnshieldRequestBatchObject) MarkDelete() {
	t.deleted = true
}

// reset all shard committee value into default value
func (t *ProcessUnshieldRequestBatchObject) Reset() bool {
	t.processedUnshieldRequestBatch = NewProcessedUnshieldRequestBatch()
	return true
}

func (t ProcessUnshieldRequestBatchObject) IsDeleted() bool {
	return t.deleted
}

// value is either default or nil
func (t ProcessUnshieldRequestBatchObject) IsEmpty() bool {
	temp := NewProcessedUnshieldRequestBatch()
	return reflect.DeepEqual(temp, t.processedUnshieldRequestBatch) || t.processedUnshieldRequestBatch == nil
}
