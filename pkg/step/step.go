package step

import (
	"context"
	"time"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/shell"
	"github.com/dfeldman/spiffelink/pkg/slerror"
	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/sirupsen/logrus"
)

type State interface{}

// TODO move StepFuncOutputMessage to spiffelinkcore and give it a better name

// Each StepFunc in a Step gets the same inputs. There are a lot
// so they're combined into a struct for convenience.
type StepFuncInput struct {
	// This is the global SpiffeLink state. Primarily it's here so Steps can access global config.
	Sl *spiffelinkcore.SpiffeLinkCore
	// This is the database config for the currently executing database.
	Dbc *config.DatabaseConfig
	// This is an opaque State interface that can be used for whatever state needs to be saved.
	// Most of the time, this is nil.
	// For example, if we need to generate an ID to store the certificate under, we can store the ID
	// temporarily here.
	State State
	// This is a pointer to the global logger.
	Logger *logrus.Logger
	// This is the update from the SPIFFE workload API that we are currently processing.
	//   The format of this struct is: type SpiffeLinkUpdate struct {
	//     Bundles []*x509bundle.Bundle
	//     Svids   []*x509svid.SVID
	//   }
	Update *spiffelinkcore.SpiffeLinkUpdate
	// This is the shellContext that the function can use for executing OS and filesystem commands.
	ShellContext shell.ShellContext
}

type StepFuncOutputMessage struct {
	Name     string
	Id       string
	Stage    string
	Time     time.Time
	Complete bool
	Errors   slerror.SLErrorList
}

func (sfi *StepFuncOutputMessage) success() bool {
	return sfi.Errors.Empty()
}

// This is the signature for the StepFuncs that are called.
type StepFunc func(ctx context.Context, sfi StepFuncInput) (State, StepFuncOutputMessage)

// The goal of the Step package is to make a series of Steps that can be called.
// Each Step consists of well-defined, preconditions, postconditions, execution, and
// undo/rollback substeps. By having these in a uniform format we can implement logging,
// telemetry, web interfaces , and error handling in ways that apply to every
// supported database.
// Each of the StepFuncs gets the same input, the StepFuncInput structure above.
type Step struct {
	// Human-readable name for the step
	Name string
	// Unique ID for this step
	Id string
	// Telemtry ID for the step (we report time and error status to telemetry)
	TelemetryID string
	// Check the binary dependencies and configuration for the step
	CheckDependencies StepFunc
	// Check any preconditions for the step
	Pre StepFunc
	// Execute the step itself
	Execute StepFunc
	// Check any postconditions for the step (it will fail if this fails)
	Post StepFunc
	// Undo the step
	Undo StepFunc
}

// Several Steps make a StepList
type StepList struct {
	DatastoreName string
	ID            string
	Steps         []Step
}

type Mode string

const (
	Execute Mode = "execute"
	DryRun  Mode = "dryrun"
	Undo    Mode = "undo"
)

type StepBuilder interface {
	BuildSteps(sl spiffelinkcore.SpiffeLinkCore, dbc *config.DatabaseConfig) (StepList, slerror.SLError)
}

func NullStepFunc(state State, update spiffelinkcore.SpiffeLinkUpdate) (State, StepFuncOutputMessage) {
	return state, StepFuncOutputMessage{}
}

// Run a list of steps
// TODO This code is incomplete. It should:
// 1: Supply output messages on an output channel
// 2: Actually check if the context has been cancelled after calling each sub-step.
func Run(ctx context.Context, sl *spiffelinkcore.SpiffeLinkCore, dbc *config.DatabaseConfig, steps []Step, mode Mode) []StepFuncOutputMessage {
	logger := sl.Logger
	sfi := StepFuncInput{
		Logger: logger,
		State:  nil,
		Dbc:    dbc,
		Sl:     sl,
		Update: nil,
	}
	var state State
	for _, step := range steps {
		logger.Printf("Running step: %s", step.Name)
		outputs := []StepFuncOutputMessage{}
		output := StepFuncOutputMessage{}
		switch mode {
		case Execute:
			if step.Pre != nil {
				state, output := runWithLogging(ctx, step.Pre, sfi, "pre")
				output.Stage = "Pre"
				outputs = append(outputs, output)
				if !output.Errors.Empty() {

					return outputs
				}
				sfi.State = &state
			}
			if step.Execute != nil {
				state, output = runWithLogging(ctx, step.Execute, sfi, "execute")
				output.Stage = "Execute"
				outputs = append(outputs, output)
				if !output.Errors.Empty() {
					return outputs
				}
				sfi.State = &state
			}
			if step.Post != nil {
				_, output = runWithLogging(ctx, step.Post, sfi, "post")
				output.Stage = "post"
				outputs = append(outputs, output)
				if !output.Errors.Empty() {
					return outputs
				}
			}
		case DryRun:
			if step.Pre != nil {
				_, output := runWithLogging(ctx, step.Pre, sfi, "pre")
				output.Stage = "pre"
				outputs = append(outputs, output)
				if !output.Errors.Empty() {
					return outputs
				}
			}
		case Undo:
			if step.Undo != nil {
				_, output := runWithLogging(ctx, step.Undo, sfi, "undo")
				output.Stage = "undo"
				outputs = append(outputs, output)
				if !output.Errors.Empty() {
					return outputs
				}
			}
		default:
			return outputs
		}
	}
	return nil
}

func runWithLogging(ctx context.Context, fn StepFunc, sfi StepFuncInput, stage string) (State, StepFuncOutputMessage) {
	start := time.Now()
	state, output := fn(ctx, sfi)
	duration := time.Since(start)
	if !output.Errors.Empty() {
		sfi.Logger.WithFields(logrus.Fields{
			"stage":    stage,
			"duration": duration,
		}).Error("Error executing stage")

		return state, output
	}
	sfi.Logger.WithFields(logrus.Fields{
		"stage":    stage,
		"duration": duration,
	}).Info("Successfully executed stage")
	return state, output
}
