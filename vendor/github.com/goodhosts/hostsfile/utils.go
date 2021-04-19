package hostsfile

import (
	"fmt"
)

func itemInSlice(item string, list []string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
}

func removeFromSlice(s string, slice []string) []string {
	for key, value := range slice {
		if value == s {
			return append(slice[:key], slice[key+1:]...)
		}
	}
	return nil
}

//func sliceContainsItem(item string, list []string) bool {
//	for _, i := range list {
//		if strings.Contains(i, item) {
//			return true
//		}
//	}
//
//	return false
//}

func buildRawLine(ip string, hosts []string) string {
	output := ip
	for _, host := range hosts {
		output = fmt.Sprintf("%s %s", output, host)
	}

	return output
}
