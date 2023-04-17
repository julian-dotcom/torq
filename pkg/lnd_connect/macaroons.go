package lnd_connect

import (
	"context"
	"encoding/hex"

	"github.com/cockroachdb/errors"
	"gopkg.in/macaroon.v2"
)

// MacaroonCredential wraps a macaroon to implement the
// credentials.PerRPCCredentials interface.
type MacaroonCredential struct {
	*macaroon.Macaroon
}

// RequireTransportSecurity implements the PerRPCCredentials interface.
func (m MacaroonCredential) RequireTransportSecurity() bool {
	return true
}

// GetRequestMetadata implements the PerRPCCredentials interface. This method
// is required in order to pass the wrapped macaroon into the gRPC context.
// With this, the macaroon will be available within the request handling scope
// of the ultimate gRPC server implementation.
func (m MacaroonCredential) GetRequestMetadata(ctx context.Context,
	uri ...string) (map[string]string, error) {

	macBytes, err := m.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "marshal of the macaroon for request metadata")
	}

	md := make(map[string]string)
	md["macaroon"] = hex.EncodeToString(macBytes)
	return md, nil
}

// NewMacaroonCredential returns a copy of the passed macaroon wrapped in a
// MacaroonCredential struct which implements PerRPCCredentials.
func NewMacaroonCredential(m *macaroon.Macaroon) (MacaroonCredential, error) {
	ms := MacaroonCredential{}

	// The macaroon library's Clone() method has a subtle bug that doesn't
	// correctly clone all caveats. We need to use our own, safe clone
	// function instead.
	var err error
	ms.Macaroon, err = SafeCopyMacaroon(m)
	if err != nil {
		return ms, errors.Wrap(err, "safe copy of the macaroon")
	}

	return ms, nil
}

// SafeCopyMacaroon creates a copy of a macaroon that is safe to be used and
// modified. This is necessary because the macaroon library's own Clone() method
// is unsafe for certain edge cases, resulting in both the cloned and the
// original macaroons to be modified.
func SafeCopyMacaroon(mac *macaroon.Macaroon) (*macaroon.Macaroon, error) {
	macBytes, err := mac.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "marshal for safe copy of the macaroon")
	}

	newMac := &macaroon.Macaroon{}
	if err := newMac.UnmarshalBinary(macBytes); err != nil {
		return nil, errors.Wrap(err, "unmarshal for safe copy of the macaroon")
	}

	return newMac, nil
}
