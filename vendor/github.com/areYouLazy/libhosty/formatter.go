package libhosty

import (
	"fmt"
	"strings"
)

// lineFormatter return a readable form for the given HostsFileLine object
func lineFormatter(hfl HostsFileLine) string {

	// returns raw if we don't need to edit the line
	// this is for UNKNOWN, EMPTY and COMMENT linetypes
	if hfl.Type < LineTypeAddress {
		return hfl.Raw
	}

	// check if it's a commented line
	if hfl.IsCommented {
		// check if there's a comment for that line
		if len(hfl.Comment) > 0 {
			return fmt.Sprintf("# %-16s %s #%s", hfl.Address, strings.Join(hfl.Hostnames, " "), hfl.Comment)
		}

		return fmt.Sprintf("# %-16s %s", hfl.Address, strings.Join(hfl.Hostnames, " "))
	}

	// return the actual hosts entry
	if len(hfl.Comment) > 0 {
		return fmt.Sprintf("%-16s %s #%s", hfl.Address, strings.Join(hfl.Hostnames, " "), hfl.Comment)
	}

	return fmt.Sprintf("%-16s %s", hfl.Address, strings.Join(hfl.Hostnames, " "))
}
