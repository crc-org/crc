package bundle

import (
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine"

	"github.com/spf13/cobra"
)

func getGenerateCmd(config *crcConfig.Config) *cobra.Command {
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

func runGenerate(config *crcConfig.Config, forceStop bool) error {
	preset := crcConfig.GetPreset(config)
	client := machine.NewClient(constants.InstanceName(preset), logging.IsDebug(), config)

	return client.GenerateBundle(forceStop)
}
