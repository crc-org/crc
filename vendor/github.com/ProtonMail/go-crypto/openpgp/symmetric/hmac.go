package symmetric

import (
	"crypto"
	"crypto/hmac"
	"crypto/sha256"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/internal/algorithm"
)

type HMACPublicKey struct {
	Hash algorithm.Hash
	BindingHash [32]byte
	// While this is a "public" key, the symmetric key needs to be present here.
	// Symmetric cryptographic operations use the same key material for
	// signing and verifying, and go-crypto assumes that a public key type will
	// be used for verification. Thus, this `Key` field must never be exported
	// publicly.
	Key []byte
}

type HMACPrivateKey struct {
	PublicKey HMACPublicKey
	HashSeed [32]byte
	Key []byte
}

func HMACGenerateKey(rand io.Reader, hash algorithm.Hash) (priv *HMACPrivateKey, err error) {
	priv, err = generatePrivatePartHMAC(rand, hash)
	if err != nil {
		return
	}

	priv.generatePublicPartHMAC(hash)
	return
}

func generatePrivatePartHMAC(rand io.Reader, hash algorithm.Hash) (priv *HMACPrivateKey, err error) {
	priv = new(HMACPrivateKey)
	var seed [32] byte
	_, err = rand.Read(seed[:])
	if err != nil {
		return
	}

	key := make([]byte, hash.Size())
	_, err = rand.Read(key)
	if err != nil {
		return
	}

	priv.HashSeed = seed
	priv.Key = key
	return
}

func (priv *HMACPrivateKey) generatePublicPartHMAC(hash algorithm.Hash) (err error) {
	priv.PublicKey.Hash = hash

	bindingHash := ComputeBindingHash(priv.HashSeed)
	copy(priv.PublicKey.BindingHash[:], bindingHash)

	priv.PublicKey.Key = make([]byte, len(priv.Key))
	copy(priv.PublicKey.Key, priv.Key)
	return
}

func ComputeBindingHash(seed [32]byte) []byte {
	bindingHash := sha256.New()
	bindingHash.Write(seed[:])

	return bindingHash.Sum(nil)
}

func (priv *HMACPrivateKey) Public() crypto.PublicKey {
	return &priv.PublicKey
}

func (priv *HMACPrivateKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	expectedMAC, err := calculateMAC(priv.PublicKey.Hash, priv.Key, digest)
	if err != nil {
		return
	}
	signature = make([]byte, len(expectedMAC))
	copy(signature, expectedMAC)
	return
}

func (pub *HMACPublicKey) Verify(digest []byte, signature []byte) (bool, error) {
	expectedMAC, err := calculateMAC(pub.Hash, pub.Key, digest)
	if err != nil {
		return false, err
	}
	return hmac.Equal(expectedMAC, signature), nil
}

func calculateMAC(hash algorithm.Hash, key []byte, data []byte) ([]byte, error) {
	hashFunc := hash.HashFunc()
	if !hashFunc.Available() {
		return nil, errors.UnsupportedError("hash function")
	}

	mac := hmac.New(hashFunc.New, key)
	mac.Write(data)

	return mac.Sum(nil), nil
}
