package zpay32

import (
	"github.com/btcsuite/btcd/btcec/v2"
)

// Signature is an interface for objects that can populate signatures during
// witness construction.
type Signature interface {
	// Serialize returns a DER-encoded ECDSA signature.
	Serialize() []byte

	// Verify return true if the ECDSA signature is valid for the passed
	// message digest under the provided public key.
	Verify([]byte, *btcec.PublicKey) bool
}
