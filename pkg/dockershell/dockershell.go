package dockershell

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

// Wrapper to be able to mock Docker
type DockerClientInterface interface {
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
}

type DockerContext struct {
	containerID string
	cli         DockerClientInterface
	logger      *logrus.Logger
}

func NewDockerContext(containerID string, logger *logrus.Logger) (*DockerContext, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockerContext{
		containerID: containerID,
		cli:         cli,
		logger:      logger,
	}, nil
}

func (dc *DockerContext) execCommand(ctx context.Context, cmd ...string) (string, error) {
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execIDResp, err := dc.cli.ContainerExecCreate(ctx, dc.containerID, execConfig)
	if err != nil {
		return "", err
	}

	execStartCheck := types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	}
	execAttachResp, err := dc.cli.ContainerExecAttach(ctx, execIDResp.ID, execStartCheck)
	if err != nil {
		return "", err
	}
	defer execAttachResp.Close()

	output := ""
	buf := make([]byte, 1024)
	for {
		n, err := execAttachResp.Reader.Read(buf)
		if err != nil {
			break
		}
		output += string(buf[:n])
	}

	// Check the exit code of the executed command
	execInspect, err := dc.cli.ContainerExecInspect(ctx, execIDResp.ID)
	if err != nil {
		return output, err
	}
	if execInspect.ExitCode != 0 {
		return output, fmt.Errorf("command exited with code %d: %s", execInspect.ExitCode, output)
	}

	return output, nil
}

func (dc *DockerContext) FindExecutable(ctx context.Context, paths []string, name string) (string, error) {
	for _, path := range paths {
		execPath := path + "/" + name
		output, err := dc.execCommand(ctx, "test", "-x", execPath)
		if err != nil {
			return "", err
		}
		if err == nil && strings.TrimSpace(output) == "" {
			return execPath, nil
		}
	}
	return "", fmt.Errorf("executable %s not found in paths", name)
}

func (dc *DockerContext) CheckExecutable(ctx context.Context, execPath string) error {
	// Check if the file exists and is executable
	output, err := dc.execCommand(ctx, "test", "-x", execPath)
	fmt.Println("Output from execCommand:", output)
	if err != nil || strings.TrimSpace(output) != "" {
		return fmt.Errorf("current user does not have execute permission for '%s'", execPath)
	}
	return nil
}

func (dc *DockerContext) FindPaths(ctx context.Context, paths []string) ([]string, error) {
	var validPaths []string
	for _, path := range paths {
		output, err := dc.execCommand(ctx, "test", "-d", path)
		if err != nil {
			// If there's an error, we stop immediately. This is probably a failure to connect to the container.
			return validPaths, err
		}
		if strings.TrimSpace(output) == "" {
			validPaths = append(validPaths, path)
		}
	}
	return validPaths, nil
}

func (dc *DockerContext) RunCmd(ctx context.Context, path string, args []string, environ []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := append([]string{path}, args...)
	output, err := dc.execCommand(ctx, cmd...)
	if ctx.Err() == context.DeadlineExceeded {
		dc.logger.Error("Command timed out")
		return "", fmt.Errorf("command timed out")
	}
	if err != nil {
		dc.logger.Error("Command failed: ", output)
		return "", fmt.Errorf("command failed: %s", err)
	}
	return output, nil
}

func (dc *DockerContext) CheckPathWriteable(ctx context.Context, path string) error {
	// Check if the path is writeable
	output, err := dc.execCommand(ctx, "test", "-w", path)
	if err != nil || strings.TrimSpace(output) != "" {
		return fmt.Errorf("path %s is not writable", path)
	}
	return nil
}
