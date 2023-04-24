package services_helpers

import (
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/core"
)

func (s *ServiceStatus) String() string {
	if s == nil {
		return core.UnknownEnumString
	}
	switch *s {
	case Inactive:
		return "Inactive"
	case Active:
		return "Active"
	case Pending:
		return "Pending"
	case Initializing:
		return "Initializing"
	}
	return core.UnknownEnumString
}

func GetCoreServiceTypes() []ServiceType {
	return []ServiceType{
		RootService,
		MaintenanceService,
		AutomationIntervalTriggerService,
		AutomationChannelBalanceEventTriggerService,
		AutomationChannelEventTriggerService,
		AutomationScheduledTriggerService,
		CronService,
		NotifierService,
		SlackService,
		TelegramHighService,
		TelegramLowService,
	}
}

func GetLndServiceTypes() []ServiceType {
	return []ServiceType{
		LndServiceVectorService,
		LndServiceAmbossService,
		LndServiceRebalanceService,
		LndServiceChannelEventStream,
		LndServiceGraphEventStream,
		LndServiceTransactionStream,
		LndServiceHtlcEventStream,
		LndServiceForwardsService,
		LndServiceInvoiceStream,
		LndServicePaymentsService,
		LndServicePeerEventStream,
		LndServiceInFlightPaymentsService,
		LndServiceChannelBalanceCacheService,
	}
}

func GetClnServiceTypes() []ServiceType {
	return []ServiceType{
		ClnServiceVectorService,
		ClnServiceAmbossService,
		ClnServicePeersService,
		ClnServiceChannelsService,
		ClnServiceFundsService,
		ClnServiceNodesService,
	}
}

func (st *ServiceType) String() string {
	if st == nil {
		return core.UnknownEnumString
	}
	switch *st {
	case LndServiceVectorService:
		return "LndServiceVectorService"
	case LndServiceAmbossService:
		return "LndServiceAmbossService"
	case RootService:
		return "RootService"
	case AutomationChannelBalanceEventTriggerService:
		return "AutomationChannelBalanceEventTriggerService"
	case AutomationChannelEventTriggerService:
		return "AutomationChannelEventTriggerService"
	case AutomationIntervalTriggerService:
		return "AutomationIntervalTriggerService"
	case AutomationScheduledTriggerService:
		return "AutomationScheduledTriggerService"
	case LndServiceRebalanceService:
		return "LndServiceRebalanceService"
	case MaintenanceService:
		return "MaintenanceService"
	case CronService:
		return "CronService"
	case NotifierService:
		return "NotifierService"
	case SlackService:
		return "SlackService"
	case TelegramHighService:
		return "TelegramHighService"
	case TelegramLowService:
		return "TelegramLowService"
	case LndServiceChannelEventStream:
		return "LndServiceChannelEventStream"
	case LndServiceGraphEventStream:
		return "LndServiceGraphEventStream"
	case LndServiceTransactionStream:
		return "LndServiceTransactionStream"
	case LndServiceHtlcEventStream:
		return "LndServiceHtlcEventStream"
	case LndServiceForwardsService:
		return "LndServiceForwardsService"
	case LndServiceInvoiceStream:
		return "LndServiceInvoiceStream"
	case LndServicePaymentsService:
		return "LndServicePaymentsService"
	case LndServicePeerEventStream:
		return "LndServicePeerEventStream"
	case LndServiceInFlightPaymentsService:
		return "LndServiceInFlightPaymentsService"
	case LndServiceChannelBalanceCacheService:
		return "LndServiceChannelBalanceCacheService"
	case ClnServiceVectorService:
		return "ClnServiceVectorService"
	case ClnServiceAmbossService:
		return "ClnServiceAmbossService"
	case ClnServicePeersService:
		return "ClnServicePeersService"
	case ClnServiceChannelsService:
		return "ClnServiceChannelsService"
	case ClnServiceFundsService:
		return "ClnServiceFundsService"
	case ClnServiceNodesService:
		return "ClnServiceNodesService"
	}
	return core.UnknownEnumString
}

func (st *ServiceType) IsChannelBalanceCache() bool {
	if st != nil && (*st == LndServiceForwardsService ||
		*st == LndServiceInvoiceStream ||
		*st == LndServicePaymentsService ||
		*st == LndServicePeerEventStream ||
		*st == LndServiceChannelEventStream ||
		*st == LndServiceGraphEventStream ||
		*st == ClnServicePeersService ||
		*st == ClnServiceChannelsService ||
		*st == ClnServiceFundsService ||
		*st == ClnServiceNodesService) {
		return true
	}
	return false
}

func (st *ServiceType) IsLndService() bool {
	if st != nil && (*st == LndServiceVectorService ||
		*st == LndServiceAmbossService ||
		*st == LndServiceRebalanceService ||
		*st == LndServiceChannelEventStream ||
		*st == LndServiceGraphEventStream ||
		*st == LndServiceTransactionStream ||
		*st == LndServiceHtlcEventStream ||
		*st == LndServiceForwardsService ||
		*st == LndServiceInvoiceStream ||
		*st == LndServicePaymentsService ||
		*st == LndServicePeerEventStream ||
		*st == LndServiceInFlightPaymentsService ||
		*st == LndServiceChannelBalanceCacheService) {
		return true
	}
	return false
}

func (st *ServiceType) IsClnService() bool {
	if st != nil && (*st == ClnServiceVectorService ||
		*st == ClnServiceAmbossService ||
		*st == ClnServicePeersService ||
		*st == ClnServiceChannelsService ||
		*st == ClnServiceFundsService ||
		*st == ClnServiceNodesService) {
		return true
	}
	return false
}

func (st *ServiceType) GetImplementation() *core.Implementation {
	if st == nil {
		return nil
	}
	if (*st).IsLndService() {
		lnd := core.LND
		return &lnd
	}
	if (*st).IsClnService() {
		cln := core.CLN
		return &cln
	}
	return nil
}

func (st *ServiceType) GetNodeConnectionDetailCustomSettings() []core.NodeConnectionDetailCustomSettings {
	if st == nil {
		return nil
	}
	switch *st {
	case LndServicePaymentsService:
		return []core.NodeConnectionDetailCustomSettings{core.ImportFailedPayments, core.ImportPayments}
	case LndServiceHtlcEventStream:
		return []core.NodeConnectionDetailCustomSettings{core.ImportHtlcEvents}
	case LndServiceTransactionStream:
		return []core.NodeConnectionDetailCustomSettings{core.ImportTransactions}
	case LndServiceInvoiceStream:
		return []core.NodeConnectionDetailCustomSettings{core.ImportInvoices}
	case LndServiceForwardsService:
		return []core.NodeConnectionDetailCustomSettings{core.ImportForwards, core.ImportHistoricForwards}
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: ServiceType not supported")
		return nil
	}
}

func GetNodeConnectionDetailServiceType(cs core.NodeConnectionDetailCustomSettings) *ServiceType {
	switch {
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportFailedPayments),
		cs.HasNodeConnectionDetailCustomSettings(core.ImportPayments):
		lndServicePaymentStream := LndServicePaymentsService
		return &lndServicePaymentStream
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportHtlcEvents):
		lndServiceHtlcEventStream := LndServiceHtlcEventStream
		return &lndServiceHtlcEventStream
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportTransactions):
		lndServiceTransactionStream := LndServiceTransactionStream
		return &lndServiceTransactionStream
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportInvoices):
		lndServiceInvoiceStream := LndServiceInvoiceStream
		return &lndServiceInvoiceStream
	case cs.HasNodeConnectionDetailCustomSettings(core.ImportForwards),
		cs.HasNodeConnectionDetailCustomSettings(core.ImportHistoricForwards):
		lndServiceForwardStream := LndServiceForwardsService
		return &lndServiceForwardStream
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: NodeConnectionDetailCustomSettings not supported")
		return nil
	}
}

func (st *ServiceType) GetPingSystem() *core.PingSystem {
	if st == nil {
		return nil
	}
	switch *st {
	case LndServiceAmbossService, ClnServiceAmbossService:
		amboss := core.Amboss
		return &amboss
	case LndServiceVectorService, ClnServiceVectorService:
		vector := core.Vector
		return &vector
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: ServiceType not supported")
		return nil
	}
}

func GetPingSystemServiceType(ps core.PingSystem) *ServiceType {
	switch {
	case ps.HasPingSystem(core.Vector):
		vectorService := LndServiceVectorService
		return &vectorService
	case ps.HasPingSystem(core.Amboss):
		ambossService := LndServiceAmbossService
		return &ambossService
	default:
		log.Error().Msgf("DEVELOPMENT ERROR: PingSystem not supported")
		return nil
	}
}

func (i ImportType) String() string {
	switch i {
	case ImportChannelRoutingPolicies:
		return "ImportChannelRoutingPolicies"
	case ImportNodeInformation:
		return "ImportNodeInformation"
	case ImportAllChannels:
		return "ImportAllChannels"
	case ImportPendingChannels:
		return "ImportPendingChannels"
	case ImportPeerStatus:
		return "ImportPeerStatus"
	}
	return core.UnknownEnumString
}
