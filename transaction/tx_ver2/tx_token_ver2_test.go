package tx_ver2

import (
	"bytes"
	// "encoding/json"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPrivacyV2TxToken(t *testing.T) {
	var err error
	var numOfPrivateKeys int
	var numOfInputs int
	var dummyPrivateKeys []*privacy.PrivateKey
	var keySets []*incognitokey.KeySet
	var paymentInfo []*privacy.PaymentInfo
	var pastCoins, pastTokenCoins []privacy.Coin
	var txParams *tx_generic.TxTokenParams
	var msgCipherText []byte
	var boolParams map[string]bool
	tokenID := &common.Hash{56}
	tx2 := &TxToken{}

	Convey("Tx Token Main Test", t, func() {
		numOfPrivateKeys = RandInt()%(maxPrivateKeys-minPrivateKeys+1) + minPrivateKeys
		numOfInputs = 2
		Convey("prepare keys", func() {
			dummyPrivateKeys, keySets, paymentInfo = preparePaymentKeys(numOfPrivateKeys)
			boolParams = make(map[string]bool)
		})

		Convey("create & store PRV UTXOs", func() {
			pastCoins = make([]privacy.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
			for i := range pastCoins {
				tempCoin, err := privacy.NewCoinFromPaymentInfo(privacy.NewCoinParams().FromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)]))
				So(err, ShouldBeNil)
				So(tempCoin.IsEncrypted(), ShouldBeFalse)
				tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
				So(tempCoin.IsEncrypted(), ShouldBeTrue)
				So(tempCoin.GetSharedRandom() == nil, ShouldBeTrue)
				pastCoins[i] = tempCoin
			}
			// store a bunch of sample OTA coins in PRV
			So(storeCoins(dummyDB, pastCoins, 0, common.PRVCoinID), ShouldBeNil)
		})

		Convey("create & store Token UTXOs", func() {
			// now store the token
			err := statedb.StorePrivacyToken(dummyDB, *tokenID, "NameName", "SYM", statedb.InitToken, false, uint64(100000), []byte{}, common.Hash{66})
			So(err, ShouldBeNil)

			pastTokenCoins = make([]privacy.Coin, (10+numOfInputs)*len(dummyPrivateKeys))
			for i := range pastTokenCoins {
				tempCoin, _, err := privacy.NewCoinCA(privacy.NewCoinParams().FromPaymentInfo(paymentInfo[i%len(dummyPrivateKeys)]), tokenID)
				So(err, ShouldBeNil)
				So(tempCoin.IsEncrypted(), ShouldBeFalse)
				tempCoin.ConcealOutputCoin(keySets[i%len(dummyPrivateKeys)].PaymentAddress.GetPublicView())
				So(tempCoin.IsEncrypted(), ShouldBeTrue)
				So(tempCoin.GetSharedRandom() == nil, ShouldBeTrue)
				pastTokenCoins[i] = tempCoin
			}
			// store a bunch of sample OTA coins in PRV
			So(storeCoins(dummyDB, pastTokenCoins, 0, common.ConfidentialAssetID), ShouldBeNil)
		})

		Convey("create salary transaction", func() {
			testTxTokenV2Salary(tokenID, dummyPrivateKeys, keySets, paymentInfo, dummyDB)
		})

		Convey("transfer token", func() {
			Convey("create TX with params", func() {
				txParams, _ = getParamForTxTokenTransfer(pastCoins, pastTokenCoins, keySets, dummyDB, tokenID)
				exists := statedb.PrivacyTokenIDExisted(dummyDB, *tokenID)
				So(exists, ShouldBeTrue)
				err = tx2.Init(txParams)
				So(err, ShouldBeNil)
			})

			Convey("should verify & accept transaction", func() {
				msgCipherText = []byte("doing a transfer")
				So(bytes.Equal(msgCipherText, tx2.GetTxNormal().GetProof().GetOutputCoins()[0].GetInfo()), ShouldBeTrue)
				var err error
				tx2, err = tx2.startVerifyTx(dummyDB)
				So(err, ShouldBeNil)

				isValidSanity, err := tx2.ValidateSanityData(nil, nil, nil, 0)
				So(isValidSanity, ShouldBeTrue)
				So(err, ShouldBeNil)

				boolParams["hasPrivacy"] = hasPrivacyForToken
				// before the token init tx is written into db, this should not pass
				isValidTxItself, err := tx2.ValidateTxByItself(boolParams, dummyDB, nil, nil, shardID, nil, nil)
				So(isValidTxItself, ShouldBeTrue)
				So(err, ShouldBeNil)
				err = tx2.ValidateTxWithBlockChain(nil, nil, nil, shardID, dummyDB)
				So(err, ShouldBeNil)
			})

			Convey("should reject tampered TXs", func() {
				testTxTokenV2JsonMarshaler(tx2, 10, dummyDB)
				testTxTokenV2DeletedProof(tx2, dummyDB)
				testTxTokenV2InvalidFee(tx2, dummyDB)
				myParams, _ := getParamForTxTokenTransfer(pastCoins, pastTokenCoins, keySets, dummyDB, tokenID)
				testTxTokenV2OneFakeOutput(tx2, keySets, dummyDB, myParams, *tokenID)
				myParams, _ = getParamForTxTokenTransfer(pastCoins, pastTokenCoins, keySets, dummyDB, tokenID)
				indexForAnotherCoinOfMine := len(dummyPrivateKeys)
				testTxTokenV2OneDoubleSpentInput(myParams, pastCoins[indexForAnotherCoinOfMine], pastTokenCoins[indexForAnotherCoinOfMine], keySets, dummyDB)
			})
		})
	})
}

func testTxTokenV2DeletedProof(txv2 *TxToken, db *statedb.StateDB) {
	// try setting the proof to nil, then verify
	// it should not go through
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isBatch"] = false
	txn, ok := txv2.GetTxNormal().(*Tx)
	So(ok, ShouldBeTrue)
	savedProof := txn.GetProof()
	txn.SetProof(nil)
	txv2.SetTxNormal(txn)
	isValid, _ := txv2.ValidateSanityData(nil, nil, nil, 0)
	So(isValid, ShouldBeTrue)
	isValidTxItself, err := txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeFalse)
	logger.Infof("TEST RESULT : Missing token proof -> %v", err)
	txn.SetProof(savedProof)
	txv2.SetTxNormal(txn)
	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)

	savedProof = txv2.GetTxBase().GetProof()
	txv2.GetTxBase().SetProof(nil)
	isValid, _ = txv2.ValidateSanityData(nil, nil, nil, 0)
	So(isValid, ShouldBeTrue)

	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeFalse)
	logger.Infof("TEST RESULT : Missing PRV proof -> %v", err)
	// undo the tampering
	txv2.GetTxBase().SetProof(savedProof)
	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)

}

func testTxTokenV2InvalidFee(txv2 *TxToken, db *statedb.StateDB) {
	// a set of init params where fee is changed so mlsag should verify to false
	// let's say someone tried to use this invalid fee for tx
	// we should encounter an error here

	// set fee to increase by 1000PRV
	savedFee := txv2.GetTxBase().GetTxFee()
	txv2.GetTxBase().SetTxFee(savedFee + 1000)

	// sanity should pass
	isValidSanity, err := txv2.ValidateSanityData(nil, nil, nil, 0)
	So(isValidSanity, ShouldBeTrue)
	So(err, ShouldBeNil)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	boolParams["isBatch"] = false

	// should reject at signature since fee & output doesn't sum to input
	isValidTxItself, err := txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeFalse)
	logger.Infof("TEST RESULT : Invalid fee -> %v", err)

	// undo the tampering
	txv2.GetTxBase().SetTxFee(savedFee)
	isValidTxItself, _ = txv2.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)
}

func testTxTokenV2OneFakeOutput(txv2 *TxToken, keySets []*incognitokey.KeySet, db *statedb.StateDB, params *tx_generic.TxTokenParams, fakingTokenID common.Hash) {
	// similar to the above. All these verifications should fail
	var err error
	var isValid bool
	txn, ok := txv2.GetTxNormal().(*Tx)
	So(ok, ShouldBeTrue)
	outs := txn.Proof.GetOutputCoins()
	tokenOutput, ok := outs[0].(*coin.CoinV2)
	savedCoinBytes := tokenOutput.Bytes()
	So(ok, ShouldBeTrue)
	tokenOutput.Decrypt(keySets[0])
	// set amount from 69 to 690
	tokenOutput.SetValue(690)
	tokenOutput.SetSharedRandom(operation.RandomScalar())
	tokenOutput.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	txv2.SetTxNormal(txn)
	// here ring is broken so signing will err
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldNotBeNil)
	// isValid, err = txv2.ValidateTxByItself(hasPrivacyForPRV, db, nil, nil, 0, false, nil, nil)
	// verify must fail
	// So(isValid, ShouldBeFalse)
	logger.Infof("TEST RESULT : Fake output (wrong amount) -> %v", err)
	// undo the tampering
	tokenOutput.SetBytes(savedCoinBytes)
	outs[0] = tokenOutput
	txn.Proof.SetOutputCoins(outs)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldBeNil)

	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = true
	boolParams["isBatch"] = false

	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, 0, nil, nil)
	So(isValid, ShouldBeTrue)

	// now instead of changing amount, we change the OTA public key
	outs = txn.GetProof().GetOutputCoins()
	tokenOutput, ok = outs[0].(*coin.CoinV2)
	savedCoinBytes = tokenOutput.Bytes()
	So(ok, ShouldBeTrue)
	payInf := &privacy.PaymentInfo{PaymentAddress: keySets[0].PaymentAddress, Amount: uint64(69), Message: []byte("doing a transfer")}
	// totally fresh OTA of the same amount, meant for the same PaymentAddress
	newCoin, _, err := privacy.NewCoinCA(privacy.NewCoinParams().FromPaymentInfo(payInf), &fakingTokenID)
	So(err, ShouldBeNil)
	newCoin.ConcealOutputCoin(keySets[0].PaymentAddress.GetPublicView())
	txn.GetProof().(*privacy.ProofV2).GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2).GetCommitments()[0] = newCoin.GetCommitment()
	outs[0] = newCoin
	txn.GetProof().SetOutputCoins(outs)
	txv2.SetTxNormal(txn)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldBeNil)
	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, 0, nil, nil)
	// verify must fail
	So(isValid, ShouldBeFalse)
	logger.Infof("Fake output (wrong receiving OTA) -> %v", err)
	// undo the tampering
	tokenOutput.SetBytes(savedCoinBytes)
	outs[0] = tokenOutput
	txn.GetProof().(*privacy.ProofV2).GetAggregatedRangeProof().(*privacy.AggregatedRangeProofV2).GetCommitments()[0] = tokenOutput.GetCommitment()
	txn.GetProof().SetOutputCoins(outs)
	txv2.SetTxNormal(txn)
	err = resignUnprovenTxToken([]*incognitokey.KeySet{keySets[0]}, txv2, params, nil)
	So(err, ShouldBeNil)
	isValid, err = txv2.ValidateTxByItself(boolParams, db, nil, nil, 0, nil, nil)
	So(isValid, ShouldBeTrue)
}

// happens after txTransfer in test
// we create a second transfer, then try to reuse fee input / token input
func testTxTokenV2OneDoubleSpentInput(pr *tx_generic.TxTokenParams, dbCoin privacy.Coin, dbTokenCoin privacy.Coin, keySets []*incognitokey.KeySet, db *statedb.StateDB) {
	feeOutputSerialized := dbCoin.Bytes()
	tokenOutputSerialized := dbTokenCoin.Bytes()

	// now we try to use them as input
	doubleSpendingFeeInput := &coin.CoinV2{}
	doubleSpendingFeeInput.SetBytes(feeOutputSerialized)
	_, err := doubleSpendingFeeInput.Decrypt(keySets[0])
	So(err, ShouldBeNil)
	doubleSpendingTokenInput := &coin.CoinV2{}
	doubleSpendingTokenInput.SetBytes(tokenOutputSerialized)
	_, err = doubleSpendingTokenInput.Decrypt(keySets[0])
	So(err, ShouldBeNil)
	// save both fee&token outputs from previous tx
	otaBytes := [][]byte{doubleSpendingFeeInput.GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, common.PRVCoinID, otaBytes, 0)
	otaBytes = [][]byte{doubleSpendingTokenInput.GetKeyImage().ToBytesS()}
	statedb.StoreSerialNumbers(db, common.ConfidentialAssetID, otaBytes, 0)

	pc := doubleSpendingFeeInput
	pr.InputCoin = []coin.PlainCoin{pc}
	tx := &TxToken{}
	err = tx.Init(pr)
	So(err, ShouldBeNil)
	tx, err = tx.startVerifyTx(db)
	So(err, ShouldBeNil)
	isValidSanity, err := tx.ValidateSanityData(nil, nil, nil, 0)
	So(isValidSanity, ShouldBeTrue)
	So(err, ShouldBeNil)
	boolParams := make(map[string]bool)
	boolParams["hasPrivacy"] = hasPrivacyForPRV
	isValidTxItself, err := tx.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)
	So(err, ShouldBeNil)
	err = tx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	logger.Infof("Swap with spent Fee Input -> %v", err)
	So(err, ShouldNotBeNil)

	// now we try to swap in a used token input
	pc = doubleSpendingTokenInput
	pr.TokenParams.TokenInput = []coin.PlainCoin{pc}
	tx = &TxToken{}
	err = tx.Init(pr)
	So(err, ShouldBeNil)
	tx, err = tx.startVerifyTx(db)
	So(err, ShouldBeNil)
	isValidSanity, err = tx.ValidateSanityData(nil, nil, nil, 0)
	So(isValidSanity, ShouldBeTrue)
	So(err, ShouldBeNil)
	isValidTxItself, err = tx.ValidateTxByItself(boolParams, db, nil, nil, shardID, nil, nil)
	So(isValidTxItself, ShouldBeTrue)
	So(err, ShouldBeNil)
	err = tx.ValidateTxWithBlockChain(nil, nil, nil, 0, db)
	logger.Infof("Swap with spent Token Input of same TokenID underneath -> %v", err)
	So(err, ShouldNotBeNil)
}

func getParamForTxTokenTransfer(dbCoins []privacy.Coin, dbTokenCoins []privacy.Coin, keySets []*incognitokey.KeySet, db *statedb.StateDB, specifiedTokenID *common.Hash) (*tx_generic.TxTokenParams, *tx_generic.TokenParam) {
	transferAmount := uint64(69)
	msgCipherText := []byte("doing a transfer")
	paymentInfo2 := []*privacy.PaymentInfo{{PaymentAddress: keySets[0].PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

	feeOutputs := dbCoins[:1]
	tokenOutputs := dbTokenCoins[:1]
	prvCoinsToPayTransfer := make([]coin.PlainCoin, 0)
	tokenCoinsToTransfer := make([]coin.PlainCoin, 0)
	for _, c := range feeOutputs {
		pc, err := c.Decrypt(keySets[0])
		So(err, ShouldBeNil)
		prvCoinsToPayTransfer = append(prvCoinsToPayTransfer, pc)
	}
	for _, c := range tokenOutputs {
		pc, err := c.Decrypt(keySets[0])
		So(err, ShouldBeNil)
		tokenCoinsToTransfer = append(tokenCoinsToTransfer, pc)
	}

	tokenParam2 := &tx_generic.TokenParam{
		PropertyID:  specifiedTokenID.String(),
		Amount:      transferAmount,
		TokenTxType: utils.CustomTokenTransfer,
		Receiver:    paymentInfo2,
		TokenInput:  tokenCoinsToTransfer,
		Mintable:    false,
		Fee:         0,
	}

	txParams := tx_generic.NewTxTokenParams(&keySets[0].PrivateKey,
		[]*key.PaymentInfo{}, prvCoinsToPayTransfer, 15, tokenParam2, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return txParams, tokenParam2
}

func testTxTokenV2Salary(tokenID *common.Hash, privateKeys []*privacy.PrivateKey, keySets []*incognitokey.KeySet, paymentInfo []*privacy.PaymentInfo, db *statedb.StateDB) {
	Convey("Tx Salary Test", func() {
		Convey("create salary coins", func() {
			var err error
			var salaryCoin *privacy.CoinV2
			for {
				salaryCoin, _, err = privacy.NewCoinCA(privacy.NewCoinParams().FromPaymentInfo(paymentInfo[0]), tokenID)
				So(err, ShouldBeNil)
				otaPublicKeyBytes := salaryCoin.GetPublicKey().ToBytesS()
				// want an OTA in shard 0
				if otaPublicKeyBytes[31] == 0 {
					break
				}
			}
			var c privacy.Coin = salaryCoin
			So(salaryCoin.IsEncrypted(), ShouldBeFalse)
			So(storeCoins(db, []privacy.Coin{c}, 0, common.ConfidentialAssetID), ShouldBeNil)
			Convey("create salary TX", func() {
				txsal := &TxToken{}
				// actually making the salary TX
				err := txsal.InitTxTokenSalary(salaryCoin, privateKeys[0], db, nil, tokenID, "Token 1")
				So(err, ShouldBeNil)
				testTxTokenV2JsonMarshaler(txsal, 10, db)
				// ptoken minting requires valid signed metadata, so we skip validation here
				SkipConvey("verify salary TX", func() {
					isValid, err := txsal.ValidateTxSalary(db)
					So(err, ShouldBeNil)
					So(isValid, ShouldBeTrue)
					// malTx := &TxToken{}
					// this other coin is already in db so it must be rejected
					// err = malTx.InitTxTokenSalary(salaryCoin, privateKeys[0], db, nil, tokenID, "Token 1")
					// So(err, ShouldNotBeNil)
				})
			})
		})

	})
}

func resignUnprovenTxToken(decryptingKeys []*incognitokey.KeySet, txToken *TxToken, params *tx_generic.TxTokenParams, nonPrivacyParams *tx_generic.TxPrivacyInitParams) error {
	var err error
	txOuter := &txToken.Tx
	txOuter.SetCachedHash(nil)

	txn, ok := txToken.GetTxNormal().(*Tx)
	if !ok {
		logger.Errorf("Test Error : cast")
		return utils.NewTransactionErr(-1000, nil, "Cast failed")
	}
	txn.SetCachedHash(nil)

	// NOTE : hasPrivacy has been deprecated in the real flow.
	if nonPrivacyParams == nil {
		propertyID, _ := common.TokenStringToHash(params.TokenParams.PropertyID)
		paramsInner := tx_generic.NewTxPrivacyInitParams(
			params.SenderKey,
			params.TokenParams.Receiver,
			params.TokenParams.TokenInput,
			params.TokenParams.Fee,
			true,
			params.TransactionStateDB,
			propertyID,
			nil,
			nil,
		)
		_ = paramsInner
		paramsOuter := tx_generic.NewTxPrivacyInitParams(
			params.SenderKey,
			params.PaymentInfo,
			params.InputCoin,
			params.FeeNativeCoin,
			false,
			params.TransactionStateDB,
			&common.PRVCoinID,
			params.MetaData,
			params.Info,
		)
		err = resignUnprovenTx(decryptingKeys, txOuter, paramsOuter, &txToken.TokenData, false)
		err = resignUnprovenTx(decryptingKeys, txn, paramsInner, nil, true)
		txToken.SetTxNormal(txn)
		txToken.Tx = *txOuter
		if err != nil {
			return err
		}
	} else {
		paramsOuter := nonPrivacyParams
		err := resignUnprovenTx(decryptingKeys, txOuter, paramsOuter, &txToken.TokenData, false)
		txToken.Tx = *txOuter
		if err != nil {
			return err
		}
	}

	temp, err := txToken.startVerifyTx(params.TransactionStateDB)
	if err != nil {
		return err
	}
	*txToken = *temp
	return nil
}

func createTokenTransferParams(inputCoins []privacy.Coin, db *statedb.StateDB, tokenID, tokenName, symbol string, keySet *incognitokey.KeySet) (*tx_generic.TxTokenParams, *tx_generic.TokenParam, error) {
	var err error

	msgCipherText := []byte("Testing Transfer Token")
	transferAmount := uint64(0)
	plainInputCoins := make([]coin.PlainCoin, len(inputCoins))
	for i, inputCoin := range inputCoins {
		plainInputCoins[i], err = inputCoin.Decrypt(keySet)
		if err != nil {
			return nil, nil, err
		}
		if i != 0 {
			transferAmount += plainInputCoins[i].GetValue()
		}
	}

	tokenPayments := []*privacy.PaymentInfo{{PaymentAddress: keySet.PaymentAddress, Amount: transferAmount, Message: msgCipherText}}

	inputCoinsPRV := []coin.PlainCoin{plainInputCoins[0]}
	paymentInfoPRV := []*privacy.PaymentInfo{key.InitPaymentInfo(keySet.PaymentAddress, uint64(10), []byte("test out"))}

	// token param for init new token
	tokenParam := &tx_generic.TokenParam{
		PropertyID:     tokenID,
		PropertyName:   tokenName,
		PropertySymbol: symbol,
		Amount:         transferAmount,
		TokenTxType:    utils.CustomTokenTransfer,
		Receiver:       tokenPayments,
		TokenInput:     plainInputCoins[1:len(inputCoins)],
		Mintable:       false,
		Fee:            0,
	}

	paramToCreateTx := tx_generic.NewTxTokenParams(&keySet.PrivateKey,
		paymentInfoPRV, inputCoinsPRV, 10, tokenParam, db, nil,
		hasPrivacyForPRV, hasPrivacyForToken, shardID, []byte{}, db)
	return paramToCreateTx, tokenParam, nil
}
