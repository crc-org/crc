package libauth

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type AuthRpc struct {
	F   io.ReadWriteCloser
	Arg []byte
	Ai  *AuthInfo
}

type AuthInfo struct {
	Cuid   string
	Suid   string
	Cap    string
	Secret []byte
}

type AuthRet int

const (
	ARok AuthRet = iota /* rpc return values */
	ARdone
	ARerror
	ARneedkey
	ARbadkey
	ARwritenext
	ARtoosmall
	ARtoobig
	ARrpcfailure
	ARphase

	AuthRpcMax = 4096
)

var artab = map[string]AuthRet{
	"ok":       ARok,
	"done":     ARdone,
	"error":    ARerror,
	"needkey":  ARneedkey,
	"badkey":   ARbadkey,
	"phase":    ARphase,
	"toosmall": ARtoosmall,
}

func gbit16(p []byte) uint16 {
	return uint16(p[0]) | uint16(p[1])<<8
}

func gstring(buf []byte) ([]byte, string) {
	if len(buf) < 2 {
		return buf, ""
	}
	n := int(gbit16(buf))
	buf = buf[2:]
	if len(buf) < n {
		return buf, ""
	}
	return buf[n:], string(buf[:n])
}

func garray(buf []byte) ([]byte, []byte) {
	if len(buf) < 2 {
		return buf, nil
	}
	n := int(gbit16(buf))
	buf = buf[2:]
	if len(buf) < n {
		return buf, nil
	}
	return buf[n:], buf[:n]
}

func convM2AI(buf []byte) *AuthInfo {
	ai := new(AuthInfo)

	buf, ai.Cuid = gstring(buf)
	buf, ai.Suid = gstring(buf)
	buf, ai.Cap = gstring(buf)
	buf, ai.Secret = garray(buf)

	return ai
}

func (rpc *AuthRpc) GetInfo() error {
	if r, msg := rpc.Rpc("authinfo", ""); r != ARok {
		return errors.New(msg)
	}
	ai := convM2AI(rpc.Arg)
	if ai == nil {
		return errors.New("bad auth info from factotum")
	}

	rpc.Ai = ai
	return nil
}

func (rpc *AuthRpc) Rpc(overb string, arg string) (AuthRet, string) {
	if len(overb)+1+len(arg) > AuthRpcMax {
		return ARtoobig, "rpc too big"
	}

	if _, err := rpc.F.Write([]byte(overb + " " + arg)); err != nil {
		return ARrpcfailure, "write: " + err.Error()
	}

	ibuf := make([]byte, AuthRpcMax)
	if n, err := rpc.F.Read(ibuf); err != nil {
		return ARrpcfailure, "read: " + err.Error()
	} else {
		ibuf = ibuf[:n]
	}

	var iverb string
	if i := bytes.IndexByte(ibuf, ' '); i > 0 {
		iverb = string(ibuf[:i])
		rpc.Arg = ibuf[i+1:]
	} else {
		iverb = string(ibuf)
		rpc.Arg = []byte{}
	}

	ar, ok := artab[iverb]
	if !ok {
		return ARrpcfailure, "malformed rpc response: " + string(ibuf)
	}

	switch ar {
	case ARok:
		return ARok, string(rpc.Arg)
	case ARrpcfailure:
		return ARrpcfailure, ""
	case ARerror:
		if string(rpc.Arg) == "" {
			return ARerror, "unspecified rpc error"
		} else {
			return ARerror, string(rpc.Arg)
		}
	case ARneedkey:
		return ARneedkey, string(ibuf)
	case ARbadkey:
		return ARbadkey, string(ibuf)
	case ARphase:
		return ARphase, "phase error " + string(rpc.Arg)
	default:
		return ar, fmt.Sprintf("unknown rpc type %d", ar)
	}
	panic("can't happen")
}

func fauth_proxy(rw io.ReadWriter, rpc *AuthRpc, params string) (*AuthInfo, error) {
	if r, msg := rpc.Rpc("start", params); r != ARok {
		return nil, errors.New("fauth_proxy start: " + msg)
	}

	for {
		switch r, msg := rpc.Rpc("read", ""); r {
		case ARdone:
			if err := rpc.GetInfo(); err != nil {
				return nil, err
			}
			return rpc.Ai, nil
		case ARok:
			if _, err := rw.Write(rpc.Arg); err != nil {
				return nil, errors.New("write: " + err.Error())
			}
		case ARphase:
			var r AuthRet
			var msg string
			n := 0
			buf := make([]byte, AuthRpcMax)
			for {
				r, msg = rpc.Rpc("write", string(buf[:n]))
				if r != ARtoosmall {
					break
				}

				i, err := strconv.Atoi(string(rpc.Arg))
				if err != nil {
					return nil, errors.New("phase atoi: " + err.Error() + ": " + string(rpc.Arg))
				}
				if i > AuthRpcMax {
					break
				}

				m, err := rw.Read(buf[n:])
				if err != nil {
					return nil, fmt.Errorf("phase read: %d %s", m, err.Error())
				}
				if m == 0 {
					return nil, errors.New("phase short read: " + string(buf))
				}
				n += m
			}

			if r != ARok {
				return nil, errors.New("phase write: " + string(buf) + ": " + msg)
			}
		default:
			return nil, errors.New("rpc: " + msg)
		}
	}

	return nil, errors.New("fauth_proxy stub")
}

func OpenRPC() (io.ReadWriteCloser, error) {
	return openRPC()
}

func Proxy(rw io.ReadWriter, format string, a ...interface{}) (*AuthInfo, error) {
	f, err := openRPC()
	if err != nil {
		return nil, errors.New("openRPC: " + err.Error())
	}
	defer f.Close()

	rpc := &AuthRpc{
		F: f,
	}

	keyspec := fmt.Sprintf(format, a...)
	return fauth_proxy(rw, rpc, keyspec)
}

func RsaSign(sha1 []byte) (signed []byte, err error) {
	f, err := openRPC()
	if err != nil {
		return nil, err
	}

	defer f.Close()

	rpc := &AuthRpc{
		F: f,
	}

	if ar, str := rpc.Rpc("start", "role=sign proto=rsa"); ar != ARok {
		return nil, errors.New("start: " + str)
	}
	if ar, str := rpc.Rpc("write", string(sha1)); ar != ARok {
		return nil, errors.New("write: " + str)
	}
	if ar, str := rpc.Rpc("read", ""); ar != ARok || rpc.Arg == nil || len(rpc.Arg) <= 0 {
		return nil, errors.New("read: " + str)
	}

	return rpc.Arg, nil
}
