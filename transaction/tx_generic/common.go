package tx_generic //nolint:revive

import (
	"errors"
	"math"
	"math/big"
	"strconv"

	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	zku "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/utils"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

type RandomCommitmentsProcessParam struct {
	usableInputCoins []coin.PlainCoin
	randNum          int
	stateDB          *statedb.StateDB
	shardID          byte
	tokenID          *common.Hash
}

func NewRandomCommitmentsProcessParam(usableInputCoins []coin.PlainCoin, randNum int, stateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) *RandomCommitmentsProcessParam {
	return &RandomCommitmentsProcessParam{
		tokenID:          tokenID,
		shardID:          shardID,
		stateDB:          stateDB,
		randNum:          randNum,
		usableInputCoins: usableInputCoins,
	}
}

// RandomCommitmentsProcess - process list commitments and useable tx to create
// a list commitment random which be used to create a proof for new tx
// result contains
// commitmentIndexs = [{1,2,3,4,myindex1,6,7,8}{9,10,11,12,13,myindex2,15,16}...]
// myCommitmentIndexs = [4, 13, ...]
func RandomCommitmentsProcess(param *RandomCommitmentsProcessParam) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	if len(param.usableInputCoins) == 0 {
		return
	}
	if param.randNum == 0 {
		param.randNum = privacy.CommitmentRingSize // default
	}
	// loop to create list usable commitments from usableInputCoins
	listUsableCommitments := make(map[common.Hash][]byte)
	listUsableCommitmentsIndices := make([]common.Hash, len(param.usableInputCoins))
	// tick index of each usable commitment with full db commitments
	mapIndexCommitmentsInUsableTx := make(map[string]*big.Int)
	for i, in := range param.usableInputCoins {
		usableCommitment := in.GetCommitment().ToBytesS()
		commitmentInHash := common.HashH(usableCommitment)
		listUsableCommitments[commitmentInHash] = usableCommitment
		listUsableCommitmentsIndices[i] = commitmentInHash
		index, err := statedb.GetCommitmentIndex(param.stateDB, *param.tokenID, usableCommitment, param.shardID)
		if err != nil {
			utils.Logger.Log.Error(err)
			return
		}
		commitmentInBase58Check := base58.Base58Check{}.Encode(usableCommitment, common.ZeroByte)
		mapIndexCommitmentsInUsableTx[commitmentInBase58Check] = index
	}
	// loop to random commitmentIndexs
	cpRandNum := (len(listUsableCommitments) * param.randNum) - len(listUsableCommitments)
	lenCommitment, err1 := statedb.GetCommitmentLength(param.stateDB, *param.tokenID, param.shardID)
	if err1 != nil {
		utils.Logger.Log.Error(err1)
		return
	}
	if lenCommitment == nil {
		utils.Logger.Log.Error(errors.New("Commitments is empty"))
		return
	}
	if lenCommitment.Uint64() == 1 && len(param.usableInputCoins) == 1 {
		commitmentIndexs = []uint64{0, 0, 0, 0, 0, 0, 0}
		temp := param.usableInputCoins[0].GetCommitment().ToBytesS()
		commitments = [][]byte{temp, temp, temp, temp, temp, temp, temp}
	} else {
		for i := 0; i < cpRandNum; i++ {
			maxRetries := 5000
			var retries int
			for retries = maxRetries; retries > 0; retries-- {
				lenCommitment, _ = statedb.GetCommitmentLength(param.stateDB, *param.tokenID, param.shardID)
				index, _ := common.RandBigIntMaxRange(lenCommitment)
				ok, err := statedb.HasCommitmentIndex(param.stateDB, *param.tokenID, index.Uint64(), param.shardID)
				if ok && err == nil {
					temp, _ := statedb.GetCommitmentByIndex(param.stateDB, *param.tokenID, index.Uint64(), param.shardID)
					if _, found := listUsableCommitments[common.HashH(temp)]; !found {
						// random commitment not in commitments of usableinputcoin
						commitmentIndexs = append(commitmentIndexs, index.Uint64())
						commitments = append(commitments, temp)
						break
					}
				} else {
					continue
				}
			}
			// no usable commitment from db
			if retries == 0 {
				utils.Logger.Log.Errorf("Error : No available commitment in db")
				return
			}
		}
	}
	// loop to insert usable commitments into commitmentIndexs for every group
	j := 0
	for _, commitmentInHash := range listUsableCommitmentsIndices {
		commitmentValue := listUsableCommitments[commitmentInHash]
		index := mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(commitmentValue, common.ZeroByte)]
		randInt, _ := common.RandBigIntMaxRange(big.NewInt(int64(param.randNum)))
		i := (j * param.randNum) + int(randInt.Int64())
		commitmentIndexs = append(commitmentIndexs[:i], append([]uint64{index.Uint64()}, commitmentIndexs[i:]...)...)
		commitments = append(commitments[:i], append([][]byte{commitmentValue}, commitments[i:]...)...)
		myCommitmentIndexs = append(myCommitmentIndexs, uint64(i)) // create myCommitmentIndexs
		j++
	}
	return commitmentIndexs, myCommitmentIndexs, commitments
}

type EstimateTxSizeParam struct {
	version                  int
	numInputCoins            int
	numPayments              int
	hasPrivacy               bool
	metadata                 metadata.Metadata
	privacyCustomTokenParams *TokenParam
	limitFee                 uint64
}

func NewEstimateTxSizeParam(version, numInputCoins, numPayments int,
	hasPrivacy bool, metadata metadata.Metadata,
	privacyCustomTokenParams *TokenParam,
	limitFee uint64) *EstimateTxSizeParam {
	estimateTxSizeParam := &EstimateTxSizeParam{
		version:                  version,
		numInputCoins:            numInputCoins,
		numPayments:              numPayments,
		hasPrivacy:               hasPrivacy,
		limitFee:                 limitFee,
		metadata:                 metadata,
		privacyCustomTokenParams: privacyCustomTokenParams,
	}
	return estimateTxSizeParam
}

func toB64Len(numOfBytes uint64) uint64 {
	l := (numOfBytes*4 + 2) / 3
	l = ((l + 3) / 4) * 4
	return l
}

func EstimateProofSizeV2(numIn, numOut uint64) uint64 {
	coinSizeBound := uint64(257) + (privacy.Ed25519KeySize+1)*7 + privacy.TxRandomGroupSize + 1
	ipProofLRLen := uint64(math.Log2(float64(numOut))) + 1
	aggProofSizeBound := uint64(4) + 1 + privacy.Ed25519KeySize*uint64(7+numOut) + 1 + uint64(2*ipProofLRLen+3)*privacy.Ed25519KeySize
	// add 10 for rounding
	result := uint64(1) + (coinSizeBound+1)*uint64(numIn+numOut) + 2 + aggProofSizeBound + 10
	return toB64Len(result)
}

func EstimateTxSizeV2(estimateTxSizeParam *EstimateTxSizeParam) uint64 {
	jsonKeysSizeBound := uint64(20*10 + 2)
	sizeVersion := uint64(1)      // int8
	sizeType := uint64(5)         // string, max : 5
	sizeLockTime := uint64(8) * 3 // int64
	sizeFee := uint64(8) * 3      // uint64
	sizeInfo := toB64Len(uint64(512))

	numIn := uint64(estimateTxSizeParam.numInputCoins)
	numOut := uint64(estimateTxSizeParam.numPayments)

	sizeSigPubKey := uint64(numIn)*privacy.RingSize*9 + 2
	sizeSigPubKey = toB64Len(sizeSigPubKey)
	sizeSig := uint64(1) + numIn + (numIn+2)*privacy.RingSize
	sizeSig = sizeSig*33 + 3

	sizeProof := EstimateProofSizeV2(numIn, numOut)

	sizePubKeyLastByte := uint64(1) * 3
	sizeMetadata := uint64(0)
	if estimateTxSizeParam.metadata != nil {
		sizeMetadata += estimateTxSizeParam.metadata.CalculateSize()
	}

	sizeTx := jsonKeysSizeBound + sizeVersion + sizeType + sizeLockTime + sizeFee + sizeInfo + sizeSigPubKey + sizeSig + sizeProof + sizePubKeyLastByte + sizeMetadata
	if estimateTxSizeParam.privacyCustomTokenParams != nil {
		tokenKeysSizeBound := uint64(20*8 + 2)
		tokenSize := toB64Len(uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyID)))
		tokenSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertySymbol))
		tokenSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyName))
		tokenSize += 2
		numIn = uint64(len(estimateTxSizeParam.privacyCustomTokenParams.TokenInput))
		numOut = uint64(len(estimateTxSizeParam.privacyCustomTokenParams.Receiver))

		// shadow variable names
		sizeSigPubKey := uint64(numIn)*privacy.RingSize*9 + 2
		sizeSigPubKey = toB64Len(sizeSigPubKey)
		sizeSig := uint64(1) + numIn + (numIn+2)*privacy.RingSize
		sizeSig = sizeSig*33 + 3

		sizeProof := EstimateProofSizeV2(numIn, numOut)
		tokenSize += tokenKeysSizeBound + sizeSigPubKey + sizeSig + sizeProof
		sizeTx += tokenSize
	}
	return sizeTx
}

// EstimateTxSize returns the estimated size of the tx in kilobyte
func EstimateTxSize(estimateTxSizeParam *EstimateTxSizeParam) uint64 {
	if estimateTxSizeParam.version == 2 {
		return uint64(math.Ceil(float64(EstimateTxSizeV2(estimateTxSizeParam)) / 1024))
	}
	sizeVersion := uint64(1)  // int8
	sizeType := uint64(5)     // string, max : 5
	sizeLockTime := uint64(8) // int64
	sizeFee := uint64(8)      // uint64
	sizeInfo := uint64(512)

	sizeSigPubKey := uint64(common.SigPubKeySize)
	sizeSig := uint64(common.SigNoPrivacySize)
	if estimateTxSizeParam.hasPrivacy {
		sizeSig = uint64(common.SigPrivacySize)
	}

	sizeProof := uint64(0)
	if estimateTxSizeParam.numInputCoins != 0 || estimateTxSizeParam.numPayments != 0 {
		sizeProof = zku.EstimateProofSize(estimateTxSizeParam.numInputCoins, estimateTxSizeParam.numPayments, estimateTxSizeParam.hasPrivacy)
	} else if estimateTxSizeParam.limitFee > 0 {
		sizeProof = zku.EstimateProofSize(1, 1, estimateTxSizeParam.hasPrivacy)
	}

	sizePubKeyLastByte := uint64(1)

	sizeMetadata := uint64(0)
	if estimateTxSizeParam.metadata != nil {
		sizeMetadata += estimateTxSizeParam.metadata.CalculateSize()
	}

	sizeTx := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeInfo + sizeSigPubKey + sizeSig + sizeProof + sizePubKeyLastByte + sizeMetadata

	// size of privacy custom token  data
	if estimateTxSizeParam.privacyCustomTokenParams != nil {
		customTokenDataSize := uint64(0)

		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyID))
		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertySymbol))
		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyName))

		customTokenDataSize += 8 // for amount
		customTokenDataSize += 4 // for TokenTxType

		customTokenDataSize += uint64(1) // int8 version
		customTokenDataSize += uint64(5) // string, max : 5 type
		customTokenDataSize += uint64(8) // int64 locktime
		customTokenDataSize += uint64(8) // uint64 fee

		customTokenDataSize += uint64(64) // info

		customTokenDataSize += uint64(common.SigPubKeySize)  // sig pubkey
		customTokenDataSize += uint64(common.SigPrivacySize) // sig

		// Proof
		customTokenDataSize += zku.EstimateProofSize(len(estimateTxSizeParam.privacyCustomTokenParams.TokenInput), len(estimateTxSizeParam.privacyCustomTokenParams.Receiver), true)

		customTokenDataSize += uint64(1) // PubKeyLastByte

		sizeTx += customTokenDataSize
	}

	return uint64(math.Ceil(float64(sizeTx) / 1024))
}

func CalculateSumOutputsWithFee(outputCoins []coin.Coin, fee uint64) *operation.Point {
	sumOutputsWithFee := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i++ {
		sumOutputsWithFee.Add(sumOutputsWithFee, outputCoins[i].GetCommitment())
	}
	feeCommitment := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenValueIndex],
		new(operation.Scalar).FromUint64(fee),
	)
	sumOutputsWithFee.Add(sumOutputsWithFee, feeCommitment)
	return sumOutputsWithFee
}

func ValidateTxParams(params *TxPrivacyInitParams) error {
	if len(params.InputCoins) > 255 {
		return utils.NewTransactionErr(utils.InputCoinIsVeryLargeError, nil, strconv.Itoa(len(params.InputCoins)))
	}
	if len(params.PaymentInfo) > 254 {
		return utils.NewTransactionErr(utils.PaymentInfoIsVeryLargeError, nil, strconv.Itoa(len(params.PaymentInfo)))
	}
	if params.TokenID == nil {
		// using default PRV
		params.TokenID = &common.Hash{}
		err := params.TokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return utils.NewTransactionErr(utils.TokenIDInvalidError, err, params.TokenID.String())
		}
	}
	return nil
}

func ParseTokenID(tokenID *common.Hash) (*common.Hash, error) {
	if tokenID == nil {
		result := new(common.Hash)
		err := result.SetBytes(common.PRVCoinID[:])
		if err != nil {
			utils.Logger.Log.Error(err)
			return nil, utils.NewTransactionErr(utils.TokenIDInvalidError, err, tokenID.String())
		}
		return result, nil
	}
	return tokenID, nil
}

func VerifySigNoPrivacy(sig []byte, sigPubKey []byte, hashedMessage []byte) (bool, error) {
	// check input transaction
	if sig == nil || sigPubKey == nil {
		return false, utils.NewTransactionErr(utils.UnexpectedError, errors.New("input transaction must be an signed one"))
	}

	var err error
	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verifyKey := new(privacy.SchnorrPublicKey)
	sigPublicKey, err := new(operation.Point).FromBytesS(sigPubKey)

	if err != nil {
		utils.Logger.Log.Error(err)
		return false, utils.NewTransactionErr(utils.DecompressSigPubKeyError, err)
	}
	verifyKey.Set(sigPublicKey)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	err = signature.SetBytes(sig)
	if err != nil {
		utils.Logger.Log.Error(err)
		return false, utils.NewTransactionErr(utils.InitTxSignatureFromBytesError, err)
	}

	// verify signature
	/*utils.Logger.log.Debugf(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash()[:])
	if tx.Proof != nil {
		utils.Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX Proof bytes before verifing the signature: %v\n", tx.Proof.Bytes())
	}
	utils.Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX meta: %v\n", tx.Metadata)*/
	res := verifyKey.Verify(signature, hashedMessage)
	return res, nil
}

func SignNoPrivacy(privKey *privacy.PrivateKey, hashedMessage []byte) (signatureBytes []byte, sigPubKey []byte, err error) {
	/****** using Schnorr signature *******/
	// sign with sigPrivKey
	// prepare private key for Schnorr
	sk := new(operation.Scalar).FromBytesS(*privKey)
	r := new(operation.Scalar).FromUint64(0)
	sigKey := new(schnorr.SchnorrPrivateKey)
	sigKey.Set(sk, r)
	signature, err := sigKey.Sign(hashedMessage)
	if err != nil {
		return nil, nil, err
	}

	signatureBytes = signature.Bytes()
	sigPubKey = sigKey.GetPublicKey().GetPublicKey().ToBytesS()
	return signatureBytes, sigPubKey, nil
}
