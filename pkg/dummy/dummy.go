package dummy

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/dfeldman/spiffelink/pkg/shell"
	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/dfeldman/spiffelink/pkg/step"
)

// TODO write unit tests for this module
type Dummy struct {
}

func (*Dummy) GetName() string {
	return "dummy"
}

// GetUpdateSteps(context.Context, config.DatabaseConfig, shell.ShellContext, spiffelinkcore.SpiffeLinkUpdate) step.StepList
func (*Dummy) GetUpdateSteps(ctx context.Context, conf config.DatabaseConfig, shellContext shell.ShellContext, update spiffelinkcore.SpiffeLinkUpdate) step.StepList {
	return step.StepList{
		DatastoreName: "dummy",
		ID:            "dummy",
		Steps: []step.Step{
			{
				Name:        "Save the certificate to disk",
				TelemetryID: "DUMMY_SAVE_CERTIFICATE",
				Execute:     executeDummyDatastore,
			},
		},
	}
}

// The dummy package implements a simple "dummy" database. Every time the cert or trust
// bundle are updated, dummy writes them to a temp file. Then it checks that the contents
// of the file are as expected. It has similar configuration to a real database.
// This is for test and dev use only.

func saveSpiffeLinkUpdateToDisk(update spiffelinkcore.SpiffeLinkUpdate, savePath string) error {
	os.Mkdir(savePath, 0644)
	for i, bundle := range update.Bundles {
		bundleData := getCertificatesPEM(bundle.X509Authorities())
		err := ioutil.WriteFile(fmt.Sprintf("%s/bundle_%d.pem", savePath, i), bundleData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write bundle to file: %w", err)
		}
	}

	for i, svid := range update.Svids {
		svidData := getCertificatesPEM(svid.Certificates)
		err := ioutil.WriteFile(fmt.Sprintf("%s/svid_%d.pem", savePath, i), svidData, 0644)
		if err != nil {
			return fmt.Errorf("failed to write svid to file: %w", err)
		}
	}

	return nil
}

func getCertificatesPEM(certs []*x509.Certificate) []byte {
	pemData := make([]byte, 0)
	for _, cert := range certs {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}

		pemData = append(pemData, pem.EncodeToMemory(block)...)
	}

	return pemData
}

// TO do : add an output channel
func executeDummyDatastore(ctx context.Context, sfi step.StepFuncInput) (step.State, step.StepFuncOutputMessage) {
	//saveSpiffeLinkUpdateToDisk(*sfi.Update, "/dummy/")
	return nil, step.StepFuncOutputMessage{}
}
