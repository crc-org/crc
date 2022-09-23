package types

type AddRequest struct {
	IP    string   `json:"ip"`
	Hosts []string `json:"hosts"`
}

type RemoveRequest struct {
	Hosts []string `json:"hosts"`
}

type CleanRequest struct {
	Domains []string `json:"domains"`
}
