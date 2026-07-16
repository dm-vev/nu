package storage

import (
	"encoding/base64"
	"fmt"

	"google.golang.org/api/option"
)

func gcsCredentialsClientOptions(cfg GCSStorageConfig) []option.ClientOption {
	// CredentialsJSON takes precedence over CredentialsFile
	if cfg.CredentialsJSON != "" {
		credentialsJSON := parseGCSCredentialsJSON(cfg.CredentialsJSON)
		fmt.Printf("[gcs] Using credentials JSON (length=%d, starts_with_brace=%v)\n",
			len(credentialsJSON), len(credentialsJSON) > 0 && credentialsJSON[0] == '{')
		//nolint:staticcheck // SA1019: WithCredentialsJSON is deprecated but needed for programmatic credentials
		return []option.ClientOption{option.WithCredentialsJSON([]byte(credentialsJSON))}
	}
	if cfg.CredentialsFile != "" {
		fmt.Printf("[gcs] Using credentials file: %s\n", cfg.CredentialsFile)
		//nolint:staticcheck // SA1019: WithCredentialsFile is deprecated but needed for file-based credentials
		return []option.ClientOption{option.WithCredentialsFile(cfg.CredentialsFile)}
	}

	fmt.Println("[gcs] No credentials provided, using Application Default Credentials")
	return nil
}

// parseCredentialsJSON parses credentials that may be base64 encoded or raw JSON
func parseGCSCredentialsJSON(creds string) string {
	// Try to decode as base64 first
	if decoded, err := base64.StdEncoding.DecodeString(creds); err == nil {
		// Check if decoded content looks like JSON
		if len(decoded) > 0 && decoded[0] == '{' {
			return string(decoded)
		}
	}
	// Return as-is (assuming it's raw JSON)
	return creds
}
