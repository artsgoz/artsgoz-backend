package auth

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseClient defines the subset of Firebase Authentication methods used by the service.
type FirebaseClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*firebaseauth.Token, error)
	RevokeRefreshTokens(ctx context.Context, uid string) error
}

type realFirebaseClient struct {
	client *firebaseauth.Client
}

func (c *realFirebaseClient) VerifyIDToken(ctx context.Context, idToken string) (*firebaseauth.Token, error) {
	return c.client.VerifyIDTokenAndCheckRevoked(ctx, idToken)
}

func (c *realFirebaseClient) RevokeRefreshTokens(ctx context.Context, uid string) error {
	return c.client.RevokeRefreshTokens(ctx, uid)
}

// InitFirebase initializes the Firebase Admin SDK app and Auth client.
// If the credentials file does not exist, it falls back to Application Default Credentials
// and returns an error when Firebase cannot be initialized.
func InitFirebase(credentialsPath string, projectID string) (FirebaseClient, error) {
	ctx := context.Background()

	if projectID == "" {
		return nil, fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}
	var opts []option.ClientOption

	if raw := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"); raw != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(raw)))
	} else if credentialsPath != "" {
		if _, err := os.Stat(credentialsPath); err == nil {
			opts = append(opts, option.WithCredentialsFile(credentialsPath))
			log.Printf("[Firebase] Using credentials file from: %s", credentialsPath)
		} else {
			log.Printf("[Firebase] Credentials file %q not found. Attempting Application Default Credentials...", credentialsPath)
		}
	}

	cfg := &firebase.Config{}
	if projectID != "" {
		cfg.ProjectID = projectID
	}

	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase auth client: %w", err)
	}

	log.Printf("[Firebase] Firebase Auth client initialized successfully")
	return &realFirebaseClient{client: authClient}, nil
}
