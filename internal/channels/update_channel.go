package channels

import (
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/pkg/commons"
)

func routingPolicyUpdate(request commons.RoutingPolicyUpdateRequest,
	eventChannel chan interface{}) (commons.RoutingPolicyUpdateResponse, error) {
	if request.NodeId == 0 {
		return commons.RoutingPolicyUpdateResponse{}, errors.New("Node id is missing")
	}
	if (request.ChannelId == nil || *request.ChannelId == 0) && request.TimeLockDelta == nil {
		return commons.RoutingPolicyUpdateResponse{}, errors.New("TimeLockDelta is missing")
	}
	responseChannel := make(chan interface{})
	request.ResponseChannel = responseChannel

	if eventChannel != nil {
		eventChannel <- request
		response := <-responseChannel
		if updateResponse, ok := response.(commons.RoutingPolicyUpdateResponse); ok {
			return updateResponse, nil
		}
	}
	return commons.RoutingPolicyUpdateResponse{},
		errors.New(
			fmt.Sprintf("Sending request to the broadcaster for nodeId: %v, channelId: %v",
				request.NodeId, request.ChannelId))
}
