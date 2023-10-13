/*
Copyright © 2023 Sentima
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/datastore"
	"github.com/dfeldman/spiffelink/pkg/slerror"
	"github.com/dfeldman/spiffelink/pkg/taskmanager"
	"github.com/dfeldman/spiffelink/pkg/updater"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func handleErrors(errors []slerror.SLError, logger *logrus.Logger) {
	shouldExit := false

	// If we have errors, log and print each one
	for _, err := range errors {
		logger.Error(err)
		fmt.Println(err)
		if err.Severity == slerror.SeverityFatal {
			shouldExit = true
		}
	}

	// Since there were fatal errors, exit with status 1
	if shouldExit {
		os.Exit(1)
	}
}

// runCmd represents the run command
func NewRunCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run spiffelink",
		Long: `Start the spiffelink process, connecting to SPIRE and any configured databases
		to send them SVIDs.`,
		Run: func(cmd *cobra.Command, args []string) {
			config, errs := config.ReadConfig(logger)
			handleErrors(errs, logger)
			// TODO Workloadapi.New will use an env var by default. We need to check that we default to the same env var.
			api, err := workloadapi.New(context.Background(), workloadapi.WithAddr(config.SpiffeAgentSocketPath))
			if err != nil {
				// TODO this is a common error, need a much better error here
				fmt.Printf("Unable to connect to socket path %s due to %v\n", config.SpiffeAgentSocketPath, err)
			}
			client := updater.NewRealWorkloadAPIClient(api)
			updater := updater.NewUpdater(&config, client, taskmanager.NewManager(logger), datastore.GetDatastores(), logger)
			updater.Start(context.Background())
		},
	}

	// Define the config flag
	cmd.Flags().StringP("config", "c", "", "Path to the configuration file")

	// Bind the flag to Viper
	viper.BindPFlag("config", cmd.Flags().Lookup("config"))
	return cmd
}
