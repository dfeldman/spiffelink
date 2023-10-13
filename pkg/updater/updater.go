package updater

import (
	"context"
	"time"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/datastore"
	"github.com/dfeldman/spiffelink/pkg/shell"
	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/dfeldman/spiffelink/pkg/step"
	"github.com/dfeldman/spiffelink/pkg/taskmanager"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// WorkloadAPIClient defines the interface for interacting with the SPIFFE Workload API.
type WorkloadAPIClient interface {
	// WatchX509Context watches for updates to the X.509 context.
	WatchX509Context(ctx context.Context, w workloadapi.X509ContextWatcher) error
	// You can add other methods you might use from the workloadapi.Client here...
}

// RealWorkloadAPIClient is an actual implementation of the WorkloadAPIClient that uses go-spiffe's workloadapi.Client.
type RealWorkloadAPIClient struct {
	client *workloadapi.Client
}

func NewRealWorkloadAPIClient(client *workloadapi.Client) *RealWorkloadAPIClient {
	return &RealWorkloadAPIClient{
		client: client,
	}
}

func (rwac *RealWorkloadAPIClient) WatchX509Context(ctx context.Context, w workloadapi.X509ContextWatcher) error {
	return rwac.client.WatchX509Context(ctx, w)
}

type Updater struct {
	config *config.Config
	client WorkloadAPIClient
	tm     taskmanager.ManagerInterface
	stores []datastore.Datastore
	logger *logrus.Logger
	sl     *spiffelinkcore.SpiffeLinkCore
}

func NewUpdater(config *config.Config, client WorkloadAPIClient, tm taskmanager.ManagerInterface, stores []datastore.Datastore, logger *logrus.Logger) *Updater {
	return &Updater{
		config: config,
		client: client,
		tm:     tm,
		stores: stores,
		logger: logger,
	}
}

func (u *Updater) Start(ctx context.Context) {
	u.logger.Info("Starting SPIFFE updater...")
	// TODO this should probably be passed in higher up the stack
	u.sl = &spiffelinkcore.SpiffeLinkCore{
		Logger: u.logger,
		Config: u.config,
	}
	err := u.client.WatchX509Context(ctx, u)
	if err != nil {
		u.logger.Fatalf("Error watching X.509 context: %v", err)
	}

}

func (u *Updater) OnX509ContextUpdate(c *workloadapi.X509Context) {
	u.logger.Info("Received SPIFFE update.")
	for _, dbConfig := range u.config.Databases {
		for _, store := range u.stores {
			if dbConfig.Name == store.GetName() {
				update := spiffelinkcore.SpiffeLinkUpdate{
					Svids:   c.SVIDs,
					Bundles: c.Bundles.Bundles(),
				}
				// TODO handle errors in GetShellContext (there are none defined right now, but in the future there might be)
				shellContext, _ := shell.GetShellContextFromConfig(dbConfig.Shell, u.logger)
				taskFunc := store.GetUpdateSteps(context.TODO(), dbConfig, shellContext, update)
				_, err := u.tm.NewTask("databaseUpdate", time.Duration(dbConfig.Timeout)*time.Second, u.stepListTaskFuncBuilder(taskFunc, &dbConfig, step.Execute))
				if err != nil {
					u.logger.Errorf("Error starting task for database %s: %v", dbConfig.Name, err)
				}
			}
		}
	}
}

func (u *Updater) OnX509ContextWatchError(err error) {
	u.logger.Errorf("OnX509ContextWatchError error: %v", err)
}

// This is just an adapter that converts the task function used in TaskManager to the format used in the Step package
func (u *Updater) stepListTaskFuncBuilder(sl step.StepList, dbc *config.DatabaseConfig, mode step.Mode) taskmanager.TaskFunc {
	return func(logger *logrus.Logger, ctx context.Context, out chan step.StepFuncOutputMessage) {
		// TODO wire in the output channel here
		// TODO make the Mode option work properly
		step.Run(ctx, u.sl, dbc, sl.Steps, mode)
	}
}
