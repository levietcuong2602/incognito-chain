package tx_ver2

import (
	"fmt"
	"github.com/levietcuong2602/incognito-chain/wallet"
	"math/big"

	"github.com/levietcuong2602/incognito-chain/common"
	"github.com/levietcuong2602/incognito-chain/dataaccessobject/statedb"
	"github.com/levietcuong2602/incognito-chain/incognitokey"
	"github.com/levietcuong2602/incognito-chain/privacy"
	"github.com/levietcuong2602/incognito-chain/privacy/privacy_v2/mlsag"
	"github.com/levietcuong2602/incognito-chain/transaction/tx_generic"
	"github.com/levietcuong2602/incognito-chain/transaction/utils"
	// "github.com/levietcuong2602/incognito-chain/wallet"
)

func createPrivKeyMlsagCA(inputCoins []privacy.PlainCoin, outputCoins []*privacy.CoinV2, outputSharedSecrets []*privacy.Point, params *tx_generic.TxPrivacyInitParams, shardID byte, commitmentsToZero []*privacy.Point) ([]*privacy.Scalar, error) {
	senderSK := params.SenderSK
	// db := params.StateDB
	tokenID := params.TokenID
	if tokenID == nil {
		tokenID = &common.PRVCoinID
	}
	rehashed := privacy.HashToPoint(tokenID[:])
	sumRand := new(privacy.Scalar).FromUint64(0)

	privKeyMlsag := make([]*privacy.Scalar, len(inputCoins)+2)
	sumInputAssetTagBlinders := new(privacy.Scalar).FromUint64(0)
	numOfInputs := new(privacy.Scalar).FromUint64(uint64(len(inputCoins)))
	numOfOutputs := new(privacy.Scalar).FromUint64(uint64(len(outputCoins)))
	mySkBytes := (*senderSK)[:]
	for i := 0; i < len(inputCoins); i++ {
		var err error
		privKeyMlsag[i], err = inputCoins[i].ParsePrivateKeyOfCoin(*senderSK)
		if err != nil {
			utils.Logger.Log.Errorf("Cannot parse private key of coin %v", err)
			return nil, err
		}

		inputCoin_specific, ok := inputCoins[i].(*privacy.CoinV2)
		if !ok || inputCoin_specific.GetAssetTag() == nil {
			return nil, fmt.Errorf("cannot cast a coin as v2-CA")
		}

		isUnblinded := privacy.IsPointEqual(rehashed, inputCoin_specific.GetAssetTag())
		if isUnblinded {
			utils.Logger.Log.Infof("Signing TX : processing an unblinded input coin")
		}

		sharedSecret := new(privacy.Point).Identity()
		bl := new(privacy.Scalar).FromUint64(0)
		if !isUnblinded {
			sharedSecret, err = inputCoin_specific.RecomputeSharedSecret(mySkBytes)
			if err != nil {
				utils.Logger.Log.Errorf("Cannot recompute shared secret : %v", err)
				return nil, err
			}

			bl, err = privacy.ComputeAssetTagBlinder(sharedSecret)
			if err != nil {
				return nil, err
			}
		}

		utils.Logger.Log.Infof("CA-MLSAG : processing input asset tag %s", string(inputCoin_specific.GetAssetTag().MarshalText()))
		utils.Logger.Log.Debugf("Shared secret is %s", string(sharedSecret.MarshalText()))
		utils.Logger.Log.Debugf("Blinder is %s", string(bl.MarshalText()))
		v := inputCoin_specific.GetAmount()
		utils.Logger.Log.Debugf("Value is %d", v.ToUint64Little())
		effectiveRCom := new(privacy.Scalar).Mul(bl, v)
		effectiveRCom.Add(effectiveRCom, inputCoin_specific.GetRandomness())

		sumInputAssetTagBlinders.Add(sumInputAssetTagBlinders, bl)
		sumRand.Add(sumRand, effectiveRCom)
	}
	sumInputAssetTagBlinders.Mul(sumInputAssetTagBlinders, numOfOutputs)

	sumOutputAssetTagBlinders := new(privacy.Scalar).FromUint64(0)

	var err error
	for i, oc := range outputCoins {
		if oc.GetAssetTag() == nil {
			return nil, fmt.Errorf("cannot cast a coin as v2-CA")
		}
		// lengths between 0 and len(outputCoins) were rejected before
		bl := new(privacy.Scalar).FromUint64(0)
		isUnblinded := privacy.IsPointEqual(rehashed, oc.GetAssetTag())
		if isUnblinded {
			utils.Logger.Log.Infof("Signing TX : processing an unblinded output coin")
		} else {
			utils.Logger.Log.Debugf("Shared secret is %s", string(outputSharedSecrets[i].MarshalText()))
			bl, err = privacy.ComputeAssetTagBlinder(outputSharedSecrets[i])
			if err != nil {
				return nil, err
			}
		}
		utils.Logger.Log.Infof("CA-MLSAG : processing output asset tag %s", string(oc.GetAssetTag().MarshalText()))
		utils.Logger.Log.Debugf("Blinder is %s", string(bl.MarshalText()))

		v := oc.GetAmount()
		utils.Logger.Log.Debugf("Value is %d", v.ToUint64Little())
		effectiveRCom := new(privacy.Scalar).Mul(bl, v)
		effectiveRCom.Add(effectiveRCom, oc.GetRandomness())
		sumOutputAssetTagBlinders.Add(sumOutputAssetTagBlinders, bl)
		sumRand.Sub(sumRand, effectiveRCom)
	}
	sumOutputAssetTagBlinders.Mul(sumOutputAssetTagBlinders, numOfInputs)

	// 2 final elements in `private keys` for MLSAG
	assetSum := new(privacy.Scalar).Sub(sumInputAssetTagBlinders, sumOutputAssetTagBlinders)
	firstCommitmentToZeroRecomputed := new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenRandomnessIndex], assetSum)
	secondCommitmentToZeroRecomputed := new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenRandomnessIndex], sumRand)
	if len(commitmentsToZero) != 2 {
		utils.Logger.Log.Errorf("Received %d points to check when signing MLSAG", len(commitmentsToZero))
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("error : need exactly 2 points for MLSAG double-checking"))
	}
	match1 := privacy.IsPointEqual(firstCommitmentToZeroRecomputed, commitmentsToZero[0])
	match2 := privacy.IsPointEqual(secondCommitmentToZeroRecomputed, commitmentsToZero[1])
	if !match1 || !match2 {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("error : asset tag sum or commitment sum mismatch"))
	}

	utils.Logger.Log.Debugf("Last 2 private keys will correspond to points %s and %s", firstCommitmentToZeroRecomputed.MarshalText(), secondCommitmentToZeroRecomputed.MarshalText())

	privKeyMlsag[len(inputCoins)] = assetSum
	privKeyMlsag[len(inputCoins)+1] = sumRand
	return privKeyMlsag, nil
}

func generateMlsagRingWithIndexesCA(inputCoins []privacy.PlainCoin, outputCoins []*privacy.CoinV2, params *tx_generic.TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, []*privacy.Point, error) {

	lenOTA, err := statedb.GetOTACoinLength(params.StateDB, common.ConfidentialAssetID, shardID)
	if err != nil || lenOTA == nil {
		utils.Logger.Log.Errorf("Getting length of commitment error, either database length ota is empty or has error, error = %v", err)
		return nil, nil, nil, err
	}
	outputCoinsAsGeneric := make([]privacy.Coin, len(outputCoins))
	for i := 0; i < len(outputCoins); i++ {
		outputCoinsAsGeneric[i] = outputCoins[i]
	}
	sumOutputsWithFee := tx_generic.CalculateSumOutputsWithFee(outputCoinsAsGeneric, params.Fee)
	inCount := new(privacy.Scalar).FromUint64(uint64(len(inputCoins)))
	outCount := new(privacy.Scalar).FromUint64(uint64(len(outputCoins)))

	sumOutputAssetTags := new(privacy.Point).Identity()
	for _, oc := range outputCoins {
		if oc.GetAssetTag() == nil {
			utils.Logger.Log.Errorf("CA error: missing asset tag for signing in output coin - %v", oc.Bytes())
			err := utils.NewTransactionErr(utils.SignTxError, fmt.Errorf("cannot sign CA token : an output coin does not have asset tag"))
			return nil, nil, nil, err
		}
		sumOutputAssetTags.Add(sumOutputAssetTags, oc.GetAssetTag())
	}
	sumOutputAssetTags.ScalarMult(sumOutputAssetTags, inCount)

	indexes := make([][]*big.Int, ringSize)
	ring := make([][]*privacy.Point, ringSize)
	var lastTwoColumnsCommitmentToZero []*privacy.Point
	attempts := 0
	for i := 0; i < ringSize; i++ {
		sumInputs := new(privacy.Point).Identity()
		sumInputs.Sub(sumInputs, sumOutputsWithFee)
		sumInputAssetTags := new(privacy.Point).Identity()

		row := make([]*privacy.Point, len(inputCoins))
		rowIndexes := make([]*big.Int, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j++ {
				row[j] = inputCoins[j].GetPublicKey()
				publicKeyBytes := inputCoins[j].GetPublicKey().ToBytesS()
				if rowIndexes[j], err = statedb.GetOTACoinIndex(params.StateDB, common.ConfidentialAssetID, publicKeyBytes); err != nil {
					utils.Logger.Log.Errorf("Getting commitment index error %v ", err)
					return nil, nil, nil, err
				}
				sumInputs.Add(sumInputs, inputCoins[j].GetCommitment())
				inputCoin_specific, ok := inputCoins[j].(*privacy.CoinV2)
				if !ok {
					return nil, nil, nil, fmt.Errorf("cannot cast a coin as v2")
				}
				if inputCoin_specific.GetAssetTag() == nil {
					utils.Logger.Log.Errorf("CA error: missing asset tag for signing in input coin - %v", inputCoin_specific.Bytes())
					err := utils.NewTransactionErr(utils.SignTxError, fmt.Errorf("cannot sign CA token : an input coin does not have asset tag"))
					return nil, nil, nil, err
				}
				sumInputAssetTags.Add(sumInputAssetTags, inputCoin_specific.GetAssetTag())
			}
		} else {
			for j := 0; j < len(inputCoins); j++ {
				coinDB := new(privacy.CoinV2)
				for attempts < privacy.MaxPrivacyAttempts { // The chance of infinite loop is negligible
					rowIndexes[j], _ = common.RandBigIntMaxRange(lenOTA)
					coinBytes, err := statedb.GetOTACoinByIndex(params.StateDB, common.ConfidentialAssetID, rowIndexes[j].Uint64(), shardID)
					if err != nil {
						utils.Logger.Log.Errorf("Get coinv2 by index error %v ", err)
						return nil, nil, nil, err
					}

					if err = coinDB.SetBytes(coinBytes); err != nil {
						utils.Logger.Log.Errorf("Cannot parse coinv2 byte error %v ", err)
						return nil, nil, nil, err
					}

					if coinDB.GetAssetTag() == nil {
						utils.Logger.Log.Errorf("CA error: missing asset tag for signing in DB coin - %v", coinBytes)
						err := utils.NewTransactionErr(utils.SignTxError, fmt.Errorf("cannot sign CA token : a CA coin in DB does not have asset tag"))
						return nil, nil, nil, err
					}

					// we do not use burned coins since they will reduce the privacy level of the transaction.
					if !wallet.IsPublicKeyBurningAddress(coinDB.GetPublicKey().ToBytesS()) {
						break
					}
					attempts++
				}

				if attempts == privacy.MaxPrivacyAttempts {
					return nil, nil, nil, fmt.Errorf("cannot form decoys CA")
				}

				row[j] = coinDB.GetPublicKey()
				sumInputs.Add(sumInputs, coinDB.GetCommitment())
				sumInputAssetTags.Add(sumInputAssetTags, coinDB.GetAssetTag())
			}
		}
		sumInputAssetTags.ScalarMult(sumInputAssetTags, outCount)

		assetSum := new(privacy.Point).Sub(sumInputAssetTags, sumOutputAssetTags)
		row = append(row, assetSum)
		row = append(row, sumInputs)
		if i == pi {
			utils.Logger.Log.Debugf("Last 2 columns in ring are %s and %s", assetSum.MarshalText(), sumInputs.MarshalText())
			lastTwoColumnsCommitmentToZero = []*privacy.Point{assetSum, sumInputs}
		}

		ring[i] = row
		indexes[i] = rowIndexes
	}
	return mlsag.NewRing(ring), indexes, lastTwoColumnsCommitmentToZero, nil
}

func (tx *Tx) proveCA(params *tx_generic.TxPrivacyInitParams) (bool, error) {
	var err error
	var outputCoins []*privacy.CoinV2
	var sharedSecrets []*privacy.Point
	var numOfCoinsBurned uint = 0
	var isBurning bool = false
	var senderKeySet incognitokey.KeySet
	_ = senderKeySet.InitFromPrivateKey(params.SenderSK)
	b := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	for _, inf := range params.PaymentInfo {
		c, ss, err := createUniqueOTACoinCA(inf, int(common.GetShardIDFromLastByte(b)), params.TokenID, params.StateDB)
		if err != nil {
			utils.Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
			return false, err
		}
		// the only way err!=nil but ss==nil is a coin meant for burning address
		if ss == nil {
			isBurning = true
			numOfCoinsBurned++
		}
		sharedSecrets = append(sharedSecrets, ss)
		outputCoins = append(outputCoins, c)
	}
	// first, reject the invalid case. After this, isBurning will correctly determine if TX is burning
	if numOfCoinsBurned > 1 {
		utils.Logger.Log.Errorf("Cannot burn multiple coins")
		return false, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("output must not have more than 1 burned coin"))
	}
	// outputCoins, err := newCoinV2ArrayFromPaymentInfoArray(params.PaymentInfo, params.TokenID, params.StateDB)

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.InputCoins
	tx.Proof, err = privacy.ProveV2(inputCoins, outputCoins, sharedSecrets, true, params.PaymentInfo)
	if err != nil {
		utils.Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return false, err
	}

	err = tx.signCA(inputCoins, outputCoins, sharedSecrets, params, tx.Hash()[:])
	return isBurning, err
}

func (tx *Tx) signCA(inp []privacy.PlainCoin, out []*privacy.CoinV2, outputSharedSecrets []*privacy.Point, params *tx_generic.TxPrivacyInitParams, hashedMessage []byte) error {
	if tx.Sig != nil {
		return utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("input transaction must be an unsigned one"))
	}
	ringSize := privacy.RingSize

	// Generate Ring
	piBig, piErr := common.RandBigIntMaxRange(big.NewInt(int64(ringSize)))
	if piErr != nil {
		return piErr
	}
	var pi int = int(piBig.Int64())
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	ring, indexes, commitmentsToZero, err := generateMlsagRingWithIndexesCA(inp, out, params, pi, shardID, ringSize)
	if err != nil {
		utils.Logger.Log.Errorf("generateMlsagRingWithIndexes got error %v ", err)
		return err
	}

	// Set SigPubKey
	txSigPubKey := new(SigPubKey)
	txSigPubKey.Indexes = indexes
	tx.SigPubKey, err = txSigPubKey.Bytes()
	if err != nil {
		utils.Logger.Log.Errorf("tx.SigPubKey cannot parse from Bytes, error %v ", err)
		return err
	}

	// Set sigPrivKey
	privKeysMlsag, err := createPrivKeyMlsagCA(inp, out, outputSharedSecrets, params, shardID, commitmentsToZero)
	if err != nil {
		utils.Logger.Log.Errorf("Cannot create private key of mlsag: %v", err)
		return err
	}
	sag := mlsag.NewMlsag(privKeysMlsag, ring, pi)
	sk, err := privacy.ArrayScalarToBytes(&privKeysMlsag)
	if err != nil {
		utils.Logger.Log.Errorf("tx.SigPrivKey cannot parse arrayScalar to Bytes, error %v ", err)
		return err
	}
	tx.SetPrivateKey(sk)

	// Set Signature
	mlsagSignature, err := sag.SignConfidentialAsset(hashedMessage)
	if err != nil {
		utils.Logger.Log.Errorf("Cannot signOnMessage mlsagSignature, error %v ", err)
		return err
	}
	// inputCoins already hold keyImage so set to nil to reduce size
	mlsagSignature.SetKeyImages(nil)
	tx.Sig, err = mlsagSignature.ToBytes()

	return err
}

// overwrite tokenID to fit existing prototype
func (tx *Tx) verifySigCA(transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("input transaction must be a signed one"))
	}

	var err error
	// Reform Ring
	sumOutputsWithFee := tx_generic.CalculateSumOutputsWithFee(tx.Proof.GetOutputCoins(), tx.Fee)
	sumOutputAssetTags := new(privacy.Point).Identity()
	for _, oc := range tx.Proof.GetOutputCoins() {
		output_specific, ok := oc.(*privacy.CoinV2)
		if !ok {
			utils.Logger.Log.Errorf("Error when casting coin as v2")
			return false, fmt.Errorf("error when casting coin as v2")
		}
		sumOutputAssetTags.Add(sumOutputAssetTags, output_specific.GetAssetTag())
	}
	inCount := new(privacy.Scalar).FromUint64(uint64(len(tx.GetProof().GetInputCoins())))
	outCount := new(privacy.Scalar).FromUint64(uint64(len(tx.GetProof().GetOutputCoins())))
	sumOutputAssetTags.ScalarMult(sumOutputAssetTags, inCount)

	ring, err := reconstructRingCAV2(tx.GetValidationEnv(), sumOutputsWithFee, sumOutputAssetTags, outCount, transactionStateDB)
	if err != nil {
		utils.Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	// Reform MLSAG Signature
	inputCoins := tx.Proof.GetInputCoins()
	keyImages := make([]*privacy.Point, len(inputCoins)+2)
	for i := 0; i < len(inputCoins); i++ {
		if inputCoins[i].GetKeyImage() == nil {
			utils.Logger.Log.Errorf("Error when reconstructing mlsagSignature: missing keyImage")
			return false, err
		}
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = privacy.RandomPoint()
	keyImages[len(inputCoins)+1] = privacy.RandomPoint()
	mlsagSignature, err := getMLSAGSigFromTxSigAndKeyImages(tx.Sig, keyImages)
	if err != nil {
		return false, err
	}

	return mlsag.VerifyConfidentialAsset(mlsagSignature, ring, tx.Hash()[:])
}

func createUniqueOTACoinCA(paymentInfo *privacy.PaymentInfo, senderShardID int, tokenID *common.Hash, stateDB *statedb.StateDB) (*privacy.CoinV2, *privacy.Point, error) {
	if tokenID == nil {
		tokenID = &common.PRVCoinID
	}
	for i := privacy.MaxPrivacyAttempts; i > 0; i-- {
		c, sharedSecret, err := privacy.NewCoinCA(privacy.NewCoinParams().From(paymentInfo, senderShardID, privacy.CoinPrivacyTypeTransfer), tokenID)
		if tokenID != nil && sharedSecret != nil && c != nil && c.GetAssetTag() != nil {
			utils.Logger.Log.Infof("Created a new coin with tokenID %s, shared secret %s, asset tag %s", tokenID.String(), sharedSecret.MarshalText(), c.GetAssetTag().MarshalText())
		}
		if err != nil {
			utils.Logger.Log.Errorf("Cannot parse coin based on payment info err: %v", err)
			return nil, nil, err
		}
		// If previously created coin is burning address
		if sharedSecret == nil {
			// assetTag := privacy.HashToPoint(tokenID[:])
			// c.SetAssetTag(assetTag)
			return c, nil, nil // No need to check db
		}
		// Onetimeaddress should be unique
		publicKeyBytes := c.GetPublicKey().ToBytesS()
		// here tokenID should always be TokenConfidentialAssetID (for db storage)
		found, _, err := statedb.HasOnetimeAddress(stateDB, common.ConfidentialAssetID, publicKeyBytes)
		if err != nil {
			utils.Logger.Log.Errorf("Cannot check public key existence in DB, err %v", err)
			return nil, nil, err
		}
		if !found {
			return c, sharedSecret, nil
		}
	}
	// MaxPrivacyAttempts could be exceeded if the OS's RNG or the statedb is corrupted
	utils.Logger.Log.Errorf("Cannot create unique OTA after %d attempts", privacy.MaxPrivacyAttempts)
	return nil, nil, fmt.Errorf("cannot create unique OTA")
}
