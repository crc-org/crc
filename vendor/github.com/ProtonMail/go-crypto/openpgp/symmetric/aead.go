package symmetric

import (
	"github.com/ProtonMail/go-crypto/openpgp/internal/algorithm"
	"io"
)

type AEADPublicKey struct {
	Cipher algorithm.CipherFunction
	BindingHash [32]byte
	Key []byte
}

type AEADPrivateKey struct {
	PublicKey AEADPublicKey
	HashSeed [32]byte
	Key []byte
}

func AEADGenerateKey(rand io.Reader, cipher algorithm.CipherFunction) (priv *AEADPrivateKey, err error) {
	priv, err = generatePrivatePartAEAD(rand, cipher)
	if err != nil {
		return
	}

	priv.generatePublicPartAEAD(cipher)
	return
}

func generatePrivatePartAEAD(rand io.Reader, cipher algorithm.CipherFunction) (priv *AEADPrivateKey, err error) {
	priv = new(AEADPrivateKey)
	var seed [32] byte
	_, err = rand.Read(seed[:])
	if err != nil {
		return
	}

	key := make([]byte, cipher.KeySize())
	_, err = rand.Read(key)
	if err != nil {
		return
	}

	priv.HashSeed = seed
	priv.Key = key
	return
}

func (priv *AEADPrivateKey) generatePublicPartAEAD(cipher algorithm.CipherFunction) (err error) {
	priv.PublicKey.Cipher = cipher

	bindingHash := ComputeBindingHash(priv.HashSeed)

	priv.PublicKey.Key = make([]byte, len(priv.Key))
	copy(priv.PublicKey.Key, priv.Key)
	copy(priv.PublicKey.BindingHash[:], bindingHash)
	return
}

func (pub *AEADPublicKey) Encrypt(rand io.Reader, data []byte, mode algorithm.AEADMode) (nonce []byte, ciphertext []byte, err error) {
	block := pub.Cipher.New(pub.Key)
	aead := mode.New(block)
	nonce = make([]byte, aead.NonceSize())
	rand.Read(nonce)
	ciphertext = aead.Seal(nil, nonce, data, nil)
	return
}

func (priv *AEADPrivateKey) Decrypt(nonce []byte, ciphertext []byte, mode algorithm.AEADMode) (message []byte, err error) {

	block := priv.PublicKey.Cipher.New(priv.Key)
	aead := mode.New(block)
	message, err = aead.Open(nil, nonce, ciphertext, nil)
	return
}

