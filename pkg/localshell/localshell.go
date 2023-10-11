package localshell

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// This is a ShellContext just for running local commands

type LocalShellContext struct {
	Logger *logrus.Logger
}

func NewLocalShell(logger *logrus.Logger) *LocalShellContext {
	return &LocalShellContext{
		Logger: logger,
	}
}

func (lsc *LocalShellContext) FindExecutable(ctx context.Context, paths []string, name string) (string, error) {
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

func (lsc *LocalShellContext) CheckExecutable(ctx context.Context, execPath string) error {
	// This won't build on Windows
	fileInfo, err := os.Stat(execPath)
	if err != nil {
		return err
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", execPath)
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return err
	}
	fileMode := fileInfo.Mode()
	if stat.Uid == uint32(uid) && fileMode&0100 != 0 { // User is owner and owner has execute permission.
		return nil
	}
	if stat.Gid == uint32(gid) && fileMode&0010 != 0 { // User is in group and group has execute permission.
		return nil
	}
	if fileMode&0001 != 0 { // Anyone has execute permission.
		return nil
	}
	return fmt.Errorf("current user does not have execute permission for '%s'", execPath)
}

func (lsc *LocalShellContext) FindPaths(ctx context.Context, paths []string) ([]string, error) {
	var validPaths []string
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		validPaths = append(validPaths, path)
	}
	return validPaths, nil
}

func (lsc *LocalShellContext) RunCmd(ctx context.Context, path string, args []string, environ []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// Create the command
	cmd := exec.CommandContext(ctx, path, args...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Check for errors
	if err != nil {
		// If the context was canceled, the command timed out
		if ctx.Err() == context.DeadlineExceeded {
			lsc.Logger.Error("Command timed out")
			return "", fmt.Errorf("command timed out")
		}

		// The command failed for another reason
		lsc.Logger.Error("Command failed: ", stderr.String())
		return "", fmt.Errorf("command failed: %w", err)
	}

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out")
	}

	// Return the command's output
	return stdout.String(), nil
}

func (lsc *LocalShellContext) CheckPathWriteable(ctx context.Context, path string) error {
	// If the file exists, try writing to it
	if _, err := os.Stat(path); err == nil {
		testFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("file %s is not writable: %w", path, err)
		}
		testFile.Close()
		return nil
	}

	// If the file doesn't exist, check if we can create it in the parent directory
	parentDir := filepath.Dir(path)
	testFile, err := os.OpenFile(filepath.Join(parentDir, "test_write.tmp"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("parent directory %s is not writable: %w", parentDir, err)
	}
	testFile.Close()
	os.Remove(filepath.Join(parentDir, "test_write.tmp"))

	return nil
}
