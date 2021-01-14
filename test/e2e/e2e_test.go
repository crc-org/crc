package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/code-ready/clicumber/testsuite"
	"github.com/code-ready/crc/test/e2e/crcsuite"
	"github.com/cucumber/godog"
)

func TestMain(m *testing.M) {

	parseFlags()

	status := godog.RunWithOptions("crc", func(s *godog.Suite) {
		getFeatureContext(s)
	}, godog.Options{
		Format:              testsuite.GodogFormat,
		Paths:               strings.Split(testsuite.GodogPaths, ","),
		Tags:                testsuite.GodogTags,
		ShowStepDefinitions: testsuite.GodogShowStepDefinitions,
		StopOnFailure:       testsuite.GodogStopOnFailure,
		NoColors:            testsuite.GodogNoColors,
	})

	os.Exit(status)
}

func getFeatureContext(s *godog.Suite) {
	// load default step definitions of clicumber testsuite
	testsuite.FeatureContext(s)

	// here you can load additional step definitions, for example:
	crcsuite.FeatureContext(s) // CRC specific step definitions
}

func parseFlags() {

	// NOTE:
	// testsuite.ParseFlags() needs to be last: it calls flag.Parse()
	crcsuite.ParseFlags()
	testsuite.ParseFlags()
}
