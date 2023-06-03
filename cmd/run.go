/*
Copyright © 2023 Sentima

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/dfeldman/spiffelink/pkg/config"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func handleErrors(errors []error, logger *logrus.Logger) {
	if len(errors) == 0 {
		// No errors, exit with status 0
		os.Exit(0)
	}

	// If we have errors, log and print each one
	for _, err := range errors {
		logger.Error(err)
		fmt.Println(err)
	}

	// Since there were errors, exit with status 1
	os.Exit(1)
}

// runCmd represents the run command
func newRunCmd(logger *logrus.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run spiffelink",
		Long: `Start the spiffelink process, connecting to SPIRE and any configured databases
		to send them SVIDs.`,
		Run: func(cmd *cobra.Command, args []string) {
			_, errs := config.ReadConfig(logger)
			handleErrors(errs, logger)
			fmt.Print("got here3")

		},
	}
}

// func init() {
// 	rootCmd.AddCommand(runCmd)

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
