package util

import (
	"context"

	"google.golang.org/grpc/credentials"
	"golang.org/x/oauth2"
)

type insecurePerRPCCredentials struct {
	token *oauth2.Token
}

func (ic insecurePerRPCCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": ic.token.Type() + " " + ic.token.AccessToken,
	}, nil
}

func (ic insecurePerRPCCredentials) RequireTransportSecurity() bool {
	//overwrite
	return false
}

func NewInsecurePerRPCCredentials(token string) credentials.PerRPCCredentials {
	oathToken := &oauth2.Token{
		AccessToken: token,
	}
	return &insecurePerRPCCredentials{token: oathToken}
}