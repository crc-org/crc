package network

import (
	"fmt"
	"net/url"
)

func URIStringForDisplay(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if u.User != nil {
		return fmt.Sprintf("%s://%s:xxx@%s", u.Scheme, u.User.Username(), u.Host), nil
	}
	return uri, nil
}
