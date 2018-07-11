package BLC

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

type Wallet struct {
	//私钥 椭圆曲线加密库
	ZQ_PrivateKey ecdsa.PrivateKey
	//公钥
	ZQ_PublicKey []byte
}

const version = byte(0x00)
const addressChecksumLen = 4

//创建钱包
func ZQ_NewWallet() *Wallet {

	privateKey, publicKey := ZQ_newKeyPair()

	return &Wallet{privateKey, publicKey}
}

//获取地址
func (wallet *Wallet) ZQ_GetAddress() []byte {

	ripemd160 := ZQ_Ripemd160Hash(wallet.ZQ_PublicKey)

	version_ripemd160 := append([]byte{version}, ripemd160...)

	checkSum := ZQ_CheckSum(version_ripemd160)

	version_ripemd160_checksum := append(version_ripemd160, checkSum...)

	return ZQ_Base58Encode(version_ripemd160_checksum)
}

//验证地址有效性
func ZQ_IsValidForAddress(address []byte) bool {

	version_ripemd160_checksum := ZQ_Base58Decode(address)

	version_ripemd160 := version_ripemd160_checksum[:len(version_ripemd160_checksum)-addressChecksumLen]
	checksum := version_ripemd160_checksum[len(version_ripemd160_checksum)-addressChecksumLen:]

	checkRes := ZQ_CheckSum(version_ripemd160)

	if bytes.Compare(checksum, checkRes) == 0 {
		return true
	}

	return false
}

func ZQ_Ripemd160Hash(publicKey []byte) []byte {
	// 256Hash
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)

	// 160Hash
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)

	return ripemd160.Sum(nil)
}

//通过私钥产生公钥
func ZQ_newKeyPair() (ecdsa.PrivateKey, []byte) {

	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)

	if err != nil {
		log.Panic()
	}

	//fmt.Println(private.D)

	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, public
}

func ZQ_CheckSum(payload []byte) []byte {

	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])

	return hash2[:addressChecksumLen]
}
