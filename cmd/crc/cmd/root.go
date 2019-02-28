package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: descriptionShort,
	Long:  descriptionLong,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		runPrerun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot()
	},
}

func init() {
	// nothing for now
}

func runPrerun() {
	fmt.Println(fmt.Sprintf("%s - %s", commandName, descriptionShort))
}

func runRoot() {
	fmt.Println("No command given")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("ERR:", err.Error())
		os.Exit(1)
	}
}
