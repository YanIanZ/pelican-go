package auth

import (
	"golang.org/x/oauth2"
)

type OAuthProviders struct {
	Discord   *oauth2.Config
	Steam     *oauth2.Config
	Authentik *oauth2.Config
}

func NewOAuthProviders(baseURL string) *OAuthProviders {
	providers := &OAuthProviders{}
	return providers
}
