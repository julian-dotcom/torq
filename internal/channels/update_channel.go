package channels

import (
	"fmt"

	"github.com/lncapital/torq/pkg/commons"
)

func SetRoutingPolicy(request commons.RoutingPolicyUpdateRequest,
	lightningRequestChannel chan interface{}) commons.RoutingPolicyUpdateResponse {

	if lightningRequestChannel == nil {
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
	if commons.RunningServices[commons.LightningCommunicationService].GetStatus(request.NodeId) != commons.Active {
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
	responseChannel := make(chan commons.RoutingPolicyUpdateResponse)
	request.ResponseChannel = responseChannel
	lightningRequestChannel <- request
	return <-responseChannel
}

func SetRebalance(request commons.RebalanceRequest,
	rebalanceRequestChannel chan commons.RebalanceRequest) commons.RebalanceResponse {

	if rebalanceRequestChannel == nil {
		message := fmt.Sprintf("Lightning request channel is nil for nodeId: %v", request.NodeId)
		return commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Inactive,
				Message: message,
				Error:   message,
			},
		}
	}
	if commons.RunningServices[commons.RebalanceService].GetStatus(request.NodeId) != commons.Active {
		message := fmt.Sprintf("Lightning communication service is not active for nodeId: %v", request.NodeId)
		return commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.Inactive,
				Message: message,
				Error:   message,
			},
		}
	}
	responseChannel := make(chan commons.RebalanceResponse)
	request.ResponseChannel = responseChannel
	rebalanceRequestChannel <- request
	return <-responseChannel
}
