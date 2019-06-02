package daemon

import (
	"github.com/code-ready/crc/pkg/crc/services/dns"
	"github.com/spf13/cobra"
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
