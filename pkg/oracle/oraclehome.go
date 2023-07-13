package oracle

import (
	"os"
	"fmt"
)

// oracleVersions is a list of possible oracle versions we're interested in.
var oracleVersions = []string{
	"19c", "18c", "12c", "11g", "10g", "9i", // Add other versions here
}

// defaultOracleHomePaths are the common paths where oracle home might exist.
var defaultOracleHomePaths = []string{
	"/opt/oracle/product/%s/dbhome_1", // Add other default paths here
}

// FindOracleHome finds the oracle home directory from environment variable or default locations.
func FindOracleHome() (string, error) {
	// Check ORACLE_HOME environment variable first
	oracleHome := os.Getenv("ORACLE_HOME")
	if oracleHome != "" {
		if isDirReadable(oracleHome) {
			return oracleHome, nil
		}
		return "", fmt.Errorf("ORACLE_HOME is set, but the directory is not readable")
	}

	// If ORACLE_HOME environment variable is not set, check the default locations
	for _, version := range oracleVersions {
		for _, pathPattern := range defaultOracleHomePaths {
			path := fmt.Sprintf(pathPattern, version)
			if _, err := os.Stat(path); err == nil && isDirReadable(path) {
				return path, nil
			}
		}
	}

	// If no readable oracle home directory is found, return an error
	return "", fmt.Errorf("No readable ORACLE_HOME directory found")
}

// isDirReadable checks if the current user has read permission on a directory.
func isDirReadable(path string) bool {
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return false
		}
		return false
	}
	file.Close()
	return true
}
