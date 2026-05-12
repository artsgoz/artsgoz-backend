package utils

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// AccessSecretVersion fetches the payload for the given secret version from
// GCP Secret Manager. Version may be "latest" or a numeric version string.
func AccessSecretVersion(ctx context.Context, projectID, secretID, version string) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("secretmanager: new client: %w", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, version),
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("secretmanager: access %q: %w", req.Name, err)
	}

	return string(result.Payload.Data), nil
}
