package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

const minimumAccessTokenInactivityTimeout = 300

type TokenValidationOptions struct {
	AccessTokenInactivityTimeout time.Duration
	APIAudiences                 []string
}

func NewTokenValidationOptions() *TokenValidationOptions {
	return &TokenValidationOptions{}
}

func (o *TokenValidationOptions) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&o.AccessTokenInactivityTimeout, "accesstoken-inactivity-timeout", 0, ""+
		"defines the default token inactivity timeout for tokens granted by any client. "+
		"The value represents the maximum amount of time that can occur between "+
		"consecutive uses of the token. Tokens become invalid if they are not "+
		"used within this temporal window (or within the temporal window "+
		"specified by an oauthclient). The user will need to acquire a new "+
		"token to regain access once a token times out.\n"+
		"Valid values are integer values:\n"+
		"\tx = 0  Tokens never time out (default)\n"+
		"\tx > 0  Tokens time out if there is no activity for x seconds",
	)
	fs.StringSliceVar(&o.APIAudiences, "api-audiences", o.APIAudiences, ""+
		"Identifiers of the API. The service account token authenticator will validate that "+
		"tokens used against the API are bound to at least one of these audiences. If the "+
		"--service-account-issuer flag is configured and this flag is not, this field "+
		"defaults to a single element list containing the issuer URL.")
}

func (o *TokenValidationOptions) Validate() []error {
	errs := []error{}

	errs = append(errs, validateAccessTokenInactivityTimeout(o.AccessTokenInactivityTimeout)...)

	return errs
}

func validateAccessTokenInactivityTimeout(timeout time.Duration) []error {
	errs := []error{}

	// int32 will always round down to units, but that's ok
	timeoutSeconds := int32(timeout.Seconds())
	if timeoutSeconds < 0 || (timeoutSeconds > 0 && timeoutSeconds < minimumAccessTokenInactivityTimeout) {
		errs = append(errs, fmt.Errorf("accesstoken-inactivity-timeout must either be 0 or greater than %d", minimumAccessTokenInactivityTimeout))
	}

	return errs
}
