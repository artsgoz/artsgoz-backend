package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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
// Credentials can be supplied as inline JSON or a file path.
// If the file is unavailable, Application Default Credentials are attempted.
func InitFirebase(credentialsPath string, projectID string) (FirebaseClient, error) {
	ctx := context.Background()

	if projectID == "" {
		return nil, fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}
	opts, err := firebaseCredentialOptions(credentialsPath, os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"))
	if err != nil {
		return nil, err
	}

	cfg := &firebase.Config{}
	if projectID != "" {
		cfg.ProjectID = projectID
	}

	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize Firebase app; check FIREBASE_PROJECT_ID and Firebase Admin credentials")
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize Firebase authentication; configure Firebase Admin credentials using FIREBASE_SERVICE_ACCOUNT_JSON, a readable FIREBASE_CREDENTIALS file, or Application Default Credentials")
	}

	log.Printf("[Firebase] Firebase Auth client initialized successfully")
	return &realFirebaseClient{client: authClient}, nil
}

// Credential contents and paths can contain secrets or mistaken inline JSON.
// Never include either value, or SDK credential parsing errors, in logs/errors.
func firebaseCredentialOptions(credentialsPath, inlineJSON string) ([]option.ClientOption, error) {
	raw := strings.TrimSpace(inlineJSON)
	path := strings.TrimSpace(credentialsPath)
	// Accept existing deployments that put service-account JSON in FIREBASE_CREDENTIALS.
	if raw == "" && strings.HasPrefix(path, "{") {
		raw = path
	}
	if raw != "" {
		var credentials struct {
			Type        string `json:"type"`
			ProjectID   string `json:"project_id"`
			PrivateKey  string `json:"private_key"`
			ClientEmail string `json:"client_email"`
		}
		if json.Unmarshal([]byte(raw), &credentials) != nil || credentials.Type != "service_account" ||
			credentials.ProjectID == "" || credentials.PrivateKey == "" || credentials.ClientEmail == "" {
			return nil, fmt.Errorf("invalid Firebase Admin service-account JSON; provide the complete service-account JSON in FIREBASE_SERVICE_ACCOUNT_JSON")
		}
		return []option.ClientOption{option.WithCredentialsJSON([]byte(raw))}, nil
	}
	if path != "" {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			log.Print("[Firebase] Using service-account credentials file")
			return []option.ClientOption{option.WithCredentialsFile(path)}, nil
		}
		log.Print("[Firebase] Credentials file unavailable; attempting Application Default Credentials")
	}
	return nil, nil
}
