package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
)

func main() {
	evmcaller.InitCacher()

	keybuf := new(bytes.Buffer)
	keybuf.Reset()
	rlp.Encode(keybuf, 0)
	blockHash := rCommon.HexToHash("0x766809cfb638b96d4f6804960c1f63c9eabdb426eb2e6d5aad0e52ce6feacec7")
	hosts := []string{"http://62.146.229.117:8545"}
	evmHeaderResult, err := evmcaller.GetEVMHeaderResult(blockHash, hosts, 15, "")
	if err != nil {
		fmt.Println("GetEVMHeaderResult")
		return
	}
	jsData, _ := json.Marshal(evmHeaderResult)
	fmt.Println("EVM Header Result:", string(jsData))
	receiptHashBytes := evmHeaderResult.Header.ReceiptHash.Bytes()
	receiptHashBase64 := base64.StdEncoding.EncodeToString(receiptHashBytes)
	fmt.Printf("ReceiptHash in base64: %s\n", receiptHashBase64)

	proofStrs := []string{

		"+QJnAYJ+VbkBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD5AV35AVqUi5F2t7ABwlT0UzDh+yS4T+S4DWvhoC1LWXk1881n+y7r8dtN68k0zuXHuqcVP5gP2+sudAhOuQEgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABvBbWdOyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJQxMnNlbmRLUkY1SlhUU3pGb0hOUzRKZHRXWUxjeFZzaGFTd2lSVGVMUkJLdjRIRFJwM0V3ckhZNVpXUHhjZGI5dmU1ajFwUDE2N004Z0hiTUJpMkFMVnplTHluQmU1c3JjVWZDNDhHc2o3b3NLdkxjQjJ0VE5lZEpxYzZ2b3BNa241a280eUI2anV4ZDk3cEY3Q0x3AAAAAAAAAAAAAAAA",
	}
	nodeList := new(light.NodeList)
	for _, proofStr := range proofStrs {
		proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
		if err != nil {
			fmt.Println("DecodeString", err)
			return
		}
		nodeList.Put([]byte{}, proofBytes)
	}

	proof := nodeList.NodeSet()

	val, _, err := trie.VerifyProof(evmHeaderResult.Header.ReceiptHash, keybuf.Bytes(), proof)
	if err != nil {
		fmt.Println("VerifyProof", err)
		return
	}

	fmt.Println(val)

}

// package main

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/light"
// 	"github.com/ethereum/go-ethereum/rlp"
// 	"github.com/ethereum/go-ethereum/trie"
// 	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
// )

// func main() {
// 	evmcaller.InitCacher()

// 	// The key should be the transaction index, not RLP-encoded 0
// 	txIndex := 0 // Your transaction index from JavaScript

// 	// Encode the transaction index as the key for the receipt trie
// 	var key []byte
// 	if txIndex == 0 {
// 		key = []byte{0x80} // RLP encoding of 0
// 	} else {
// 		keybuf := new(bytes.Buffer)
// 		rlp.Encode(keybuf, uint64(txIndex))
// 		key = keybuf.Bytes()
// 	}

// 	blockHash := common.HexToHash("0x766809cfb638b96d4f6804960c1f63c9eabdb426eb2e6d5aad0e52ce6feacec7")
// 	hosts := []string{"http://62.146.229.117:8545"}

// 	evmHeaderResult, err := evmcaller.GetEVMHeaderResult(blockHash, hosts, 15, "")
// 	if err != nil {
// 		fmt.Println("GetEVMHeaderResult error:", err)
// 		return
// 	}

// 	jsData, _ := json.Marshal(evmHeaderResult)
// 	fmt.Println("EVM Header Result:", string(jsData))

// 	receiptHashBytes := evmHeaderResult.Header.ReceiptHash.Bytes()
// 	receiptHashBase64 := base64.StdEncoding.EncodeToString(receiptHashBytes)
// 	fmt.Printf("ReceiptHash in base64: %s\n", receiptHashBase64)

// 	// You need actual trie proof nodes here, not receipt data
// 	// These should come from your JavaScript code's trie proof generation
// 	proofStrs := []string{
// 		// Replace with actual trie proof nodes from your JavaScript
// 		// The current data is receipt data, not trie proof
// 		"+QJnAYJ+VbkBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD5AV35AVqUi5F2t7ABwlT0UzDh+yS4T+S4DWvhoC1LWXk1881n+y7r8dtN68k0zuXHuqcVP5gP2+sudAhOuQEgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABvBbWdOyAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAJQxMnNlbmRLUkY1SlhUU3pGb0hOUzRKZHRXWUxjeFZzaGFTd2lSVGVMUkJLdjRIRFJwM0V3ckhZNVpXUHhjZGI5dmU1ajFwUDE2N004Z0hiTUJpMkFMVnplTHluQmU1c3JjVWZDNDhHc2o3b3NLdkxjQjJ0VE5lZEpxYzZ2b3BNa241a280eUI2anV4ZDk3cEY3Q0x3AAAAAAAAAAAAAAAA",
// 	}

// 	if len(proofStrs) == 0 {
// 		fmt.Println("No proof data provided. You need to generate actual trie proof nodes.")
// 		return
// 	}

// 	nodeList := new(light.NodeList)
// 	for i, proofStr := range proofStrs {
// 		proofBytes, err := base64.StdEncoding.DecodeString(proofStr)
// 		if err != nil {
// 			fmt.Printf("DecodeString error for proof %d: %v\n", i, err)
// 			return
// 		}

// 		// For trie proofs, you need the hash of each node as the key
// 		// This is typically provided by the proof generation method
// 		nodeHash := common.BytesToHash(proofBytes[:32]) // First 32 bytes should be the hash
// 		nodeList.Put(nodeHash[:], proofBytes[32:])      // Rest is the node data
// 	}

// 	proof := nodeList.NodeSet()

// 	fmt.Printf("Verifying with key: %x\n", key)
// 	fmt.Printf("Root hash: %x\n", evmHeaderResult.Header.ReceiptHash)

// 	val, nodeCount, err := trie.VerifyProof(evmHeaderResult.Header.ReceiptHash, key, proof)
// 	if err != nil {
// 		fmt.Printf("VerifyProof error: %v\n", err)
// 		fmt.Printf("Verified %d nodes before error\n", nodeCount)
// 		return
// 	}

// 	fmt.Printf("Proof verified successfully! Nodes verified: %d\n", nodeCount)
// 	fmt.Printf("Receipt data length: %d bytes\n", len(val))

// 	// Optional: Decode the receipt if verification succeeds
// 	if len(val) > 0 {
// 		fmt.Printf("Receipt data (base64): %s\n", base64.StdEncoding.EncodeToString(val))
// 	}
// }
