package libauth

import (
	"os"
	"strings"
	"unicode/utf8"
)

func getkey(params string) error {
	p, e := os.StartProcess(factotum, []string{"getkey", "-g", params},
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if e != nil {
		return e
	}
	_, e = p.Wait()
	return e
}

var qsep = " \t\r\n"

func qtoken(t, sep string) (tok, rest string) {
	quoting := false

	for len(t) > 0 {
		r, width := utf8.DecodeRuneInString(t)
		if !(quoting || !strings.ContainsRune(sep, r)) {
			break
		}

		if r != '\'' {
			tok += t[:width]
			t = t[width:]
			continue
		}

		/* r is a quote */
		if !quoting {
			quoting = true
			t = t[width:]
			continue
		}

		/* quoting and we're on a quote */
		if len(t) > 1 && t[1] != '\'' {
			quoting = false
			t = t[width:]
			continue
		}

		/* doubled quote; fold one quote into two */
		t = t[width:]
		_, width = utf8.DecodeRuneInString(t)
		tok += t[:width]
	}

	return tok, t
}

func tokenize(s string) []string {
	var ss []string

	for len(s) > 0 {
		// skip ws
		for {
			r, width := utf8.DecodeRuneInString(s)
			if !strings.ContainsRune(qsep, r) {
				break
			}
			s = s[width:]
		}

		tok, rest := qtoken(s, qsep)
		ss = append(ss, tok)
		s = rest
	}

	return ss
}

// given a string of the form 'proto=foo service=bar user=baz', tokenize it into a map.
func attrmap(s string) map[string]string {
	attrmap := make(map[string]string)

	strs := tokenize(s)
	for _, av := range strs {
		a := strings.Split(av, "=")
		if len(a) == 1 {
			attrmap[a[0]] = ""
		} else {
			attrmap[a[0]] = a[1]
		}
	}

	return attrmap
}
