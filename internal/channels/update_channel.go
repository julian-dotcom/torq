package channels

import (
	"fmt"
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func SetRoutingPolicyWithTimeout(request commons.RoutingPolicyUpdateRequest,
	lightningRequestChannel chan interface{}) commons.RoutingPolicyUpdateResponse {

	if lightningRequestChannel != nil {
		if commons.RunningServices[commons.LightningCommunicationService].GetStatus(request.NodeId) == commons.Active {
			responseChannel := make(chan commons.RoutingPolicyUpdateResponse, 1)
			request.ResponseChannel = responseChannel
			lightningRequestChannel <- request
			time.AfterFunc(2*time.Second, func() {
				responseChannel <- commons.RoutingPolicyUpdateResponse{
					Request: request,
					CommunicationResponse: commons.CommunicationResponse{
						Status:  commons.TimedOut,
						Message: "Routing policy update timed out after 2 seconds.",
						Error:   "Routing policy update timed out after 2 seconds.",
					},
				}
			})
			return <-responseChannel
		} else {
			message := fmt.Sprintf("Lightning communication service is not active for nodeId: %v, channelId: %v",
				request.NodeId, request.ChannelId)
			return commons.RoutingPolicyUpdateResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status:  commons.Inactive,
					Message: message,
					Error:   message,
				},
			}
		}
	}
	message := fmt.Sprintf("Lightning request channel is nil for nodeId: %v, channelId: %v",
		request.NodeId, request.ChannelId)
	return commons.RoutingPolicyUpdateResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status:  commons.Inactive,
			Message: message,
			Error:   message,
		},
	}
}
