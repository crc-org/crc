package results

type Result int

// man systemd.exec for more information
// https://man.archlinux.org/man/core/systemd/systemd.exec.5.en
// search for SERVICE_RESULT for more information
const (
	Unknown Result = iota
	Success
	ExitCode
	Signal
	CoreDump
	Watchdog
	StartLimitHit
	Timeout
	Resources
)

var results = []string{
	"unknown",
	"success",
	"exit-code",
	"signal",
	"core-dump",
	"watchdog",
	"start-limit-hit",
	"timeout",
	"resources",
}

func (r Result) String() string {
	if int(r) >= 0 && int(r) < len(results) {
		return results[r]
	}
	return ""
}

func (r Result) IsSuccess() bool {
	return r == Success
}

// Make sure input is trimmed and lowercase before parsing
func Parse(input string) Result {
	for i, result := range results {
		if result == input {
			return Result(i)
		}
	}
	return Unknown
}
