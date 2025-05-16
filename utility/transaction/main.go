package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	db, err := incdb.Open("leveldb", filepath.Join("./", "./"))
	if err != nil {
		fmt.Print("could not open connection to leveldb")
		fmt.Print(err)
		panic(err)
	}

	// init an genesis tx
	initGenesisTx(db)

	// init thank tx
	initThankTx(db)
}

func initGenesisTx(db incdb.Database) {
	var initTxs []string
	testUserkeyList := map[string]uint64{
		"112t8rnXUsZd9VE6GakgR3prjHDW71vG2gKPVt7hcpXXPtFUh4Nd8Fav8w8X7KbAX4MYEnnUV1knYCxYtYnxZa6eJkDkc8HxJcnXSzytTEec": uint64(1e18),
	}
	for privateKey, initAmount := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(privateKey)
		testUserKey.KeySet.InitFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := transaction.TxVersion1{}

		// TODO Privacy
		stateDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
		testSalaryTX.InitTxSalary(initAmount, &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey, stateDB, nil)
		initTx, _ := json.MarshalIndent(testSalaryTX, "", "  ")
		initTxs = append(initTxs, string(initTx))
	}
	fmt.Println(initTxs)
}

func initThankTx(db incdb.Database) {
	var initTxs []string
	testUserkeyList := map[string]string{
		"112t8rnXBS7jJ4iqFon5rM66ex1Fc7sstNrJA9iMKgNURMUf3rywYfJ4c5Kpxw1BgL1frj9Nu5uL5vpemn9mLUW25CD1w7khX88WdauTVyKa": "@abc",
	}
	for privateKey, info := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(privateKey)
		testUserKey.KeySet.InitFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := transaction.TxVersion1{}

		// TODO Privacy
		stateDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
		testSalaryTX.InitTxSalary(0, &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey, stateDB, nil)
		testSalaryTX.Info = []byte(info)
		initTx, _ := json.MarshalIndent(testSalaryTX, "", "  ")
		initTxs = append(initTxs, string(initTx))
	}
	fmt.Println(initTxs)
}
