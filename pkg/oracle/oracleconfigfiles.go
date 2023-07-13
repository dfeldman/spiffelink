package oracle

import (
	"fmt"
	"os"
	"path/filepath"
)

// defaultOracleFiles are the file names we're interested in.
var defaultOracleFiles = []string{
	"tnsnames.ora",
	"listener.ora",
	"sqlnet.ora",
}

// defaultOracleFileDirs are the default relative directories where oracle files might exist.
var defaultOracleFileDirs = []string{
	"network/admin", // Add other default directories here if needed
}

// OracleFileLocations represents the locations of Oracle files.
type OracleFileLocations struct {
	TnsnamesOra string
	ListenerOra string
	SqlnetOra   string
}

// FindOracleFiles finds the oracle files from their default locations.
func FindOracleFiles(oracleHome string) (OracleFileLocations, error) {
	var locations OracleFileLocations

	// Check the default locations for each file
	for _, file := range defaultOracleFiles {
		var found bool
		for _, dir := range defaultOracleFileDirs {
			path := filepath.Join(oracleHome, dir, file)
			if _, err := os.Stat(path); err == nil {
				switch file {
				case "tnsnames.ora":
					locations.TnsnamesOra = path
				case "listener.ora":
					locations.ListenerOra = path
				case "sqlnet.ora":
					locations.SqlnetOra = path
				}
				found = true
				break
			}
		}

		if !found {
			return OracleFileLocations{}, fmt.Errorf("No readable %s file found in ORACLE_HOME", file)
		}
	}

	return locations, nil
}
