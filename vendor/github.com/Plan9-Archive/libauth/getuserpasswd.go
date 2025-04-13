package libauth

import (
	"bufio"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
)

type UserPasswd struct {
	User, Password string
}

// Get a password. params i.e. proto=pass service=ssh role=client server=%s user=%s
func Getuserpasswd(params string, args ...interface{}) (*UserPasswd, error) {
	var buf [4096]byte
	f, e := openRPC()
	if e != nil {
		return nil, e
	}
	defer f.Close()

retry0:
	cmd := fmt.Sprintf("start "+params, args...)
	_, e = io.WriteString(f, cmd)
	if e != nil {
		return nil, e
	}
	n, e := f.Read(buf[:])
	if e != nil {
		return nil, e
	}
	s := string(buf[0:n])
	ss := tokenize(s)
	switch ss[0] {
	case "ok":
	case "needkey":
		getkey(strings.Join(ss[1:], " "))
		goto retry0
	default:
		return nil, errors.New(s)
	}
retry1:
	cmd = "read"
	_, e = io.WriteString(f, cmd)
	if e != nil {
		return nil, e
	}
	n, e = f.Read(buf[:])
	if e != nil {
		return nil, e
	}
	s = string(buf[0:n])
	ss = tokenize(s)
	switch ss[0] {
	case "needkey":
		getkey(strings.Join(ss[1:], " "))
		goto retry1
	case "ok":
		if len(ss) != 3 {
			return nil, fmt.Errorf("expected 3 fields in ok, got %d", len(ss))
		}
		return &UserPasswd{ss[1], ss[2]}, nil
	default:
		return nil, errors.New(s)
	}

	return nil, nil
}

// find our rsa public keys
func Listkeys() ([]rsa.PublicKey, error) {
	var keys []rsa.PublicKey

	fctl, err := openCtl()
	if err != nil {
		return nil, err
	}
	defer fctl.Close()

	scan := bufio.NewScanner(fctl)

	for scan.Scan() {
		l := scan.Text()
		spl := tokenize(l)

		// ignore 'key'
		if spl[0] == "key" {
			spl = spl[1:]
		}

		attrs := attrmap(strings.Join(spl, " "))

		if proto, ok := attrs["proto"]; ok && proto == "rsa" {
			if exp, ok := attrs["ek"]; ok {
				if modulus, ok := attrs["n"]; ok {
					var pk rsa.PublicKey
					var eb bool
					var expint int64

					if expint, err = strconv.ParseInt(exp, 16, 0); err != nil {
						return nil, err
					}

					pk.E = int(expint)

					N := new(big.Int)
					if pk.N, eb = N.SetString(modulus, 16); !eb {
						return nil, fmt.Errorf("failed to read modulus")
					}

					keys = append(keys, pk)
				}
			}
		}

	}

	return keys, nil
}

// Get a private key. params i.e. proto=rsa service=ssh role=client
func Getkey(params string, args ...interface{}) {
}
