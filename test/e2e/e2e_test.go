package e2e

import (
	"os"
	"testing"

	"github.com/crc-org/crc/v2/test/e2e/testsuite"
	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
)

var opts = godog.Options{
	Format:              "pretty",
	Paths:               []string{"./features"},
	Tags:                "",
	ShowStepDefinitions: false,
	StopOnFailure:       false,
	NoColors:            false,
}

func init() {

	pflag.StringVar(&opts.Format, "godog.format", "pretty", "Sets which format godog will use")
	pflag.StringSliceVar(&opts.Paths, "godog.paths", []string{"./features"}, "Relative location of feature files")
	pflag.StringVar(&opts.Tags, "godog.tags", "", "Tags for godog test")
	pflag.BoolVar(&opts.ShowStepDefinitions, "godog.definitions", false, "")
	pflag.BoolVar(&opts.StopOnFailure, "godog.stop-on-failure", false, "Stop when failure is found")
	pflag.BoolVar(&opts.NoColors, "godog.no-colors", false, "Disable colors in godog output")

	testsuite.ParseFlags()
}

func TestMain(_ *testing.M) {

	pflag.Parse()
	testsuite.GodogTags = opts.Tags
	status := godog.TestSuite{
		Name:                 "crc",
		TestSuiteInitializer: testsuite.InitializeTestSuite,
		ScenarioInitializer:  testsuite.InitializeScenario,
		Options:              &opts,
	}.Run()

	os.Exit(status)
}
