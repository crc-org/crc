package libhosty

import (
	"io/ioutil"
	"net"
	"strings"
)

//ParseHostsFile parse a hosts file from the given location.
// error is not nil if something goes wrong
func ParseHostsFile(path string) ([]HostsFileLine, error) {
	byteData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parser(byteData)
}

//ParseHostsFileAsString parse a hosts file from a given string.
// error is not nil if something goes wrong
func ParseHostsFileAsString(stringData string) ([]HostsFileLine, error) {
	bytesData := []byte(stringData)
	return parser(bytesData)
}

func parser(bytesData []byte) ([]HostsFileLine, error) {
	byteDataNormalized := strings.Replace(string(bytesData), "\r\n", "\n", -1)
	fileLines := strings.Split(byteDataNormalized, "\n")
	hostsFileLines := make([]HostsFileLine, len(fileLines))

	// trim leading an trailing whitespace
	for i, l := range fileLines {
		curLine := &hostsFileLines[i]
		curLine.Number = i
		curLine.Raw = l

		// trim line
		curLine.trimed = strings.TrimSpace(l)

		// check if it's an empty line
		if curLine.trimed == "" {
			curLine.Type = LineTypeEmpty
			continue
		}

		// check if line starts with a #
		if strings.HasPrefix(curLine.trimed, "#") {
			// this can be a comment or a commented host line
			// so remove the 1st char (#), trim spaces
			// and try to parse the line as a host line
			noCommentLine := strings.TrimPrefix(curLine.trimed, "#")
			tmpParts := strings.Fields(strings.TrimSpace(noCommentLine))

			// check what we have
			switch len(tmpParts) {
			case 0:
				// empty line, comment line
				curLine.Type = LineTypeComment
				continue
			default:
				// non-empty line, try to parse as address
				address := net.ParseIP(tmpParts[0])

				// if address is nil this line is a comment
				if address == nil {
					curLine.Type = LineTypeComment
					continue
				}
			}

			// otherwise it is a commented line so let's try to parse it as a normal line
			curLine.IsCommented = true
			curLine.trimed = noCommentLine
		}

		// not a comment or empty line so try to parse it
		// check if it contains a comment
		curLineSplit := strings.SplitN(curLine.trimed, "#", 2)
		if len(curLineSplit) > 1 {
			// trim spaces from comments
			curLine.Comment = strings.TrimSpace(curLineSplit[1])
		}

		curLine.trimed = curLineSplit[0]
		curLine.Parts = strings.Fields(curLine.trimed)

		if len(curLine.Parts) > 1 {
			// parse address to ensure we have a valid address line
			tmpIP := net.ParseIP(curLine.Parts[0])
			if tmpIP != nil {

				curLine.Type = LineTypeAddress
				curLine.Address = tmpIP
				// lower case all
				for _, p := range curLine.Parts[1:] {
					curLine.Hostnames = append(curLine.Hostnames, strings.ToLower(p))
				}

				continue
			}
		}

		// if we can't figure out what this line is mark it as unknown
		curLine.Type = LineTypeUnknown
	}

	// normalize slice
	hostsFileLines = hostsFileLines[:]

	return hostsFileLines, nil
}
