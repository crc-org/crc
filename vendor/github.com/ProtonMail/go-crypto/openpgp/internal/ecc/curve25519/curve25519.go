// Package curve25519 implements custom field operations without clamping for forwarding.
package curve25519

import (
	"crypto/subtle"
	"github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/internal/ecc/curve25519/field"
	x25519lib "github.com/cloudflare/circl/dh/x25519"
	"math/big"
)

var curveGroupByte = [x25519lib.Size]byte{
	0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x14, 0xde, 0xf9, 0xde, 0xa2, 0xf7, 0x9c, 0xd6, 0x58, 0x12, 0x63, 0x1a, 0x5c, 0xf5, 0xd3, 0xed,
}

const ParamSize = x25519lib.Size

func DeriveProxyParam(recipientSecretByte, forwardeeSecretByte []byte) (proxyParam []byte, err error) {
	curveGroup := new(big.Int).SetBytes(curveGroupByte[:])
	recipientSecret := new(big.Int).SetBytes(recipientSecretByte)
	forwardeeSecret := new(big.Int).SetBytes(forwardeeSecretByte)

	proxyTransform := new(big.Int).Mod(
		new(big.Int).Mul(
			new(big.Int).ModInverse(forwardeeSecret, curveGroup),
			recipientSecret,
		),
		curveGroup,
	)

	rawProxyParam := proxyTransform.Bytes()

	// pad and convert to small endian
	proxyParam = make([]byte, x25519lib.Size)
	l := len(rawProxyParam)
	for i := 0; i < l; i++ {
		proxyParam[i] = rawProxyParam[l-i-1]
	}

	return proxyParam, nil
}

func ProxyTransform(ephemeral, proxyParam []byte) ([]byte, error) {
	var transformed, safetyCheck [x25519lib.Size]byte

	var scalarEight = make([]byte, x25519lib.Size)
	scalarEight[0] = 0x08
	err := ScalarMult(&safetyCheck, scalarEight, ephemeral)
	if err != nil {
		return nil, err
	}

	err = ScalarMult(&transformed, proxyParam, ephemeral)
	if err != nil {
		return nil, err
	}

	return transformed[:], nil
}

func ScalarMult(dst *[32]byte, scalar, point []byte) error {
	var in, base, zero [32]byte
	copy(in[:], scalar)
	copy(base[:], point)

	scalarMult(dst, &in, &base)
	if subtle.ConstantTimeCompare(dst[:], zero[:]) == 1 {
		return errors.InvalidArgumentError("invalid ephemeral: low order point")
	}

	return nil
}

func scalarMult(dst, scalar, point *[32]byte) {
	var e [32]byte

	copy(e[:], scalar[:])

	var x1, x2, z2, x3, z3, tmp0, tmp1 field.Element
	x1.SetBytes(point[:])
	x2.One()
	x3.Set(&x1)
	z3.One()

	swap := 0
	for pos := 254; pos >= 0; pos-- {
		b := e[pos/8] >> uint(pos&7)
		b &= 1
		swap ^= int(b)
		x2.Swap(&x3, swap)
		z2.Swap(&z3, swap)
		swap = int(b)

		tmp0.Subtract(&x3, &z3)
		tmp1.Subtract(&x2, &z2)
		x2.Add(&x2, &z2)
		z2.Add(&x3, &z3)
		z3.Multiply(&tmp0, &x2)
		z2.Multiply(&z2, &tmp1)
		tmp0.Square(&tmp1)
		tmp1.Square(&x2)
		x3.Add(&z3, &z2)
		z2.Subtract(&z3, &z2)
		x2.Multiply(&tmp1, &tmp0)
		tmp1.Subtract(&tmp1, &tmp0)
		z2.Square(&z2)

		z3.Mult32(&tmp1, 121666)
		x3.Square(&x3)
		tmp0.Add(&tmp0, &z3)
		z3.Multiply(&x1, &z2)
		z2.Multiply(&tmp1, &tmp0)
	}

	x2.Swap(&x3, swap)
	z2.Swap(&z3, swap)

	z2.Invert(&z2)
	x2.Multiply(&x2, &z2)
	copy(dst[:], x2.Bytes())
}
