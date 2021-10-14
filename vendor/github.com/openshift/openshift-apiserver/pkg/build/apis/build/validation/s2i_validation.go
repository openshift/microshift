package validation

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var urlSchemeRegexp = regexp.MustCompile("(?i)^[a-z][-a-z0-9+.]*:") // matches scheme: according to RFC3986
var dosDriveRegexp = regexp.MustCompile("(?i)^[a-z]:")
var scpRegexp = regexp.MustCompile("^" +
	"(?:([^@/]*)@)?" + // user@ (optional)
	"([^/]*):" + //            host:
	"(.*)" + //                     path
	"$")

// URLType indicates the type of the URL (see above)
type URLType int

const (
	// URLTypeURL is the URL type (see above)
	URLTypeURL URLType = iota
	// URLTypeSCP is the SCP type (see above)
	URLTypeSCP
	// URLTypeLocal is the local type (see above)
	URLTypeLocal
)

// String returns a string representation of the URLType
func (t URLType) String() string {
	switch t {
	case URLTypeURL:
		return "URLTypeURL"
	case URLTypeSCP:
		return "URLTypeSCP"
	case URLTypeLocal:
		return "URLTypeLocal"
	}
	panic("unknown URLType")
}

// GoString returns a Go string representation of the URLType
func (t URLType) GoString() string {
	return t.String()
}

// URL represents a "Git URL"
type URL struct {
	URL  url.URL
	Type URLType
}

// Parse parses a "Git URL"
func parseGitURL(rawurl string) (*URL, error) {
	if urlSchemeRegexp.MatchString(rawurl) &&
		(runtime.GOOS != "windows" || !dosDriveRegexp.MatchString(rawurl)) {
		u, err := url.Parse(rawurl)
		if err != nil {
			return nil, err
		}
		if u.Scheme == "file" && u.Opaque == "" {
			if u.Host != "" {
				return nil, fmt.Errorf("file url %q has non-empty host %q", rawurl, u.Host)
			}
			if runtime.GOOS == "windows" && (len(u.Path) == 0 || !filepath.IsAbs(u.Path[1:])) {
				return nil, fmt.Errorf("file url %q has non-absolute path %q", rawurl, u.Path)
			}
		}

		return &URL{
			URL:  *u,
			Type: URLTypeURL,
		}, nil
	}

	s, fragment := splitOnByte(rawurl, '#')

	if m := scpRegexp.FindStringSubmatch(s); m != nil &&
		(runtime.GOOS != "windows" || !dosDriveRegexp.MatchString(s)) {
		u := &url.URL{
			Host:     m[2],
			Path:     m[3],
			Fragment: fragment,
		}
		if m[1] != "" {
			u.User = url.User(m[1])
		}

		return &URL{
			URL:  *u,
			Type: URLTypeSCP,
		}, nil
	}

	return &URL{
		URL: url.URL{
			Path:     s,
			Fragment: fragment,
		},
		Type: URLTypeLocal,
	}, nil
}

func splitOnByte(s string, c byte) (string, string) {
	if i := strings.IndexByte(s, c); i != -1 {
		return s[:i], s[i+1:]
	}
	return s, ""
}
