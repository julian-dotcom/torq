package channels

import (
	"fmt"
	"time"

	"github.com/lncapital/torq/pkg/commons"
)

func SetRoutingPolicyWithTimeout(request commons.RoutingPolicyUpdateRequest,
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
	responseChannel := make(chan commons.RoutingPolicyUpdateResponse, 1)
	request.ResponseChannel = responseChannel
	lightningRequestChannel <- request
	time.AfterFunc(commons.LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS*time.Second, func() {
		message := fmt.Sprintf("Routing policy update timed out after %v seconds.", commons.LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS)
		responseChannel <- commons.RoutingPolicyUpdateResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.TimedOut,
				Message: message,
				Error:   message,
			},
		}
	})
	return <-responseChannel
}

func SetRebalanceWithTimeout(request commons.RebalanceRequest,
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
	responseChannel := make(chan commons.RebalanceResponse, 1)
	request.ResponseChannel = responseChannel
	rebalanceRequestChannel <- request
	time.AfterFunc(commons.LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS*time.Second, func() {
		message := fmt.Sprintf("Routing policy update timed out after %v seconds.", commons.LIGHTNING_COMMUNICATION_TIMEOUT_SECONDS)
		responseChannel <- commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status:  commons.TimedOut,
				Message: message,
				Error:   message,
			},
		}
	})
	return <-responseChannel
}
