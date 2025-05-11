package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/incognitochain/incognito-chain/wallet"
)

type user struct {
	privatekey, paymentaddress string
}

func main() {
	/* TEST SETUP */
	user1 := user{
		privatekey:     "112t8rnXakYzsUFGtHfYWjx97vahcgKgMGYCmmvDmsXDbsv7WXVgzp7L3BfZzYFKojbYRf6yGKz6naH7BKkf1vKgNtBxS3Zcxw85hxGdRU2R",
		paymentaddress: "12sendKRF5JXTSzFoHNS4JdtWYLcxVshaSwiRTeLRBKv4HDRp3EwrHY5ZWPxcdb9ve5j1pP167M8gHbMBi2ALVzeLynBe5srcUfC48Gsj7osKvLcB2tTNedJqc6vopMkn5ko4yB6juxd97pF7CLw",
	}
	user2 := user{
		privatekey:     "112t8rnZ7BKZEw1GwRfhUADRjwoiKT2pw6PrsD9kitvvaTW3qb3d5PEPJNpezHe1BbYGMtbfNKcMzxmeoGYRB7t9WidfJbfdZyzgsgzc5f6n",
		paymentaddress: "12sdGdquGtzmWkhsUz6MxE8sSLVZJHnAkg3MQRg4SPnr6goiHvUZvLmZPu2Q8cdQPKnWR7tbYJsXXrbBjK2PsgKMvqKQAFC1Vvx4PCWdhkLtDBjoHNaTcUUn27M98zJDa969fFQtbmoYRQ3JaMCm",
	}
	var RPCServer string = "http://localhost:9354"

	amount := uint64(50000)

	// Step 0: submitKey
	err := submitKey(RPCServer, user1.privatekey)
	if err != nil {
		log.Println(err)
	}
	err = submitKey(RPCServer, user2.privatekey)
	if err != nil {
		log.Println(err)
	}

	// Step 1: Check balance of address 1 and address 2 before sending PRV
	fmt.Println("// Step 1: Check balance of address 1 and address 2 before sending PRV")
	balance1Before, err := getBalanceByPrivateKey(RPCServer, user1.privatekey)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Get Balance Address-1:", balance1Before)
	}

	balance2Before, err := getBalanceByPrivateKey(RPCServer, user2.privatekey)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Get Balance Address-2:", balance2Before)

	}
	// Step 2: Sending PRV from address 1 to address 2
	fmt.Println("// Step 2: Sending PRV from address 1 to address 2")
	fmt.Println("-Sending ", amount)
	sendTX, err := createAndSendTransaction(RPCServer, user1.privatekey, user2.paymentaddress, amount)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("TxID: ", sendTX)
	}

	fmt.Println("-Sleep 1 minute, wait for balance update")
	time.Sleep(1 * time.Minute)

	// Step 3: Get Transaction Fee
	fmt.Println("// Step 3: Get Transaction Fee")
	txID := sendTX.(string)
	fee, err := getTransactionFeeByHash(RPCServer, txID)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Transaction Fee: ", fee)
	}

	// Step 4: Check balance of address 1 and address 2 after sent PRV
	fmt.Println("// Step 3: Check balance of address 1 and address 2 after sent PRV")

	balance1After, err := getBalanceByPrivateKey(RPCServer, user1.privatekey)
	if err != nil {
		fmt.Println("Error go here: ", err)
	} else {
		fmt.Println("Get Balance Address-1:", balance1After)
	}

	balance2After, err := getBalanceByPrivateKey(RPCServer, user2.privatekey)
	if err != nil {
		fmt.Printf("Error go here: %v", err)
	} else {
		fmt.Println("Get Balance Address-2:", balance2After)

	}

	/* TEST RESULT */
	fmt.Println("// TEST RESULT:")
	if (balance1After+amount+fee != balance1Before) || (balance2After-amount != balance2Before) {
		fmt.Println("FAILED")
	} else {
		fmt.Println("PASSED")
	}

	/* TEST CLEAN UP*/
	// fmt.Println("// TEST CLEAN UP")
	// cleanupTX, err := createAndSendTransaction(RPCServer, user2.privatekey, user1.paymentaddress, balance2After-2) // default tx fee = 2
	// if err != nil {
	// 	fmt.Print("Clean up error: ", err)
	// } else {
	// 	fmt.Println("Clean up success ", cleanupTX)
	// }
}

func privateKeyToPaymentAddress(privateKey string, keyType int) string {
	keyWallet, _ := wallet.Base58CheckDeserialize(privateKey)
	err := keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		return ""
	}
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	switch keyType {
	case 0: //Old address, old encoding
		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, false)
		return addr
	case 1:
		addr, _ := wallet.GetPaymentAddressV1(paymentAddStr, true)
		return addr
	default:
		return paymentAddStr
	}
}

func getBalanceByPrivateKey(RPC string, privatekey string) (uint64, error) {
	var params []interface{}
	params = append(params, privatekey)

	result, err := sendHttpRequest(RPC, "getbalancebyprivatekey", params, false)
	if err != nil {
		fmt.Println("Error go here: ", err)
		return 0, err
	}
	balance := result.(float64)
	return uint64(balance), err
}

func getTransactionFeeByHash(RPC string, txID string) (uint64, error) {
	var params []interface{}
	params = append(params, txID)

	result, err := sendHttpRequest(RPC, "gettransactionbyhash", params, false)
	if err != nil {
		fmt.Println("Error go here: ", err)
		return 0, err
	}
	fee := result.(map[string]interface{})["Fee"].(float64)
	return uint64(fee), err
}

func submitKey(RPC string, privatekey string) error {
	var params []interface{}
	params = append(params, privatekey)

	_, err := sendHttpRequest(RPC, "submitkey", params, true)
	return err
}

// createAndSendTransaction creates and sends a transaction
// using the provided RPC server, private key, payment address, and amount.
// It returns the transaction ID and any error encountered.
// The function sends a JSON-RPC request to the server and parses the response.
func createAndSendTransaction(RPC string, privatekey string, paymentaddress string, amount uint64) (interface{}, error) {
	// param #1: private key of sender
	// param #2: list receivers
	// param #3: estimation fee nano P per kb
	// param #4: hasPrivacyCoin flag: 1 or -1
	// default: -1 (has no privacy) (if missing this param)
	// param #5: meta data (optional) don't do anything
	// param#6: info (optional)
	var params []interface{}
	params = append(params, privatekey)
	params = append(params, map[string]uint64{paymentaddress: amount})
	params = append(params, 10)
	params = append(params, -1) // hasPrivacyCoin flag: 1 or -1

	log.Println("Params: ", params)
	result, err := sendHttpRequest(RPC, "createandsendtransaction", params, true)
	log.Println("Result: ", result)
	log.Println("ERR: ", err)
	if err != nil {
		fmt.Println("Error go here: ", err)
		return nil, err
	}
	txID := result.(map[string]interface{})["TxID"].(string)
	fmt.Println("Transaction ID: ", txID)
	return txID, nil
}

// utility functions
type JsonRpcFormat struct {
	ID      string        `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type Response struct {
	Result interface{} `json:"Result"`
	Error  interface{} `json:"Error"`
}

func sendHttpRequest(url, method string, params []interface{}, isToConsole bool) (interface{}, error) {
	payload := &JsonRpcFormat{
		ID:      "1",
		JsonRpc: "jsonrpc",
		Method:  method,
		Params:  params,
	}
	payloadData, err := json.Marshal(payload)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadData))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if isToConsole {
		log.Println(string(body))
	}
	response := Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("error from server: %v", response.Error)
	}
	return response.Result, nil
}
