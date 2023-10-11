package localshell

// TODO switch to using os. functions instead of ioutil.
import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newMockLogger() *logrus.Logger {
	ml := logrus.New()
	ml.Out = ioutil.Discard // Ensures that the logger does not print anything
	return ml
}

func TestFindExecutable(t *testing.T) {
	shell := NewLocalShell(newMockLogger())

	tmpDir := os.TempDir()

	// Create a fake executable file
	execPath := filepath.Join(tmpDir, "fake_executable")
	os.WriteFile(execPath, []byte("#!/bin/sh\necho hello"), 0755)

	// Test finding the executable
	foundPath, err := shell.FindExecutable(context.TODO(), []string{tmpDir}, "fake_executable")
	assert.NoError(t, err)
	assert.Equal(t, execPath, foundPath)

	// Test not finding the executable
	_, err = shell.FindExecutable(context.TODO(), []string{tmpDir}, "nonexistent")
	assert.Error(t, err)
}

func TestCheckExecutable(t *testing.T) {
	shell := NewLocalShell(newMockLogger())

	// Create a temporary directory for testing
	tmpDir := os.TempDir()

	// Create a fake executable file
	execPath := filepath.Join(tmpDir, "fake_executable")
	os.WriteFile(execPath, []byte("#!/bin/sh\necho hello"), 0755)

	// Test checking the executable
	err := shell.CheckExecutable(context.TODO(), execPath)
	assert.NoError(t, err)

	// Test checking a nonexistent file
	err = shell.CheckExecutable(context.TODO(), filepath.Join(tmpDir, "nonexistent"))
	assert.Error(t, err)
}

func TestFindPaths(t *testing.T) {
	shell := NewLocalShell(newMockLogger())

	// Create a temporary directory for testing
	tmpDir := os.TempDir()
	defer os.Remove(tmpDir + "/nonexistentpath")

	paths, err := shell.FindPaths(context.TODO(), []string{tmpDir, "/nonexistentpath"})
	assert.NoError(t, err)
	assert.Contains(t, paths, tmpDir)
	assert.NotContains(t, paths, "/nonexistentpath")
}

func TestRunCmd(t *testing.T) {
	shell := NewLocalShell(newMockLogger())

	//output, err := shell.RunCmd(context.TODO(), "echo", []string{"hello"}, nil, 5*time.Second)
	output, err := shell.RunCmd(context.TODO(), "/bin/sh", []string{"-c", "echo hello"}, nil, 5*time.Second)

	assert.NoError(t, err)
	assert.Equal(t, "hello\n", output)

	// Test with a timeout
	_, err = shell.RunCmd(context.TODO(), "sleep", []string{"10"}, nil, 1*time.Second)
	assert.Error(t, err)
}

func TestCheckPathWriteable(t *testing.T) {
	shell := NewLocalShell(newMockLogger())

	tmpDir := os.TempDir()
	defer os.Remove(tmpDir + "/nonexistentpath")

	// Check a writeable path
	err := shell.CheckPathWriteable(context.TODO(), tmpDir+"tempfile")
	assert.NoError(t, err)

	// Check a non-writeable path (root is typically non-writeable)
	err = shell.CheckPathWriteable(context.TODO(), "/xyz")
	assert.Error(t, err)
}
