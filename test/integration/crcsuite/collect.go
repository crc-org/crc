package crcsuite

import (
	"archive/zip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/ssh"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

func runDiagnose(dir string) error {
	filename := filepath.Join(dir, fmt.Sprintf("diagnose-%d.zip", time.Now().Unix()))
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	fmt.Printf("Writing diagnostics at %s\n", filename)
	zipFile, err := os.Create(absolute)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	var collectors = []Collector{
		&TreeCollector{
			Dir:    constants.CrcBaseDir,
			Target: "tree.txt",
		},
		file(constants.ConfigPath),
		file(constants.LogFilePath),
		file(constants.DaemonLogFilePath),
		file(filepath.Join(constants.MachineBaseDir, "machines", constants.DefaultName, "config.json")),
		file("/etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf"),
		file("/etc/NetworkManager/dnsmasq.d/crc.conf"),
		file("/etc/hosts"),
		vmFile("/etc/resolv.conf"),
		vmFile("/etc/hosts"),
		command("journalctl -n 1000 --no-pager", "journalctl.txt"),
		command("ip addr", "ip-addr.txt"),
		vmCommand("journalctl -n 1000 --no-pager", "journalctl.txt"),
		vmCommand("ip addr", "ip-addr.txt"),
		vmCommand("dmesg", "dmesg.txt"),
		vmCommand("ps aux", "ps.txt"),
		vmCommand("df -h", "df.txt"),
		vmCommand("free -m", "free.txt"),
		vmCommand("sudo podman ps -a", "podman-ps.txt"),
		vmCommand("sudo crictl ps -a", "crictl-ps.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc get clusterversion", "get-clusterversion.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc describe clusterversion", "describe-clusterversion.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc get co", "get-clusteroperators.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc describe co", "describe-clusteroperators.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc get nodes", "get-nodes.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc describe nodes", "describe-nodes.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc get pods --all-namespaces", "get-pods.txt"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc describe pods --all-namespaces", "describe-pods.txt"),
		&ContainerLogCollector{Process: "podman"},
		&ContainerLogCollector{Process: "crictl"},
	}

	start := time.Now()

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(4)

	fw := &FileWriter{
		writer: writer,
	}
	for i := range collectors {
		wg.Add(1)
		collector := collectors[i]
		go func() {
			defer wg.Done()
			_ = sem.Acquire(context.Background(), 1)
			defer sem.Release(1)
			if err := collector.Collect(fw); err != nil {
				logging.Errorf("error while collecting: %v", err)
			}
		}()
	}
	wg.Wait()
	logging.Infof("diagnose took %s", time.Since(start))
	return nil
}

type Writer interface {
	Write(filename string, bin []byte) error
}

type FileWriter struct {
	writer    *zip.Writer
	writeLock sync.Mutex
}

func (fw *FileWriter) Write(filename string, bin []byte) error {
	fw.writeLock.Lock()
	defer fw.writeLock.Unlock()

	f, err := fw.writer.Create(filename)
	if err != nil {
		return err
	}

	_, err = f.Write(bin)
	return err
}

type Collector interface {
	Collect(Writer) error
}

func file(path string) *FileCollector {
	return &FileCollector{Source: path, Target: filepath.Base(path)}
}

type FileCollector struct {
	Source, Target string
}

func (collector *FileCollector) Collect(w Writer) error {
	bin, err := ioutil.ReadFile(collector.Source)
	if err != nil {
		return err
	}
	return w.Write(collector.Target, bin)
}

type TreeCollector struct {
	Dir    string
	Target string
}

func (collector *TreeCollector) Collect(w Writer) error {
	var list []string
	err := filepath.Walk(collector.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		list = append(list, fmt.Sprintf("%s: %d %s", path, info.Size(), info.Mode().String()))
		return nil
	})
	if err != nil {
		return err
	}
	return w.Write(collector.Target, []byte(strings.Join(list, "\n")))
}

func vmFile(file string) *VMCommandCollector {
	return vmCommand(fmt.Sprintf("cat %s", file), filepath.Base(file))
}

func vmCommand(command, target string) *VMCommandCollector {
	return &VMCommandCollector{
		Command: command,
		Target:  target,
	}
}

type VMCommandCollector struct {
	Command, Target string
}

func (collector *VMCommandCollector) Collect(w Writer) error {
	client := machine.NewClient(constants.DefaultName, true, network.DefaultMode)
	ip, err := client.IP()
	if err != nil {
		return err
	}
	ssh, err := ssh.NewClient(constants.DefaultSSHUser, ip, 22, &ssh.Auth{
		Keys: []string{constants.GetPrivateKeyPath()},
	})
	if err != nil {
		return err
	}
	out, err := ssh.Output(collector.Command)
	if err != nil {
		return errors.Wrapf(err, "collecting %s", collector.Target)
	}
	return w.Write(path.Join("vm", collector.Target), []byte(out))
}

type ContainerLogCollector struct {
	Process string
}

func (collector *ContainerLogCollector) Collect(w Writer) error {
	client := machine.NewClient(constants.DefaultName, true, network.DefaultMode)
	ip, err := client.IP()
	if err != nil {
		return err
	}
	ssh, err := ssh.NewClient(constants.DefaultSSHUser, ip, 22, &ssh.Auth{
		Keys: []string{constants.GetPrivateKeyPath()},
	})
	if err != nil {
		return err
	}
	out, err := ssh.Output(fmt.Sprintf("sudo %s ps -a -q", collector.Process))
	if err != nil {
		return err
	}
	for _, id := range strings.Split(strings.TrimSpace(out), "\n") {
		inspect, err := ssh.Output(fmt.Sprintf("sudo %s inspect %s", collector.Process, id))
		if err != nil {
			logging.Errorf("error while inspecting %s: %v", id, err)
			continue
		}
		if err := w.Write(path.Join("vm", collector.Process, fmt.Sprintf("%s-inspect.txt", id)), []byte(inspect)); err != nil {
			logging.Errorf("error while inspecting %s: %v", id, err)
			continue
		}
		logs, err := ssh.Output(fmt.Sprintf("sudo %s logs --tail 200 %s", collector.Process, id))
		if err != nil {
			logging.Errorf("error while getting logs %s: %v", id, err)
			continue
		}
		if err := w.Write(path.Join("vm", collector.Process, fmt.Sprintf("%s-logs.txt", id)), []byte(logs)); err != nil {
			logging.Errorf("error while getting logs %s: %v", id, err)
			continue
		}
	}
	return nil
}

func command(command, target string) *CommandCollector {
	return &CommandCollector{
		Command: command,
		Target:  target,
	}
}

type CommandCollector struct {
	Command, Target string
}

func (collector *CommandCollector) Collect(w Writer) error {
	// #nosec G204
	out, err := exec.Command("/bin/sh", "-c", collector.Command).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "collecting %s", collector.Target)
	}
	return w.Write(collector.Target, out)
}
