package types

type Handshake struct {
	MTU     int    `json:"mtu"`
	Gateway string `json:"gateway"`
	VM      string `json:"vm"`
}

type ExposeRequest struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

type UnexposeRequest struct {
	Local string `json:"local"`
}
