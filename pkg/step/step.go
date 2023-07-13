package step

import (
	"context"
	"errors"
	"time"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/sirupsen/logrus"
)

type State interface{}

// Each time a StepFunc is called, it gets the same inputs. There are a lot
// so they're combined into a struct for convenience.
type StepFuncInput struct {
	// This is the global SpiffeLink state. Primarily it's here so Steps can access global config.
	sl *spiffelinkcore.SpiffeLinkCore
	// This is the database config for the currently executing database.
	dbc *config.DatabaseConfig
	// This is an opaque State interface that can be used for whatever state needs to be saved.
	// Most of the time, this is nil.
	// For example, if we need to generate an ID to store the certificate under, we can store the ID
	// temporarily here.
	state State
	// This is a pointer to the global logger.
	logger *logrus.Logger
	// This is the update from the SPIFFE workload API that we are currently processing
	update *spiffelinkcore.SpiffeLinkUpdate
}

// This is the signature for the StepFuncs that are called.
type StepFunc func(ctx context.Context, sfi StepFuncInput) (State, error)

// The goal of the Step package is to make a series of Steps that can be called.
// Each Step consists of well-defined, preconditions, postconditions, execution, and
// undo/rollback substeps. By having these in a uniform format we can implement logging,
// telemetry, web interfaces , and error handling in ways that apply to every
// supported database.
type Step struct {
	// Human-readable name for the step
	Name string
	// Telemtry ID for the step (we report time and error status to telemetry)
	TelemetryID string
	// Check the binary dependencies and configuration for the step
	CheckDependencies StepFunc
	// Check any preconditions for the step
	Pre StepFunc
	// Execute the step itself
	Execute StepFunc
	// Check any postcodnitions for the step (it will fail if this fails)
	Post StepFunc
	// Undo the step
	Undo StepFunc
}

type Mode string

const (
	Execute Mode = "execute"
	DryRun  Mode = "dryrun"
	Undo    Mode = "undo"
)

type StepBuilder interface {
	BuildSteps(sl spiffelinkcore.SpiffeLinkCore, dbc *config.DatabaseConfig) ([]Step, error)
}

func NullStepFunc(state State, update spiffelinkcore.SpiffeLinkUpdate) (State, error) {
	return state, nil
}

func Run(ctx context.Context, sl *spiffelinkcore.SpiffeLinkCore, dbc *config.DatabaseConfig, steps []Step, mode Mode, state State) error {
	logger := sl.Logger
	sfi := StepFuncInput{
		logger: logger,
		state:  state,
		dbc:    dbc,
		sl:     sl,
	}
	for _, step := range steps {
		logger.Printf("Running step: %s", step.Name)
		switch mode {
		case Execute:
			state, err := runWithLogging(ctx, step.Pre, sfi, "pre")
			if err != nil {
				return err
			}
			sfi.state = &state
			state, err = runWithLogging(ctx, step.Execute, sfi, "execute")
			if err != nil {
				return err
			}
			_, err = runWithLogging(ctx, step.Post, sfi, "post")
			if err != nil {
				return err
			}
		case DryRun:
			_, err := runWithLogging(ctx, step.Pre, sfi, "pre")
			if err != nil {
				return err
			}
		case Undo:
			_, err := runWithLogging(ctx, step.Undo, sfi, "undo")
			if err != nil {
				return err
			}
		default:
			return errors.New("unknown mode")
		}
	}
	return nil
}

func runWithLogging(
	fn StepFunc, sl *spiffelinkcore.SpiffeLinkCore, dbc *config.DatabaseConfig, state State, stage string) (State, error) {
	start := time.Now()
	state, err := fn(sl, dbc, state)
	duration := time.Since(start)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"stage":    stage,
			"error":    err,
			"duration": duration,
		}).Error("Error executing stage")
		return state, err
	}
	logger.WithFields(logrus.Fields{
		"stage":    stage,
		"duration": duration,
	}).Info("Successfully executed stage")
	return state, nil
}
