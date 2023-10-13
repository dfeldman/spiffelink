package shell

import (
	"context"
	"time"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/dockershell"
	"github.com/dfeldman/spiffelink/pkg/localshell"
	"github.com/sirupsen/logrus"
)

// ShellContext is the interface that SL uses to interact with all files and executables.
// The idea is that it will be able to interact with:
//  * the local system that SL is running in
//  * Docker containers
//  * Kubernetes containers
//  * Remote systems via SSH
// All using the same interface.
// No other part of SL should interact with the fileystem or run executables.

type ShellContext interface {
	// Find the path of an executable based on a search path
	FindExecutable(ctx context.Context, paths []string, name string) (string, error)
	// Check that a path really is a valid executable and we can run it
	// Note -- in some implementations this may be a no-op, it is just to provide better error handling if the executable
	// does not exist
	CheckExecutable(ctx context.Context, execPath string) error
	// Given a list of directories, identify the ones that really exist -- used since databases may install in many candidate locations
	FindPaths(ctx context.Context, paths []string) ([]string, error)
	// Run an executable
	// The path must be complete as returned by FindExecutable
	RunCmd(ctx context.Context, path string, args []string, environ []string, timeout time.Duration) (string, error)
	// Check that a path is writeable (for writing the client cert)
	CheckPathWriteable(ctx context.Context, path string) error
}

func GetShellContextFromConfig(conf config.ShellContextConfig, logger *logrus.Logger) (ShellContext, error) {
	switch conf.ShellType {
	case "LocalShell":
		return localshell.NewLocalShell(logger), nil
	case "DockerShell":
		return dockershell.NewDockerContext(conf.ContainerID, logger)
	default:
		return localshell.NewLocalShell(logger), nil
	}
}
