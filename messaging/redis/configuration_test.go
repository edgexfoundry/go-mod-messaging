package redis

import (
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg"
	"github.com/stretchr/testify/assert"
)

func TestBuilderMethods(t *testing.T) {
	tests := []struct {
		name           string
		builder        *redisOptionalConfigurationBuilder
		expectedValues map[string]string
	}{
		{
			name:           "Password",
			builder:        NewRedisOptionalConfigurationBuilder().Password("MyPassword"),
			expectedValues: map[string]string{pkg.Password: "MyPassword"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observedConfig := test.builder.Build()

			for key, value := range test.expectedValues {
				assert.Equal(t, value, observedConfig[key])
			}

		})
	}
}
