package lnd

import (
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/lightning_helpers"
)

func ImportAllChannels(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := ImportAllChannelsRequest{
		ImportRequest: ImportRequest{
			CommunicationRequest: lightning_helpers.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := ImportAllChannelsUnshared(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportPendingChannels(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := ImportPendingChannelsRequest{
		ImportRequest: ImportRequest{
			CommunicationRequest: lightning_helpers.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := ImportPendingChannelsUnshared(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportChannelRoutingPolicies(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := ImportChannelRoutingPoliciesRequest{
		ImportRequest: ImportRequest{
			CommunicationRequest: lightning_helpers.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := ImportChannelRoutingPoliciesUnshared(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportNodeInformation(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := ImportNodeInformationRequest{
		ImportRequest: ImportRequest{
			CommunicationRequest: lightning_helpers.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := ImportNodeInformationUnshared(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func ImportPeerStatus(db *sqlx.DB,
	force bool,
	nodeId int) error {

	request := ImportPeerStatusRequest{
		ImportRequest: ImportRequest{
			CommunicationRequest: lightning_helpers.CommunicationRequest{
				NodeId: nodeId,
			},
			Db:    db,
			Force: force,
		},
	}
	response := ImportPeerStatusUnshared(request)
	if response.Error != nil {
		return response.Error
	}
	return nil
}
