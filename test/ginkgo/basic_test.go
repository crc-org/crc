package test_test

//. "github.com/onsi/ginkgo"
//. "github.com/onsi/gomega"

var statusInfo StatusAnswer

/*
var _ = Describe("basic commands", func() {

	Describe("version", func() {
		It("show", func() {
			Expect(RunCRCExpectSuccess("version")).To(ContainSubstring("CodeReady Containers version"))
		})

		It("show as json", func() {

			raw := RunCRCExpectSuccess("version", "-o", "json")
			err := json.Unmarshal([]byte(raw), &versionInfo)

			Expect(err).NotTo(HaveOccurred())
			Expect(versionInfo.Version).Should(MatchRegexp(`\d+\.\d+\.\d+.*`))

		})

	})

	Describe("help", func() {
		It("show", func() {
			Expect(RunCRCExpectSuccess("help")).To(ContainSubstring("Usage"))
		})
	})

	Describe("cleanup", func() {

		var stdout, stderr string

		Context("1st run", func() {
			It("should not error", func() {
				Expect(RunCRCExpectSuccess("cleanup")).To(ContainSubstring("Cleanup finished"))
			})
		})

		Context("2nd run", func() {

			It("should not error", func() {
				stdout, stderr = RunCRCExpectSuccessWithLogs("cleanup")
				Expect(stdout).To(ContainSubstring("Cleanup finished"))
			})

			switch os := runtime.GOOS; os {
			case "linux":
				It("should do these steps on Linux", func() {
					Expect(RunOnHostWithPrivilege("virsh", "list", "--all")).NotTo(ContainSubstring("crc")) // VM should not exist
					Expect("/etc/NetworkManager/dnsmasq.d/crc.conf").NotTo(BeAnExistingFile())
					Expect("/etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf").NotTo(BeAnExistingFile())
					Expect(RunOnHost("virsh", "net-list")).NotTo(ContainSubstring("crc")) // remove 'crc' net from libvirt
				})
			case "darwin":
				It("should do these steps on Darwin", func() {

					Expect(stderr).To(ContainSubstring("Unload CodeReady Containers tray"))
					Expect(stderr).To(ContainSubstring("Unload CodeReady Containers daemon"))
					Expect(stderr).To(ContainSubstring("Remove launchd configuration for tray"))
					Expect(stderr).To(ContainSubstring("Remove launchd configuration for daemon"))
					Expect("/etc/resolver/testing").NotTo(BeAnExistingFile())
				})
			case "windows":
				It("should do these steps on Windows", func() {
					Expect(stderr).To(ContainSubstring("Uninstalling tray if installed"))
					Expect(stderr).To(ContainSubstring("Will run as admin: Uninstalling CodeReady Containers System Tray"))
					Expect(stderr).To(ContainSubstring("Removing the crc VM if exists"))
					Expect(stderr).To(ContainSubstring("Removing dns server from interface"))
					Expect(stderr).To(ContainSubstring("Will run as admin: Remove dns entry for default switch"))
				})
			}
		})
	})

	Describe("setup", func() {

		var stdout, stderr string

		It("should run without error", func() {
			stdout, stderr = RunCRCExpectSuccessWithLogs("setup")
			Expect(stdout).To(ContainSubstring("Setup is complete"))
		})

		It("should cache oc binary and say how to start CRC", func() {
			ocLocation := filepath.Join(userHome, ".crc", "bin", "oc")
			_, err := os.Stat(ocLocation)
			Expect(err).NotTo(HaveOccurred())

			if !versionInfo.Embedded {
				Expect(stdout).To(ContainSubstring("crc start -b"))
			} else {
				Expect(stdout).NotTo(ContainSubstring("crc start -b"))
				Expect(stdout).To(ContainSubstring("crc start"))
			}

		})

		switch os := runtime.GOOS; os {
		case "linux":
			It("should do these steps on Linux", func() {
				Expect(stderr).To(ContainSubstring("Checking if oc binary is cached"))
				Expect(stderr).To(ContainSubstring("Checking if podman remote binary is cached"))
				Expect(stderr).To(ContainSubstring("Checking if goodhosts binary is cached"))
				Expect(stderr).To(ContainSubstring("Checking if CRC bundle is cached"))
				Expect(stderr).To(ContainSubstring("Checking minimum RAM requirements"))
				Expect(stderr).To(ContainSubstring("Checking if running as non-root"))
				Expect(stderr).To(ContainSubstring("Checking if Virtualization is enabled"))
				Expect(stderr).To(ContainSubstring("Checking if KVM is enabled"))
				Expect(stderr).To(ContainSubstring("Checking if libvirt is installed"))
				Expect(stderr).To(ContainSubstring("Checking if user is part of libvirt group"))
				Expect(stderr).To(ContainSubstring("Checking if libvirt daemon is running"))
				Expect(stderr).To(ContainSubstring("Checking if a supported libvirt version is installed"))
				Expect(stderr).To(ContainSubstring("Checking if crc-driver-libvirt is installed"))
				Expect(stderr).To(ContainSubstring("Checking if libvirt 'crc' network is available"))
				Expect(stderr).To(ContainSubstring("Checking if libvirt 'crc' network is active"))
				Expect(stderr).To(ContainSubstring("Checking if NetworkManager is installed"))
				Expect(stderr).To(ContainSubstring("Checking if NetworkManager service is running"))
				Expect(stderr).To(ContainSubstring("Checking if /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf exists"))
				Expect(stderr).To(ContainSubstring("Checking if /etc/NetworkManager/dnsmasq.d/crc.conf exists"))
			})
		case "darwin":
			It("should do these steps on Darwin", func() {
				Expect(stderr).To(ContainSubstring("Checking if oc binary is cached"))
				Expect(stderr).To(ContainSubstring("Checking if podman remote binary is cached"))
				Expect(stderr).To(ContainSubstring("Checking if goodhosts binary is cached"))
				Expect(stderr).To(ContainSubstring("Checking if CRC bundle is cached"))
				Expect(stderr).To(ContainSubstring("Checking minimum RAM requirements"))
				Expect(stderr).To(ContainSubstring("Checking if running as non-root"))
				Expect(stderr).To(ContainSubstring("Checking if HyperKit is installed"))
				Expect(stderr).To(ContainSubstring("Checking if crc-driver-hyperkit is installed"))
				Expect(stderr).To(ContainSubstring("Checking file permissions for /etc/hosts"))
				Expect(stderr).To(ContainSubstring("Checking file permissions for /etc/resolver/testing"))
			})
		case "windows":
			It("should do these steps on Windows", func() {
				Expect(stderr).To(ContainSubstring("Checking if oc binary is cached"))
				Expect(stderr).To(ContainSubstring("Checking if podman remote binary is cached"))
				//Expect(stderr).To(ContainSubstring("Checking if CRC bundle is cached")) would probably fail under different start conditions
				Expect(stderr).To(ContainSubstring("Checking minimum RAM requirements"))
				Expect(stderr).To(ContainSubstring("Checking if running as normal user"))
				Expect(stderr).To(ContainSubstring("Checking Windows 10 release"))
				Expect(stderr).To(ContainSubstring("Checking Windows edition"))
				Expect(stderr).To(ContainSubstring("Checking if Hyper-V is installed and operational"))
				Expect(stderr).To(ContainSubstring("Checking if user is a member of the Hyper-V Administrators group"))
				Expect(stderr).To(ContainSubstring("Checking if Hyper-V service is enabled"))
				Expect(stderr).To(ContainSubstring("Checking if the Hyper-V virtual switch exists"))
				Expect(stderr).To(ContainSubstring("Found Virtual Switch to use: Default Switch"))
			})
		}
	})

	Describe("start", func() {

		// start clean
		Context("from scratch or after delete", func() {

			var stdout, stderr string
			It("should run without error", func() {

				if versionInfo.Embedded {
					stdout, stderr = RunCRCExpectSuccessWithLogs("start", "--pull-secret-file", pullSecretLocation)
				} else {
					stdout, stderr = RunCRCExpectSuccessWithLogs("start", "--pull-secret-file", pullSecretLocation, "--bundle", bundleLocation)
				}

				Expect(stdout).To(ContainSubstring("Started the OpenShift cluster"))
				Expect(stdout).To(ContainSubstring("To access the cluster, first set up your environment"))
				Expect(stdout).To(ContainSubstring("Then you can access it by running"))
				Expect(stdout).To(ContainSubstring("To login as an admin, run"))
				Expect(stdout).To(ContainSubstring("You can now run 'crc console' and use these credentials to access the OpenShift web console"))

			})

			It("should result in a running cluster", func() {

				raw := RunCRCExpectSuccess("status", "-o", "json")
				err := json.Unmarshal([]byte(raw), &statusInfo)

				Expect(err).NotTo(HaveOccurred())
				Expect(statusInfo.CRCStatus).Should(MatchRegexp(`Running`))
				Expect(statusInfo.OpenshiftStatus).Should(MatchRegexp(`Running`))

			})

			switch os := runtime.GOOS; os {
			case "linux":
				It("should do these steps on Linux", func() {
					Expect(stderr).To(ContainSubstring("Checking if oc binary is cached"))
					Expect(stderr).To(ContainSubstring("Checking if podman remote binary is cached"))
					Expect(stderr).To(ContainSubstring("Checking if goodhosts binary is cached"))
					Expect(stderr).To(ContainSubstring("Checking minimum RAM requirements"))
					Expect(stderr).To(ContainSubstring("Checking if running as non-root"))
					Expect(stderr).To(ContainSubstring("Checking if Virtualization is enabled"))
					Expect(stderr).To(ContainSubstring("Checking if KVM is enabled"))
					Expect(stderr).To(ContainSubstring("Checking if libvirt is installed"))
					Expect(stderr).To(ContainSubstring("Checking if user is part of libvirt group"))
					Expect(stderr).To(ContainSubstring("Checking if libvirt daemon is running"))
					Expect(stderr).To(ContainSubstring("Checking if a supported libvirt version is installed"))
					Expect(stderr).To(ContainSubstring("Checking if crc-driver-libvirt is installed"))
					//Expect(stderr).To(ContainSubstring("Checking for obsolete crc-driver-libvirt"))
					Expect(stderr).To(ContainSubstring("Checking if libvirt 'crc' network is available"))
					Expect(stderr).To(ContainSubstring("Checking if libvirt 'crc' network is active"))
					Expect(stderr).To(ContainSubstring("Checking if NetworkManager is installed"))
					Expect(stderr).To(ContainSubstring("Checking if NetworkManager service is running"))
					Expect(stderr).To(ContainSubstring("Checking if /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf exists"))
					Expect(stderr).To(ContainSubstring("Checking if /etc/NetworkManager/dnsmasq.d/crc.conf exists"))
					//Expect(stderr).To(ContainSubstring("Loading bundle")) // this will fail if bundle not extracted yet
					Expect(stderr).To(ContainSubstring("Checking size of the disk image"))
					Expect(stderr).To(ContainSubstring("Creating CodeReady Containers VM for OpenShift"))
					Expect(stderr).To(ContainSubstring("CodeReady Containers VM is running"))
					Expect(stderr).To(ContainSubstring("Generating new SSH Key pair"))
					Expect(stderr).To(ContainSubstring("Updating authorized keys"))
					Expect(stderr).To(ContainSubstring("Copying kubeconfig file to instance dir"))
					Expect(stderr).To(ContainSubstring("Starting network time synchronization in CodeReady Containers VM"))
					Expect(stderr).To(ContainSubstring("Check internal and public DNS query"))
					Expect(stderr).To(ContainSubstring("Check DNS query from host"))
					Expect(stderr).To(ContainSubstring("Adding user's pull secret to instance disk"))
					Expect(stderr).To(ContainSubstring("Verifying validity of the kubelet certificates"))
					Expect(stderr).To(ContainSubstring("Starting OpenShift kubelet service"))
					Expect(stderr).To(ContainSubstring("Configuring cluster for first start"))
					Expect(stderr).To(ContainSubstring("Adding user's pull secret to the cluster"))
					Expect(stderr).To(ContainSubstring("Starting OpenShift cluster ... [waiting 3m]"))
					Expect(stderr).To(ContainSubstring("Updating kubeconfig"))
				})
			case "darwin":
				It("should do these steps on Darwin", func() {
					Expect(stderr).To(ContainSubstring("Checking if oc binary is cached"))
					Expect(stderr).To(ContainSubstring("Checking if podman remote binary is cached"))
					Expect(stderr).To(ContainSubstring("Checking if goodhosts binary is cached"))
					Expect(stderr).To(ContainSubstring("Checking if CRC bundle is cached"))
					Expect(stderr).To(ContainSubstring("Checking minimum RAM requirements"))
					Expect(stderr).To(ContainSubstring("Checking if running as non-root"))
					Expect(stderr).To(ContainSubstring("Checking if HyperKit is installed"))
					Expect(stderr).To(ContainSubstring("Checking if crc-driver-hyperkit is installed"))
					Expect(stderr).To(ContainSubstring("Checking file permissions for /etc/hosts"))
					Expect(stderr).To(ContainSubstring("Checking file permissions for /etc/resolver/testing"))
					Expect(stderr).To(ContainSubstring("Checking size of the disk image"))
					Expect(stderr).To(ContainSubstring("Creating CodeReady Containers VM for OpenShift"))
					Expect(stderr).To(ContainSubstring("CodeReady Containers VM is running"))
					Expect(stderr).To(ContainSubstring("Generating new SSH Key pair"))
					Expect(stderr).To(ContainSubstring("Updating authorized keys"))
					Expect(stderr).To(ContainSubstring("Copying kubeconfig file to instance dir"))
					Expect(stderr).To(ContainSubstring("Starting network time synchronization in CodeReady Containers VM"))
					Expect(stderr).To(ContainSubstring("Restarting the host network"))
					Expect(stderr).To(ContainSubstring("Check internal and public DNS query"))
					Expect(stderr).To(ContainSubstring("Check DNS query from host"))
					Expect(stderr).To(ContainSubstring("Adding user's pull secret to instance disk"))
					Expect(stderr).To(ContainSubstring("Verifying validity of the kubelet certificates"))
					Expect(stderr).To(ContainSubstring("Starting OpenShift kubelet service"))
					Expect(stderr).To(ContainSubstring("Configuring cluster for first start"))
					Expect(stderr).To(ContainSubstring("Adding user's pull secret to the cluster"))
					Expect(stderr).To(ContainSubstring("Updating cluster ID"))
					Expect(stderr).To(ContainSubstring("Starting OpenShift cluster ... [waiting 3m]"))
					Expect(stderr).To(ContainSubstring("Updating kubeconfig"))
				})
			case "windows":
				It("should do these steps on Windows", func() {
					Expect(stderr).To(ContainSubstring("Checking if oc binary is cached"))
					Expect(stderr).To(ContainSubstring("Checking if podman remote binary is cached"))
					//Expect(stderr).To(ContainSubstring("Checking if CRC bundle is cached")) // would probably fail under different start conditions
					Expect(stderr).To(ContainSubstring("Checking minimum RAM requirements"))
					Expect(stderr).To(ContainSubstring("Checking if running as normal user"))
					Expect(stderr).To(ContainSubstring("Checking Windows 10 release"))
					Expect(stderr).To(ContainSubstring("Checking Windows edition"))
					Expect(stderr).To(ContainSubstring("Checking if Hyper-V is installed and operational"))
					Expect(stderr).To(ContainSubstring("Checking if user is a member of the Hyper-V Administrators group"))
					Expect(stderr).To(ContainSubstring("Checking if Hyper-V service is enabled"))
					Expect(stderr).To(ContainSubstring("Checking if the Hyper-V virtual switch exists"))
					Expect(stderr).To(ContainSubstring("Found Virtual Switch to use: Default Switch"))
					Expect(stderr).To(ContainSubstring("Checking size of the disk image"))
					Expect(stderr).To(ContainSubstring("Creating CodeReady Containers VM for OpenShift"))
					Expect(stderr).To(ContainSubstring("CodeReady Containers VM is running"))
					Expect(stderr).To(ContainSubstring("Generating new SSH Key pair"))
					Expect(stderr).To(ContainSubstring("Updating authorized keys"))
					Expect(stderr).To(ContainSubstring("Copying kubeconfig file to instance dir"))
					Expect(stderr).To(ContainSubstring("Starting network time synchronization in CodeReady Containers VM"))
					Expect(stderr).To(ContainSubstring("Will run as admin: add dns server address to interface vEthernet"))
					Expect(stderr).To(ContainSubstring("Check internal and public DNS query"))
					Expect(stderr).To(ContainSubstring("Adding user's pull secret to instance disk"))
					Expect(stderr).To(ContainSubstring("Verifying validity of the kubelet certificates"))
					Expect(stderr).To(ContainSubstring("Starting OpenShift kubelet service"))
					Expect(stderr).To(ContainSubstring("Configuring cluster for first start"))
					Expect(stderr).To(ContainSubstring("Adding user's pull secret to the cluster"))
					Expect(stderr).To(ContainSubstring("Updating cluster ID"))
					Expect(stderr).To(ContainSubstring("Starting OpenShift cluster ... [waiting 3m]"))
					Expect(stderr).To(ContainSubstring("Updating kubeconfig"))
				})
			}
		})

		// start existing VM
		Context("after stop", func() {

			It("(1. stop the cluster first)", func() {
				Expect(RunCRCExpectSuccess("stop")).To(ContainSubstring("Stopped the OpenShift cluster"))
			})

			It("2. start without pull secret)", func() {
				Expect(RunCRCExpectSuccess("start")).To(ContainSubstring("Started the OpenShift cluster"))
			})

			It("3. cluster is running", func() {

				raw := RunCRCExpectSuccess("status", "-o", "json")
				err := json.Unmarshal([]byte(raw), &statusInfo)

				Expect(err).NotTo(HaveOccurred())
				Expect(statusInfo.CRCStatus).Should(Equal("Running"))
				Expect(statusInfo.OpenshiftStatus).Should(Equal("Running"))

			})
		})
	})

	Describe("stop", func() {
		It("the cluster forcibly works", func() {
			stdout, stderr := RunCRCExpectSuccessWithLogs("stop", "-f")
			Expect(stderr).To(ContainSubstring("Stopping the OpenShift cluster, this may take a few minutes..."))
			Expect(stdout).To(ContainSubstring("Stopped the OpenShift cluster"))
		})
	})

	Describe("delete", func() {
		It("the VM works", func() {
			Expect(RunCRCExpectSuccess("delete", "-f")).To(ContainSubstring("Deleted the OpenShift cluster"))
		})
	})
})
*/
