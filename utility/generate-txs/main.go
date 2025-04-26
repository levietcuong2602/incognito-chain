package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func main() {
	//==========Write
	transactions := []string{}
	db, err := incdb.Open("leveldb", filepath.Join("./", "./"))
	if err != nil {
		fmt.Print("could not open connection to leveldb")
		fmt.Print(err)
		panic(err)
	}
	stateDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
	privateKeys := readTxsFromFile("private-keys-shard-1-1.json")
	fmt.Println(len(privateKeys))
	for _, privateKey := range privateKeys {
		txs := initTx("1000", privateKey, stateDB)
		transactions = append(transactions, txs[0])
	}
	fmt.Println(len(transactions))
	file, _ := json.MarshalIndent(transactions, "", " ")
	_ = ioutil.WriteFile("shard1-1-init-txs.json", file, 0644)
}
func readTxsFromFile(filename string) []string {
	// Open our jsonFile
	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened ", filename)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result []string
	json.Unmarshal([]byte(byteValue), &result)
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	return result
}
func initTx(amount string, privateKey string, stateDB *statedb.StateDB) []string {
	var initTxs []string
	var initAmount, _ = strconv.Atoi(amount) // amount init
	testUserkeyList := []string{
		privateKey,
	}
	for _, val := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(val)
		testUserKey.KeySet.InitFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := transaction.Tx{}

		// TODO Privacy
		testSalaryTX.InitTxSalary(uint64(initAmount), coin.NewTxRandom(), &testUserKey.KeySet.PaymentAddress, &testUserKey.KeySet.PrivateKey,
			stateDB,
			nil,
		)
		initTx, _ := json.Marshal(testSalaryTX)
		initTxs = append(initTxs, string(initTx))
	}
	return initTxs
}
