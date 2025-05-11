package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
	"github.com/incognitochain/incognito-chain/privacy/coin"
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
	// initThankTx(db)
}

func initGenesisTx(db incdb.Database) {
	var initTxs []string
	testUserkeyList := map[string]uint64{
		// Shard 0
		"112t8rnfXYskvWnHAXKs8dXLtactxRqpPTYJ6PzwkVHnF1begkenMviATTJVM6gVAgSdXsN5DEpTkLFPHtFVnS5RePi6aqTStdpb3St3uRni": uint64(1e18),
		"112t8rngZ1rZ3eWHZucwf9vrpD1DNUAmrTTARSsptNDFrEoHv3QsDY3dZe8LXy3GeKXmeso8nUPsNwUM2qmZibQVXxGzstF4v4vbfQvgk5Ci": uint64(1e18),
		"112t8rnpXg6CLjvBg2ZiyMDgpgQoZuAjYGzbm6b2eXVSHUKjZUyb2LVJmJDPw4yNaP5M14DomzC514joTH3EVknRwnnGViWuH1QucebGtVxd": uint64(1e18),
		"112t8rnqijhT2AqiS8NBVgifb86sqjfwQwf4MHLMAxK3gr1mwxMaeUWQtR1MfxHscrKQ2MsyQMvJ3LEu49LEcZzTzoJCkCiewApeZP48v3no": uint64(1e18),

		// Shard 1
		"112t8rnYRAAQ9BqLA9CF7ESWQzAAUBL1EZQwVPx4z5gPstyNpLk9abFp7iXQFu1rQ5xKukKtvorrxyetpP6Crs7Hj7GeVaVPDaHCkXeHGHSM": uint64(1e18),
		"112t8rnZkBMAJ2DYpYpnmLVJB7YkCWU7NvxxWaETLnKvdMZbhKxVU5iP97GRUBCVZbsknVsGvrdfiajD3d4Av44MXSZQd6DfGiDKdrw9SmtV": uint64(1e18),
		"112t8rnasPw9nNQqLJ4oposEYxzos63dzDUv33yJTXEaFsNfESFHenv3j32gp9DujciWXouvzPbnP3CFnpysqSPGwrYqfswb4nM1pDLofRAF": uint64(1e18),
		"112t8rncy1vEiCMxvev5EkUQyfH9HLeManjS4kbcsSiMgp4FEiddsiMunhYL2pa8wciCAWxYtt9USgCv21fe2PkSxfnRkiq4AQxTz4KgvLvB": uint64(1e18),

		// Beacon
		"112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5Jqx3A2tef8qVj1": uint64(1e18),
		"112t8rnY3WLfkE9MsKyW9s3Z5qGnPgCkeutTXJzcT5KJgAMS3vgTL9YbaJ7wyc52CzMnrj8QtwHuCpDzo47PV1qCnrui2dfJzVPU1Wn8q2Jm": uint64(1e18),
		"112t8rnX5AVkpTZtBo97KdyuDavCtufutWu8tBdDvt6D4WvULd4yyQtiVACadFdDZ28XTGgdfHkmf7wKY9iHo5gsKGwSTnsXZEHW6G7WaPss": uint64(1e18),
		"112t8rnaXH1znBqZX1Ry6xvE5hFbQUCuWJb9oiEVfCJDUWbVA9mD4NpL3dLW3TMEUtFajEsu3oKgPLMQyDPEWBuB6JfP4fXEnnAU9hdW1yCb": uint64(1e18),
		// // Accounts
		// "112t8rnXakYzsUFGtHfYWjx97vahcgKgMGYCmmvDmsXDbsv7WXVgzp7L3BfZzYFKojbYRf6yGKz6naH7BKkf1vKgNtBxS3Zcxw85hxGdRU2R": uint64(1e18),
		// "112t8rnZ7BKZEw1GwRfhUADRjwoiKT2pw6PrsD9kitvvaTW3qb3d5PEPJNpezHe1BbYGMtbfNKcMzxmeoGYRB7t9WidfJbfdZyzgsgzc5f6n": uint64(1e18),
		// "112t8rndXfrknVYTSSZ7WYWMQjtfL6BCfNw9SbtEk3poF3woD3oq9z8AZ9D37qc3BEYYMnJK8DJnF47fMGV9kBvTTESg5s1tA4e7hGiCpKSZ": uint64(1e18),
		// "112t8rnfBGykpGzRa3eBv3Hvhg8WvoHeviRVRdzqoiAGMyGJqDDUAPzj7AWd4qsxSy1aThPC2JNA2gQ5ma1RqM5HXusxEiJ7UT29vT5r7ce5": uint64(1e18),
	}
	for privateKey, initAmount := range testUserkeyList {

		testUserKey, _ := wallet.Base58CheckDeserialize(privateKey)
		testUserKey.KeySet.InitFromPrivateKey(&testUserKey.KeySet.PrivateKey)

		testSalaryTX := transaction.TxVersion2{}
		otaCoin, err := coin.NewCoinFromAmountAndReceiver(uint64(initAmount), testUserKey.KeySet.PaymentAddress)
		if err != nil {
			return
		}
		// TODO Privacy
		stateDB, _ := statedb.NewWithPrefixTrie(common.EmptyRoot, statedb.NewDatabaseAccessWarper(db))
		testSalaryTX.InitTxSalary(otaCoin, &testUserKey.KeySet.PrivateKey, stateDB, nil)
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
