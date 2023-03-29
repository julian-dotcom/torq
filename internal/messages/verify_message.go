package messages

import (
	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/pkg/lightning"
)

func verifyMessage(req VerifyMessageRequest) (VerifyMessageResponse, error) {
	if req.NodeId == 0 {
		return VerifyMessageResponse{}, errors.New("Node Id missing")
	}
	publicKey, valid, err := lightning.SignatureVerification(req.NodeId, req.Message, req.Signature)
	if err != nil {
		return VerifyMessageResponse{}, errors.Wrapf(err, "Signature Verification (nodeId: %v)", req.NodeId)
	}
	return VerifyMessageResponse{
		Valid:  valid,
		PubKey: publicKey,
	}, nil
}
