package bundle

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/spf13/cobra"
)

func getGenerateCmd(config *config.Config) *cobra.Command {
	var forceStop bool
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a custom bundle from the running OpenShift cluster",
		Long:  "Generate a custom bundle from the running OpenShift cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(config, forceStop)
		},
	}
	generateCmd.PersistentFlags().BoolVarP(&forceStop, "force-stop", "f", false, "Forcefully stop the instance")
	return generateCmd
}

func runGenerate(config *config.Config, forceStop bool) error {
	client := machine.NewClient(constants.DefaultName, logging.IsDebug(), config)

	return client.GenerateBundle(forceStop)
}
