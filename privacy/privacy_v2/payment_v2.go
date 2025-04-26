//nolint:revive // skip linter for this package name
package privacy_v2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/env"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/bulletproofs"
	"github.com/incognitochain/incognito-chain/privacy/proof/agg_interface"
)

// PaymentProofV2 contains the input & output coins, along with the Bulletproofs for output coin range.
// This is what shows up in a transaction's Proof field.
type PaymentProofV2 struct {
	Version              uint8
	aggregatedRangeProof *bulletproofs.AggregatedRangeProof
	inputCoins           []coin.PlainCoin
	outputCoins          []*coin.CoinV2
}

func (proof *PaymentProofV2) SetVersion()       { proof.Version = 2 }
func (proof *PaymentProofV2) GetVersion() uint8 { return 2 }

// GetInputCoins is the getter for input coins.
func (proof PaymentProofV2) GetInputCoins() []coin.PlainCoin { return proof.inputCoins }

// GetOutputCoins is the getter for output coins.
func (proof PaymentProofV2) GetOutputCoins() []coin.Coin {
	res := make([]coin.Coin, len(proof.outputCoins))
	for i := 0; i < len(proof.outputCoins); i++ {
		res[i] = proof.outputCoins[i]
	}
	return res
}

// GetAggregatedRangeProof returns the Bulletproof in this, but as a generic range proof object.
func (proof PaymentProofV2) GetAggregatedRangeProof() agg_interface.AggregatedRangeProof {
	return proof.aggregatedRangeProof
}

func (proof *PaymentProofV2) SetInputCoins(v []coin.PlainCoin) error {
	var err error
	proof.inputCoins = make([]coin.PlainCoin, len(v))
	for i := 0; i < len(v); i++ {
		b := v[i].Bytes()
		if proof.inputCoins[i], err = coin.NewPlainCoinFromByte(b); err != nil {
			Logger.Log.Errorf("Proofv2 cannot create inputCoins from new plain coin from bytes: err %v", err)
			return err
		}
	}
	return nil
}

func (proof *PaymentProofV2) SetOutputCoinsV2(v []*coin.CoinV2) error {
	var err error
	proof.outputCoins = make([]*coin.CoinV2, len(v))
	for i := 0; i < len(v); i++ {
		b := v[i].Bytes()
		proof.outputCoins[i] = new(coin.CoinV2)
		if err = proof.outputCoins[i].SetBytes(b); err != nil {
			Logger.Log.Errorf("Proofv2 cannot set byte to outputCoins : err %v", err)
			return err
		}
	}
	return nil
}

// v should be all coinv2 or else it would crash
func (proof *PaymentProofV2) SetOutputCoins(v []coin.Coin) error {
	var err error
	proof.outputCoins = make([]*coin.CoinV2, len(v))
	for i := 0; i < len(v); i++ {
		proof.outputCoins[i] = new(coin.CoinV2)
		b := v[i].Bytes()
		if err = proof.outputCoins[i].SetBytes(b); err != nil {
			Logger.Log.Errorf("Proofv2 cannot set byte to outputCoins : err %v", err)
			return err
		}
	}
	return nil
}

func (proof *PaymentProofV2) SetAggregatedRangeProof(aggregatedRangeProof *bulletproofs.AggregatedRangeProof) {
	proof.aggregatedRangeProof = aggregatedRangeProof
}

// Init allocates and zeroes all fields in this proof.
func (proof *PaymentProofV2) Init() {
	aggregatedRangeProof := &bulletproofs.AggregatedRangeProof{}
	aggregatedRangeProof.Init()
	proof.Version = 2
	proof.aggregatedRangeProof = aggregatedRangeProof
	proof.inputCoins = []coin.PlainCoin{}
	proof.outputCoins = []*coin.CoinV2{}
}

// MarshalJSON implements JSON Marshaller
func (proof PaymentProofV2) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	// temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	temp := base64.StdEncoding.EncodeToString(data)
	return json.Marshal(temp)
}

// UnmarshalJSON implements JSON Unmarshaller
func (proof *PaymentProofV2) UnmarshalJSON(data []byte) error {
	dataStr := common.EmptyString
	errJson := json.Unmarshal(data, &dataStr)
	if errJson != nil {
		Logger.Log.Errorf("PaymentProofV2 unmarshalling dataStr error: %v\n", errJson)
		return errJson
	}
	temp, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		Logger.Log.Errorf("PaymentProofV2 decodeing string dataStr error: %v\n", err)
		return err
	}

	errSetBytes := proof.SetBytes(temp)
	if errSetBytes != nil {
		Logger.Log.Errorf("PaymentProofV2 setbytes error: %v\n", errSetBytes)
		return errSetBytes
	}
	return nil
}

// Bytes does byte serialization for this payment proof
func (proof PaymentProofV2) Bytes() []byte {
	var bytes []byte
	bytes = append(bytes, proof.GetVersion())

	comOutputMultiRangeProof := proof.aggregatedRangeProof.Bytes()
	var rangeProofLength uint32 = uint32(len(comOutputMultiRangeProof))
	bytes = append(bytes, common.Uint32ToBytes(rangeProofLength)...)
	bytes = append(bytes, comOutputMultiRangeProof...)

	// InputCoins
	bytes = append(bytes, byte(len(proof.inputCoins)))
	for i := 0; i < len(proof.inputCoins); i++ {
		inputCoins := proof.inputCoins[i].Bytes()
		lenInputCoins := len(inputCoins)
		var lenInputCoinsBytes []byte
		if lenInputCoins < 256 {
			lenInputCoinsBytes = []byte{byte(lenInputCoins)}
		} else {
			lenInputCoinsBytes = common.IntToBytes(lenInputCoins)
		}

		bytes = append(bytes, lenInputCoinsBytes...)
		bytes = append(bytes, inputCoins...)
	}

	// OutputCoins
	bytes = append(bytes, byte(len(proof.outputCoins)))
	for i := 0; i < len(proof.outputCoins); i++ {
		outputCoins := proof.outputCoins[i].Bytes()
		lenOutputCoins := len(outputCoins)
		var lenOutputCoinsBytes []byte
		if lenOutputCoins < 256 {
			lenOutputCoinsBytes = []byte{byte(lenOutputCoins)}
		} else {
			lenOutputCoinsBytes = common.IntToBytes(lenOutputCoins)
		}

		bytes = append(bytes, lenOutputCoinsBytes...)
		bytes = append(bytes, outputCoins...)
	}

	return bytes
}

// SetBytes does byte deserialization for this payment proof
func (proof *PaymentProofV2) SetBytes(proofbytes []byte) *errhandler.PrivacyError {
	if len(proofbytes) == 0 {
		return errhandler.NewPrivacyErr(errhandler.InvalidInputToSetBytesErr, errors.New("Proof bytes is zero"))
	}
	if proofbytes[0] != proof.GetVersion() {
		Logger.Log.Errorf("proof bytes version is incorrect: %v != %v\n", proofbytes[0], proof.GetVersion())
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Proof bytes version is incorrect"))
	}
	proof.SetVersion()
	offset := 1

	// ComOutputMultiRangeProofSize *aggregatedRangeProof
	if offset+common.Uint32Size >= len(proofbytes) {
		Logger.Log.Errorf("out of range aggregated range proof: %v + %v >= %v\n", offset, common.Uint32Size, len(proofbytes))
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range aggregated range proof"))
	}
	lenComOutputMultiRangeUint32, _ := common.BytesToUint32(proofbytes[offset : offset+common.Uint32Size])
	lenComOutputMultiRangeProof := int(lenComOutputMultiRangeUint32)
	offset += common.Uint32Size

	if offset+lenComOutputMultiRangeProof > len(proofbytes) {
		Logger.Log.Errorf("out of range aggregated range proof: %v + %v >= %v\n", offset, lenComOutputMultiRangeProof, len(proofbytes))
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range aggregated range proof"))
	}
	if lenComOutputMultiRangeProof > 0 {
		bulletproof := &bulletproofs.AggregatedRangeProof{}
		bulletproof.Init()
		proof.aggregatedRangeProof = bulletproof
		err := proof.aggregatedRangeProof.SetBytes(proofbytes[offset : offset+lenComOutputMultiRangeProof])
		if err != nil {
			Logger.Log.Errorf("aggregated range proof setbytes error: %v\n", err)
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenComOutputMultiRangeProof
	}

	if offset >= len(proofbytes) {
		Logger.Log.Errorf("out of range input coin: %v >= %v\n", offset, len(proofbytes))
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
	}
	lenInputCoinsArray := int(proofbytes[offset])
	offset++
	proof.inputCoins = make([]coin.PlainCoin, lenInputCoinsArray)
	var err error
	for i := 0; i < lenInputCoinsArray; i++ {
		// try get 1-byte for len
		if offset >= len(proofbytes) {
			Logger.Log.Errorf("out of range input coin: %v >= %v\n", offset, len(proofbytes))
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		lenInputCoin := int(proofbytes[offset])
		offset++

		if offset+lenInputCoin > len(proofbytes) {
			Logger.Log.Errorf("out of range input coin: %v + %v >= %v\n", offset, lenInputCoin, len(proofbytes))
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		proof.inputCoins[i], err = coin.NewPlainCoinFromByte(proofbytes[offset : offset+lenInputCoin])
		if err != nil {
			// 1-byte is wrong
			// try get 2-byte for len
			if offset+1 > len(proofbytes) {
				Logger.Log.Error("out of range input coin")
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
			}
			lenInputCoin = common.BytesToInt(proofbytes[offset-1 : offset+1])
			offset++

			if offset+lenInputCoin > len(proofbytes) {
				Logger.Log.Error("out of range input coin")
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
			}
			proof.inputCoins[i], err = coin.NewPlainCoinFromByte(proofbytes[offset : offset+lenInputCoin])
			if err != nil {
				Logger.Log.Errorf("input coin setbytes error: %v\n", err)
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
			}
		}
		offset += lenInputCoin
	}

	if offset >= len(proofbytes) {
		Logger.Log.Error("out of range output coin: %v >= %v\n", offset, len(proofbytes))
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
	}
	lenOutputCoinsArray := int(proofbytes[offset])
	offset++
	proof.outputCoins = make([]*coin.CoinV2, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		proof.outputCoins[i] = new(coin.CoinV2)
		// try get 1-byte for len
		if offset >= len(proofbytes) {
			Logger.Log.Error("out of range output coin: %v >= %v\n", offset, len(proofbytes))
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		lenOutputCoin := int(proofbytes[offset])
		offset++

		if offset+lenOutputCoin > len(proofbytes) {
			Logger.Log.Error("out of range output coin: %v + %v >= %v\n", offset, lenOutputCoin, len(proofbytes))
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		err := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
		if err != nil {
			// 1-byte is wrong
			// try get 2-byte for len
			if offset+1 > len(proofbytes) {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
			}
			lenOutputCoin = common.BytesToInt(proofbytes[offset-1 : offset+1])
			offset++

			if offset+lenOutputCoin > len(proofbytes) {
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
			}
			err1 := proof.outputCoins[i].SetBytes(proofbytes[offset : offset+lenOutputCoin])
			if err1 != nil {
				Logger.Log.Errorf("output coin setbytes error: %v\n", err1)
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err1)
			}
		}
		offset += lenOutputCoin
	}

	return nil
}

// IsPrivacy is a helper that returns true when an output is encrypted,
// which means the transaction is of "privacy" type. It says nothing about the validity of this proof.
//
// This is not a tight classifier between "privacy" and "non-privacy",
// and should not be called before sanity check.
func (proof *PaymentProofV2) IsPrivacy() bool {
	for _, outCoin := range proof.GetOutputCoins() {
		if !outCoin.IsEncrypted() {
			return false
		}
	}
	return true
}

// IsConfidentialAsset returns true if this is a Confidential Asset transaction (all coins in it must have asset tag field).
// An error means the proof is simply invalid. After this function returns, check the error first.
func (proof *PaymentProofV2) IsConfidentialAsset() (bool, error) {
	// asset tag consistency check
	assetTagCount := 0
	inputCoins := proof.GetInputCoins()
	for _, c := range inputCoins {
		coinSpecific, ok := c.(*coin.CoinV2)
		if !ok {
			Logger.Log.Errorf("Cannot cast a coin to v2 : %v", c.Bytes())
			return false, errhandler.NewPrivacyErr(errhandler.UnexpectedErr, errors.New("Casting error : CoinV2"))
		}
		if coinSpecific.GetAssetTag() != nil {
			assetTagCount++
		}
	}
	outputCoins := proof.GetOutputCoins()
	for _, c := range outputCoins {
		coinSpecific, ok := c.(*coin.CoinV2)
		if !ok {
			Logger.Log.Errorf("Cannot cast a coin to v2 : %v", c.Bytes())
			return false, errhandler.NewPrivacyErr(errhandler.UnexpectedErr, errors.New("Casting error : CoinV2"))
		}
		if coinSpecific.GetAssetTag() != nil {
			assetTagCount++
		}
	}

	if assetTagCount == len(inputCoins)+len(outputCoins) {
		return true, nil
	} else if assetTagCount == 0 {
		return false, nil
	}
	return false, errhandler.NewPrivacyErr(errhandler.UnexpectedErr, errors.New("Error : TX contains both confidential asset & non-CA coins"))
}

// ValidateSanity performs sanity check for this proof.
// The input parameter is ingored.
func (proof PaymentProofV2) ValidateSanity(vEnv env.ValidationEnviroment) (bool, error) {
	if len(proof.GetInputCoins()) > 255 {
		return false, errors.New("Input coins in tx are very large:" + strconv.Itoa(len(proof.GetInputCoins())))
	}

	if len(proof.GetOutputCoins()) > 255 {
		return false, errors.New("Output coins in tx are very large:" + strconv.Itoa(len(proof.GetOutputCoins())))
	}

	if !proof.aggregatedRangeProof.ValidateSanity() {
		return false, errors.New("validate sanity Aggregated range proof failed")
	}

	// check output coins with privacy
	duplicatePublicKeys := make(map[string]bool)
	outputCoins := proof.GetOutputCoins()
	// cmsValues := proof.aggregatedRangeProof.GetCommitments()
	for i, outputCoin := range outputCoins {
		if outputCoin.GetPublicKey() == nil || !outputCoin.GetPublicKey().PointValid() {
			return false, errors.New("validate sanity Public key of output coin failed")
		}

		// check duplicate output addresses
		pubkeyStr := string(outputCoin.GetPublicKey().ToBytesS())
		if _, ok := duplicatePublicKeys[pubkeyStr]; ok {
			return false, errors.New("Cannot have duplicate publickey ")
		}
		duplicatePublicKeys[pubkeyStr] = true

		if !outputCoin.GetCommitment().PointValid() {
			return false, errors.New("validate sanity Coin commitment of output coin failed")
		}

		// re-compute the commitment if the output coin's address is the burning address
		// burn TX cannot use confidential asset]
		// BOOKMARK
		if common.IsPublicKeyBurningAddress(outputCoins[i].GetPublicKey().ToBytesS()) {
			value := outputCoin.GetValue()
			rand := outputCoin.GetRandomness()
			commitment := operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(value), rand, coin.PedersenValueIndex)
			outputCoinSpecific, ok := outputCoin.(*coin.CoinV2)
			if !ok {
				return false, errors.New("Validate sanity - Cannot cast a coin to v2")
			}
			if outputCoinSpecific.GetAssetTag() != nil {
				com, err := outputCoinSpecific.ComputeCommitmentCA()
				if err != nil {
					return false, errors.New("Cannot compute commitment for confidential asset")
				}
				commitment = com
			}
			if !operation.IsPointEqual(commitment, outputCoin.GetCommitment()) {
				return false, errors.New("validate sanity Coin commitment of burned coin failed")
			}
		}
	}
	return true, nil
}

// Prove returns the payment proof object.
// It generates Bulletproofs for output commitments,
// then conceals sensitive information in the coins.
// The parameter hasConfidentialAsset will determine the type of Bulletproof prover to use.
func Prove(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, sharedSecrets []*operation.Point, hasConfidentialAsset bool, paymentInfo []*key.PaymentInfo) (*PaymentProofV2, error) {
	var err error

	proof := new(PaymentProofV2)
	proof.SetVersion()
	// aggregateproof := new(bulletproofs.AggregatedRangeProof)
	// aggregateproof.Init()
	// proof.aggregatedRangeProof = aggregateproof
	if err = proof.SetInputCoins(inputCoins); err != nil {
		Logger.Log.Errorf("Cannot set input coins in payment_v2 proof: err %v", err)
		return nil, err
	}
	if err = proof.SetOutputCoinsV2(outputCoins); err != nil {
		Logger.Log.Errorf("Cannot set output coins in payment_v2 proof: err %v", err)
		return nil, err
	}

	// Prepare range proofs
	n := len(outputCoins)
	outputValues := make([]uint64, n)
	outputRands := make([]*operation.Scalar, n)
	for i := 0; i < n; i++ {
		outputValues[i] = outputCoins[i].GetValue()
		outputRands[i] = outputCoins[i].GetRandomness()
	}

	wit := new(bulletproofs.AggregatedRangeWitness)
	wit.Set(outputValues, outputRands)
	if hasConfidentialAsset {
		blinders := make([]*operation.Scalar, len(sharedSecrets))
		for i := range sharedSecrets {
			if sharedSecrets[i] == nil {
				blinders[i] = new(operation.Scalar).FromUint64(0)
			} else {
				blinders[i], err = coin.ComputeAssetTagBlinder(sharedSecrets[i])
				if err != nil {
					return nil, err
				}
			}
		}
		var err error
		wit, err = bulletproofs.TransformWitnessToCAWitness(wit, blinders)
		if err != nil {
			return nil, err
		}

		theBase, err := bulletproofs.GetFirstAssetTag(outputCoins)
		if err != nil {
			return nil, err
		}
		proof.aggregatedRangeProof, err = wit.ProveUsingBase(theBase)

		outputCommitments := make([]*operation.Point, n)
		for i := 0; i < n; i++ {
			com, err := outputCoins[i].ComputeCommitmentCA()
			if err != nil {
				return nil, err
			}
			outputCommitments[i] = com
		}
		proof.aggregatedRangeProof.SetCommitments(outputCommitments)
		if err != nil {
			return nil, err
		}
	} else {
		proof.aggregatedRangeProof, err = wit.Prove()
		if err != nil {
			return nil, err
		}
	}

	// After Prove, we should hide all information in coin details.
	for i, outputCoin := range proof.outputCoins {
		if !common.IsPublicKeyBurningAddress(outputCoin.GetPublicKey().ToBytesS()) {
			if err = outputCoin.ConcealOutputCoin(paymentInfo[i].PaymentAddress.GetPublicView()); err != nil {
				return nil, err
			}

			// OutputCoin.GetKeyImage should be nil even though we do not have it
			// Because otherwise the RPC server will return the Bytes of [1 0 0 0 0 ...] (the default byte)
			proof.outputCoins[i].SetKeyImage(nil)
		}

	}

	for _, inputCoin := range proof.GetInputCoins() {
		c, ok := inputCoin.(*coin.CoinV2)
		if !ok {
			return nil, errors.New("Input c of PaymentProofV2 must be CoinV2")
		}
		c.ConcealInputCoin()
	}

	return proof, nil
}

func (proof PaymentProofV2) verifyHasConfidentialAsset(isBatch bool) (bool, error) {
	cmsValues := proof.aggregatedRangeProof.GetCommitments()
	if len(proof.GetOutputCoins()) != len(cmsValues) {
		return false, errors.New("Commitment length mismatch")
	}
	// Verify the proof that output values and sum of them do not exceed v_max
	for i := 0; i < len(proof.outputCoins); i++ {

		if !proof.outputCoins[i].IsEncrypted() {
			if common.IsPublicKeyBurningAddress(proof.outputCoins[i].GetPublicKey().ToBytesS()) {
				continue
			}
			return false, errors.New("Verify has privacy should have every coin encrypted")
		}
		// check if output coins' commitment is the same as in the proof
		if !operation.IsPointEqual(cmsValues[i], proof.outputCoins[i].GetCommitment()) {
			return false, errors.New("Coin & Proof Commitments mismatch")
		}
	}
	// for CA, batching is not supported
	theBase, err := bulletproofs.GetFirstAssetTag(proof.outputCoins)
	if err != nil {
		return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err)
	}
	valid, err := proof.aggregatedRangeProof.VerifyUsingBase(theBase)
	if !valid {
		Logger.Log.Errorf("VERIFICATION PAYMENT PROOF V2: Multi-range failed")
		return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err)
	}
	return true, nil
}

func (proof PaymentProofV2) verifyHasNoCA(isBatch bool) (bool, error) {
	cmsValues := proof.aggregatedRangeProof.GetCommitments()
	if len(proof.GetOutputCoins()) != len(cmsValues) {
		return false, errors.New("Commitment length mismatch")
	}
	// Verify the proof that output values and sum of them do not exceed v_max
	for i := 0; i < len(proof.outputCoins); i++ {

		if !proof.outputCoins[i].IsEncrypted() {
			if common.IsPublicKeyBurningAddress(proof.outputCoins[i].GetPublicKey().ToBytesS()) {
				continue
			}
			return false, errors.New("Verify has privacy should have every coin encrypted")
		}
		// check if output coins' commitment is the same as in the proof
		if !operation.IsPointEqual(cmsValues[i], proof.outputCoins[i].GetCommitment()) {
			return false, errors.New("Coin & Proof Commitments mismatch")
		}
	}
	if !isBatch {
		valid, err := proof.aggregatedRangeProof.Verify()
		if !valid {
			Logger.Log.Errorf("VERIFICATION PAYMENT PROOF V2: Multi-range failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifyAggregatedProofFailedErr, err)
		}
	}
	return true, nil
}

// Verify performs verification on this payment proof.
// It verifies the Bulletproof inside & checks for duplicates among outputs.
// It works with Bulletproof batching (in which case it skips that verification, since that is handled in batchTransaction struct).
func (proof PaymentProofV2) Verify(boolParams map[string]bool, _ key.PublicKey, _ uint64, _ byte, _ *common.Hash, _ interface{}) (bool, error) {
	hasConfidentialAsset, ok := boolParams["hasConfidentialAsset"]
	if !ok {
		hasConfidentialAsset = true
	}

	isBatch, ok := boolParams["isBatch"]
	if !ok {
		isBatch = false
	}

	inputCoins := proof.GetInputCoins()
	dupMap := make(map[string]bool)
	for _, inCoin := range inputCoins {
		identifier := base64.StdEncoding.EncodeToString(inCoin.GetKeyImage().ToBytesS())
		_, exists := dupMap[identifier]
		if exists {
			return false, errors.New("Duplicate input inCoin in PaymentProofV2")
		}
		dupMap[identifier] = true
	}

	if !hasConfidentialAsset {
		return proof.verifyHasNoCA(isBatch)
	}
	return proof.verifyHasConfidentialAsset(isBatch)
}
