package wallet

import (
	"bytes"
	"crypto/sha256"
	"math/big"

	"github.com/levietcuong2602/incognito-chain/common"

	"github.com/levietcuong2602/incognito-chain/common/base58"
)

// padByteSlice returns a byte slice of the given size with contents of the
// given slice left padded and any empty spaces filled with 0's.
func padByteSlice(slice []byte, length int) []byte {
	offset := length - len(slice)
	if offset <= 0 {
		return slice
	}
	newSlice := make([]byte, length)
	copy(newSlice[offset:], slice)
	return newSlice
}

// compareByteSlices returns true of the byte slices have equal contents and
// returns false otherwise.
func compareByteSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// computeChecksum returns hashing of data using SHA256
func computeChecksum(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// Appends to data the first (len(data) / 32)bits of the result of sha256(data)
// Currently only supports data up to 32 bytes
func addChecksum(data []byte) []byte {
	// Get first byte of sha256
	hash := computeChecksum(data)
	firstChecksumByte := hash[0]

	// len() is in bytes so we divide by 4
	checksumBitLength := uint(len(data) / 4)

	// For each bit of check sum we want we shift the data one the left
	// and then set the (new) right most bit equal to checksum bit at that index
	// staring from the left
	dataBigInt := new(big.Int).SetBytes(data)
	for i := uint(0); i < checksumBitLength; i++ {
		// Bitshift 1 left
		dataBigInt.Mul(dataBigInt, bigTwo)

		// Set rightmost bit if leftmost checksum bit is set
		if uint8(firstChecksumByte&(1<<(7-i))) > 0 {
			dataBigInt.Or(dataBigInt, bigOne)
		}
	}

	return dataBigInt.Bytes()
}

// GetBurningPublicKey returns the public key of the burning address.
func GetBurningPublicKey() []byte {
	// get burning address
	w, err := Base58CheckDeserialize(common.BurningAddress2)
	if err != nil {
		return nil
	}

	return w.KeySet.PaymentAddress.Pk
}

func IsPublicKeyBurningAddress(publicKey []byte) bool {
	// get burning address
	keyWalletBurningAdd1, err := Base58CheckDeserialize(common.BurningAddress)
	if err != nil {
		return false
	}
	if bytes.Equal(publicKey, keyWalletBurningAdd1.KeySet.PaymentAddress.Pk) {
		return true
	}
	keyWalletBurningAdd2, err := Base58CheckDeserialize(common.BurningAddress2)
	if err != nil {
		return false
	}
	if bytes.Equal(publicKey, keyWalletBurningAdd2.KeySet.PaymentAddress.Pk) {
		return true
	}

	return false
}

func InitPublicKeyBurningAddressByte() error {
	keyWalletBurningAdd1, err := Base58CheckDeserialize(common.BurningAddress)
	if err != nil {
		return err
	}
	common.BurningAddressByte = keyWalletBurningAdd1.KeySet.PaymentAddress.Pk

	keyWalletBurningAdd2, err := Base58CheckDeserialize(common.BurningAddress2)
	if err != nil {
		return err
	}
	common.BurningAddressByte2 = keyWalletBurningAdd2.KeySet.PaymentAddress.Pk
	return nil
}

func GetPublicKeysFromPaymentAddresses(payments []string) []string {
	res := []string{}
	for _, paymentAddressStr := range payments {
		keyWallet, err := Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return res
		}
		if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
			return res
		}
		pkStr := base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, common.Base58Version)
		res = append(res, pkStr)
	}
	return res
}
