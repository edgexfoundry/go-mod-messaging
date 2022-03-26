package redis

import "github.com/edgexfoundry/go-mod-messaging/v2/internal/pkg"

type redisOptionalConfigurationBuilder struct {
	options map[string]string
}

// NewRedisOptionalConfigurationBuilder creates a new builder which aids in creating the map that can be used as
// MessageBusConfig.Optional field to provide additional configuration options.
func NewRedisOptionalConfigurationBuilder() *redisOptionalConfigurationBuilder {
	return &redisOptionalConfigurationBuilder{
		options: make(map[string]string),
	}
}

// Build constructs a map with the configured configuration options.
func (r *redisOptionalConfigurationBuilder) Build() map[string]string {
	return r.options
}

// Password adds a password to the optional configuration properties.
func (r *redisOptionalConfigurationBuilder) Password(password string) *redisOptionalConfigurationBuilder {
	r.options[pkg.Password] = password

	return r
}
