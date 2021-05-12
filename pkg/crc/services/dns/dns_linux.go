package dns

func runPostStartForOS(serviceConfig ServicePostStartConfig) error {
	// We might need to set the firewall here to forward
	// Update /etc/hosts file for host
	return addOpenShiftHosts(serviceConfig)
}
