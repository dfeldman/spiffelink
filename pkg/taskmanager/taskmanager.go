package taskmanager

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dfeldman/spiffelink/pkg/step"
	"github.com/sirupsen/logrus"
)

type ManagerInterface interface {
	NewTask(taskType string, timeout time.Duration, taskFunc TaskFunc) (*Task, error)
	KillTask(id string) error
	KillOverdueTasks()
	KillAllTasks()
	GetRunningTasks() []*Task
	GetTaskByID(id string) (*Task, error)
	GetTasksByType(taskType string) []*Task
	Shutdown()
}

// Task represents a background task with its properties and channels.
type Task struct {
	ID          string
	Type        string
	StartTime   time.Time
	Timeout     time.Duration
	OutputChan  chan step.StepFuncOutputMessage
	cancelFunc  context.CancelFunc
	IsCompleted bool
	Logger      *logrus.Logger
	Logs        []string
	Zombie      bool
}

type TaskFunc func(logger *logrus.Logger, ctx context.Context, out chan step.StepFuncOutputMessage)

// Manager manages all the tasks.
type Manager struct {
	globalLogger *logrus.Logger
	tasks        map[string]*Task
	mu           sync.RWMutex
	shutdown     chan struct{}
}

// NewManager returns a new instance of the task manager.
func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{
		tasks:        make(map[string]*Task),
		shutdown:     make(chan struct{}),
		globalLogger: logger,
	}
}

// TaskLogHook is a custom Logrus hook to capture logs for a task.
type TaskLogHook struct {
	Task *Task
}

func (hook *TaskLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *TaskLogHook) Fire(entry *logrus.Entry) error {
	hook.Task.Logs = append(hook.Task.Logs, entry.Message)

	// Forward the log entry to the global logger.
	switch entry.Level {
	case logrus.PanicLevel:
		logrus.Panic(entry.Message)
	case logrus.FatalLevel:
		logrus.Fatal(entry.Message)
	case logrus.ErrorLevel:
		logrus.Error(entry.Message)
	case logrus.WarnLevel:
		logrus.Warn(entry.Message)
	case logrus.InfoLevel:
		logrus.Info(entry.Message)
	case logrus.DebugLevel:
		logrus.Debug(entry.Message)
	case logrus.TraceLevel:
		logrus.Trace(entry.Message)
	}
	return nil
}

// NewTask creates a new task and adds it to the manager.
func (m *Manager) NewTask(taskType string, timeout time.Duration, taskFunc TaskFunc) (*Task, error) {
	id := taskType + " " + time.Now().String()
	taskCtx, cancelFunc := context.WithTimeout(context.Background(), timeout)

	// Create a new logger instance and attach the Logrus hook to capture logs.
	logger := logrus.New()
	logHook := &TaskLogHook{}
	logger.AddHook(logHook)

	task := &Task{
		ID:          id,
		Type:        taskType,
		StartTime:   time.Now(),
		Timeout:     timeout,
		OutputChan:  make(chan step.StepFuncOutputMessage),
		cancelFunc:  cancelFunc,
		IsCompleted: false,
		Logger:      logger,
		Logs:        []string{},
	}

	logHook.Task = task

	// Add the task to the manager's tasks map.
	m.mu.Lock()
	m.tasks[id] = task
	m.mu.Unlock()

	go func() {
		defer close(task.OutputChan)
		defer m.removeTask(id)
		logrus.Infof("About to start task %v", id)
		taskFunc(logger, taskCtx, task.OutputChan)
		logrus.Infof("Completed task %v", id)
	}()

	return task, nil
}

// removeTask removes a task from the manager after completion or cancellation.
func (m *Manager) removeTask(id string) {
	m.mu.Lock()
	m.tasks[id].Zombie = true // It can take a moment to shut down the function
	defer m.mu.Unlock()

	// Remove the task from the tasks map.
	delete(m.tasks, id)

	// Log the task removal.
	logrus.WithField("taskID", id).Info("Task removed")
}

// KillTask kills a specific task based on its ID.
func (m *Manager) KillTask(id string) error {
	m.mu.RLock()
	task, exists := m.tasks[id]
	m.mu.RUnlock()

	if !exists {
		return errors.New("task not found")
	}

	task.cancelFunc()
	return nil
}

// KillOverdueTasks kills all tasks that have exceeded their timeout.
func (m *Manager) KillOverdueTasks() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, task := range m.tasks {
		if time.Since(task.StartTime) > task.Timeout {
			logrus.Infof("Cancel overdue task %v", task.ID)
			task.cancelFunc()
		}
	}
}

// KillAllTasks kills all the running tasks.
func (m *Manager) KillAllTasks() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, task := range m.tasks {
		logrus.Infof("About to kill task %v", task.ID)
		task.cancelFunc()
	}
}

// GetRunningTasks returns all the currently running tasks.
func (m *Manager) GetRunningTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetTaskByID retrieves a task based on its ID.
func (m *Manager) GetTaskByID(id string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, errors.New("task not found")
	}

	return task, nil
}

// GetTasksByType retrieves tasks based on their type.
func (m *Manager) GetTasksByType(taskType string) []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []*Task
	for _, task := range m.tasks {
		if task.Type == taskType {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// Shutdown gracefully shuts down the manager and all its tasks.
func (m *Manager) Shutdown() {
	m.KillAllTasks()
	close(m.shutdown)
}
