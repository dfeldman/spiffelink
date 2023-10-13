package config

import (
	"fmt"
	"regexp"

	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/dfeldman/spiffelink/pkg/slerror"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const DEFAULT_TIMEOUT_SECONDS = 300

type ShellContextConfig struct {
	ShellType   string
	ContainerID string
}

// Structures to store all the config information
type DatabaseConfig struct {
	Name             string
	Type             string
	ConnectionString string
	SpiffeID         string
	ParsedSpiffeID   spiffeid.ID
	Timeout          int
	Shell            ShellContextConfig
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

var debugMode = true

func isValidName(s string) bool {
	r := regexp.MustCompile("^[a-zA-Z0-9]+$")
	return (len(s) < 255) && r.MatchString(s)
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
			return slerror.AgentSocketPathDoesNotExistError(log, socketPath)
		}
		return slerror.AgentSocketPathStatFailedError(log, socketPath)
	}

	if info.Mode()&os.ModeSocket == 0 {
		return slerror.AgentSocketPathInvalidError(log, socketPath)
	}

	// Check for read permission
	if info.Mode().Perm()&0600 == 0 {
		return slerror.InvalidPermissionsAgentSocketPathError(log, socketPath)
	}

	log.Infof("Using SPIFFE agent socket: %s", socketPath)

	return nil
}

func ReadConfig(log *logrus.Logger) (Config, []slerror.SLError) {
	var errs []slerror.SLError

	viper.SetConfigType("yaml")
	//viper.AddConfigPath(".")
	viper.SetConfigName("spiffelink")
	viper.AddConfigPath("config/")
	viper.AddConfigPath("/etc/spiffelink/")

	// If the user specifies a config path on the command line, use it instead
	if viper.IsSet("config") {
		viper.AddConfigPath(viper.GetString("config"))
	}

	err := viper.ReadInConfig()
	if err != nil {
		errs = append(errs, slerror.UnableToReadConfigFileError(log, err))
		return Config{}, errs
	}

	return ParseConfig(log)
}

// Checks the fields in the DatabaseConfig structure and parses out the contents of the SpiffeID field.
func parseDatabaseConfigFields(log *logrus.Logger, db *DatabaseConfig) []error {
	var errs []error
	// Check there are no empty fields
	if db.Name == "" {
		errs = append(errs, slerror.DatabaseNameEmptyError(log))
	}
	if db.Type == "" {
		errs = append(errs, slerror.DatabaseTypeEmptyError(log))
	}
	if db.ConnectionString == "" {
		errs = append(errs, slerror.ConnectionStringEmptyError(log))
	}
	if db.SpiffeID == "" {
		errs = append(errs, slerror.SpiffeIDEmptyError(log))
	}
	if db.Timeout == 0 {
		db.Timeout = DEFAULT_TIMEOUT_SECONDS
	}
	// TODO We need to check for errors in the Shell configuration here.
	if db.Shell.ShellType == "" {
		db.Shell.ShellType = "LocalShell"
	}

	// Check the database name is alphanumeric
	if !isValidName(db.Name) {
		errs = append(errs, fmt.Errorf("invalid database name %s", db.Name))
	}

	// Check that the Type field is a supported database type
	switch db.Type {
	case "oracle":
		break
	case "dummy":
		break
	default:
		errs = append(errs, slerror.InvalidDatabaseType(log))
	}

	parsedSpiffeId, err := spiffeid.FromString(db.SpiffeID)
	if err != nil {
		errs = append(errs, fmt.Errorf("cannot parse spiffe ID %s", db.SpiffeID))
	}
	db.ParsedSpiffeID = parsedSpiffeId
	return errs
}

// Parse the config file. It is automatically read from a location set with viper.AddConfigPath.
func ParseConfig(log *logrus.Logger) (Config, []slerror.SLError) {
	var config Config
	var errs []slerror.SLError

	err := viper.Unmarshal(&config)
	if err != nil {
		errs = append(errs, slerror.CantParseConfigFile(err, log))
		return Config{}, errs
	}

	if debugMode {
		spew.Dump(config)
	}

	setLogLevel(log)

	spiffeIDs := make(map[string]bool)
	for _, db := range config.Databases {
		errs := parseDatabaseConfigFields(log, &db)
		errs = append(errs, errs...)

		spiffeIDs[db.SpiffeID] = true
	}

	if len(spiffeIDs) == 0 {
		// TODO need to handle this error better
		errs = append(errs, slerror.CantParseConfigFile(fmt.Errorf("no databases given in configuration file"), log))
	}
	otel := config.OpenTelemetry
	if otel.OtlpExporter.Endpoint == "" {
		log.Debug("no OpenTelemetry exporter specified; telemetry is disabled")
	}

	return config, errs
}

// Perform additional validation on the config file that requires network and FS access.
// In particular, check that files, binaries and addresses referred to actually exist.
func ValidateConfig(log *logrus.Logger, config Config) []error {
	var errs []error
	// TODO check the SPIFFE_AGENT_SOCKET_PATH environment variable
	if config.SpiffeAgentSocketPath == "" {
		errs = append(errs, fmt.Errorf("spiffe agent socket path is empty"))
	}

	err := checkSpiffeAgentSocket(log, config.SpiffeAgentSocketPath)
	if err != nil {
		errs = append(errs, err)
	}
	return errs
}
