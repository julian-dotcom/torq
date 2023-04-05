package messages

import (
	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/internal/lightning"
)

func signMessage(req SignMessageRequest) (SignMessageResponse, error) {
	if req.NodeId == 0 {
		return SignMessageResponse{}, errors.New("Node Id missing")
	}
	signature, err := lightning.SignMessage(req.NodeId, req.Message, req.SingleHash)
	if err != nil {
		return SignMessageResponse{}, errors.Wrapf(err, "Signing message (nodeId: %v)", req.NodeId)
	}
	return SignMessageResponse{
		Signature: signature,
	}, nil
}
