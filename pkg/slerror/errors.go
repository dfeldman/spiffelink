package slerror

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func NoError() SLError {
	return SLError{}
}

var dbNameInvalidMessage = `
The database name %s is invalid. Database names must consist of alphanumeric characters and
be less than 255 characters.`

// TODO sometimes we want to create an error and NOT immediately log it. May need to rethink this.

func DatabaseNameInvalidError(log *logrus.Logger, name string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CONFIG_DATABASE_NAME_INVALID",
		Err:             fmt.Errorf("invalid database name %s", name),
		Heading:         "Database name contains invalid characters",
		DetailedMessage: fmt.Sprintf(dbNameInvalidMessage, name),
		Severity:        "Fatal",
	})
}

var dbNameEmpty = `
The database name is empty. Check the configuration file formatting (including indentation).`

func DatabaseNameEmptyError(log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CONFIG_DATABASE_NAME_INVALID",
		Err:             fmt.Errorf("empty database name"),
		Heading:         "Empty database name",
		DetailedMessage: dbNameEmpty,
		Severity:        "Fatal",
	})
}

var dbTypeEmpty = `
The database type is empty. Check the configuration file formatting (including indentation).`

func DatabaseTypeEmptyError(log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CONFIG_DATABASE_TYPE_EMPTY",
		Err:             fmt.Errorf("empty database type"),
		Heading:         "Empty database type",
		DetailedMessage: dbTypeEmpty,
		Severity:        "Fatal",
	})
}

var dbTypeInvalid = `
The database type %s is invalid. Check the configuration file for the correct database type.`

func DatabaseTypeInvalidError(log *logrus.Logger, dbType string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CONFIG_DATABASE_TYPE_INVALID",
		Err:             fmt.Errorf("invalid database type %s", dbType),
		Heading:         "Invalid database type",
		DetailedMessage: fmt.Sprintf(dbTypeInvalid, dbType),
		Severity:        "Fatal",
	})
}

var connStrEmpty = `
The connection string is empty. Check the configuration file formatting (including indentation).`

func ConnectionStringEmptyError(log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CONFIG_CONNECTION_STRING_EMPTY",
		Err:             fmt.Errorf("empty connection string"),
		Heading:         "Empty connection string",
		DetailedMessage: connStrEmpty,
		Severity:        "Fatal",
	})
}

var connStrInvalid = `
The connection string %s is invalid. Check the configuration file for the correct connection string.`

func ConnectionStringInvalidError(log *logrus.Logger, connStr string) SLError {
	return SLError{
		Code:            "CONFIG_CONNECTION_STRING_INVALID",
		Err:             fmt.Errorf("invalid connection string %s", connStr),
		Heading:         "Invalid connection string",
		DetailedMessage: fmt.Sprintf(connStrInvalid, connStr),
		Severity:        "Fatal",
	}
}

var cantConnect = `
Cannot connect to the database using connection string %s. Please verify the database is running and the connection string is correct.`

func ConnectionFailedError(log *logrus.Logger, connStr string) error {
	return LogAndReturn(log, SLError{
		Code:            "DATABASE_CONNECTION_FAILED",
		Err:             fmt.Errorf("cannot connect to the database using connection string %s", connStr),
		Heading:         "Failed to connect to the database",
		DetailedMessage: fmt.Sprintf(cantConnect, connStr),
		Severity:        "Fatal",
	})
}

var spiffeIDEmpty = `
The SPIFFE ID is empty. Check the configuration file formatting (including indentation).`

func SpiffeIDEmptyError(log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "SPIFFE_ID_EMPTY",
		Err:             fmt.Errorf("SPIFFE ID is empty"),
		Heading:         "Empty SPIFFE ID",
		DetailedMessage: spiffeIDEmpty,
		Severity:        "Fatal",
	})
}

var spiffeIDInvalid = `
The SPIFFE ID %s is invalid. SPIFFE IDs must adhere to the standard SPIFFE ID format.`

func SpiffeIDInvalidError(log *logrus.Logger, spiffeID string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "SPIFFE_ID_INVALID",
		Err:             fmt.Errorf("invalid SPIFFE ID %s", spiffeID),
		Heading:         "Invalid SPIFFE ID",
		DetailedMessage: fmt.Sprintf(spiffeIDInvalid, spiffeID),
		Severity:        "Fatal",
	})
}

var mismatchedSpiffeIDs = `
The SPIFFE IDs %s and %s do not match. SPIFFE IDs must match.`

func MismatchedSpiffeIDsError(log *logrus.Logger, spiffeID1, spiffeID2 string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "SPIFFE_ID_MISMATCH",
		Err:             fmt.Errorf("SPIFFE IDs %s and %s do not match", spiffeID1, spiffeID2),
		Heading:         "Mismatched SPIFFE IDs",
		DetailedMessage: fmt.Sprintf(mismatchedSpiffeIDs, spiffeID1, spiffeID2),
		Severity:        "Fatal",
	})
}

var agentSocketPathEmpty = `
The agent socket path is empty. Check the configuration file formatting (including indentation).`

func AgentSocketPathEmptyError(log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "AGENT_SOCKET_PATH_EMPTY",
		Err:             fmt.Errorf("agent socket path is empty"),
		Heading:         "Empty agent socket path",
		DetailedMessage: agentSocketPathEmpty,
		Severity:        "Fatal",
	})
}

var agentSocketPathDoesNotExist = `
The agent socket path %s does not exist. Ensure the path is correct.`

func AgentSocketPathDoesNotExistError(log *logrus.Logger, path string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "AGENT_SOCKET_PATH_NOT_EXIST",
		Err:             fmt.Errorf("agent socket path %s does not exist", path),
		Heading:         "Agent socket path does not exist",
		DetailedMessage: fmt.Sprintf(agentSocketPathDoesNotExist, path),
		Severity:        "Fatal",
	})
}

var agentSocketPathStatFailed = `
The agent socket path %s cannot be accessed. Ensure the path is correct.`

func AgentSocketPathStatFailedError(log *logrus.Logger, path string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "AGENT_SOCKET_PATH_STAT_FAILED",
		Err:             fmt.Errorf("agent socket path %s cannot stat", path),
		Heading:         "Cannot access agent socket path",
		DetailedMessage: fmt.Sprintf(agentSocketPathInvalid, path),
		Severity:        "Fatal",
	})
}

var agentSocketPathInvalid = `
The agent socket path %s does not appear to be a socket. Ensure the path is correct.`

func AgentSocketPathInvalidError(log *logrus.Logger, path string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "AGENT_SOCKET_PATH_INVALID",
		Err:             fmt.Errorf("agent socket path %s is invalid", path),
		Heading:         "Invalid agent socket path",
		DetailedMessage: fmt.Sprintf(agentSocketPathInvalid, path),
		Severity:        "Fatal",
	})
}

var invalidPermissionsAgentSocketPath = `
Invalid permissions on agent socket path %s. Ensure the application has the necessary permissions.`

func InvalidPermissionsAgentSocketPathError(log *logrus.Logger, path string) SLError {
	return LogAndReturn(log, SLError{
		Code:            "AGENT_SOCKET_PATH_INVALID_PERMISSIONS",
		Err:             fmt.Errorf("invalid permissions on agent socket path %s", path),
		Heading:         "Invalid permissions on agent socket path",
		DetailedMessage: fmt.Sprintf(invalidPermissionsAgentSocketPath, path),
		Severity:        "Fatal",
	})
}

var unableToReadConfigFile = `
Unable to read the configuration file. Please ensure the config file is in a valid format.
The specific error that occured was: 
%s.`

func UnableToReadConfigFileError(log *logrus.Logger, err error) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CONFIG_FILE_UNREADABLE",
		Err:             err,
		Heading:         "Unable to read config file",
		DetailedMessage: unableToReadConfigFile,
		Severity:        "Fatal",
	})
}

var unableToParseConfigFile = `
Unable to parse the configuration file. Please check the formatting and contents of the file.`

func UnableToParseConfigFileError(log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CONFIG_FILE_UNPARSABLE",
		Err:             fmt.Errorf("unable to parse configuration file"),
		Heading:         "Unable to parse config file",
		DetailedMessage: unableToParseConfigFile,
		Severity:        "Fatal",
	})
}

var invalidDatabaseType = `
Invalid database type %s. Expected values are "oracle" or "testdb".`

func InvalidDatabaseType(log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "INVALID_DATABASE_TYPE",
		Err:             fmt.Errorf("invalid database type"),
		Heading:         "Invalid database type",
		DetailedMessage: unableToParseConfigFile,
		Severity:        "Fatal",
	})
}

func CantParseConfigFile(err error, log *logrus.Logger) SLError {
	return LogAndReturn(log, SLError{
		Code:            "CANT_PARSE_CONFIG_FILE",
		Err:             err,
		Heading:         "Invalid database type",
		DetailedMessage: unableToParseConfigFile,
		Severity:        "Fatal",
	})
}
