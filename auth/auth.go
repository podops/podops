package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/txsvc/stdlib/v2/settings"
)

const (
	// default scopes
	ScopeRead  = "api:read"
	ScopeWrite = "api:write"
	ScopeAdmin = "api:admin"
)

var (
	// ErrNotAuthorized indicates that the API caller is not authorized
	ErrNotAuthorized     = errors.New("not authorized")
	ErrAlreadyAuthorized = errors.New("already authorized")

	// ErrNoToken indicates that no bearer token was provided
	ErrNoToken = errors.New("no token provided")
	// ErrNoScope indicates that no scope was provided
	ErrNoScope = errors.New("no scope provided")

	// different types of lookup tables
	tokenToAuth map[string]*settings.DialSettings
	idToAuth    map[string]*settings.DialSettings
)

func init() {
	tokenToAuth = make(map[string]*settings.DialSettings)
	idToAuth = make(map[string]*settings.DialSettings)
}

func RegisterAuthorization(cfg *settings.DialSettings) {
	tokenToAuth[cfg.Credentials.Token] = cfg
	idToAuth[namedKey(cfg.Credentials.ProjectID, cfg.Credentials.UserID)] = cfg
}

func LookupAuthorization(ctx context.Context, realm, userid string) (*settings.DialSettings, error) {
	if a, ok := idToAuth[namedKey(realm, userid)]; ok {
		return a, nil
	}
	return nil, nil
}

// FindAuthorizationByToken looks for an authorization by the token
func FindAuthorizationByToken(ctx context.Context, token string) (*settings.DialSettings, error) {
	if token == "" {
		return nil, ErrNoToken
	}
	if a, ok := tokenToAuth[token]; ok {
		return a, nil
	}
	return nil, nil
}

// CheckAuthorization relies on the presence of a bearer token and validates the
// matching authorization against a list of requested scopes.
// If everything checks out, the function returns the authorization or an error otherwise.
func CheckAuthorization(ctx context.Context, c echo.Context, scope string) (*settings.DialSettings, error) {
	token, err := GetBearerToken(c.Request())
	if err != nil {
		return nil, err
	}

	auth, err := FindAuthorizationByToken(ctx, token)
	if err != nil || auth == nil {
		return nil, ErrNotAuthorized
	}

	if hasScope(auth.Scopes, ScopeAdmin) {
		return auth, nil
	}
	if !hasScope(auth.Scopes, scope) {
		return nil, ErrNotAuthorized
	}

	return auth, nil
}

// GetClientID extracts the ClientID from the token
func GetClientID(ctx context.Context, r *http.Request) (string, error) {
	token, err := GetBearerToken(r)
	if err != nil {
		return "", err
	}

	// FIXME optimize this, e.g. implement caching

	auth, err := FindAuthorizationByToken(ctx, token)
	if err != nil {
		return "", err
	}
	if auth == nil {
		return "", ErrNotAuthorized
	}

	return auth.Credentials.UserID, nil
}

// GetBearerToken extracts the bearer token
func GetBearerToken(r *http.Request) (string, error) {

	// FIXME optimize this !!

	auth := r.Header.Get("Authorization")
	if len(auth) == 0 {
		return "", ErrNoToken
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return "", ErrNoToken
	}
	if parts[0] == "Bearer" {
		return parts[1], nil
	}

	return "", ErrNoToken
}

// hasScope FIXME this is a VERY simple implementation
func hasScope(target []string, scope string) bool {

	scopes := strings.Split(scope, ",")
	mustMatch := len(scopes)

	for _, s := range scopes {
		for _, ss := range target {
			if s == ss {
				mustMatch--
				break
			}
		}
	}

	return mustMatch == 0
}

func namedKey(part1, part2 string) string {
	return strings.ToLower(part1 + "." + part2)
}
