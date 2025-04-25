//nolint:revive // skip linter for this package name
package privacy_v2

import (
	"bytes"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	numOfLoops = 3
	minOutCoinCount = 1
	maxOutCoinCount = 12

)

var _ = func() (_ struct{}) {
	fmt.Println("This runs before init() starting payment v2 logger for test !")
	Logger.Init(common.NewBackend(nil).Logger("test", true))
	common.MaxShardNumber = 1
	return
}()


func TestPaymentV2InitAndMarshalling(t *testing.T) {
	for loop:=0;loop<=numOfLoops;loop++{
		outCoinCount := common.RandInt() % (maxOutCoinCount-minOutCoinCount+1) + minOutCoinCount
		// make some dummy private keys for our dummy users
		dummyPrivateKeys := make([]*operation.Scalar,outCoinCount)
		for i,_ := range dummyPrivateKeys{
			dummyPrivateKeys[i] = operation.RandomScalar()
		}
		// each of these dummy users are provided a (not confirmed by blockchain) coinv2 of value 3000
		// paymentAdress is persistent and held by this user, while the OTA is inside the coin
		paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i, pk := range dummyPrivateKeys {
			pkb := pk.ToBytes()
			paymentInfo[i] = key.InitPaymentInfo(key.GeneratePaymentAddress(pkb[:]),3000,[]byte{})
		}
		inputCoins := make([]coin.PlainCoin, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			inputCoins[i],err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
		}
		// in this test, each user will send themselves 2000 and the rest is txfee
		for _,pInf := range paymentInfo{
			pInf.Amount = 2000
		}
		outputCoins := make([]*coin.CoinV2, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			outputCoins[i],err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
		}
	// prove and verify without privacy (no bulletproof)
	// also marshal to byte and back
		proof, err := Prove(inputCoins, outputCoins, nil, false, paymentInfo)
		assert.Equal(t, nil, err)
		b := proof.Bytes()

		temp := new(PaymentProofV2)
		err = temp.SetBytes(b)
		b2 := temp.Bytes()
		assert.Equal(t, true, bytes.Equal(b2, b))

		// correct,err := proof.Verify(false, nil, uint64(1000*outCoinCount), byte(0), nil, false, nil)
		// assert.Equal(t, nil, err)
		// assert.Equal(t,true,correct)
	}
}

func TestPaymentV2ProveWithPrivacy(t *testing.T) {
	outCoinCount := common.RandInt() % (maxOutCoinCount-minOutCoinCount+1) + minOutCoinCount
	for loop:=0;loop<numOfLoops;loop++{
		// make some dummy private keys for our dummy users
		dummyPrivateKeys := make([]*key.PrivateKey,outCoinCount)
		for i := 0; i < outCoinCount; i += 1 {
			privateKey := key.GeneratePrivateKey(common.RandBytes(32))
			dummyPrivateKeys[i] = &privateKey
		}
		// each of these dummy users are provided a (not confirmed by blockchain) coinv2 of value 3000
		// paymentAdress is persistent and held by this user, while the OTA is inside the coin
		paymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		keySets := make([]*incognitokey.KeySet,len(dummyPrivateKeys))
		for i, _ := range dummyPrivateKeys {
			keySets[i] = new(incognitokey.KeySet)
			err := keySets[i].InitFromPrivateKey(dummyPrivateKeys[i])
			assert.Equal(t, nil, err)

			paymentInfo[i] = key.InitPaymentInfo(keySets[i].PaymentAddress,3000,[]byte{})
		}
		inputCoins := make([]coin.PlainCoin, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			inputCoins[i],err = coin.NewCoinFromPaymentInfo(paymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
			ic_specific,ok := inputCoins[i].(*coin.CoinV2)
			assert.Equal(t, true, ok)
			ic_specific.ConcealOutputCoin(keySets[i].PaymentAddress.GetPublicView())
			ic_specific.Decrypt(keySets[i])
		}
		// in this test, each user will send some other rando 2500 and the rest is txfee
		outPaymentInfo := make([]*key.PaymentInfo, len(dummyPrivateKeys))
		for i, _ := range dummyPrivateKeys {
			otherPriv := operation.RandomScalar()
			pkb := otherPriv.ToBytes()
			outPaymentInfo[i] = key.InitPaymentInfo(key.GeneratePaymentAddress(pkb[:]),2500,[]byte{})
		}
		outputCoins := make([]*coin.CoinV2, outCoinCount)
		for i:=0;i<outCoinCount;i++ {
			var err error
			outputCoins[i],err = coin.NewCoinFromPaymentInfo(outPaymentInfo[i])
			if err!=nil{
				fmt.Println(err)
			}
		}
		boolParams := make(map[string]bool)
		boolParams["isNewTransaction"] = true
		boolParams["hasConfidentialAsset"] = false
		// prove and verify with privacy using bulletproof
		// note that bulletproofs only assure each outcoin amount is in uint64 range
		// while the equality suminput = suminput + sumfee must be checked using mlsag later
		// here our mock scenario has out+fee>in but passes anyway
		proof, err := Prove(inputCoins, outputCoins, nil, false, paymentInfo)
		assert.Equal(t, nil, err)
		isSane, err := proof.ValidateSanity(nil)
		assert.Equal(t,nil,err)
		assert.Equal(t,true,isSane)

		isValid,err := proof.Verify(boolParams, nil, uint64(200*outCoinCount), byte(0), nil, nil)
		assert.Equal(t, nil, err)
		assert.Equal(t,true,isValid)

		pBytes := proof.Bytes()
		// try `corrupting` one byte in the proof
		for i:=0; i<10; i++{ 
			b := make([]byte, len(pBytes))
			copy(b, pBytes)
			corruptedIndex := common.RandInt() % len(b)
			// random in 1..255 (not zero)
			diff := common.RandInt() % 255 + 1
			b[corruptedIndex] = byte((int(b[corruptedIndex]) + diff) % 256)
			reconstructedProof := new(PaymentProofV2)
			// it's a corrupted proof so it must fail one of these 3 checks
			err = reconstructedProof.SetBytes(b)
			if err != nil{
				continue
			}
			isSane, err = reconstructedProof.ValidateSanity(nil)
			if !isSane{
				continue
			}
			isValid,err = reconstructedProof.Verify(boolParams, nil, uint64(200*outCoinCount), byte(0), nil, nil)
			if !isValid{
				continue
			}
			fmt.Printf("Corrupted proof %v",reconstructedProof)
			assert.Equal(t,false,isValid)
		}
		// try completely made up proof
		for i:=0; i<10; i++{
			// length from 0..299
			randomLength := common.RandInt() % 300
			// 10% of the time the stupid proof is very long
			if i==0{
				randomLength += 3000000
			}
			bs := common.RandBytes(randomLength)
			reconstructedProof := new(PaymentProofV2)
			err = reconstructedProof.SetBytes(bs)
			// it's a bs proof so it must fail one of these 3 checks
			if err != nil{
				continue
			}
			isSane, err = reconstructedProof.ValidateSanity(nil)
			if !isSane{
				continue
			}
			isValid,err = reconstructedProof.Verify(boolParams, nil, uint64(200*outCoinCount), byte(0), nil, nil)
			if !isValid{
				continue
			}
			fmt.Printf("Forged proof %v",reconstructedProof)
			assert.Equal(t,false,isValid)
		}
	}
}
