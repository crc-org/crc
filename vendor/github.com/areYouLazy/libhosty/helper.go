package libhosty

//RestoreDefaultWindowsHostsFile loads the default windows hosts file
func (h *HostsFile) RestoreDefaultWindowsHostsFile() {
	hfl, _ := ParseHostsFileAsString(windowsHostsTemplate)
	h.HostsFileLines = hfl
}

//RestoreDefaultLinuxHostsFile loads the default linux hosts file
func (h *HostsFile) RestoreDefaultLinuxHostsFile() {
	hfl, _ := ParseHostsFileAsString(linuxHostsTemplate)
	h.HostsFileLines = hfl
}

//RestoreDefaultDarwinHostsFile loads the default darwin hosts file
func (h *HostsFile) RestoreDefaultDarwinHostsFile() {
	hfl, _ := ParseHostsFileAsString(darwinHostsTemplate)
	h.HostsFileLines = hfl
}

//AddDockerDesktopTemplate adds the dockerDesktopTemplate to the actual hostsFile
func (h *HostsFile) AddDockerDesktopTemplate() {
	hfl, _ := ParseHostsFileAsString(dockerDesktopTemplate)
	h.HostsFileLines = append(h.HostsFileLines, hfl...)
}
