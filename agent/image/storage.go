package image

import (
	"fmt"
	"time"

	"github.com/dm-vev/nu/agent/config"
	"github.com/dm-vev/nu/internal/data/storage"
	"github.com/dm-vev/nu/internal/data/storage/gcs"
	"github.com/dm-vev/nu/internal/data/storage/local"
)

// createStorageFromConfig creates an image storage backend from YAML configuration.
func createStorageFromConfig(config *config.ImageStorageYAML) (storage.Storage, error) {
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
		localCfg := local.Config{}
		if config.Local != nil {
			localCfg.Path = config.Local.Path
			localCfg.BaseURL = config.Local.BaseURL
		}
		return local.New(localCfg)

	case "gcs":
		if config.GCS == nil {
			return nil, fmt.Errorf("GCS storage configuration is required when type is 'gcs'")
		}
		gcsCfg := gcs.Config{
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
		return gcs.New(gcsCfg)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s (only 'local' and 'gcs' are supported)", storageType)
	}
}
