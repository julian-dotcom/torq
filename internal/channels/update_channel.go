package channels

import (
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/pkg/commons"
)

func routingPolicyUpdate(request commons.RoutingPolicyUpdateRequest,
	lightningRequestChannel chan interface{}) (commons.RoutingPolicyUpdateResponse, error) {

	if lightningRequestChannel != nil {
		if commons.RunningServices[commons.LightningCommunicationService].GetStatus(request.NodeId) == commons.Active {
			responseChannel := make(chan commons.RoutingPolicyUpdateResponse)
			request.ResponseChannel = responseChannel
			lightningRequestChannel <- request
			response := <-responseChannel
			if response.Error != "" {
				return commons.RoutingPolicyUpdateResponse{}, errors.New(response.Error)
			}
			return response, nil
		} else {
			return commons.RoutingPolicyUpdateResponse{},
				errors.New(
					fmt.Sprintf("Lightning communication service is not active for nodeId: %v, channelId: %v",
						request.NodeId, request.ChannelId))
		}
	}
	return commons.RoutingPolicyUpdateResponse{},
		errors.New(
			fmt.Sprintf("Sending request to the broadcaster for nodeId: %v, channelId: %v",
				request.NodeId, request.ChannelId))
}
