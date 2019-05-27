package preflight

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	preflightCheckSucceedsOrFails(false,
		checkOcBinaryCached,
		"Checking if oc binary is cached",
		false,
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	preflightCheckAndFix(false,
		checkOcBinaryCached,
		fixOcBinaryCached,
		"Caching oc binary",
		false,
	)
}
