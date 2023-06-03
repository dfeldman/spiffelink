package config

import (
	"fmt"

	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Structures to store all the config information
type DatabaseConfig struct {
	Type             string
	ConnectionString string
	SpiffeID         string
}

type OTLPExporterConfig struct {
	Endpoint       string
	Insecure       bool
	Timeout        string
	RetryOnFailure struct {
		Enabled         bool
		InitialInterval string
		MaxInterval     string
		MaxElapsedTime  string
	}
	SendingQueue struct {
		Enabled      bool
		NumConsumers int
		QueueSize    int
	}
}

type OpenTelemetryConfig struct {
	OtlpExporter OTLPExporterConfig
}

type Config struct {
	SpiffeAgentSocketPath string
	Databases             []DatabaseConfig
	OpenTelemetry         OpenTelemetryConfig
}

func setLogLevel(log *logrus.Logger) {
	levelStr := viper.GetString("log.level")

	if levelStr == "" {
		log.SetLevel(logrus.DebugLevel)
		log.Info("No log level specified in configuration, defaulting to DEBUG")
		return
	}

	// Convert the string level to a logrus.Level
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		log.Infof("Could not parse log level %v, defaulting to DEBUG", levelStr)
		// If we couldn't parse the level, default to DEBUG
		level = logrus.DebugLevel
	}
	log.SetLevel(level)
}

func checkSpiffeAgentSocket(log *logrus.Logger, socketPath string) error {
	info, err := os.Stat(socketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("socket does not exist: %v", err)
		}
		return fmt.Errorf("failed to stat socket: %v", err)
	}

	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("not a socket file")
	}

	// Check for read permission
	if info.Mode().Perm()&0600 == 0 {
		return fmt.Errorf("socket file is not readable and writeable")
	}

	log.Infof("Using SPIFFE agent socket: %s", socketPath)

	return nil
}

func ReadConfig(log *logrus.Logger) (Config, []error) {
	var errs []error

	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("spiffelink")
	viper.AddConfigPath("config/")
	viper.AddConfigPath("/etc/spiffelink/")

	// If the user specifies a config path on the command line, use it instead
	if viper.IsSet("config") {
		viper.AddConfigPath(viper.GetString("config"))
	}

	err := viper.ReadInConfig()
	if err != nil {
		errs = append(errs, fmt.Errorf("Unable to find config file: %v", err))
		return Config{}, errs
	}

	return ParseConfig(log)
}

func ParseConfig(log *logrus.Logger) (Config, []error) {
	var config Config
	var errs []error

	err := viper.Unmarshal(&config)
	if err != nil {
		errs = append(errs, fmt.Errorf("unable to parse configuration file, %v", err))
		return Config{}, errs
	}

	setLogLevel(log)

	if config.SpiffeAgentSocketPath == "" {
		errs = append(errs, fmt.Errorf("spiffe agent socket path is empty"))
	}

	err = checkSpiffeAgentSocket(log, config.SpiffeAgentSocketPath)
	if err != nil {
		errs = append(errs, err)
	}

	spiffeIDs := make(map[string]bool)
	for _, db := range config.Databases {
		if db.Type == "" || db.ConnectionString == "" || db.SpiffeID == "" {
			errs = append(errs, fmt.Errorf("empty fields in database configuration"))
		}
		if _, exists := spiffeIDs[db.SpiffeID]; exists {
			errs = append(errs, fmt.Errorf("duplicate spiffe ID %s", db.SpiffeID))
		}
		spiffeIDs[db.SpiffeID] = true
	}

	if len(spiffeIDs) == 0 {
		errs = append(errs, fmt.Errorf("no databases given in configuration file"))
	}
	otel := config.OpenTelemetry
	if otel.OtlpExporter.Endpoint == "" {
		log.Debug("no OpenTelemetry exporter specified; telemetry is disabled")
	}

	return config, errs
}
