package updater_test

import (
	"context"
	"testing"
	"time"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/datastore"
	"github.com/dfeldman/spiffelink/pkg/shell"
	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/dfeldman/spiffelink/pkg/step"
	"github.com/dfeldman/spiffelink/pkg/taskmanager"
	"github.com/dfeldman/spiffelink/pkg/updater"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/stretchr/testify/mock"
)

// MockWorkloadAPIClient
type MockWorkloadAPIClient struct {
	mock.Mock
}

func (m *MockWorkloadAPIClient) WatchX509Context(ctx context.Context, w workloadapi.X509ContextWatcher) error {
	args := m.Called(ctx, w)
	return args.Error(0)
}

type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) NewTask(taskType string, timeout time.Duration, taskFunc taskmanager.TaskFunc) (*taskmanager.Task, error) {
	args := m.Called(taskType, timeout, taskFunc)
	return args.Get(0).(*taskmanager.Task), args.Error(1)
}

func (m *MockTaskManager) KillTask(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskManager) KillOverdueTasks() {
	m.Called()
}

func (m *MockTaskManager) KillAllTasks() {
	m.Called()
}

func (m *MockTaskManager) GetRunningTasks() []*taskmanager.Task {
	args := m.Called()
	return args.Get(0).([]*taskmanager.Task)
}

func (m *MockTaskManager) GetTaskByID(id string) (*taskmanager.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*taskmanager.Task), args.Error(1)
}

func (m *MockTaskManager) GetTasksByType(taskType string) []*taskmanager.Task {
	args := m.Called(taskType)
	return args.Get(0).([]*taskmanager.Task)
}

func (m *MockTaskManager) Shutdown() {
	m.Called()
}

// MockDatastore
type MockDatastore struct {
	mock.Mock
}

func (m *MockDatastore) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDatastore) GetUpdateSteps(ctx context.Context, dbConfig config.DatabaseConfig, shellContext shell.ShellContext, update spiffelinkcore.SpiffeLinkUpdate) step.StepList {
	args := m.Called(ctx, dbConfig, update)
	return args.Get(0).(step.StepList)
}

func TestUpdater_OnX509ContextUpdate(t *testing.T) {
	// Mock setup
	mockClient := new(MockWorkloadAPIClient)
	mockTM := new(MockTaskManager)
	mockDatastore := new(MockDatastore)

	// Config setup
	dbConfig := config.DatabaseConfig{
		Name: "mockDB",
	}
	cfg := config.Config{
		Databases: []config.DatabaseConfig{dbConfig},
	}
	stores := []datastore.Datastore{mockDatastore}

	// Mock expectations
	mockDatastore.On("GetName").Return("mockDB")
	mockDatastore.On("GetUpdateSteps", mock.Anything, dbConfig, mock.Anything).Return(step.StepList{})
	mockTM.On("NewTask", mock.Anything, mock.Anything, mock.Anything).Return(&taskmanager.Task{}, nil).Once()

	// Create the Updater
	u := updater.NewUpdater(cfg, mockClient, mockTM, stores, logrus.New())

	// Call OnX509ContextUpdate
	bundlesSet := x509bundle.NewSet()
	context := &workloadapi.X509Context{
		SVIDs:   []*x509svid.SVID{},
		Bundles: bundlesSet,
	}
	u.OnX509ContextUpdate(context)

	// Assertions
	mockDatastore.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}
