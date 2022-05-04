package internalversion

import (
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"

	oauthv1 "github.com/openshift/api/oauth/v1"
	oauthapi "github.com/openshift/oauth-apiserver/pkg/oauth/apis/oauth"
	"github.com/openshift/oauth-apiserver/pkg/printers"
)

func AddOAuthOpenShiftHandler(h printers.PrintHandler) {
	addOAuthClient(h)
	addOAuthAccessToken(h)
	addUserOAuthAccessToken(h)
	addOAuthAuthorizeToken(h)
	addOAuthClientAuthorization(h)
}

// formatResourceName receives a resource kind, name, and boolean specifying
// whether or not to update the current name to "kind/name"
func formatResourceName(kind schema.GroupKind, name string, withKind bool) string {
	if !withKind || kind.Empty() {
		return name
	}

	return strings.ToLower(kind.String()) + "/" + name
}

func addOAuthClient(h printers.PrintHandler) {
	// oauthClientColumns              = []string{"Name", "Secret", "WWW-Challenge", "Token-Max-Age", "Redirect URIs"}
	oauthClientColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Secret", Type: "string", Description: oauthv1.OAuthClient{}.SwaggerDoc()["secret"]},
		{Name: "WWW-Challenge", Type: "bool", Description: oauthv1.OAuthClient{}.SwaggerDoc()["respondWithChallenges"]},
		{Name: "Token-Max-Age", Type: "string", Description: oauthv1.OAuthClient{}.SwaggerDoc()["accessTokenMaxAgeSeconds"]},
		{Name: "Redirect URIs", Type: "string", Description: oauthv1.OAuthClient{}.SwaggerDoc()["redirectURIs"]},
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthClient); err != nil {
		panic(err)
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthClientList); err != nil {
		panic(err)
	}
}

func printOAuthClient(oauthClient *oauthapi.OAuthClient, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: oauthClient},
	}

	var maxAge string
	switch {
	case oauthClient.AccessTokenMaxAgeSeconds == nil:
		maxAge = "default"
	case *oauthClient.AccessTokenMaxAgeSeconds == 0:
		maxAge = "unexpiring"
	default:
		duration := time.Duration(*oauthClient.AccessTokenMaxAgeSeconds) * time.Second
		maxAge = duration.String()
	}

	row.Cells = append(row.Cells, oauthClient.Name, oauthClient.Secret, oauthClient.RespondWithChallenges, maxAge, strings.Join(oauthClient.RedirectURIs, ","))

	return []metav1.TableRow{row}, nil
}

func printOAuthClientList(oauthClientList *oauthapi.OAuthClientList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(oauthClientList.Items))
	for i := range oauthClientList.Items {
		r, err := printOAuthClient(&oauthClientList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func addOAuthClientAuthorization(h printers.PrintHandler) {
	oauthClientColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "User Name", Type: "string", Format: "name", Description: oauthv1.OAuthClientAuthorization{}.SwaggerDoc()["userName"]},
		{Name: "Client Name", Type: "string", Format: "name", Description: oauthv1.OAuthClientAuthorization{}.SwaggerDoc()["clientName"]},
		{Name: "Scopes", Type: "string", Description: oauthv1.OAuthClientAuthorization{}.SwaggerDoc()["scopes"]},
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthClientAuthorization); err != nil {
		panic(err)
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthClientAuthorizationList); err != nil {
		panic(err)
	}
}

func printOAuthClientAuthorization(oauthClientAuthorization *oauthapi.OAuthClientAuthorization, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: oauthClientAuthorization},
	}
	row.Cells = append(row.Cells,
		oauthClientAuthorization.Name,
		oauthClientAuthorization.UserName,
		oauthClientAuthorization.ClientName,
		strings.Join(oauthClientAuthorization.Scopes, ","),
	)
	return []metav1.TableRow{row}, nil
}

func printOAuthClientAuthorizationList(oauthClientAuthorizationList *oauthapi.OAuthClientAuthorizationList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(oauthClientAuthorizationList.Items))
	for i := range oauthClientAuthorizationList.Items {
		r, err := printOAuthClientAuthorization(&oauthClientAuthorizationList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func addOAuthAccessToken(h printers.PrintHandler) {
	oauthClientColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "User Name", Type: "string", Format: "name", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["userName"]},
		{Name: "Client Name", Type: "string", Format: "name", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["clientName"]},
		{Name: "Created", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "Expires", Type: "string", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["expiresIn"]},
		{Name: "Redirect URI", Type: "string", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["redirectURI"]},
		{Name: "Scopes", Type: "string", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["scopes"]},
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthAccessToken); err != nil {
		panic(err)
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthAccessTokenList); err != nil {
		panic(err)
	}
}

func addUserOAuthAccessToken(h printers.PrintHandler) {
	userOAuthTokenColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "Client Name", Type: "string", Format: "name", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["clientName"]},
		{Name: "Created", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "Expires", Type: "string", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["expiresIn"]},
		{Name: "Redirect URI", Type: "string", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["redirectURI"]},
		{Name: "Scopes", Type: "string", Description: oauthv1.OAuthAccessToken{}.SwaggerDoc()["scopes"]},
	}
	if err := h.TableHandler(userOAuthTokenColumnsDefinitions, printUserOAuthAccessToken); err != nil {
		panic(err)
	}
	if err := h.TableHandler(userOAuthTokenColumnsDefinitions, printUserOAuthAccessTokenList); err != nil {
		panic(err)
	}
}

func printOAuthAccessToken(oauthAccessToken *oauthapi.OAuthAccessToken, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: oauthAccessToken},
	}
	created := oauthAccessToken.CreationTimestamp
	expires := "never"
	if oauthAccessToken.ExpiresIn > 0 {
		expires = created.Add(time.Duration(oauthAccessToken.ExpiresIn) * time.Second).String()
	}
	row.Cells = append(row.Cells,
		oauthAccessToken.Name,
		oauthAccessToken.UserName,
		oauthAccessToken.ClientName,
		translateTimestampSince(created),
		expires,
		oauthAccessToken.RedirectURI,
		strings.Join(oauthAccessToken.Scopes, ","),
	)
	return []metav1.TableRow{row}, nil
}

func printUserOAuthAccessToken(personalAccessToken *oauthapi.UserOAuthAccessToken, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: personalAccessToken},
	}
	created := personalAccessToken.CreationTimestamp
	expires := "never"
	if personalAccessToken.ExpiresIn > 0 {
		expires = created.Add(time.Duration(personalAccessToken.ExpiresIn) * time.Second).String()
	}
	row.Cells = append(row.Cells,
		personalAccessToken.Name,
		personalAccessToken.ClientName,
		translateTimestampSince(created),
		expires,
		personalAccessToken.RedirectURI,
		strings.Join(personalAccessToken.Scopes, ","),
	)
	return []metav1.TableRow{row}, nil
}

func printOAuthAccessTokenList(oauthAccessTokenList *oauthapi.OAuthAccessTokenList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(oauthAccessTokenList.Items))
	for i := range oauthAccessTokenList.Items {
		r, err := printOAuthAccessToken(&oauthAccessTokenList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printUserOAuthAccessTokenList(personalAccessTokenList *oauthapi.UserOAuthAccessTokenList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(personalAccessTokenList.Items))
	for i := range personalAccessTokenList.Items {
		r, err := printUserOAuthAccessToken(&personalAccessTokenList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil

}

func addOAuthAuthorizeToken(h printers.PrintHandler) {
	oauthClientColumnsDefinitions := []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: metav1.ObjectMeta{}.SwaggerDoc()["name"]},
		{Name: "User Name", Type: "string", Format: "name", Description: oauthv1.OAuthAuthorizeToken{}.SwaggerDoc()["userName"]},
		{Name: "Client Name", Type: "string", Format: "name", Description: oauthv1.OAuthAuthorizeToken{}.SwaggerDoc()["userName"]},
		{Name: "Created", Type: "string", Description: metav1.ObjectMeta{}.SwaggerDoc()["creationTimestamp"]},
		{Name: "Expires", Type: "string", Description: oauthv1.OAuthAuthorizeToken{}.SwaggerDoc()["expiresIn"]},
		{Name: "Redirect URI", Type: "string", Description: oauthv1.OAuthAuthorizeToken{}.SwaggerDoc()["redirectURI"]},
		{Name: "Scopes", Type: "string", Description: oauthv1.OAuthAuthorizeToken{}.SwaggerDoc()["scopes"]},
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthAuthorizeToken); err != nil {
		panic(err)
	}
	if err := h.TableHandler(oauthClientColumnsDefinitions, printOAuthAuthorizeTokenList); err != nil {
		panic(err)
	}
}

func printOAuthAuthorizeToken(oauthAuthorizeToken *oauthapi.OAuthAuthorizeToken, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: oauthAuthorizeToken},
	}
	created := oauthAuthorizeToken.CreationTimestamp
	expires := "never"
	if oauthAuthorizeToken.ExpiresIn > 0 {
		expires = created.Add(time.Duration(oauthAuthorizeToken.ExpiresIn) * time.Second).String()
	}
	row.Cells = append(row.Cells,
		oauthAuthorizeToken.Name,
		oauthAuthorizeToken.UserName,
		oauthAuthorizeToken.ClientName,
		translateTimestampSince(created),
		expires,
		oauthAuthorizeToken.RedirectURI,
		strings.Join(oauthAuthorizeToken.Scopes, ","),
	)
	return []metav1.TableRow{row}, nil
}

func printOAuthAuthorizeTokenList(oauthAuthorizeTokenList *oauthapi.OAuthAuthorizeTokenList, options printers.GenerateOptions) ([]metav1.TableRow, error) {
	rows := make([]metav1.TableRow, 0, len(oauthAuthorizeTokenList.Items))
	for i := range oauthAuthorizeTokenList.Items {
		r, err := printOAuthAuthorizeToken(&oauthAuthorizeTokenList.Items[i], options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
