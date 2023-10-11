package datastore

import (
	"context"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/dummy"
	"github.com/dfeldman/spiffelink/pkg/shell"
	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/dfeldman/spiffelink/pkg/step"
)

//import "github.com/dfeldman/spiffelink/pkg/spiffelinkcore"

type Datastore interface {
	// Return the name of this datastore implementation
	GetName() string
	// Return the steps to update SPIFFE certs in this datastore implementation
	// TODO This interface needs to be rethought. You definitely need the database, the shell context, and the update.
	// But you don't need a context, you probably DO need some other params, and the result should have more info.
	GetUpdateSteps(context.Context, config.DatabaseConfig, shell.ShellContext, spiffelinkcore.SpiffeLinkUpdate) step.StepList
}

// TODO This is not a good pattern. Instead this should work like GetShellContextFromConfig.
func GetDatastores() []Datastore {
	return []Datastore{&dummy.Dummy{}}
}
