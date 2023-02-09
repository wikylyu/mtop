package mtop

import (
	"fmt"
	"net/url"
)

/*
 * Generate a url like
 * mtop://user:password@example.com:443/quic?proto=mtop-example
 * it's used for easy share
 */
func GenerateMTopURL(username, password, host string, port uint16, stype, proto string) *url.URL {
	u := url.URL{
		Scheme: "mtop",
		User:   url.UserPassword(username, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/" + url.QueryEscape(stype),
	}
	q := u.Query()
	q.Set("proto", url.QueryEscape(proto))
	u.RawQuery = q.Encode()
	return &u
}
