package spiffelinkcore

import (
	"github.com/dfeldman/spiffelink/pkg/config"
	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"

	"github.com/sirupsen/logrus"
)

type SpiffeLinkCore struct {
	Logger *logrus.Logger
	Config *config.Config
}

// This aggregates all the SVIDs and bundles retrieved from the Workload API.
// There is no guarantee that all of these are valid, or that the SVIDs correspond
// to the bundles (any given SVID might not have a related bundle).
type SpiffeLinkUpdate struct {
	Bundles []*x509bundle.Bundle
	Svids   []*x509svid.SVID
}
