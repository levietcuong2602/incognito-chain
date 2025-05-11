// package main

// import (
// 	"fmt"

// 	"github.com/incognitochain/incognito-chain/metadata/common"
// 	"github.com/incognitochain/incognito-chain/wallet"
// )

// func main() {
// 	// listPrivateKeys := []string{
// 	// 	// Shard 0
// 	// 	"112t8rnfXYskvWnHAXKs8dXLtactxRqpPTYJ6PzwkVHnF1begkenMviATTJVM6gVAgSdXsN5DEpTkLFPHtFVnS5RePi6aqTStdpb3St3uRni",
// 	// 	"112t8rngZ1rZ3eWHZucwf9vrpD1DNUAmrTTARSsptNDFrEoHv3QsDY3dZe8LXy3GeKXmeso8nUPsNwUM2qmZibQVXxGzstF4v4vbfQvgk5Ci",
// 	// 	"112t8rnpXg6CLjvBg2ZiyMDgpgQoZuAjYGzbm6b2eXVSHUKjZUyb2LVJmJDPw4yNaP5M14DomzC514joTH3EVknRwnnGViWuH1QucebGtVxd",
// 	// 	"112t8rnqijhT2AqiS8NBVgifb86sqjfwQwf4MHLMAxK3gr1mwxMaeUWQtR1MfxHscrKQ2MsyQMvJ3LEu49LEcZzTzoJCkCiewApeZP48v3no",

// 	// 	// Shard 1
// 	// 	"112t8rnYRAAQ9BqLA9CF7ESWQzAAUBL1EZQwVPx4z5gPstyNpLk9abFp7iXQFu1rQ5xKukKtvorrxyetpP6Crs7Hj7GeVaVPDaHCkXeHGHSM",
// 	// 	"112t8rnZkBMAJ2DYpYpnmLVJB7YkCWU7NvxxWaETLnKvdMZbhKxVU5iP97GRUBCVZbsknVsGvrdfiajD3d4Av44MXSZQd6DfGiDKdrw9SmtV",
// 	// 	"112t8rnasPw9nNQqLJ4oposEYxzos63dzDUv33yJTXEaFsNfESFHenv3j32gp9DujciWXouvzPbnP3CFnpysqSPGwrYqfswb4nM1pDLofRAF",
// 	// 	"112t8rncy1vEiCMxvev5EkUQyfH9HLeManjS4kbcsSiMgp4FEiddsiMunhYL2pa8wciCAWxYtt9USgCv21fe2PkSxfnRkiq4AQxTz4KgvLvB",

// 	// 	// Beacon
// 	// 	"112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5Jqx3A2tef8qVj1",
// 	// 	"112t8rnY3WLfkE9MsKyW9s3Z5qGnPgCkeutTXJzcT5KJgAMS3vgTL9YbaJ7wyc52CzMnrj8QtwHuCpDzo47PV1qCnrui2dfJzVPU1Wn8q2Jm",
// 	// 	"112t8rnX5AVkpTZtBo97KdyuDavCtufutWu8tBdDvt6D4WvULd4yyQtiVACadFdDZ28XTGgdfHkmf7wKY9iHo5gsKGwSTnsXZEHW6G7WaPss",
// 	// 	"112t8rnaXH1znBqZX1Ry6xvE5hFbQUCuWJb9oiEVfCJDUWbVA9mD4NpL3dLW3TMEUtFajEsu3oKgPLMQyDPEWBuB6JfP4fXEnnAU9hdW1yCb",
// 	// }

// 	// for _, pk := range listPrivateKeys {
// 	// 	privateSeed, err := consensus_v2.LoadUserKeyFromIncPrivateKey(pk)
// 	// 	if err != nil {
// 	// 		return
// 	// 	}

// 	// 	fmt.Printf("%v - %v \n\n", pk, privateSeed)
// 	// }

// 	paymentAddr := privateKeyToPaymentAddress("112t8rnXUiPCmJ19QxbmyVYvbTPM6UcN6aJZ95mgWmb1tVGexMn6BXNHAdtK1bMP4aTspD8EcmPPJhUyc66jLBmuG4ck8rjV1VNKvzXWk44m", -1)

// 	fmt.Println("Payment Address: ", paymentAddr)

// 	test, err := common.AssertPaymentAddressAndTxVersion(paymentAddr, 1)
// 	fmt.Println(test, err)

// }

// func privateKeyToPaymentAddress(privateKey string, keyType int) string {
// 	keyWallet, _ := wallet.Base58CheckDeserialize(privateKey)
// 	err := keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
// 	if err != nil {
// 		return ""
// 	}
// 	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
// 	switch keyType {
// 	case 0: //Old address, old encoding
// 		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, false)
// 		return addr
// 	case 1:
// 		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, true)
// 		return addr
// 	default:
// 		return paymentAddStr
// 	}
// }
