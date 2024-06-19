// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ecdh implements ECDH encryption, suitable for OpenPGP,
// as specified in RFC 6637, section 8.
package ecdh

import (
	"bytes"
	"errors"
	"github.com/ProtonMail/go-crypto/openpgp/aes/keywrap"
	pgperrors "github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/internal/algorithm"
	"github.com/ProtonMail/go-crypto/openpgp/internal/ecc"
	"github.com/ProtonMail/go-crypto/openpgp/internal/ecc/curve25519"
	"io"
)

const (
	KDFVersion1 = 1
	KDFVersionForwarding = 255
)

type KDF struct {
	Version                int // Defaults to v1; 255 for forwarding
	Hash                   algorithm.Hash
	Cipher                 algorithm.Cipher
	ReplacementFingerprint []byte // (forwarding only) fingerprint to use instead of recipient's (20 octets)
}

func (kdf *KDF) Serialize(w io.Writer) (err error) {
	switch kdf.Version {
		case 0, KDFVersion1: // Default to v1 if unspecified
			return kdf.serializeForHash(w)
		case KDFVersionForwarding:
			// Length || Version || Hash || Cipher || Replacement Fingerprint
			length := byte(3 + len(kdf.ReplacementFingerprint))
			if _, err := w.Write([]byte{length, KDFVersionForwarding, kdf.Hash.Id(), kdf.Cipher.Id()}); err != nil {
				return err
			}
			if _, err := w.Write(kdf.ReplacementFingerprint); err != nil {
				return err
			}

			return nil
		default:
			return errors.New("ecdh: invalid KDF version")
	}
}

func (kdf *KDF) serializeForHash(w io.Writer) (err error) {
	// Length || Version || Hash || Cipher
	if _, err := w.Write([]byte{3, KDFVersion1, kdf.Hash.Id(), kdf.Cipher.Id()}); err != nil {
		return err
	}
	return nil
}

type PublicKey struct {
	curve ecc.ECDHCurve
	Point []byte
	KDF
}

type PrivateKey struct {
	PublicKey
	D []byte
}

func NewPublicKey(curve ecc.ECDHCurve, kdf KDF) *PublicKey {
	return &PublicKey{
		curve: curve,
		KDF:   kdf,
	}
}

func NewPrivateKey(key PublicKey) *PrivateKey {
	return &PrivateKey{
		PublicKey: key,
	}
}

func (pk *PublicKey) GetCurve() ecc.ECDHCurve {
	return pk.curve
}

func (pk *PublicKey) MarshalPoint() []byte {
	return pk.curve.MarshalBytePoint(pk.Point)
}

func (pk *PublicKey) UnmarshalPoint(p []byte) error {
	pk.Point = pk.curve.UnmarshalBytePoint(p)
	if pk.Point == nil {
		return errors.New("ecdh: failed to parse EC point")
	}
	return nil
}

func (sk *PrivateKey) MarshalByteSecret() []byte {
	return sk.curve.MarshalByteSecret(sk.D)
}

func (sk *PrivateKey) UnmarshalByteSecret(d []byte) error {
	sk.D = sk.curve.UnmarshalByteSecret(d)

	if sk.D == nil {
		return errors.New("ecdh: failed to parse scalar")
	}
	return nil
}

func GenerateKey(rand io.Reader, c ecc.ECDHCurve, kdf KDF) (priv *PrivateKey, err error) {
	priv = new(PrivateKey)
	priv.PublicKey.curve = c
	priv.PublicKey.KDF = kdf
	priv.PublicKey.Point, priv.D, err = c.GenerateECDH(rand)
	return
}

func Encrypt(random io.Reader, pub *PublicKey, msg, curveOID, fingerprint []byte) (vsG, c []byte, err error) {
	if len(msg) > 40 {
		return nil, nil, errors.New("ecdh: message too long")
	}
	// the sender MAY use 21, 13, and 5 bytes of padding for AES-128,
	// AES-192, and AES-256, respectively, to provide the same number of
	// octets, 40 total, as an input to the key wrapping method.
	padding := make([]byte, 40-len(msg))
	for i := range padding {
		padding[i] = byte(40 - len(msg))
	}
	m := append(msg, padding...)

	ephemeral, zb, err := pub.curve.Encaps(random, pub.Point)
	if err != nil {
		return nil, nil, err
	}

	vsG = pub.curve.MarshalBytePoint(ephemeral)

	z, err := buildKey(pub, zb, curveOID, fingerprint, false, false)
	if err != nil {
		return nil, nil, err
	}

	if c, err = keywrap.Wrap(z, m); err != nil {
		return nil, nil, err
	}

	return vsG, c, nil

}

func Decrypt(priv *PrivateKey, vsG, c, curveOID, fingerprint []byte) (msg []byte, err error) {
	var m []byte
	zb, err := priv.PublicKey.curve.Decaps(priv.curve.UnmarshalBytePoint(vsG), priv.D)

	// Try buildKey three times to workaround an old bug, see comments in buildKey.
	for i := 0; i < 3; i++ {
		var z []byte
		// RFC6637 §8: "Compute Z = KDF( S, Z_len, Param );"
		z, err = buildKey(&priv.PublicKey, zb, curveOID, fingerprint, i == 1, i == 2)
		if err != nil {
			return nil, err
		}

		// RFC6637 §8: "Compute C = AESKeyWrap( Z, c ) as per [RFC3394]"
		m, err = keywrap.Unwrap(z, c)
		if err == nil {
			break
		}
	}

	// Only return an error after we've tried all (required) variants of buildKey.
	if err != nil {
		return nil, err
	}

	// RFC6637 §8: "m = symm_alg_ID || session key || checksum || pkcs5_padding"
	// The last byte should be the length of the padding, as per PKCS5; strip it off.
	return m[:len(m)-int(m[len(m)-1])], nil
}

func buildKey(pub *PublicKey, zb []byte, curveOID, fingerprint []byte, stripLeading, stripTrailing bool) ([]byte, error) {
	// Param = curve_OID_len || curve_OID || public_key_alg_ID
	//         || KDF_params for AESKeyWrap
	//         || "Anonymous Sender    " || recipient_fingerprint;
	param := new(bytes.Buffer)
	if _, err := param.Write(curveOID); err != nil {
		return nil, err
	}
	algo := []byte{18}
	if _, err := param.Write(algo); err != nil {
		return nil, err
	}

	if err := pub.KDF.serializeForHash(param); err != nil {
		return nil, err
	}

	if _, err := param.Write([]byte("Anonymous Sender    ")); err != nil {
		return nil, err
	}
	if pub.KDF.ReplacementFingerprint != nil {
		fingerprint = pub.KDF.ReplacementFingerprint
	}

	if _, err := param.Write(fingerprint); err != nil {
		return nil, err
	}

	// MB = Hash ( 00 || 00 || 00 || 01 || ZB || Param );
	h := pub.KDF.Hash.New()
	if _, err := h.Write([]byte{0x0, 0x0, 0x0, 0x1}); err != nil {
		return nil, err
	}
	zbLen := len(zb)
	i := 0
	j := zbLen - 1
	if stripLeading {
		// Work around old go crypto bug where the leading zeros are missing.
		for i < zbLen && zb[i] == 0 {
			i++
		}
	}
	if stripTrailing {
		// Work around old OpenPGP.js bug where insignificant trailing zeros in
		// this little-endian number are missing.
		// (See https://github.com/openpgpjs/openpgpjs/pull/853.)
		for j >= 0 && zb[j] == 0 {
			j--
		}
	}
	if _, err := h.Write(zb[i : j+1]); err != nil {
		return nil, err
	}
	if _, err := h.Write(param.Bytes()); err != nil {
		return nil, err
	}
	mb := h.Sum(nil)

	return mb[:pub.KDF.Cipher.KeySize()], nil // return oBits leftmost bits of MB.

}

func Validate(priv *PrivateKey) error {
	return priv.curve.ValidateECDH(priv.Point, priv.D)
}

func DeriveProxyParam(recipientKey, forwardeeKey *PrivateKey) (proxyParam []byte, err error) {
	if recipientKey.GetCurve().GetCurveName() != "curve25519" {
		return nil, pgperrors.InvalidArgumentError("recipient subkey is not curve25519")
	}

	if forwardeeKey.GetCurve().GetCurveName() != "curve25519" {
		return nil, pgperrors.InvalidArgumentError("forwardee subkey is not curve25519")
	}

	c := ecc.NewCurve25519()

	// Clamp and reverse two secrets
	proxyParam, err = curve25519.DeriveProxyParam(c.MarshalByteSecret(recipientKey.D), c.MarshalByteSecret(forwardeeKey.D))

	return proxyParam, err
}

func ProxyTransform(ephemeral, proxyParam []byte) ([]byte, error) {
	c := ecc.NewCurve25519()

	parsedEphemeral := c.UnmarshalBytePoint(ephemeral)
	if parsedEphemeral == nil {
		return nil, pgperrors.InvalidArgumentError("invalid ephemeral")
	}

	if len(proxyParam) != curve25519.ParamSize {
		return nil, pgperrors.InvalidArgumentError("invalid proxy parameter")
	}

	transformed, err := curve25519.ProxyTransform(parsedEphemeral, proxyParam)
	if err != nil {
		return nil, err
	}

	return c.MarshalBytePoint(transformed), nil
}