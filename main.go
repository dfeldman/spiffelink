package main

import (
	"github.com/dfeldman/spiffelink/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	cmd.Execute(logger)
}
