package tron

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

func hexToBase58(hexStr string) string {
	addrBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return ""
	}

	hash1 := s256(addrBytes)
	hash2 := s256(hash1)
	checksum := hash2[:4]

	fullAddress := append(addrBytes, checksum...)
	base58Address := base58.Encode(fullAddress)

	return base58Address
}
func decodeContractData(data string) (methodID string, address string, value *big.Int) {
	bytes, _ := hex.DecodeString(data)
	if len(bytes) < 36 {
		return
	}
	methodID = hex.EncodeToString(bytes[:4])
	addrBytes := bytes[4:36]
	address = hex.EncodeToString(addrBytes[12:])
	value = new(big.Int)
	value.SetBytes(bytes[36:])
	return
}
func s256(s []byte) []byte {
	h := sha256.New()
	h.Write(s)
	bs := h.Sum(nil)
	return bs
}

func base58ToHex(address string) string {
	//convert base58 to hex
	decodedAddress := base58.Decode(address)
	dst := make([]byte, hex.EncodedLen(len(decodedAddress)))
	hex.Encode(dst, decodedAddress)
	dst = dst[:len(dst)-8]
	return string(dst)
}

func privateKeyToAddress(privateKeyString string) (string, error) {
	// Decode hex string to byte slice
	privateKeyBytes, err := hexutil.Decode("0x" + privateKeyString)
	if err != nil {
		return "", errors.New("hex_decode_err")
	}
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return "", errors.New("ecdsa_err")
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("public_err")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	address = "41" + address[2:]
	addb, _ := hex.DecodeString(address)
	hash1 := s256(s256(addb))
	secret := hash1[:4]
	for _, v := range secret {
		addb = append(addb, v)
	}
	addressBase58 := base58.Encode(addb)

	return addressBase58, nil
}
