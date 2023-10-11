package step

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/slerror"
	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// // Mock logger
// type mockLogger struct{}

// func (l *mockLogger) Printf(format string, args ...interface{}) {}
// func (l *mockLogger) WithFields(fields logrus.Fields) *logrus.Entry {
// 	return &logrus.Entry{}
// }

// func (l *mockLogger) Printf(format string, args ...interface{}) {}
// func (l *mockLogger) WithFields(fields logrus.Fields) *logrus.Entry {
// 	return &logrus.Entry{Logger: l.Logger}
// }

func newMockLogger() *logrus.Logger {
	ml := logrus.New()
	ml.Out = ioutil.Discard // Ensures that the logger does not print anything
	return ml
}

// Mock step function that always succeeds
func successfulStepFunc(ctx context.Context, sfi StepFuncInput) (State, StepFuncOutputMessage) {
	return nil, StepFuncOutputMessage{Errors: slerror.SLErrorList{Errors: []slerror.SLError{}}}
}

// Mock step function that always fails
func failingStepFunc(ctx context.Context, sfi StepFuncInput) (State, StepFuncOutputMessage) {
	return nil, StepFuncOutputMessage{Errors: slerror.SLErrorList{Errors: []slerror.SLError{
		slerror.New("Mock error"),
	}}}
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	sl := &spiffelinkcore.SpiffeLinkCore{Logger: newMockLogger()}
	dbc := &config.DatabaseConfig{}
	steps := []Step{
		{
			Name:              "TestStep",
			TelemetryID:       "TestTelemetry",
			CheckDependencies: successfulStepFunc,
			Pre:               successfulStepFunc,
			Execute:           successfulStepFunc,
			Post:              successfulStepFunc,
			Undo:              successfulStepFunc,
		},
	}

	// Test Execute mode
	err := Run(ctx, sl, dbc, steps, Execute)
	assert.Nil(t, err)

	// Test DryRun mode
	err = Run(ctx, sl, dbc, steps, DryRun)
	assert.Nil(t, err)

	// Test Undo mode
	err = Run(ctx, sl, dbc, steps, Undo)
	assert.Nil(t, err)

	// Test unknown mode
	err = Run(ctx, sl, dbc, steps, Mode("unknown"))
	assert.NotNil(t, err)

	// Test failing step
	steps[0].Execute = failingStepFunc
	err = Run(ctx, sl, dbc, steps, Execute)
	assert.NotNil(t, err)
}
