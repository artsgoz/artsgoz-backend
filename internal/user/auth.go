package user

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

// FirebaseAuthClient is the Firebase functionality used by login and protected routes.
// The Admin SDK's *auth.Client implements this interface.
type FirebaseAuthClient interface {
	VerifyIDTokenAndCheckRevoked(context.Context, string) (*auth.Token, error)
	RevokeRefreshTokens(context.Context, string) error
}
