package taskmanager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dfeldman/spiffelink/pkg/step"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func sampleTaskFunc(logger *logrus.Logger, ctx context.Context, out chan step.StepFuncOutputMessage) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Task ended due to context cancellation")
			return
		case t := <-ticker.C:
			msg := fmt.Sprintf("Tick at %v", t)
			logger.Info(msg)
		}
	}
}

func TestTaskManager(t *testing.T) {
	// Create a new task manager.
	logger := logrus.New()

	manager := NewManager(logger)
	assert.NotNil(t, manager)

	// Add a new task to the manager.
	task, err := manager.NewTask("sample", 2*time.Second, sampleTaskFunc)
	assert.NoError(t, err)
	assert.NotNil(t, task)

	// Fetch a task by its ID.
	fetchedTask, err := manager.GetTaskByID(task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task.ID, fetchedTask.ID)

	// Fetch tasks by type.
	tasksByType := manager.GetTasksByType("sample")
	assert.Equal(t, 1, len(tasksByType))

	// Kill a specific task by its ID.
	err = manager.KillTask(task.ID)
	assert.NoError(t, err)

	// Add more tasks to test killing overdue tasks and killing all tasks.
	task1, _ := manager.NewTask("sample1", 1*time.Second, sampleTaskFunc)
	task2, _ := manager.NewTask("sample2", 50*time.Second, sampleTaskFunc)

	// Let's wait for a bit to ensure task1 is overdue but task2 isn't.
	time.Sleep(5 * time.Second)

	// Kill overdue tasks. Only task1 should be killed.
	manager.KillOverdueTasks()
	time.Sleep(1 * time.Second)
	_, err1 := manager.GetTaskByID(task1.ID)
	_, err2 := manager.GetTaskByID(task2.ID)
	time.Sleep(1 * time.Second) // TODO this shouldn't be needed
	assert.Error(t, err1)       // task1 should be removed
	assert.NoError(t, err2)     // task2 should still exist

	// Kill all tasks.
	// manager.KillAllTasks()
	// _, err2AfterKill := manager.GetTaskByID(task2.ID)
	// assert.Error(t, err2AfterKill) // task2 should now be removed

	// Shutdown the manager.
	manager.Shutdown()
}
