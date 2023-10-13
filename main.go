package main

import (
	"log"

	"github.com/dfeldman/spiffelink/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	// TODO this initial logger we create here may get replaced after specifying different log options in config.
	// Need to investigate
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	var rootCmd = &cobra.Command{
		Use:   "spiffelink",
		Short: "Spiffelink is a tool for updating databases using SPIFFE credentials",
		Long:  `A longer description of spiffelink...`,
	}

	// Add the run command to the root command
	runCmd := cmd.NewRunCmd(logger)
	rootCmd.AddCommand(runCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
}
