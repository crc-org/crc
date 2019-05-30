package network

type NameServer struct {
	IPAddress string
}

type SearchDomain struct {
	Domain string
}

type ResolvFileValues struct {
	SearchDomains []SearchDomain
	NameServers   []NameServer
}
