package rpcserver

import (
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock cho TxService
type MockTxService struct {
	mock.Mock
}

// Mock cho BlockService
type MockBlockService struct {
	mock.Mock
}

func (m *MockBlockService) GetBlockHashByHeight(chainID int, height uint64) ([]common.Hash, error) {
	return nil, nil
}

// Mock cho hash
type MockHash struct {
	mock.Mock
}

func (m *MockHash) String() string {
	args := m.Called()
	return args.String(0)
}

func TestHttpServer_handleCreateAndSendTxWithIssuingETHReq(t *testing.T) {
	httpServer := &HttpServer{}

	closeChan := make(chan struct{})

	testCases := []struct {
		name           string
		params         interface{}
		mockSetup      func()
		expectedResult interface{}
		expectedError  *rpcservice.RPCError
	}{
		{
			name: "Valid issuing ETH request",
			params: []interface{}{
				"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjuVrf9J9GSHFN9GdKxeFTnRrGbzciY7vrVdrCHQZGQSDrNTVDYqXiQPzsWZKtThrE15wKQ1Qrh",                                     // Private key
				map[string]interface{}{"SenderAddress": "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"}, // Receiver address
				1.0,  // Fee
				true, // Privacy
				map[string]interface{}{ // Metadata
					"BlockHash":         "0x88c64caa73823c85a4b783be4b25016f57c1bbdbef001fa4059e26ab7fc86229",
					"BlockHeight":       9462294,
					"TxIndex":           0,
					"ContractAddress":   "0x3Ea1c52d4A3D17F985327C1456Fcef58CfB10a48",
					"TokenID":           "f4f36b58e38461db5836c06315a8cd785e3424b52d8e841af8412a0e8f5f4b8e",
					"IncognitoAddress":  "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA",
					"Type":              1,
					"ExternalTokenID":   "0x0000000000000000000000000000000000000000",
					"ExternalTokenName": "ETH",
					"Amount":            "1000000000000000000", // 1 ETH
				},
			},
			mockSetup: func() {
				// Setup mock for BuildRawTransaction
				//mockHash := new(MockHash)
				//mockHash.On("String").Return("tx123456789abcdef")
				//
				//mockTx := &transaction.Tx{
				//	// Populate fields as needed
				//}
				//
				//// Hack: Assign the hash function to the mockTx
				//transaction.Hash = func(tx *transaction.Tx) *common.Hash {
				//	h := common.Hash{}
				//	h.SetString("tx123456789abcdef")
				//	return &h
				//}
				//
				//// Mock BuildRawTransaction
				//mockTxService.On("BuildRawTransaction", mock.Anything, mock.MatchedBy(func(meta metadata.Metadata) bool {
				//	// Verify that metadata is correct type
				//	_, ok := meta.(*metadataBridge.IssuingETHRequest)
				//	return ok
				//})).Return(mockTx, nil)
				//
				//// Mock handleSendRawTransaction
				//mockSendResult := jsonresult.CreateTransactionResult{
				//	TxID:    "tx123456789abcdef",
				//	ShardID: 0,
				//}
				//
				//// Mock handleSendRawTransaction by overriding it temporarily
				//oldHandleSendRawTransaction := handleSendRawTransaction
				//handleSendRawTransaction = func(httpServer *HttpServer, params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
				//	return mockSendResult, nil
				//}
				//defer func() { handleSendRawTransaction = oldHandleSendRawTransaction }()
			},
			expectedResult: jsonresult.CreateTransactionResult{
				TxID:    "tx123456789abcdef",
				ShardID: 0,
			},
			expectedError: nil,
		},
		//{
		//	name: "Invalid parameters",
		//	params: []interface{}{
		//		"invalid_data", // Not enough parameters
		//	},
		//	mockSetup: func() {
		//		// No mocks needed for this test case
		//	},
		//	expectedResult: nil,
		//	expectedError: &rpcservice.RPCError{
		//		Code:    rpcservice.RPCInvalidParamsError,
		//		Message: "Invalid parameters",
		//	},
		//},
		//{
		//	name: "Error in creating raw transaction",
		//	params: []interface{}{
		//		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjuVrf9J9GSHFN9GdKxeFTnRrGbzciY7vrVdrCHQZGQSDrNTVDYqXiQPzsWZKtThrE15wKQ1Qrh",                                     // Private key
		//		map[string]interface{}{"SenderAddress": "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"}, // Receiver address
		//		1.0,  // Fee
		//		true, // Privacy
		//		map[string]interface{}{ // Metadata
		//			"BlockHash":         "0x88c64caa73823c85a4b783be4b25016f57c1bbdbef001fa4059e26ab7fc86229",
		//			"BlockHeight":       9462294,
		//			"TxIndex":           0,
		//			"ContractAddress":   "0x3Ea1c52d4A3D17F985327C1456Fcef58CfB10a48",
		//			"TokenID":           "f4f36b58e38461db5836c06315a8cd785e3424b52d8e841af8412a0e8f5f4b8e",
		//			"IncognitoAddress":  "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA",
		//			"Type":              1,
		//			"ExternalTokenID":   "0x0000000000000000000000000000000000000000",
		//			"ExternalTokenName": "ETH",
		//			"Amount":            "1000000000000000000", // 1 ETH
		//		},
		//	},
		//	mockSetup: func() {
		//		// Mock BuildRawTransaction to return error
		//		mockTxService.On("BuildRawTransaction", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error creating transaction"))
		//	},
		//	expectedResult: nil,
		//	expectedError: &rpcservice.RPCError{
		//		Code:    rpcservice.UnexpectedError,
		//		Message: "Unexpected error",
		//	},
		//},
		//{
		//	name: "Error in sending raw transaction",
		//	params: []interface{}{
		//		"112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjuVrf9J9GSHFN9GdKxeFTnRrGbzciY7vrVdrCHQZGQSDrNTVDYqXiQPzsWZKtThrE15wKQ1Qrh",                                     // Private key
		//		map[string]interface{}{"SenderAddress": "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA"}, // Receiver address
		//		1.0,  // Fee
		//		true, // Privacy
		//		map[string]interface{}{ // Metadata
		//			"BlockHash":         "0x88c64caa73823c85a4b783be4b25016f57c1bbdbef001fa4059e26ab7fc86229",
		//			"BlockHeight":       9462294,
		//			"TxIndex":           0,
		//			"ContractAddress":   "0x3Ea1c52d4A3D17F985327C1456Fcef58CfB10a48",
		//			"TokenID":           "f4f36b58e38461db5836c06315a8cd785e3424b52d8e841af8412a0e8f5f4b8e",
		//			"IncognitoAddress":  "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA",
		//			"Type":              1,
		//			"ExternalTokenID":   "0x0000000000000000000000000000000000000000",
		//			"ExternalTokenName": "ETH",
		//			"Amount":            "1000000000000000000", // 1 ETH
		//		},
		//	},
		//	mockSetup: func() {
		//		// Setup mock for BuildRawTransaction
		//		mockHash := new(MockHash)
		//		mockHash.On("String").Return("tx123456789abcdef")
		//
		//		mockTx := &transaction.Tx{
		//			// Populate fields as needed
		//		}
		//
		//		// Hack: Assign the hash function to the mockTx
		//		transaction.Hash = func(tx *transaction.Tx) *common.Hash {
		//			h := common.Hash{}
		//			h.SetString("tx123456789abcdef")
		//			return &h
		//		}
		//
		//		// Mock BuildRawTransaction
		//		mockTxService.On("BuildRawTransaction", mock.Anything, mock.Anything).Return(mockTx, nil)
		//
		//		// Mock handleSendRawTransaction with error
		//		oldHandleSendRawTransaction := handleSendRawTransaction
		//		handleSendRawTransaction = func(httpServer *HttpServer, params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
		//			return nil, &rpcservice.RPCError{
		//				Code:    rpcservice.RPCInvalidParamsError,
		//				Message: "Failed to send transaction",
		//			}
		//		}
		//		defer func() { handleSendRawTransaction = oldHandleSendRawTransaction }()
		//	},
		//	expectedResult: nil,
		//	expectedError: &rpcservice.RPCError{
		//		Code:    rpcservice.UnexpectedError,
		//		Message: "Unexpected error",
		//	},
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks for this test case
			tc.mockSetup()

			// Call the function being tested
			result, err := httpServer.handleCreateAndSendTxWithIssuingETHReq(tc.params, closeChan)

			// Check results
			if tc.expectedError != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError.Code, err.Code)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

// handleSendRawTransaction is a function variable that we'll override in tests
var handleSendRawTransaction = func(httpServer *HttpServer, params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// This will be overridden in the test
	return nil, nil
}
