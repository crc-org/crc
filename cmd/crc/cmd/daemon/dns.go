package daemon

import (
	"github.com/spf13/cobra"

	"github.com/code-ready/crc/pkg/crc/services/dns"
)

var (
	daemonDNSCmd = &cobra.Command{
		Use:    "dns",
		Short:  "Starts a DNS server on host",
		Long:   `Starts a DNS server on host`,
		Run:    runDNSDaemon,
		Hidden: true,
	}
)

func init() {
	DaemonCmd.AddCommand(daemonDNSCmd)
}

func runDNSDaemon(cmd *cobra.Command, args []string) {
	dns.RunDaemon()
}
