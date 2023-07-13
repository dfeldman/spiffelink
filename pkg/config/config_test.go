package config

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertNoErrors(t *testing.T, errs []error) {
	if len(errs) > 0 {
		t.Helper()
		for i, err := range errs {
			fmt.Printf("Error %d in ParseConfig: %v\n", i, err)
		}
		t.FailNow()
	}
}

// TestSetup initializes Logrus for testing
func testSetup(t *testing.T) (*logrus.Logger, *bytes.Buffer, func()) {
	// Create a buffer to store log output
	buffer := &bytes.Buffer{}

	// Create a new Logrus logger instance
	logger := logrus.New()

	// Set the output of the logger to the buffer
	logger.SetOutput(buffer)

	// Optionally, you can set the log level
	logger.SetLevel(logrus.DebugLevel)

	// Cleanup function to restore the standard output
	cleanup := func() {
	}

	return logger, buffer, cleanup
}

func TestParseConfig(t *testing.T) {
	logger, _, cleanup := testSetup(t)
	defer cleanup()
	viper.SetConfigFile("../../test/fixture/config/testConfig1.yaml")
	err := viper.ReadInConfig()
	assert.NoError(t, err)

	c, errs := ParseConfig(logger)

	assertNoErrors(t, errs)

	assert.Equal(t, "agent.sock", c.SpiffeAgentSocketPath)
	db1 := c.Databases[0]
	assert.Equal(t, "my-oracle-database", db1.Name)
	assert.Equal(t, "oracle", db1.Type)
	assert.Equal(t, "localhost:8080", db1.ConnectionString)
	assert.Equal(t, "spiffe://test/x", db1.SpiffeID)

	cleanup()
}

func TestParseDatabaseConfig(t *testing.T) {
	for _, tt := range []struct {
		name        string
		config      *DatabaseConfig
		expectError string
	}{
		{
			name: "no error",
			config: &DatabaseConfig{
				Name:             "valid",
				Type:             "oracle",
				ConnectionString: "postgres://username:password@localhost:8000",
				SpiffeID:         "spiffe://test.com/test",
			},
			expectError: "",
		},
		{
			name: "empty name",
			config: &DatabaseConfig{
				Name:             "",
				Type:             "oracle",
				ConnectionString: "postgres://username:password@localhost:8000",
				SpiffeID:         "spiffe://test.com/test",
			},
			expectError: "empty fields in database configuration",
		},
		{
			name: "empty type",
			config: &DatabaseConfig{
				Name:             "valid",
				Type:             "",
				ConnectionString: "postgres://username:password@localhost:8000",
				SpiffeID:         "spiffe://test.com/test",
			},
			expectError: "empty fields in database configuration",
		},
		{
			name: "empty connection string",
			config: &DatabaseConfig{
				Name:             "valid",
				Type:             "oracle",
				ConnectionString: "",
				SpiffeID:         "spiffe://test.com/test",
			},
			expectError: "empty fields in database configuration",
		},
		{
			name: "empty spiffe id",
			config: &DatabaseConfig{
				Name:             "valid",
				Type:             "oracle",
				ConnectionString: "postgres://username:password@localhost:8000",
				SpiffeID:         "",
			},
			expectError: "empty fields in database configuration",
		},
	} {
		logger, _, cleanup := testSetup(t)
		t.Run(tt.name, func(t *testing.T) {
			errs := parseDatabaseConfigFields(logger, tt.config)
			if tt.expectError != "" {
				require.NotEmpty(t, errs)
				require.Error(t, errs[0], tt.expectError)
				return
			}

			require.Empty(t, errs)
		})
		cleanup()
	}
}

// func TestValidateConfig(t *testing.T) {
// 	for _, tt := range []struct {
// 		name        string
// 		config      *Config
// 		expectError string
// 	}{
// 		{
// 			name: "no error",
// 			config: &Config{
// 				AgentAddress:       "path",
// 				SvidFileName:       "cert.pem",
// 				SvidKeyFileName:    "key.pem",
// 				SvidBundleFileName: "bundle.pem",
// 			},
// 		},
// 		{
// 			name: "no address",
// 			config: &Config{
// 				SvidFileName:       "cert.pem",
// 				SvidKeyFileName:    "key.pem",
// 				SvidBundleFileName: "bundle.pem",
// 			},
// 			expectError: "agentAddress is required",
// 		},
// 		{
// 			name: "no SVID file",
// 			config: &Config{
// 				AgentAddress:       "path",
// 				SvidKeyFileName:    "key.pem",
// 				SvidBundleFileName: "bundle.pem",
// 			},
// 			expectError: "svidFileName is required",
// 		},
// 		{
// 			name: "no key file",
// 			config: &Config{
// 				AgentAddress:       "path",
// 				SvidFileName:       "cert.pem",
// 				SvidBundleFileName: "bundle.pem",
// 			},
// 			expectError: "svidKeyFileName is required",
// 		},
// 		{
// 			name: "no bundle file",
// 			config: &Config{
// 				AgentAddress:    "path",
// 				SvidFileName:    "cert.pem",
// 				SvidKeyFileName: "key.pem",
// 			},
// 			expectError: "svidBundleFileName is required",
// 		},
// 	} {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := ValidateConfig(tt.config)
// 			if tt.expectError != "" {
// 				require.Error(t, err, tt.expectError)
// 				return
// 			}

// 			require.NoError(t, err)
// 		})
// 	}
// }
