package machine

type Client interface {
	Delete(deleteConfig DeleteConfig) (DeleteResult, error)
	Exists(name string) (bool, error)
	GetConsoleURL(consoleConfig ConsoleConfig) (ConsoleResult, error)
	IP(ipConfig IPConfig) (IPResult, error)
	PowerOff(powerOff PowerOffConfig) (PowerOffResult, error)
	Start(startConfig StartConfig) (StartResult, error)
	Status(statusConfig ClusterStatusConfig) (ClusterStatusResult, error)
	Stop(stopConfig StopConfig) (StopResult, error)
}

type client struct{}

func NewClient() Client {
	return &client{}
}
