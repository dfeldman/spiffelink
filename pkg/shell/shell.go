package shell

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func findExecutable(paths []string, name string) (string, error) {
	for _, path := range paths {
		execPath := filepath.Join(path, name)
		if fileInfo, err := os.Stat(execPath); err == nil {
			if (fileInfo.Mode() & 0111) == 0 {
				return "", fmt.Errorf("%s is not executable", execPath)
			}
			return execPath, nil
		}
	}
	return "", fmt.Errorf("executable %s not found in paths", name)
}

func RunBinaryWithTimeout(logger *logrus.Logger, paths []string, binary string, args []string, timeout time.Duration) (string, error) {
	execPath, err := findExecutable(paths, binary)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	// Create a new context with the specified timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create the command
	cmd := exec.CommandContext(ctx, execPath, args...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err = cmd.Run()

	// Check for errors
	if err != nil {
		// If the context was canceled, the command timed out
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("Command timed out")
			return "", fmt.Errorf("command timed out")
		}

		// The command failed for another reason
		logger.Error("Command failed: ", stderr.String())
		return "", fmt.Errorf("command failed: %w", err)
	}

	// Return the command's output
	return stdout.String(), nil
}
