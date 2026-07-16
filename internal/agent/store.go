package agent

import (
	"fmt"
	"time"

	agentconfig "nu/internal/agent/config"
	"nu/internal/data/storage"
)

// createImageStorageFromConfig creates an image storage backend from YAML configuration
func createImageStorageFromConfig(config *agentconfig.ImageStorageYAML) (storage.Storage, error) {
	if config == nil {
		return nil, nil
	}

	storageType := config.Type
	if storageType == "" {
		// Infer from which config is provided
		if config.Local != nil {
			storageType = "local"
		} else if config.GCS != nil {
			storageType = "gcs"
		} else {
			storageType = "local" // Default to local
		}
	}

	switch storageType {
	case "local":
		localCfg := storage.LocalStorageConfig{}
		if config.Local != nil {
			localCfg.Path = config.Local.Path
			localCfg.BaseURL = config.Local.BaseURL
		}
		return storage.NewLocal(localCfg)

	case "gcs":
		if config.GCS == nil {
			return nil, fmt.Errorf("GCS storage configuration is required when type is 'gcs'")
		}
		// Debug: log the credentials being passed
		fmt.Printf("[createImageStorageFromConfig] GCS config: bucket=%s, prefix=%s, creds_json_len=%d, creds_file=%s\n",
			config.GCS.Bucket, config.GCS.Prefix, len(config.GCS.CredentialsJSON), config.GCS.CredentialsFile)
		gcsCfg := storage.GCSStorageConfig{
			Bucket:          config.GCS.Bucket,
			Prefix:          config.GCS.Prefix,
			CredentialsFile: config.GCS.CredentialsFile,
			CredentialsJSON: config.GCS.CredentialsJSON,
		}
		// Parse signed URL expiration duration
		if config.GCS.SignedURLExpiration != "" {
			duration, err := time.ParseDuration(config.GCS.SignedURLExpiration)
			if err != nil {
				return nil, fmt.Errorf("invalid signed_url_expiration format: %w", err)
			}
			gcsCfg.SignedURLExpiration = duration
			gcsCfg.UseSignedURLs = true
		}
		return storage.NewGCS(gcsCfg)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s (only 'local' and 'gcs' are supported)", storageType)
	}
}
