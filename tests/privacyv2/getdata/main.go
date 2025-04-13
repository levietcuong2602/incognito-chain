// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"strings"
// 	"time"
// )

// func main() {

// 	ticker := time.NewTicker(10 * time.Second)
// 	for _ = range ticker.C {
// 		url := "http://localhost:9355"
// 		method := "POST"

// 		payload := strings.NewReader(`{
// 	"jsonrpc": "1.0",
//     "method": "getconsensusdata",
//     "params": [0],
//     "id": 1
// }`)

// 		client := &http.Client{}
// 		req, err := http.NewRequest(method, url, payload)

// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		res, err := client.Do(req)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		defer res.Body.Close()

// 		body, err := ioutil.ReadAll(res.Body)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		tempRes := make(map[string]interface{})
// 		json.Unmarshal(body, &tempRes)
// 		result := tempRes["Result"].(map[string]interface{})
// 		jsonResult, _ := json.Marshal(result["voteHistory"])
// 		fmt.Println(string(jsonResult))
// 		fmt.Println("---------------------------------")
// 	}
// }
