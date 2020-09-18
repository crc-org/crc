package cmd

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/machine/libmachine/ssh"
	"github.com/spf13/cobra"
)

var diagnoseFile string

func init() {
	diagnoseCmd.Flags().StringVar(&diagnoseFile, "o", "diagnose.zip", "output file")
	rootCmd.AddCommand(diagnoseCmd)
}

var diagnoseCmd = &cobra.Command{
	Use:    "diagnose",
	Hidden: true,
	Short:  "Gather information about this CodeReady Containers installation",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDiagnose(); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runDiagnose() error {
	zipFile, err := os.Create("diagnose.zip")
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
		file("/etc/hosts"),
		vmFile("/etc/resolv.conf"),
		vmFile("/etc/hosts"),
		vmCommand("journalctl -n 1000 --no-pager", "journalctl"),
		vmCommand("ifconfig", "ifconfig"),
		vmCommand("dmesg", "dmesg"),
		vmCommand("sudo podman ps", "podman-ps"),
		vmCommand("sudo crictl ps", "crictl-ps"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc get co", "clusteroperators"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc get nodes", "nodes"),
		vmCommand("KUBECONFIG=/opt/kubeconfig oc get pods --all-namespaces", "pods"),
	}

	var wg sync.WaitGroup
	var writeLock sync.Mutex

	for i := range collectors {
		wg.Add(1)
		collector := collectors[i]
		go func() {
			defer wg.Done()
			bin, err := collector.Collect()
			if err != nil {
				logging.Errorf("error while collecting %s: %s", collector.Name(), err)
				bin = append([]byte(err.Error()), bin...)
			}

			writeLock.Lock()
			defer writeLock.Unlock()

			logging.Infof("collected %s", collector.Name())
			f, err := writer.Create(collector.Name())
			if err != nil {
				logging.Error(err)
				return
			}

			_, err = f.Write(bin)
			if err != nil {
				logging.Error(err)
			}
		}()
	}
	wg.Wait()
	return nil
}

type Collector interface {
	Name() string
	Collect() ([]byte, error)
}

func file(path string) *FileCollector {
	return &FileCollector{Source: path, Target: filepath.Base(path)}
}

type FileCollector struct {
	Source, Target string
}

func (fc *FileCollector) Name() string {
	return fc.Target
}

func (fc *FileCollector) Collect() ([]byte, error) {
	return ioutil.ReadFile(fc.Source)
}

type TreeCollector struct {
	Dir    string
	Target string
}

func (fc *TreeCollector) Name() string {
	return fc.Target
}

func (fc *TreeCollector) Collect() ([]byte, error) {
	var list []string
	err := filepath.Walk(fc.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		list = append(list, fmt.Sprintf("%s: %d %s", path, info.Size(), info.Mode().String()))
		return nil
	})
	return []byte(strings.Join(list, "\n")), err
}

func vmFile(file string) *VMFileCollector {
	return &VMFileCollector{Source: file, Target: path.Join("vm", filepath.Base(file))}
}

type VMFileCollector struct {
	Source, Target string
}

func (fc *VMFileCollector) Name() string {
	return fc.Target
}

func (fc *VMFileCollector) Collect() ([]byte, error) {
	client := machine.NewClient()
	ip, err := client.IP(machine.IPConfig{
		Name: constants.DefaultName,
	})
	if err != nil {
		return nil, err
	}
	ssh, err := ssh.NewClient("core", ip.IP, 22, &ssh.Auth{
		Keys: []string{constants.GetPrivateKeyPath()},
	})
	if err != nil {
		return nil, err
	}
	out, err := ssh.Output(fmt.Sprintf("cat %s", fc.Source))
	return []byte(out), err
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

func (fc *VMCommandCollector) Name() string {
	return path.Join("vm", fc.Target)
}

func (fc *VMCommandCollector) Collect() ([]byte, error) {
	client := machine.NewClient()
	ip, err := client.IP(machine.IPConfig{
		Name: constants.DefaultName,
	})
	if err != nil {
		return nil, err
	}
	ssh, err := ssh.NewClient("core", ip.IP, 22, &ssh.Auth{
		Keys: []string{constants.GetPrivateKeyPath()},
	})
	if err != nil {
		return nil, err
	}
	out, err := ssh.Output(fc.Command)
	return []byte(out), err
}
