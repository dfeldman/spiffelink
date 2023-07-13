package shell

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

func FindExecutable(paths []string, name string) (string, error) {
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

func CheckExecutable(execPath string) error {
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

func MockExecutable(ctx context.Context, execPath string, args ...string) {
	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping mock executable.")
		default:
			fmt.Printf("Running %s with arguments: %v\n", execPath, args)
		}
	}()
}

func FindPaths(paths []string) ([]string, error) {
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

func RunBinaryWithTimeout(logger *logrus.Logger, paths []string, binary string, args []string, timeout time.Duration) (string, error) {
	execPath, err := FindExecutable(paths, binary)
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

func CheckPathWriteable(path string) error {
	info, err := os.Stat(path)

	// If the file doesn't exist, check if parent directory is writable
	if os.IsNotExist(err) {
		parentDir := filepath.Dir(path)
		info, err = os.Stat(parentDir)
		if err != nil {
			return err
		}
		if info.Mode().Perm()&(1<<(uint(7))) == 0 {
			return fmt.Errorf("parent directory %s is not writable", parentDir)
		}
	} else if err != nil {
		// An error occurred that wasn't os.IsNotExist
		return err
	} else {
		// If the file exists, check it is readable and writable
		if info.Mode().Perm()&(1<<(uint(7))) == 0 || info.Mode().Perm()&(1<<(uint(8))) == 0 {
			return fmt.Errorf("file %s is not both readable and writable", path)
		}
	}

	return nil
}
