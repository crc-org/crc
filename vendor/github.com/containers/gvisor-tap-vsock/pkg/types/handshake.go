package types

type ExposeRequest struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

type UnexposeRequest struct {
	Local string `json:"local"`
}
