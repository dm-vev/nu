package config

// Global instance of the configuration
var globalConfig *Config

// Initialize the global configuration
func init() {
	globalConfig = LoadFromEnv()
}

// Get returns the global configuration
func Get() *Config {
	return globalConfig
}

// Reload reloads the configuration from environment variables
func Reload() *Config {
	globalConfig = LoadFromEnv()
	return globalConfig
}
