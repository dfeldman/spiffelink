package oracle

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"software.sslmate.com/src/go-pkcs12"
)

const p12Password = "password"

func listenSPIFFEUpdates(ctx context.Context, spiffeID spiffeid.ID, expectedCN string) error {
	client, err := workloadapi.New(ctx)
	if err != nil {
		return err
	}

	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClient(client))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Get an X509 SVID from the SPIFFE workload API
			svid, err := source.GetX509SVID()
			if err != nil {
				return err
			}

			// Check that the leaf certificate is RSA
			_, ok := svid.PrivateKey.(*rsa.PrivateKey)
			if !ok {
				return fmt.Errorf("Oracle requires RSA certificates. SPIRE needs to be configured to deliver RSA certificates.")
			}

			// Check that the CN matches the expected CN
			if svid.Certificates[0].Subject.CommonName != expectedCN {
				return fmt.Errorf("leaf certificate CN does not match expected CN")
			}

			// Write the certificates and key to a pkcs12 file
			tempFile, err := ioutil.TempFile("", "svid_*.p12")
			if err != nil {
				return err
			}

			p12Bytes, err := pkcs12.Encode(rand.Reader, svid.PrivateKey, svid.Certificates[0], svid.Certificates[1:], p12Password)
			if err != nil {
				return err
			}

			if _, err := tempFile.Write(p12Bytes); err != nil {
				return err
			}

			// Sleep for a while before getting the next update
			time.Sleep(time.Second * 30)
		}
	}
}
