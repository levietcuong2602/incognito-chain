package statedb

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"testing"
)

func TestNewRebuildInfo(t *testing.T) {
	h1 := common.Hash{}.NewHashFromStr2("123")
	h2 := common.Hash{}.NewHashFromStr2("456")
	info1 := NewRebuildInfo(common.STATEDB_BATCH_COMMIT_MODE, h1, h2, 1800, 900)
	b, err := json.Marshal(info1)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println("marshal ====>", info1.String(), string(b), len(b), err)

	info2 := NewEmptyRebuildInfo("")
	if err := json.Unmarshal(b, &info2); err != nil {
		t.Fatal(err)
	}

	//fmt.Println("===>", info1, info2)

	if info1.mode != info2.mode {
		t.Fatal("!= mode")
	}

	if info1.rebuildRootHash.String() != info2.rebuildRootHash.String() {
		t.Fatal("!= rebuild")
	}

	if info1.pivotRootHash.String() != info2.pivotRootHash.String() {
		t.Fatal("!= pivotroot")
	}

	if info1.rebuildFFIndex != info2.rebuildFFIndex {
		t.Fatal("!= rebuildFFIndex")
	}
	if info1.pivotFFIndex != info2.pivotFFIndex {
		t.Fatal("!= pivotffindex")
	}
	//fmt.Printf("%+v", info2)
}
