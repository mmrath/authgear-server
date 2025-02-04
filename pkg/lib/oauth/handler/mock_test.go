package handler_test

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type mockURLsProvider struct{}

func (mockURLsProvider) AuthorizeURL(r protocol.AuthorizationRequest) *url.URL {
	u, _ := url.Parse("https://auth/authorize")
	return u
}

func (mockURLsProvider) FromWebAppURL(r protocol.AuthorizationRequest) *url.URL {
	u, _ := url.Parse("https://auth/from_webapp")
	return u
}

func (mockURLsProvider) AuthenticateURL(opts webapp.AuthenticateURLOptions) (httputil.Result, error) {
	return &httputil.ResultRedirect{URL: "https://auth/authenticate"}, nil
}

type mockAuthzStore struct {
	authzs []oauth.Authorization
}

func (m *mockAuthzStore) Get(userID, clientID string) (*oauth.Authorization, error) {
	for _, a := range m.authzs {
		if a.UserID == userID && a.ClientID == clientID {
			return &a, nil
		}
	}
	return nil, oauth.ErrAuthorizationNotFound
}

func (m *mockAuthzStore) GetByID(id string) (*oauth.Authorization, error) {
	for _, a := range m.authzs {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, oauth.ErrAuthorizationNotFound
}

func (m *mockAuthzStore) Create(authz *oauth.Authorization) error {
	m.authzs = append(m.authzs, *authz)
	return nil
}

func (m *mockAuthzStore) Delete(authz *oauth.Authorization) error {
	n := 0
	for _, a := range m.authzs {
		if a.ID != authz.ID {
			m.authzs[n] = a
			n++
		}
	}
	m.authzs = m.authzs[:n]
	return nil
}

func (m *mockAuthzStore) ResetAll(userID string) error {
	n := 0
	for _, a := range m.authzs {
		if a.UserID != userID {
			m.authzs[n] = a
			n++
		}
	}
	m.authzs = m.authzs[:n]
	return nil
}

func (m *mockAuthzStore) UpdateScopes(authz *oauth.Authorization) error {
	for i, a := range m.authzs {
		if a.ID == authz.ID {
			a.Scopes = authz.Scopes
			a.UpdatedAt = authz.UpdatedAt
			m.authzs[i] = a
		}
	}
	return nil
}

type mockCodeGrantStore struct {
	grants []oauth.CodeGrant
}

func (m *mockCodeGrantStore) GetCodeGrant(codeHash string) (*oauth.CodeGrant, error) {
	for _, g := range m.grants {
		if g.CodeHash == codeHash {
			return &g, nil
		}
	}
	return nil, oauth.ErrGrantNotFound
}

func (m *mockCodeGrantStore) CreateCodeGrant(grant *oauth.CodeGrant) error {
	m.grants = append(m.grants, *grant)
	return nil
}

func (m *mockCodeGrantStore) DeleteCodeGrant(grant *oauth.CodeGrant) error {
	n := 0
	for _, g := range m.grants {
		if g.CodeHash != grant.CodeHash {
			m.grants[n] = g
			n++
		}
	}
	m.grants = m.grants[:n]
	return nil
}

type mockAuthenticationInfoService struct {
	Entry *authenticationinfo.Entry
}

func (m *mockAuthenticationInfoService) Consume(entryID string) (*authenticationinfo.Entry, error) {
	return m.Entry, nil
}

type mockCookieManager struct{}

func (m *mockCookieManager) GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error) {
	return &http.Cookie{}, nil
}

func (m *mockCookieManager) ClearCookie(def *httputil.CookieDef) *http.Cookie {
	return &http.Cookie{}
}
