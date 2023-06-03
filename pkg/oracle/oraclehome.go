package oracle

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func FindOracleHome(logger *logrus.Logger) (string, error) {
	// First, check if the ORACLE_HOME environment variable is set
	envOracleHome := os.Getenv("ORACLE_HOME")
	if envOracleHome != "" {
		logger.Infof("Found ORACLE_HOME in environment: %s", envOracleHome)
		return envOracleHome, nil
	}

	// If not, check some common Oracle installation paths
	commonPaths := []string{
		"/opt/oracle/product/11.2.0.1/dbhome_1",
		"/opt/oracle/product/12.2.0.2/dbhome_1",
		"/opt/oracle/product/18c/dbhome_1",
		"/opt/oracle/product/19c/dbhome_1",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			logger.Infof("Found ORACLE_HOME at common path: %s", path)
			return path, nil
		}
	}

	// If we haven't found ORACLE_HOME yet, return an error
	logger.Error("Could not find ORACLE_HOME")
	return "", errors.New("could not find ORACLE_HOME")
}

func VerifyOracleHome(path string) bool {
	// These are some typical subdirectories found in an Oracle Home directory.
	subdirs := []string{"bin", "network", "rdbms"}

	for _, subdir := range subdirs {
		subdirPath := filepath.Join(path, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			return false
		}
	}

	return true
}
