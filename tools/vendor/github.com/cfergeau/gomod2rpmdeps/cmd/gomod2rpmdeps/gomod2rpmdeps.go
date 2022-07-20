package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

type modScanner interface {
	Scan() bool
	Text() string
	Err() error
	io.Closer
}

type govendorScanner struct {
	*exec.Cmd
	scanner *bufio.Scanner
	err     error
}

func govendorScannerNew() *govendorScanner {
	cmd := exec.Command("go", "mod", "vendor", "-v")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return &govendorScanner{err: err}
	}
	if err := cmd.Start(); err != nil {
		return &govendorScanner{err: err}
	}
	scanner := bufio.NewScanner(stderr)

	return &govendorScanner{cmd, scanner, nil}
}

func (govendor *govendorScanner) Scan() bool {
	if govendor.err != nil {
		return false
	}

	return govendor.scanner.Scan()
}

func (govendor *govendorScanner) Text() string {
	if govendor.err != nil {
		return ""
	}

	return govendor.scanner.Text()
}

func (govendor *govendorScanner) Close() error {
	if govendor.err != nil {
		return govendor.err
	}

	return govendor.Wait()
}

func (govendor *govendorScanner) Err() error {
	if govendor.err != nil {
		return govendor.err
	}

	return govendor.scanner.Err()
}

type goModuleInfo struct {
	name          string
	pseudoVersion string
}

type byModuleName []goModuleInfo

func (a byModuleName) Len() int           { return len(a) }
func (a byModuleName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byModuleName) Less(i, j int) bool { return a[i].name < a[j].name }

var errNoSharpPrefix = fmt.Errorf("Parse error - line not starting with '#'")
var errComment = fmt.Errorf("Parse error - line is a comment")
var errEmptyLine = fmt.Errorf("Parse error - line is empty")

func parseModuleLine(line string) (goModuleInfo, error) {
	substrings := strings.Split(line, " ")
	if len(substrings) == 0 {
		return goModuleInfo{}, errEmptyLine
	}
	if substrings[0] != "#" {
		return goModuleInfo{}, errNoSharpPrefix
	}
	if substrings[0] == "##" {
		return goModuleInfo{}, errComment
	}
	switch len(substrings) {
	case 3:
		modInfo := goModuleInfo{
			name:          substrings[1],
			pseudoVersion: substrings[2],
		}
		return modInfo, nil
	case 5:
		if substrings[2] != "=>" {
			return goModuleInfo{}, fmt.Errorf("Parse error - unknown format")
		}
		modInfo := goModuleInfo{
			name:          substrings[3],
			pseudoVersion: substrings[4],
		}
		return modInfo, nil
	case 6:
		if substrings[3] != "=>" {
			return goModuleInfo{}, fmt.Errorf("Parse error - unknown format")
		}
		modInfo := goModuleInfo{
			name:          substrings[4],
			pseudoVersion: substrings[5],
		}
		return modInfo, nil
	default:
		return goModuleInfo{}, fmt.Errorf("Parse error - incorrect number of substrings (%d - expected %d)", len(substrings), 3)
	}
}

func pseudoVersionToRpmVersion(pseudoVersion string) (string, error) {
	// https://golang.org/ref/mod#pseudo-versions

	versionRegexp := regexp.MustCompile("^(v[0-9]+.[0-9]+.[0-9]+)")
	dateRegexp := regexp.MustCompile("([0-9]{8})[0-9]{6}$")
	commitRegexp := regexp.MustCompile("-([0-9a-f]{12})$")
	prereleaseRegexp := regexp.MustCompile("^-(?:0.)?([a-z][a-z0-9]+).*$")

	// v1.2.3-xx+incompatible
	if strings.HasSuffix(pseudoVersion, "+incompatible") {
		pseudoVersion = strings.TrimSuffix(pseudoVersion, "+incompatible")
	}

	// v1.2.3
	var version string
	substrings := versionRegexp.FindStringSubmatch(pseudoVersion)
	if len(substrings) != 2 {
		return "", fmt.Errorf("Failed to parse version substring: %s", pseudoVersion)
	}
	version = strings.TrimPrefix(substrings[1], "v")
	pseudoVersion = pseudoVersion[len(substrings[0]):]
	debug("version: %s (left: %s)\n", version, pseudoVersion)

	if pseudoVersion == "" {
		return version, nil
	}

	// -0123456789abc
	var commit string
	substrings = commitRegexp.FindStringSubmatch(pseudoVersion)
	if substrings != nil {
		if len(substrings) != 2 {
			return "", fmt.Errorf("Failed to parse commit substring: %s", pseudoVersion)
		}
		pseudoVersion = pseudoVersion[:len(pseudoVersion)-len(substrings[0])]
		commit = substrings[1]
	}
	debug("commit: %s (left: %s)\n", commit, pseudoVersion)

	// -20200101121212
	var date string
	substrings = dateRegexp.FindStringSubmatch(pseudoVersion)
	if substrings != nil {
		if len(substrings) != 2 {
			return "", fmt.Errorf("Failed to parse date substring: %s", pseudoVersion)
		}
		pseudoVersion = pseudoVersion[:len(pseudoVersion)-len(substrings[0])]
		date = substrings[1]
	}
	debug("date: %s (left: %s)\n", date, pseudoVersion)

	// parse prerelease, everything between v1.2.3- at the beginining of
	// the version string and the date/commit at the end of the version
	// string
	var prerelease string
	substrings = prereleaseRegexp.FindStringSubmatch(pseudoVersion)
	if substrings != nil {
		if len(substrings) != 2 {
			return "", fmt.Errorf("Failed to parse prerelease substring: %s", pseudoVersion)
		}
		prerelease = substrings[1]
	}
	debug("prelease: %s\n", prerelease)

	var rpmVersion strings.Builder
	rpmVersion.WriteString(version)
	rpmVersion.WriteString("-0")
	if prerelease != "" {
		rpmVersion.WriteString(".")
		rpmVersion.WriteString(prerelease)
	}
	if date != "" && commit != "" {
		rpmVersion.WriteString(".")
		rpmVersion.WriteString(date)
		rpmVersion.WriteString("git")
		rpmVersion.WriteString(commit)
	}

	return rpmVersion.String(), nil
}

func printBundledProvides(vendoredModules []goModuleInfo) {
	sort.Sort(byModuleName(vendoredModules))
	for _, modInfo := range vendoredModules {
		rpmVersion, err := pseudoVersionToRpmVersion(modInfo.pseudoVersion)
		if err == nil {
			fmt.Printf("Provides: bundled(golang(%s)) = %s\n", modInfo.name, rpmVersion)
		} else {
			fmt.Fprintf(os.Stderr, "failed to parse pseudoversion %s for module %s: %v\n", modInfo.pseudoVersion, modInfo.name, err)
			fmt.Printf("Provides: bundled(golang(%s))\n", modInfo.name)
		}
	}
}

func fetchVendoredModules() ([]goModuleInfo, error) {
	scanner := modScanner(govendorScannerNew())

	vendoredModules := []goModuleInfo{}

	for scanner.Scan() {
		line := scanner.Text()
		modInfo, err := parseModuleLine(line)
		if err != nil {
			if errors.Is(err, errNoSharpPrefix) || errors.Is(err, errComment) || errors.Is(err, errEmptyLine) {
				continue
			}
			fmt.Fprintf(os.Stderr, "failed to parse line: %v\n\t%s\n", err, line)
			continue
		}
		vendoredModules = append(vendoredModules, modInfo)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if err := scanner.Close(); err != nil {
		return nil, err
	}
	return vendoredModules, nil
}

func main() {
	vendoredModules, err := fetchVendoredModules()
	if err != nil {
		log.Fatal(err)
	}

	printBundledProvides(vendoredModules)
}

func debug(format string, args ...interface{}) {
	if os.Getenv("GOMOD2RPMDEPS_DEBUG") != "" {
		fmt.Printf(format, args...)
	}
}
