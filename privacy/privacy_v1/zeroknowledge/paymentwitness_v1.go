package zkp

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/aggregatedrange"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/oneoutofmany"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumberprivacy"
	logUtils "github.com/incognitochain/incognito-chain/utils"
)

// PaymentWitness contains all of witness for proving when spending coins
type PaymentWitness struct {
	privateKey          *operation.Scalar
	inputCoins          []coin.PlainCoin
	outputCoins         []*coin.CoinV1
	commitmentIndices   []uint64
	myCommitmentIndices []uint64

	oneOfManyWitness             []*oneoutofmany.OneOutOfManyWitness
	serialNumberWitness          []*serialnumberprivacy.SNPrivacyWitness
	serialNumberNoPrivacyWitness []*serialnumbernoprivacy.SNNoPrivacyWitness

	aggregatedRangeWitness *aggregatedrange.AggregatedRangeWitness

	comOutputValue                 []*operation.Point
	comOutputSerialNumberDerivator []*operation.Point
	comOutputShardID               []*operation.Point

	comInputSecretKey             *operation.Point
	comInputValue                 []*operation.Point
	comInputSerialNumberDerivator []*operation.Point
	comInputShardID               *operation.Point

	randSecretKey *operation.Scalar
}

func (paymentWitness PaymentWitness) GetRandSecretKey() *operation.Scalar {
	return paymentWitness.randSecretKey
}

type PaymentWitnessParam struct {
	HasPrivacy              bool
	PrivateKey              *operation.Scalar
	InputCoins              []coin.PlainCoin
	OutputCoins             []*coin.CoinV1
	PublicKeyLastByteSender byte
	Commitments             []*operation.Point
	CommitmentIndices       []uint64
	MyCommitmentIndices     []uint64
	Fee                     uint64
}

// Build prepares witnesses for all protocol need to be proved when create tx
// if hashPrivacy = false, witness includes spending key, input coins, output coins
// otherwise, witness includes all attributes in PaymentWitness struct
func (wit *PaymentWitness) Init(PaymentWitnessParam PaymentWitnessParam) *errhandler.PrivacyError {
	defer func() {
		if r := recover(); r != nil {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init panic: %v", r)

		}
	}()

	// log payment witness param
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam: %v", PaymentWitnessParam)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.PrivateKey: %v", PaymentWitnessParam.PrivateKey)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.InputCoins: %v", PaymentWitnessParam.InputCoins)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.OutputCoins: %v", PaymentWitnessParam.OutputCoins)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.CommitmentIndices: %v", PaymentWitnessParam.CommitmentIndices)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.MyCommitmentIndices: %v", PaymentWitnessParam.MyCommitmentIndices)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.Fee: %v", PaymentWitnessParam.Fee)

	_ = PaymentWitnessParam.Fee

	// set payment witness param
	wit.privateKey = PaymentWitnessParam.PrivateKey
	wit.inputCoins = PaymentWitnessParam.InputCoins
	wit.outputCoins = PaymentWitnessParam.OutputCoins
	wit.commitmentIndices = PaymentWitnessParam.CommitmentIndices
	wit.myCommitmentIndices = PaymentWitnessParam.MyCommitmentIndices

	randInputSK := operation.RandomScalar()
	wit.randSecretKey = new(operation.Scalar).Set(randInputSK)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.HasPrivacy: %v", PaymentWitnessParam.HasPrivacy)
	if !PaymentWitnessParam.HasPrivacy {
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init PaymentWitnessParam.HasPrivacy: %v", PaymentWitnessParam.HasPrivacy)
		for _, outCoin := range wit.outputCoins {
			outCoin.CoinDetails.SetRandomness(operation.RandomScalar())
			err := outCoin.CoinDetails.CommitAll()
			if err != nil {
				logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init err: %v", err)
				return errhandler.NewPrivacyErr(errhandler.CommitNewOutputCoinNoPrivacyErr, nil)
			}
		}
		lenInputs := len(wit.inputCoins)

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init lenInputs: %v", lenInputs)
		if lenInputs > 0 {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init lenInputs: %v", lenInputs)
			wit.serialNumberNoPrivacyWitness = make([]*serialnumbernoprivacy.SNNoPrivacyWitness, lenInputs)
			for i := 0; i < len(wit.inputCoins); i++ {
				/***** Build witness for proving that serial number is derived from the committed derivator *****/
				if wit.serialNumberNoPrivacyWitness[i] == nil {
					wit.serialNumberNoPrivacyWitness[i] = new(serialnumbernoprivacy.SNNoPrivacyWitness)
				}
				wit.serialNumberNoPrivacyWitness[i].Set(wit.inputCoins[i].GetKeyImage(), wit.inputCoins[i].GetPublicKey(),
					wit.inputCoins[i].GetSNDerivator(), wit.privateKey)
			}
		}
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init done")
		return nil
	}

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init numInputCoin")
	numInputCoin := len(wit.inputCoins)
	numOutputCoin := len(wit.outputCoins)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init cmInputSK")
	cmInputSK := operation.PedCom.CommitAtIndex(wit.privateKey, randInputSK, operation.PedersenPrivateKeyIndex)
	wit.comInputSecretKey = new(operation.Point).Set(cmInputSK)

	randInputShardID := FixedRandomnessShardID
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init randInputShardID: %v", randInputShardID)
	senderShardID := common.GetShardIDFromLastByte(PaymentWitnessParam.PublicKeyLastByteSender)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init senderShardID: %v", senderShardID)
	wit.comInputShardID = operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(uint64(senderShardID)), randInputShardID, operation.PedersenShardIDIndex)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init comInputShardID: %v", wit.comInputShardID)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init comInputValue")
	wit.comInputValue = make([]*operation.Point, numInputCoin)
	wit.comInputSerialNumberDerivator = make([]*operation.Point, numInputCoin)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init comInputValue done")

	// It is used for proving 2 commitments commit to the same value (input)
	//cmInputSNDIndexSK := make([]*operation.Point, numInputCoin)

	randInputValue := make([]*operation.Scalar, numInputCoin)
	randInputSND := make([]*operation.Scalar, numInputCoin)
	//randInputSNDIndexSK := make([]*big.Int, numInputCoin)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init cmInputValueAll")
	// cmInputValueAll is sum of all input coins' value commitments
	cmInputValueAll := new(operation.Point).Identity()
	randInputValueAll := new(operation.Scalar).FromUint64(0)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init randInputValue")
	// Summing all commitments of each input coin into one commitment and proving the knowledge of its Openings
	cmInputSum := make([]*operation.Point, numInputCoin)
	randInputSum := make([]*operation.Scalar, numInputCoin)
	// randInputSumAll is sum of all randomess of coin commitments
	randInputSumAll := new(operation.Scalar).FromUint64(0)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init oneOfManyWitness")
	wit.oneOfManyWitness = make([]*oneoutofmany.OneOutOfManyWitness, numInputCoin)
	wit.serialNumberWitness = make([]*serialnumberprivacy.SNPrivacyWitness, numInputCoin)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init commitmentTemps")
	commitmentTemps := make([][]*operation.Point, numInputCoin)
	randInputIsZero := make([]*operation.Scalar, numInputCoin)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init preIndex")
	preIndex := 0
	commitments := PaymentWitnessParam.Commitments
	for i, inputCoin := range wit.inputCoins {
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init inputCoin: %v", inputCoin)
		// tx only has fee, no output, Rand_Value_Input = 0
		if numOutputCoin == 0 {
			randInputValue[i] = new(operation.Scalar).FromUint64(0)
		} else {
			randInputValue[i] = operation.RandomScalar()
		}
		// commit each component of coin commitment
		randInputSND[i] = operation.RandomScalar()

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init wit.comInputValue[i]")
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init inputCoin.GetValue(): %v", inputCoin.GetValue())
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init randInputValue[i]: %v", randInputValue[i])
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init operation.PedersenValueIndex: %v", operation.PedersenValueIndex)
		wit.comInputValue[i] = operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(inputCoin.GetValue()), randInputValue[i], operation.PedersenValueIndex)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init wit.comInputValue[i]: %v", wit.comInputValue[i])

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init wit.comInputSerialNumberDerivator[i]")
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init inputCoin.GetSNDerivator(): %v", inputCoin.GetSNDerivator())
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init randInputSND[i]: %v", randInputSND[i])
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init operation.PedersenSndIndex: %v", operation.PedersenSndIndex)
		snDerivator := inputCoin.GetSNDerivator()
		if snDerivator == nil {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init inputCoin.GetSharedConcealRandom(): %v", inputCoin.GetSharedConcealRandom())
			snDerivator = inputCoin.GetSharedConcealRandom()
		}

		wit.comInputSerialNumberDerivator[i] = operation.PedCom.CommitAtIndex(snDerivator, randInputSND[i], operation.PedersenSndIndex)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init wit.comInputSerialNumberDerivator[i]: %v", wit.comInputSerialNumberDerivator[i])

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init cmInputValueAll")
		cmInputValueAll.Add(cmInputValueAll, wit.comInputValue[i])
		randInputValueAll.Add(randInputValueAll, randInputValue[i])

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init cmInputSum")
		/***** Build witness for proving one-out-of-N commitments is a commitment to the coins being spent *****/
		cmInputSum[i] = new(operation.Point).Add(cmInputSK, wit.comInputValue[i])
		cmInputSum[i].Add(cmInputSum[i], wit.comInputSerialNumberDerivator[i])
		cmInputSum[i].Add(cmInputSum[i], wit.comInputShardID)

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init randInputSum")
		randInputSum[i] = new(operation.Scalar).Set(randInputSK)
		randInputSum[i].Add(randInputSum[i], randInputValue[i])
		randInputSum[i].Add(randInputSum[i], randInputSND[i])
		randInputSum[i].Add(randInputSum[i], randInputShardID)

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init randInputSumAll")
		randInputSumAll.Add(randInputSumAll, randInputSum[i])

		// commitmentTemps is a list of commitments for protocol one-out-of-N
		commitmentTemps[i] = make([]*operation.Point, privacy_util.CommitmentRingSize)

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init randInputIsZero")
		randInputIsZero[i] = new(operation.Scalar).FromUint64(0)
		randInputIsZero[i].Sub(inputCoin.GetRandomness(), randInputSum[i])

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init commitmentTemps")
		for j := 0; j < privacy_util.CommitmentRingSize; j++ {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init commitmentTemps[i][j]")
			commitmentTemps[i][j] = new(operation.Point).Sub(commitments[preIndex+j], cmInputSum[i])
		}

		if wit.oneOfManyWitness[i] == nil {
			wit.oneOfManyWitness[i] = new(oneoutofmany.OneOutOfManyWitness)
		}
		indexIsZero := wit.myCommitmentIndices[i] % privacy_util.CommitmentRingSize

		wit.oneOfManyWitness[i].Set(commitmentTemps[i], randInputIsZero[i], indexIsZero)
		preIndex = privacy_util.CommitmentRingSize * (i + 1)
		// ---------------------------------------------------

		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		if wit.serialNumberWitness[i] == nil {
			wit.serialNumberWitness[i] = new(serialnumberprivacy.SNPrivacyWitness)
		}
		stmt := new(serialnumberprivacy.SerialNumberPrivacyStatement)
		stmt.Set(inputCoin.GetKeyImage(), cmInputSK, wit.comInputSerialNumberDerivator[i])

		wit.serialNumberWitness[i].Set(stmt, wit.privateKey, randInputSK, snDerivator, randInputSND[i])
		// ---------------------------------------------------
	}

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init cmOutputSND")
	cmOutputSND := make([]*operation.Point, numOutputCoin)
	cmOutputSum := make([]*operation.Point, numOutputCoin)
	cmOutputValue := make([]*operation.Point, numOutputCoin)
	randOutputSum := make([]*operation.Scalar, numOutputCoin)
	randOutputSND := make([]*operation.Scalar, numOutputCoin)
	randOutputValue := make([]*operation.Scalar, numOutputCoin)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init cmOutputSND")
	// cmOutputValueAll is sum of all value coin commitments
	cmOutputSumAll := new(operation.Point).Identity()
	cmOutputValueAll := new(operation.Point).Identity()
	randOutputValueAll := new(operation.Scalar).FromUint64(0)

	cmOutputShardID := make([]*operation.Point, numOutputCoin)
	randOutputShardID := make([]*operation.Scalar, numOutputCoin)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init outputCoins")
	outputCoins := wit.outputCoins
	for i, outputCoin := range outputCoins {
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init outputCoin: %v", outputCoin)
		if i == len(outputCoins)-1 {
			randOutputValue[i] = new(operation.Scalar).Sub(randInputValueAll, randOutputValueAll)
		} else {
			randOutputValue[i] = operation.RandomScalar()
		}

		randOutputSND[i] = operation.RandomScalar()
		randOutputShardID[i] = operation.RandomScalar()

		cmOutputValue[i] = operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(outputCoin.CoinDetails.GetValue()), randOutputValue[i], operation.PedersenValueIndex)
		cmOutputSND[i] = operation.PedCom.CommitAtIndex(outputCoin.CoinDetails.GetSNDerivator(), randOutputSND[i], operation.PedersenSndIndex)

		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init outputCoins[i].GetShardID()")
		receiverShardID, err := outputCoins[i].GetShardID()
		if err != nil {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init err: %v", err)
			return errhandler.NewPrivacyErr(errhandler.UnexpectedErr, errors.New("cannot parse shardID of outputCoins"))
		}
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init receiverShardID: %v", receiverShardID)
		cmOutputShardID[i] = operation.PedCom.CommitAtIndex(new(operation.Scalar).FromUint64(uint64(receiverShardID)), randOutputShardID[i], operation.PedersenShardIDIndex)

		randOutputSum[i] = new(operation.Scalar).FromUint64(0)
		randOutputSum[i].Add(randOutputValue[i], randOutputSND[i])
		randOutputSum[i].Add(randOutputSum[i], randOutputShardID[i])

		cmOutputSum[i] = new(operation.Point).Identity()
		cmOutputSum[i].Add(cmOutputValue[i], cmOutputSND[i])
		cmOutputSum[i].Add(cmOutputSum[i], outputCoins[i].GetPublicKey())
		cmOutputSum[i].Add(cmOutputSum[i], cmOutputShardID[i])

		cmOutputValueAll.Add(cmOutputValueAll, cmOutputValue[i])
		randOutputValueAll.Add(randOutputValueAll, randOutputValue[i])

		// calculate final commitment for output coins
		outputCoins[i].CoinDetails.SetCommitment(cmOutputSum[i])
		outputCoins[i].CoinDetails.SetRandomness(randOutputSum[i])

		cmOutputSumAll.Add(cmOutputSumAll, cmOutputSum[i])
	}

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init numOutputCoin: %v", numOutputCoin)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init outputValue")
	// For Multi Range Protocol
	// proving each output value is less than vmax
	// proving sum of output values is less than vmax
	outputValue := make([]uint64, numOutputCoin)
	for i := 0; i < numOutputCoin; i++ {
		if outputCoins[i].CoinDetails.GetValue() > 0 {
			outputValue[i] = outputCoins[i].CoinDetails.GetValue()
		} else {
			return errhandler.NewPrivacyErr(errhandler.UnexpectedErr, errors.New("output coin's value is less than 0"))
		}
	}
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init wit.aggregatedRangeWitness")
	if wit.aggregatedRangeWitness == nil {
		wit.aggregatedRangeWitness = new(aggregatedrange.AggregatedRangeWitness)
	}
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init wit.aggregatedRangeWitness.Set")
	wit.aggregatedRangeWitness.Set(outputValue, randOutputValue)
	// ---------------------------------------------------

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init save partial commitments (value, input, shardID)")
	// save partial commitments (value, input, shardID)
	wit.comOutputValue = cmOutputValue
	wit.comOutputSerialNumberDerivator = cmOutputSND
	wit.comOutputShardID = cmOutputShardID

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Init done")
	return nil
}

// Prove creates big proof
func (wit *PaymentWitness) Prove(hasPrivacy bool, paymentInfo []*key.PaymentInfo) (*PaymentProof, *errhandler.PrivacyError) {
	defer func() {
		if r := recover(); r != nil {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove recover: %v", r)
		}
	}()

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove hasPrivacy: %v", hasPrivacy)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove paymentInfo: %v", paymentInfo)
	proof := new(PaymentProof)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof: %v", proof)
	proof.Init()

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.inputCoins: %v", wit.inputCoins)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.outputCoins: %v", wit.outputCoins)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.comOutputValue: %v", wit.comOutputValue)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.comOutputSerialNumberDerivator: %v", wit.comOutputSerialNumberDerivator)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.comOutputShardID: %v", wit.comOutputShardID)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.comInputSecretKey: %v", wit.comInputSecretKey)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.comInputValue: %v", wit.comInputValue)

	proof.inputCoins = wit.inputCoins
	proof.outputCoins = wit.outputCoins
	proof.commitmentOutputValue = wit.comOutputValue
	proof.commitmentOutputSND = wit.comOutputSerialNumberDerivator
	proof.commitmentOutputShardID = wit.comOutputShardID

	proof.commitmentInputSecretKey = wit.comInputSecretKey
	proof.commitmentInputValue = wit.comInputValue
	proof.commitmentInputSND = wit.comInputSerialNumberDerivator
	proof.commitmentInputShardID = wit.comInputShardID
	proof.commitmentIndices = wit.commitmentIndices

	// if hasPrivacy == false, don't need to create the zero knowledge proof
	// proving user has spending key corresponding with public key in input coins
	// is proved by signing with spending key
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove !hasPrivacy: %v", !hasPrivacy)
	if !hasPrivacy {
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove !hasPrivacy: %v", !hasPrivacy)
		// Proving that serial number is derived from the committed derivator
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.serialNumberNoPrivacyWitness: %v", wit.serialNumberNoPrivacyWitness)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.serialNumberNoPrivacyProof: %v", proof.serialNumberNoPrivacyProof)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.inputCoins: %v", proof.inputCoins)
		for i := 0; i < len(wit.inputCoins); i++ {
			snNoPrivacyProof, err := wit.serialNumberNoPrivacyWitness[i].Prove(nil)
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove snNoPrivacyProof: %v", snNoPrivacyProof)
			if err != nil {
				logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove err: %v", err)
				return nil, errhandler.NewPrivacyErr(errhandler.ProveSerialNumberNoPrivacyErr, err)
			}
			proof.serialNumberNoPrivacyProof = append(proof.serialNumberNoPrivacyProof, snNoPrivacyProof)
		}
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.serialNumberNoPrivacyProof: %v", proof.serialNumberNoPrivacyProof)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.outputCoins: %v", proof.outputCoins)
		for i := 0; i < len(proof.outputCoins); i++ {
			proof.outputCoins[i].CoinDetails.SetKeyImage(nil)
		}
		return proof, nil
	}

	// if hasPrivacy == true
	numInputCoins := len(wit.oneOfManyWitness)

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove numInputCoins: %v", numInputCoins)
	for i := 0; i < numInputCoins; i++ {
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.oneOfManyWitness[i]: %v", wit.oneOfManyWitness[i])
		// Proving one-out-of-N commitments is a commitment to the coins being spent
		oneOfManyProof, err := wit.oneOfManyWitness[i].Prove()
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove oneOfManyProof: %v", oneOfManyProof)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove err: %v", err)
		if err != nil {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove err: %v", err)
			return nil, errhandler.NewPrivacyErr(errhandler.ProveOneOutOfManyErr, err)
		}
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.oneOfManyProof before append: %v", proof.oneOfManyProof)
		proof.oneOfManyProof = append(proof.oneOfManyProof, oneOfManyProof)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.oneOfManyProof after append: %v", proof.oneOfManyProof)

		// Proving that serial number is derived from the committed derivator
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.serialNumberWitness[i].Prove(nil)")
		serialNumberProof, err := wit.serialNumberWitness[i].Prove(nil)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove serialNumberProof: %v", serialNumberProof)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove err: %v", err)
		if err != nil {
			return nil, errhandler.NewPrivacyErr(errhandler.ProveSerialNumberPrivacyErr, err)
		}
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.serialNumberProof before append: %v", proof.serialNumberProof)
		proof.serialNumberProof = append(proof.serialNumberProof, serialNumberProof)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.serialNumberProof after append: %v", proof.serialNumberProof)
	}
	var err error

	// Proving that each output values and sum of them does not exceed v_max
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove wit.aggregatedRangeWitness.Prove")
	proof.aggregatedRangeProof, err = wit.aggregatedRangeWitness.Prove()
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.aggregatedRangeProof: %v", proof.aggregatedRangeProof)
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove err: %v", err)
	if err != nil {
		return nil, errhandler.NewPrivacyErr(errhandler.ProveAggregatedRangeErr, err)
	}

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove len(proof.inputCoins): %v", len(proof.inputCoins))
	if len(proof.inputCoins) == 0 {
		proof.commitmentIndices = nil
		proof.commitmentInputSecretKey = nil
		proof.commitmentInputShardID = nil
		proof.commitmentInputSND = nil
		proof.commitmentInputValue = nil
	}

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove len(proof.outputCoins): %v", len(proof.outputCoins))
	if len(proof.outputCoins) == 0 {
		proof.commitmentOutputValue = nil
		proof.commitmentOutputSND = nil
		proof.commitmentOutputShardID = nil
	}

	// After Prove, we should hide all information in coin details.
	// encrypt coin details (Randomness)
	// hide information of output coins except coin commitments, public key, snDerivators
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove len(proof.outputCoins): %v", len(proof.outputCoins))
	for i := 0; i < len(proof.outputCoins); i++ {
		errEncrypt := proof.outputCoins[i].Encrypt(paymentInfo[i].PaymentAddress.Tk)
		if errEncrypt != nil {
			logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove errEncrypt: %v", errEncrypt)
			return nil, errEncrypt
		}
		proof.outputCoins[i].CoinDetails.SetKeyImage(nil)
		proof.outputCoins[i].CoinDetails.SetValue(0)
		proof.outputCoins[i].CoinDetails.SetRandomness(nil)
	}

	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove len(proof.GetInputCoins()): %v", len(proof.GetInputCoins()))
	for i := 0; i < len(proof.GetInputCoins()); i++ {
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.inputCoins[i].ConcealOutputCoin(nil)")
		proof.inputCoins[i].ConcealOutputCoin(nil)
		logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof.inputCoins[i].ConcealOutputCoin(nil) done")
	}
	logUtils.LogPrintf("paymentwitness_v1.PaymentWitness.Prove proof: %v", proof)
	return proof, nil
}
