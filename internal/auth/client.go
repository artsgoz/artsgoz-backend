package auth

import (
	"context"

	firebaseauth "firebase.google.com/go/v4/auth"
)

// Identity contains only identity data verified by Firebase.
type Identity struct {
	UID   string
	Email string
}

// FirebaseClient is the Firebase functionality used by login and protected routes.
// The Admin SDK's *auth.Client implements this interface.
type FirebaseClient interface {
	VerifyIDTokenAndCheckRevoked(context.Context, string) (*firebaseauth.Token, error)
	RevokeRefreshTokens(context.Context, string) error
}
