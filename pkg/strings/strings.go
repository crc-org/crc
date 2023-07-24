package strings

import (
	"bufio"
	"strings"
)

func Contains(input []string, match string) bool {
	for _, v := range input {
		if v == match {
			return true
		}
	}
	return false
}

// Split a multi line string in an array of string, one for each line
func SplitLines(input string) []string {
	output := []string{}

	s := bufio.NewScanner(strings.NewReader(input))
	for s.Scan() {
		output = append(output, s.Text())
	}

	return output
}
