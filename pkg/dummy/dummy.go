package dummy

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/dfeldman/spiffelink/pkg/spiffelinkcore"
	"github.com/dfeldman/spiffelink/pkg/step"
)

// The dummy package implements a simple "dummy" database. Every time the cert or trust
// bundle are updated, dummy writes them to a temp file. Then it checks that the contents
// of the file are as expected. It has similar configuration to a real database.

func saveSpiffeLinkUpdateToDisk(update spiffelinkcore.SpiffeLinkUpdate, savePath string) error {
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


func preDummySaveCertificate(sl *spiffelinkcore.SpiffeLinkCore, dbc *config.DatabaseConfig, state step.State, update spiffelinkcore.SpiffeLinkCore) {
	return func()
	logger := 
}

func BuildSteps(sl spiffelinkcore.SpiffeLinkCore, dbc *config.DatabaseConfig) ([]step.Step, error) {
	return []step.Step{
		step.Step{
			Name:              "Save the certificate to disk",
			TelemetryID:       "DUMMY_SAVE_CERTIFICATE",
			CheckDependencies: step.NullStepFunc,
			Pre:               preDummySaveCertificate,
			Execute:           step.NullStepFunc,
			Post:              step.NullStepFunc,
			Undo:              step.NullStepFunc,
		},
	}, nil
}
