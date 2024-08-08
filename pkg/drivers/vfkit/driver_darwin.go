/*
Copyright 2021, Red Hat, Inc - All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vfkit

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/os/darwin"
	"github.com/crc-org/machine/libmachine/drivers"
	"github.com/crc-org/machine/libmachine/state"
	"github.com/crc-org/vfkit/pkg/config"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type Driver struct {
	*drivers.VMDriver
	VfkitPath string
	VirtioNet bool

	VsockPath       string
	DaemonVsockPort uint
	QemuGAVsockPort uint
}

func NewDriver(hostName, storePath string) *Driver {
	// checks that vfdriver.Driver implements the libmachine.Driver interface
	var _ drivers.Driver = &Driver{}
	return &Driver{
		VMDriver: &drivers.VMDriver{
			BaseDriver: &drivers.BaseDriver{
				MachineName: hostName,
				StorePath:   storePath,
			},
			CPU:    DefaultCPUs,
			Memory: DefaultMemory,
		},
		// needed when loading a VM which was created before
		// DaemonVsockPort was introduced
		DaemonVsockPort: constants.DaemonVsockPort,
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return DriverName
}

// Get Version information
func (d *Driver) DriverVersion() string {
	return DriverVersion
}

// GetIP returns an IP or hostname that this host is available at
// inherited from  libmachine.BaseDriver
// func (d *Driver) GetIP() (string, error)

// GetMachineName returns the name of the machine
// inherited from  libmachine.BaseDriver
// func (d *Driver) GetMachineName() string

// GetBundleName() Returns the name of the unpacked bundle which was used to create this machine
// inherited from  libmachine.BaseDriver
// func (d *Driver) GetBundleName() (string, error)

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) getDiskPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.img", d.MachineName))
}

func (d *Driver) resize(newSize int64) error {
	diskPath := d.getDiskPath()
	fi, err := os.Stat(diskPath)
	if err != nil {
		return err
	}
	if newSize == fi.Size() {
		log.Debugf("%s is already %d bytes", diskPath, newSize)
		return nil
	}
	if newSize < fi.Size() {
		return fmt.Errorf("current disk image capacity is bigger than the requested size (%d > %d)", fi.Size(), newSize)
	}
	return os.Truncate(diskPath, newSize)

}

// Create a host using the driver's config
func (d *Driver) Create() error {
	if err := d.PreCreateCheck(); err != nil {
		return err
	}

	switch d.ImageFormat {
	case "raw":
		if err := unix.Clonefile(d.ImageSourcePath, d.getDiskPath(), 0); err != nil {
			// fall back to a regular sparse copy, the error may be caused by the filesystem not supporting unix.CloneFile
			if err := crcos.CopyFileSparse(d.ImageSourcePath, d.getDiskPath()); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%s is an unsupported disk image format", d.ImageFormat)
	}

	return d.resize(int64(d.DiskCapacity))
}

func startVfkit(vfkitPath string, args []string) (*os.Process, error) {
	log.Debugf("Running %s %s", vfkitPath, strings.Join(args, " "))
	cmd := exec.Command(vfkitPath, args...)
	cmd.Stdout = log.StandardLogger().WriterLevel(log.DebugLevel)
	cmd.Stderr = cmd.Stdout
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()

	// catch vfkit early startup failures
	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(time.Second):
		break
	}

	return cmd.Process, nil
}

// Start a host
func (d *Driver) Start() error {
	if err := d.recoverFromUncleanShutdown(); err != nil {
		return err
	}

	efiStore := d.ResolveStorePath("efistore.nvram")
	create := !crcos.FileExists(efiStore)

	bootLoader := config.NewEFIBootloader(efiStore, create)

	vm := config.NewVirtualMachine(
		uint(d.CPU),
		uint64(d.Memory),
		bootLoader,
	)

	// console
	logFile := d.ResolveStorePath("vfkit.log")
	dev, err := config.VirtioSerialNew(logFile)
	if err != nil {
		return err
	}
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	// network
	// 52:54:00 is the OUI used by QEMU
	const mac = "52:54:00:70:2b:79"
	if d.VirtioNet {
		dev, err = config.VirtioNetNew(mac)
		if err != nil {
			return err
		}
		err = vm.AddDevice(dev)
		if err != nil {
			return err
		}
	}

	// shared directories
	if d.supportsVirtiofs() {
		for _, sharedDir := range d.SharedDirs {
			// TODO: add support for 'mount.ReadOnly'
			// TODO: check format
			dev, err := config.VirtioFsNew(sharedDir.Source, sharedDir.Tag)
			if err != nil {
				return err
			}
			err = vm.AddDevice(dev)
			if err != nil {
				return err
			}
		}
	}

	// entropy
	dev, err = config.VirtioRngNew()
	if err != nil {
		return err
	}
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	// disk
	diskPath := d.getDiskPath()
	dev, err = config.VirtioBlkNew(diskPath)
	if err != nil {
		return err
	}
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	// virtio-vsock device
	dev, err = config.VirtioVsockNew(d.DaemonVsockPort, d.VsockPath, true)
	if err != nil {
		return err
	}
	err = vm.AddDevice(dev)
	if err != nil {
		return err
	}

	// when loading a VM created by a crc version predating this commit,
	// d.QemuGAVsockPort will be missing from ~/.crc/machines/crc/config.json
	// In such a case, assume the VM will not support time sync
	if d.QemuGAVsockPort != 0 {
		timesync, err := config.TimeSyncNew(d.QemuGAVsockPort)
		if err != nil {
			return err
		}
		err = vm.AddDevice(timesync)
		if err != nil {
			return err
		}
	}

	args, err := vm.ToCmdLine()
	if err != nil {
		return err
	}
	process, err := startVfkit(d.VfkitPath, args)
	if err != nil {
		return err
	}

	_ = os.WriteFile(d.getPidFilePath(), []byte(strconv.Itoa(process.Pid)), 0600)

	if !d.VirtioNet {
		return nil
	}

	getIP := func() error {
		d.IPAddress, err = GetIPAddressByMACAddress(mac)
		if err != nil {
			return &RetriableError{Err: err}
		}
		return nil
	}

	if err := RetryAfter(60, getIP, 2*time.Second); err != nil {
		return fmt.Errorf("IP address never found in dhcp leases file %v", err)
	}
	log.Debugf("IP: %s", d.IPAddress)

	return nil
}

func (d *Driver) GetSharedDirs() ([]drivers.SharedDir, error) {
	if !d.supportsVirtiofs() {
		return nil, drivers.ErrNotSupported
	}
	// check if host supports file sharing, return drivers.ErrNotSupported if not
	return d.SharedDirs, nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	p, err := d.findVfkitProcess()
	if err != nil {
		return state.Error, err
	}
	if p == nil {
		return state.Stopped, nil
	}
	return state.Running, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return d.sendSignal(syscall.SIGKILL)
}

// Remove a host
func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil || s == state.Error {
		log.Debugf("Error checking machine status: %v, assuming it has been removed already", err)
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}
	return nil
}

// UpdateConfigRaw allows to change the state (memory, ...) of an already created machine
func (d *Driver) UpdateConfigRaw(rawConfig []byte) error {
	var newDriver Driver
	err := json.Unmarshal(rawConfig, &newDriver)
	if err != nil {
		return err
	}

	err = d.resize(int64(newDriver.DiskCapacity))
	if err != nil {
		log.Debugf("failed to resize disk image: %v", err)
		return err
	}
	*d = newDriver

	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s != state.Stopped {
		err := d.sendSignal(syscall.SIGTERM)
		if err != nil {
			return errors.Wrap(err, "vfkit sigterm failed")
		}
		// wait 120s for graceful shutdown
		for i := 0; i < 60; i++ {
			time.Sleep(2 * time.Second)
			s, _ := d.GetState()
			log.Debugf("VM state: %s", s)
			if s == state.Stopped {
				_ = os.Remove(d.getPidFilePath())
				return nil
			}
		}
		return errors.New("VM Failed to gracefully shutdown, try the kill command")
	}
	_ = os.Remove(d.getPidFilePath())
	return nil
}

func (d *Driver) getPidFilePath() string {
	const pidFileName = "vfkit.pid"
	return d.ResolveStorePath(pidFileName)
}

/*
 * Returns a ps.Process instance if it could find a vfkit process with the pid
 * stored in $pidFileName
 *
 * Returns nil, nil if:
 * - if the $pidFileName file does not exist,
 * - if a process with the pid from this file cannot be found,
 * - if a process was found, but its name is not 'vfkit'
 */
func (d *Driver) findVfkitProcess() (*process.Process, error) {
	pidFile := d.getPidFilePath()
	pid, err := readPidFromFile(pidFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "error reading pidfile %s", pidFile)
	}

	exists, err := process.PidExists(int32(pid))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot find pid %d", pid))
	}
	if p == nil {
		log.Debugf("vfkit pid %d missing from process table", pid)
		// return PidNotExist error?
		return nil, nil
	}

	name, err := p.Name()
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(name, "vfkit") {
		// return InvalidExecutable error?
		log.Debugf("pid %d is stale, and is being used by %s", pid, name)
		return nil, nil
	}

	return p, nil
}

func readPidFromFile(filename string) (int, error) {
	bs, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	content := strings.TrimSpace(string(bs))
	pid, err := strconv.Atoi(content)
	if err != nil {
		return 0, errors.Wrapf(err, "parsing %s", filename)
	}

	return pid, nil
}

// recoverFromUncleanShutdown searches for an existing vfkit.pid file in
// the machine directory. If it can't find it, a clean shutdown is assumed.
// If it finds the pid file, it checks for a running vfkit process with that pid
// as the existence of a file might not indicate an unclean shutdown but an actual running
// vfkit server. This is an error situation - we shouldn't start minikube as there is likely
// an instance running already. If the PID in the pidfile does not belong to a running vfkit
// process, we can safely delete it, and there is a good chance the machine will recover when restarted.
func (d *Driver) recoverFromUncleanShutdown() error {
	proc, err := d.findVfkitProcess()
	if err == nil && proc != nil {
		/* vfkit is running, pid file can't be stale */
		return nil
	}
	pidFile := d.getPidFilePath()
	/* There might be a stale pid file, try to remove it */
	if err := os.Remove(pidFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return errors.Wrap(err, fmt.Sprintf("removing pidFile %s", pidFile))
		}
	} else {
		log.Debugf("Removed stale pid file %s...", pidFile)
	}
	return nil
}

func (d *Driver) sendSignal(s syscall.Signal) error {
	proc, err := d.findVfkitProcess()
	if err != nil {
		return err
	}

	return proc.SendSignal(s)
}

func (d *Driver) supportsVirtiofs() bool {
	supportsVirtioFS, err := darwin.AtLeast("12.0.0")
	if err != nil {
		log.Debugf("Not able to compare version: %v", err)
	}
	return supportsVirtioFS
}
