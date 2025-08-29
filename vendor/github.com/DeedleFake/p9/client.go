package p9

import (
	"errors"
	"net"
	"sync/atomic"

	"github.com/DeedleFake/p9/proto"
)

var (
	// ErrUnsupportedVersion is returned from a handshake attempt that
	// fails due to a version mismatch.
	ErrUnsupportedVersion = errors.New("unsupported version")
)

// Client provides functionality for sending requests to and receiving
// responses from a 9P server.
//
// A Client must be closed when it is no longer going to be used in
// order to free up the related resources.
type Client struct {
	*proto.Client

	fid uint32
}

// NewClient returns a client that communicates using c. The Client
// will close c when the Client is closed.
func NewClient(c net.Conn) *Client {
	return &Client{Client: proto.NewClient(Proto(), c)}
}

// Dial is a convenience function that dials and creates a client in
// the same step.
func Dial(network, addr string) (*Client, error) {
	pc, err := proto.Dial(Proto(), network, addr)
	if err != nil {
		return nil, err
	}

	return &Client{Client: pc}, nil
}

func (c *Client) nextFID() uint32 {
	return atomic.AddUint32(&c.fid, 1) - 1
}

// Handshake performs an initial handshake to establish the maximum
// allowed message size. A handshake must be performed before any
// other request types may be sent.
func (c *Client) Handshake(msize uint32) (uint32, error) {
	rsp, err := c.Send(&Tversion{
		Msize:   msize,
		Version: Version,
	})
	if err != nil {
		return 0, err
	}

	version := rsp.(*Rversion)
	if version.Version != Version {
		return 0, ErrUnsupportedVersion
	}

	c.SetMsize(version.Msize)

	return version.Msize, nil
}

// Auth requests an auth file from the server, returning a Remote
// representing it or an error if one occurred.
func (c *Client) Auth(user, aname string) (*Remote, error) {
	fid := c.nextFID()

	rsp, err := c.Send(&Tauth{
		AFID:  fid,
		Uname: user,
		Aname: aname,
	})
	if err != nil {
		return nil, err
	}
	rauth := rsp.(*Rauth)

	return &Remote{
		client: c,
		fid:    fid,
		qid:    rauth.AQID,
	}, nil
}

// Attach attaches to a filesystem provided by the connected server
// with the given attributes. If no authentication has been done,
// afile may be nil.
func (c *Client) Attach(afile *Remote, user, aname string) (*Remote, error) {
	fid := c.nextFID()

	afid := NoFID
	if afile != nil {
		afid = afile.fid
	}

	rsp, err := c.Send(&Tattach{
		FID:   fid,
		AFID:  afid,
		Uname: user,
		Aname: aname,
	})
	if err != nil {
		return nil, err
	}
	attach := rsp.(*Rattach)

	return &Remote{
		client: c,
		fid:    fid,
		qid:    attach.QID,
	}, nil
}
