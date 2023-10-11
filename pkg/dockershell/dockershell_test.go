package dockershell

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Docker returns HijackedResponse structs that contain a Conn and have a Close function
// By mocking a TCPConn we get the needed fields to have a HijackedResponse that can be Closed
type NoOpCloseConn struct {
	*net.TCPConn
}

func (c *NoOpCloseConn) Close() error {
	return nil
}

// Mock implementation of the DockerClientInterface
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	args := m.Called(ctx, container, config)
	return args.Get(0).(types.IDResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	args := m.Called(ctx, execID, config)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	args := m.Called(ctx, execID)
	return args.Get(0).(types.ContainerExecInspect), args.Error(1)
}

// TODO this is not really useful
func newMockDockerClient() DockerContext {
	mockClient := new(MockDockerClient)
	dc := DockerContext{
		containerID: "test-container-id",
		cli:         mockClient,
		logger:      logrus.New(),
	}
	return dc
}

func newMockDockerResponse(response string) types.HijackedResponse {
	return types.HijackedResponse{
		Conn:   &NoOpCloseConn{},                                   // Use the no-op close connection
		Reader: bufio.NewReader(bytes.NewBuffer([]byte(response))), // Fake output buffer with no contents
	}
}

// Test for the FindExecutable function
func TestFindExecutable(t *testing.T) {
	mockClient := new(MockDockerClient)
	dc := DockerContext{
		containerID: "test-container-id",
		cli:         mockClient,
		logger:      logrus.New(),
	}

	mockHijackedResponse := types.HijackedResponse{
		Conn:   &NoOpCloseConn{},                           // Use the no-op close connection
		Reader: bufio.NewReader(bytes.NewBuffer([]byte{})), // Fake output buffer with no contents
	}

	// Scenario: Executable is found
	t.Run("Executable found", func(t *testing.T) {

		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(mockHijackedResponse, nil)
		// Mock expectations
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil)
		mockClient.On("ContainerExecInspect", mock.Anything, "mocked-id").Return(types.ContainerExecInspect{ExitCode: 0}, nil)

		// Run the function
		resultPath, err := dc.FindExecutable(context.Background(), []string{"/test/path"}, "test-executable")

		// Asserts
		assert.NoError(t, err)
		assert.Equal(t, "/test/path/test-executable", resultPath)
		mockClient.AssertExpectations(t)
	})

	// Scenario: Executable not found
	t.Run("Executable not found", func(t *testing.T) {
		// Mock expectations for a non-existing executable
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, fmt.Errorf("executable nonexistent-executable not found in paths"))

		// Run the function
		_, err := dc.FindExecutable(context.Background(), []string{"/test/path"}, "nonexistent-executable")

		// Asserts
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "executable nonexistent-executable not found in paths")
		}
		mockClient.AssertExpectations(t)
	})

	// Scenario: Error during ContainerExecCreate
	t.Run("Error from ContainerExecCreate", func(t *testing.T) {
		// Mock expectations
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{}, fmt.Errorf("Failed to create exec context"))

		// Run the function
		_, err := dc.FindExecutable(context.Background(), []string{"/test/path"}, "test-executable")

		// Asserts
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to create exec context")
		mockClient.AssertExpectations(t)
	})

	// Scenario: Error during ContainerExecAttach
	t.Run("Error from ContainerExecAttach", func(t *testing.T) {
		// Mock expectations
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(mockHijackedResponse, fmt.Errorf("Failed to attach to exec context"))

		// Run the function
		_, err := dc.FindExecutable(context.Background(), []string{"/test/path"}, "test-executable")

		// Asserts
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to attach to exec context")
		mockClient.AssertExpectations(t)
	})
}

func TestCheckExecutable(t *testing.T) {
	mockClient := new(MockDockerClient)
	dc := DockerContext{
		containerID: "test-container-id",
		cli:         mockClient,
		logger:      logrus.New(),
	}

	mockHijackedResponse := types.HijackedResponse{
		Conn:   &NoOpCloseConn{},                           // Use the no-op close connection
		Reader: bufio.NewReader(bytes.NewBuffer([]byte{})), // Fake output buffer with no contents
	}

	// Scenario: Executable exists and is executable
	t.Run("Executable is executable", func(t *testing.T) {
		// Mock expectations
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(mockHijackedResponse, nil)

		// Run the function
		err := dc.CheckExecutable(context.Background(), "/test/path/executable")

		// Asserts
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	mockClient = new(MockDockerClient)
	dc.cli = mockClient
	// Scenario: Executable does not exist or is not executable
	t.Run("Executable not executable", func(t *testing.T) {
		// Mock expectations
		nonExecResponse := types.HijackedResponse{
			Conn:   &NoOpCloseConn{},
			Reader: bufio.NewReader(bytes.NewBufferString("error: not executable")),
		}
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(nonExecResponse, nil)

		// Run the function
		err := dc.CheckExecutable(context.Background(), "/test/path/non-executable")

		// Asserts
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "current user does not have execute permission for '/test/path/non-executable'")
		mockClient.AssertExpectations(t)
	})

}

// TODO this doesn't really mock how test works
func TestFindPaths(t *testing.T) {
	mockClient := new(MockDockerClient)
	dc := DockerContext{
		containerID: "test-container-id",
		cli:         mockClient,
		logger:      logrus.New(),
	}

	mockHijackedResponse := types.HijackedResponse{
		Conn:   &NoOpCloseConn{},                           // Use the no-op close connection
		Reader: bufio.NewReader(bytes.NewBuffer([]byte{})), // Fake output buffer with no contents
	}

	// Scenario: All paths are valid directories
	t.Run("All paths valid", func(t *testing.T) {
		// Mock expectations for valid directories
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil).Times(2)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(mockHijackedResponse, nil).Times(2)

		paths, err := dc.FindPaths(context.Background(), []string{"/valid/path1", "/valid/path2"})

		assert.NoError(t, err)
		assert.Equal(t, []string{"/valid/path1", "/valid/path2"}, paths)
		mockClient.AssertExpectations(t)
	})

	// Scenario: Some paths are valid directories, others are not
	t.Run("Some paths valid", func(t *testing.T) {
		// Mock expectations for one valid directory and one invalid
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil).Times(2)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(mockHijackedResponse, nil).Once()
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(types.HijackedResponse{
			Conn:   &NoOpCloseConn{},
			Reader: bufio.NewReader(bytes.NewBufferString("error: not a directory")),
		}, nil).Once()

		paths, err := dc.FindPaths(context.Background(), []string{"/valid/path", "/invalid/path"})

		assert.NoError(t, err)
		assert.Equal(t, []string{"/valid/path"}, paths)
		mockClient.AssertExpectations(t)
	})
}

func TestRunCmd(t *testing.T) {
	mockClient := new(MockDockerClient)
	dc := DockerContext{
		containerID: "test-container-id",
		cli:         mockClient,
		logger:      logrus.New(),
	}

	mockHijackedResponse := types.HijackedResponse{
		Conn:   &NoOpCloseConn{},                                  // Use the no-op close connection
		Reader: bufio.NewReader(bytes.NewBuffer([]byte("hello"))), // Fake output buffer with no contents
	}

	// Scenario: Command runs successfully
	t.Run("Command success", func(t *testing.T) {
		// Mock expectations
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(mockHijackedResponse, nil)
		mockClient.On("ContainerExecInspect", mock.Anything, "mocked-id").Return(types.ContainerExecInspect{ExitCode: 0}, nil)

		output, err := dc.RunCmd(context.Background(), "echo", []string{"hello"}, nil, 5*time.Second)

		assert.NoError(t, err)
		assert.Equal(t, "hello", output)
		mockClient.AssertExpectations(t)
	})

	// Scenario: Command failure
	t.Run("Command failure", func(t *testing.T) {
		// Mock expectations for command failure
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Return(types.HijackedResponse{
			Conn:   &NoOpCloseConn{},
			Reader: bufio.NewReader(bytes.NewBufferString("Command failed")),
		}, fmt.Errorf("Command execution failed"))

		_, err := dc.RunCmd(context.Background(), "failcmd", nil, nil, 5*time.Second)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command failed")
		mockClient.AssertExpectations(t)
	})

	t.Run("Command timeout", func(t *testing.T) {
		// Mock expectations for a long-running command (e.g., sleep 10 seconds)
		mockClient.On("ContainerExecCreate", mock.Anything, "test-container-id", mock.Anything).Return(types.IDResponse{ID: "mocked-id"}, nil)
		mockClient.On("ContainerExecAttach", mock.Anything, "mocked-id", mock.Anything).Run(func(args mock.Arguments) {
			time.Sleep(2 * time.Second)
		}).Return(mockHijackedResponse, nil)
		mockClient.On("ContainerExecInspect", mock.Anything, "mocked-id").Return(types.ContainerExecInspect{ExitCode: 0}, nil)

		// Run the command with a timeout of 1 second
		_, err := dc.RunCmd(context.Background(), "sleep", []string{"10"}, nil, 1*time.Second)

		// The command should be terminated by the context's timeout
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command timed out")
		mockClient.AssertExpectations(t)
	})
}
