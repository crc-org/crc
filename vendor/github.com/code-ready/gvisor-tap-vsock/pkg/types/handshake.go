package types

type Handshake struct {
	MTU     int    `json:"mtu"`
	Gateway string `json:"gateway"`
	VM      string `json:"vm"`
}
