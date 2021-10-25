package impersonatingclient

import (
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/user"
	kclientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
	"k8s.io/client-go/util/flowcontrol"
)

// NewImpersonatingConfig wraps the config's transport to impersonate a user, including user, groups, and scopes
func NewImpersonatingConfig(user user.Info, config restclient.Config) restclient.Config {
	oldWrapTransport := config.WrapTransport
	if oldWrapTransport == nil {
		oldWrapTransport = func(rt http.RoundTripper) http.RoundTripper { return rt }
	}
	newConfig := transport.ImpersonationConfig{
		UserName: user.GetName(),
		Groups:   user.GetGroups(),
		Extra:    user.GetExtra(),
	}
	config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return transport.NewImpersonatingRoundTripper(newConfig, oldWrapTransport(rt))
	}
	return config
}

// NewImpersonatingKubernetesClientset returns a Kubernetes clientset that will impersonate a user, including user, groups, and scopes
func NewImpersonatingKubernetesClientset(user user.Info, config restclient.Config) (kclientset.Interface, error) {
	impersonatingConfig := NewImpersonatingConfig(user, config)
	return kclientset.NewForConfig(&impersonatingConfig)
}

// impersonatingRESTClient sets impersonating user, groups, and scopes headers per request
type impersonatingRESTClient struct {
	user     user.Info
	delegate restclient.Interface
}

func NewImpersonatingRESTClient(user user.Info, client restclient.Interface) restclient.Interface {
	return &impersonatingRESTClient{user: user, delegate: client}
}

// Verb does the impersonation per request by setting the proper headers
func (c *impersonatingRESTClient) impersonate(req *restclient.Request) *restclient.Request {
	req.SetHeader(transport.ImpersonateUserHeader, c.user.GetName())
	req.SetHeader(transport.ImpersonateGroupHeader, c.user.GetGroups()...)
	for k, vv := range c.user.GetExtra() {
		req.SetHeader(transport.ImpersonateUserExtraHeaderPrefix+headerKeyEscape(k), vv...)
	}
	return req
}

func (c *impersonatingRESTClient) Verb(verb string) *restclient.Request {
	return c.impersonate(c.delegate.Verb(verb))
}

func (c *impersonatingRESTClient) Post() *restclient.Request {
	return c.impersonate(c.delegate.Post())
}

func (c *impersonatingRESTClient) Put() *restclient.Request {
	return c.impersonate(c.delegate.Put())
}

func (c *impersonatingRESTClient) Patch(pt types.PatchType) *restclient.Request {
	return c.impersonate(c.delegate.Patch(pt))
}

func (c *impersonatingRESTClient) Get() *restclient.Request {
	return c.impersonate(c.delegate.Get())
}

func (c *impersonatingRESTClient) Delete() *restclient.Request {
	return c.impersonate(c.delegate.Delete())
}

func (c *impersonatingRESTClient) APIVersion() schema.GroupVersion {
	return c.delegate.APIVersion()
}

func (c *impersonatingRESTClient) GetRateLimiter() flowcontrol.RateLimiter {
	return c.delegate.GetRateLimiter()
}

// the below header-escaping code is copied from k8s.io/client-go roundtrippers
func legalHeaderByte(b byte) bool {
	return int(b) < len(legalHeaderKeyBytes) && legalHeaderKeyBytes[b]
}

func shouldEscape(b byte) bool {
	// url.PathUnescape() returns an error if any '%' is not followed by two
	// hexadecimal digits, so we'll intentionally encode it.
	return !legalHeaderByte(b) || b == '%'
}

func headerKeyEscape(key string) string {
	buf := strings.Builder{}
	for i := 0; i < len(key); i++ {
		b := key[i]
		if shouldEscape(b) {
			// %-encode bytes that should be escaped:
			// https://tools.ietf.org/html/rfc3986#section-2.1
			fmt.Fprintf(&buf, "%%%02X", b)
			continue
		}
		buf.WriteByte(b)
	}
	return buf.String()
}

// legalHeaderKeyBytes was copied from net/http/lex.go's isTokenTable.
// See https://httpwg.github.io/specs/rfc7230.html#rule.token.separators
var legalHeaderKeyBytes = [127]bool{
	'%':  true,
	'!':  true,
	'#':  true,
	'$':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}
