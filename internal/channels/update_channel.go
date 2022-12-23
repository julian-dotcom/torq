package channels

import (
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/pkg/commons"
)

func routingPolicyUpdate(request commons.RoutingPolicyUpdateRequest,
	eventChannel chan interface{}) (commons.RoutingPolicyUpdateResponse, error) {

	if eventChannel != nil {
		if commons.RunningServices[commons.LightningCommunicationService].GetStatus(request.NodeId) == commons.Active {
			responseChannel := make(chan interface{})
			request.ResponseChannel = responseChannel

			eventChannel <- request
			response := <-responseChannel
			if updateResponse, ok := response.(commons.RoutingPolicyUpdateResponse); ok {
				if updateResponse.Error != "" {
					return commons.RoutingPolicyUpdateResponse{}, errors.New(updateResponse.Error)
				}
				return updateResponse, nil
			}
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
