package bundle

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/spf13/cobra"
)

func getGenerateCmd(config *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate a custom bundle from the running OpenShift cluster",
		Long:  "Generate a custom bundle from the running OpenShift cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(config)
		},
	}
}

func runGenerate(config *config.Config) error {
	client := machine.NewClient(constants.DefaultName, isDebugLog(), config)

	return client.GenerateBundle()
}

func isDebugLog() bool {
	return logging.LogLevel == "debug"
}
