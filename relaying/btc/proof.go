package btcrelaying

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

const BTCBlockConfirmations = 6

type MerkleProof struct {
	ProofHash *chainhash.Hash
	IsLeft    bool
}

type BTCProof struct {
	MerkleProofs []*MerkleProof
	BTCTx        *wire.MsgTx
	BlockHash    *chainhash.Hash
}

func ParseAndValidateSanityBTCProofFromB64EncodeStr(b64EncodedStr string) (*BTCProof, error) {
	jsonBytes, err := base64.StdEncoding.DecodeString(b64EncodedStr)
	if err != nil {
		return nil, err
	}
	var proof BTCProof
	err = json.Unmarshal(jsonBytes, &proof)
	if err != nil {
		return nil, err
	}
	isValid, err := ValidateSanityBTCProof(&proof)
	if !isValid && err != nil {
		return nil, err
	}
	return &proof, nil
}

func ValidateSanityBTCProof(btcProof *BTCProof) (bool, error) {
	if btcProof == nil {
		return false, errors.New("ValidateSanityBTCProof btcProof is nil")
	}
	if len(btcProof.MerkleProofs) == 0 {
		return false, errors.New("ValidateSanityBTCProof MerkleProofs is empty")
	}
	for _, mp := range btcProof.MerkleProofs {
		if mp.ProofHash == nil {
			return false, errors.New("ValidateSanityBTCProof ProofHash is nil")
		}
	}
	if btcProof.BTCTx == nil {
		return false, errors.New("ValidateSanityBTCProof BTCTx is nil")
	}
	if btcProof.BlockHash == nil {
		return false, errors.New("ValidateSanityBTCProof BlockHash is nil")
	}
	return true, nil
}

func buildMerkleTreeStoreFromTxHashes(txHashes []*chainhash.Hash) []*chainhash.Hash {
	nextPoT := nextPowerOfTwo(len(txHashes))
	arraySize := nextPoT*2 - 1
	merkles := make([]*chainhash.Hash, arraySize)

	for i, txHash := range txHashes {
		merkles[i] = txHash
	}

	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		case merkles[i] == nil:
			merkles[offset] = nil

		case merkles[i+1] == nil:
			newHash := HashMerkleBranches(merkles[i], merkles[i])
			merkles[offset] = newHash

		default:
			newHash := HashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}

func buildMerkleProof(txHashes []*chainhash.Hash, targetedTxHash *chainhash.Hash) []*MerkleProof {
	merkleTree := buildMerkleTreeStoreFromTxHashes(txHashes)
	nextPoT := nextPowerOfTwo(len(txHashes))
	layers := [][]*chainhash.Hash{}
	left := 0
	right := nextPoT
	for left < right {
		layers = append(layers, merkleTree[left:right])
		curLen := len(merkleTree[left:right])
		left = right
		right = right + curLen/2
	}

	merkleProofs := []*MerkleProof{}
	curHash := targetedTxHash
	for _, layer := range layers {
		if len(layer) == 1 {
			break
		}

		for i := 0; i < len(layer); i++ {
			if layer[i] == nil || layer[i].String() != curHash.String() {
				continue
			}
			if i%2 == 0 {
				if layer[i+1] == nil {
					curHash = HashMerkleBranches(layer[i], layer[i])
					merkleProofs = append(
						merkleProofs,
						&MerkleProof{
							ProofHash: layer[i],
							IsLeft:    false,
						},
					)
				} else {
					curHash = HashMerkleBranches(layer[i], layer[i+1])
					merkleProofs = append(
						merkleProofs,
						&MerkleProof{
							ProofHash: layer[i+1],
							IsLeft:    false,
						},
					)
				}
			} else {
				if layer[i-1] == nil {
					curHash = HashMerkleBranches(layer[i], layer[i])
					merkleProofs = append(
						merkleProofs,
						&MerkleProof{
							ProofHash: layer[i],
							IsLeft:    true,
						},
					)
				} else {
					curHash = HashMerkleBranches(layer[i-1], layer[i])
					merkleProofs = append(
						merkleProofs,
						&MerkleProof{
							ProofHash: layer[i-1],
							IsLeft:    true,
						},
					)
				}
			}
			break // process next layer
		}
	}
	return merkleProofs
}

// verify verifies that a tx is present in a block or not
func verify(
	merkleRoot *chainhash.Hash,
	merkleProofs []*MerkleProof,
	txHash *chainhash.Hash,
) bool {
	curHash := txHash
	for _, mklProof := range merkleProofs {
		if mklProof.IsLeft {
			curHash = HashMerkleBranches(mklProof.ProofHash, curHash)
		} else {
			curHash = HashMerkleBranches(curHash, mklProof.ProofHash)
		}
	}
	return curHash.String() == merkleRoot.String()
}

func (btcChain *BlockChain) VerifyTxWithMerkleProofs(
	btcProof *BTCProof,
) (bool, error) {
	btcBlock, err := btcChain.BlockByHash(btcProof.BlockHash)
	if err != nil {
		Logger.log.Errorf("Failed to get BTC block by hash %s - with error: %v\n", btcProof.BlockHash.String(), err)
		return false, err
	}

	bestState := btcChain.BestSnapshot()
	if bestState == nil || btcBlock == nil {
		Logger.log.Errorf("Both BTC best state and BTC block by hash (%s) should not be null, "+
			"but best state: %+v; block: %+v\n", btcProof.BlockHash.String(), bestState, btcBlock)
		return false, nil
	}
	if bestState.Height < btcBlock.Height()+BTCBlockConfirmations-1 {
		Logger.log.Errorf("Need to wait for %d btc block confirmations, best state height: %d, "+
			"targeting block height: %d\n", BTCBlockConfirmations, bestState.Height,
			btcBlock.Height()+BTCBlockConfirmations-1)
		return false, nil
	}
	merkleRoot := btcBlock.MsgBlock().Header.MerkleRoot
	txHash := btcProof.BTCTx.TxHash()
	Logger.log.Infof("VerifyTxWithMerkleProofs info - merkle root (%s)\n", merkleRoot.String())
	Logger.log.Infof("VerifyTxWithMerkleProofs info - btcProof (%+v)\n", btcProof)
	Logger.log.Infof("VerifyTxWithMerkleProofs info - txHash (%s)\n", txHash.String())
	return verify(&merkleRoot, btcProof.MerkleProofs, &txHash), nil
}

func ExtractAttachedMsgFromTx(msgTx *wire.MsgTx) (string, error) {
	opReturnPrefix := []byte{
		txscript.OP_RETURN,
	}
	for _, txOut := range msgTx.TxOut {
		if txOut.Value != 0 || !bytes.HasPrefix(txOut.PkScript, opReturnPrefix) {
			continue
		}
		opReturnPkScript := txOut.PkScript
		if len(opReturnPkScript) < 5 {
			return "", fmt.Errorf("Memo is invalid")
		}
		first_byte := opReturnPkScript[1]
		if first_byte <= 75 {
			return string(opReturnPkScript[2:]), nil
		} else if first_byte == 76 { //0x4c
			return string(opReturnPkScript[3:]), nil
		} else if first_byte == 77 { //0x4d
			return string(opReturnPkScript[4:]), nil
		}
	}
	return "", nil
}

// ExtractPaymentAddrStrFromPkScript extracts payment address string from pkscript
func (b *BlockChain) ExtractPaymentAddrStrFromPkScript(pkScript []byte) (string, error) {
	chainParams := b.GetChainParams()
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(pkScript, chainParams)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", nil
	}
	return addrs[0].EncodeAddress(), nil
}

// IsBTCAddressValid checks whether the passed btc address string is valid or not
func (btcChain *BlockChain) IsBTCAddressValid(addrStr string) bool {
	params := btcChain.GetChainParams()
	btcAddress, err := btcutil.DecodeAddress(addrStr, params)
	if err != nil {
		Logger.log.Warnf("IsBTCAddressValid - Failed to decode btc address with error: %v\n", err)
		return false
	}
	if !btcAddress.IsForNet(params) {
		Logger.log.Warn("IsBTCAddressValid - Failed to check valid network.")
		return false
	}

	return true
}
