package config

// Configuration is the root struct for config type
type Configuration struct {
	Server   ServerConfiguration
	Database DatabaseConfiguration
}

// ServerConfiguration configuration of server
type ServerConfiguration struct {
	Port   int
	APIKey string
	JWTKey string
}

// DatabaseConfiguration config for mongodb
type DatabaseConfiguration struct {
	Host string
	Port int
}
