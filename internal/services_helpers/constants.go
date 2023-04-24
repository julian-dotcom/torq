package services_helpers

import "github.com/lncapital/torq/internal/core"

type ServiceType int

// When adding here also add to GetServiceTypes
const (
	RootService = ServiceType(iota)
	MaintenanceService
	AutomationIntervalTriggerService
	AutomationChannelEventTriggerService
	AutomationChannelBalanceEventTriggerService
	AutomationScheduledTriggerService
	CronService
	LndServiceVectorService
	LndServiceAmbossService
	LndServiceRebalanceService
	SlackService
	TelegramHighService
	TelegramLowService
	NotifierService
	LndServiceChannelEventStream
	LndServiceGraphEventStream
	LndServiceTransactionStream
	LndServiceHtlcEventStream
	LndServiceForwardsService
	LndServiceInvoiceStream
	LndServicePaymentsService
	LndServicePeerEventStream
	LndServiceInFlightPaymentsService
	LndServiceChannelBalanceCacheService
	ClnServiceVectorService
	ClnServiceAmbossService
	ClnServicePeersService
	ClnServiceChannelsService
	ClnServiceFundsService
	ClnServiceNodesService
	ClnServiceClosedChannelsService
	ClnServiceForwardsService
	ClnServiceInvoicesService
	ClnServiceHtlcsService
	ClnServiceTransactionsService
)

type ServiceStatus int

const (
	// Inactive is the initial state of a service
	Inactive = ServiceStatus(core.Inactive)
	// Active is the state when the service is fully operational
	// (so not bootstrapping but working with live data)
	Active = ServiceStatus(core.Active)
	// Pending is when the service is booted but waiting for some connection to become active
	// (not all services have this state)
	Pending = ServiceStatus(core.Pending)
	// Initializing is when a service is operational but it's performing an initialization task
	// (an example here would be importing historic data before starting to process live data)
	Initializing = ServiceStatus(core.Initializing)
)

type ImportType int

const (
	ImportChannelRoutingPolicies = ImportType(iota)
	ImportNodeInformation
	ImportAllChannels
	ImportPendingChannels
	ImportPeerStatus
)
