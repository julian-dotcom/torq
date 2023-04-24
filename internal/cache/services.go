package cache

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/services_helpers"
)

var ServicesCacheChannel = make(chan ServiceCache) //nolint:gochecknoglobals

type ServiceCacheOperationType uint
type nodeIdType int

const (
	initStates ServiceCacheOperationType = iota
	readCurrentCoreServiceState
	readCurrentNodeServiceState
	readDesiredCoreServiceState
	readDesiredNodeServiceState
	readCoreServiceTime
	readNodeServiceTime
	readLatestCoreFailedAttempt
	readLatestNodeFailedAttempt
	readSuccessTimes
	readNodeConnectionDetails
	readActiveState
	readActiveStateChannelBalanceCache
	writeCurrentCoreServiceState
	writeCurrentNodeServiceState
	writeDesiredCoreServiceState
	writeDesiredNodeServiceState
	writeCurrentCoreServiceFailure
	writeCurrentNodeServiceFailure
	readLndNodeIds
	readClnNodeIds
	writeSuccessTimes
	writeNodeConnectionDetails
	cancelCoreService
	cancelNodeService
)

type ServiceCache struct {
	Type                               ServiceCacheOperationType
	NoSubscriptions                    bool
	ServiceType                        services_helpers.ServiceType
	NodeId                             int
	ServiceStatus                      services_helpers.ServiceStatus
	CancelFunc                         context.CancelFunc
	SuccessTimes                       map[services_helpers.ImportType]time.Time
	NodeConnectionDetails              NodeConnectionDetails
	NodeConnectionDetailCustomSettings core.NodeConnectionDetailCustomSettings
	SystemServiceStateOut              chan<- SystemServiceState
	ServiceStateOut                    chan<- ServiceState
	TimeOut                            chan<- *time.Time
	BoolOut                            chan<- bool
	IntsOut                            chan<- []int
	SuccessTimesOut                    chan<- map[services_helpers.ImportType]time.Time
	NodeConnectionDetailsOut           chan<- NodeConnectionDetails
}

type SystemServiceState struct {
	NoSubscriptions   bool
	CoreServiceStates map[services_helpers.ServiceType]ServiceState
	NodeServiceStates map[core.Implementation]map[nodeIdType]map[services_helpers.ServiceType]ServiceState
}

type ServiceState struct {
	Status             services_helpers.ServiceStatus
	CancelFunc         *context.CancelFunc
	ActiveTime         *time.Time
	PendingTime        *time.Time
	InitializationTime *time.Time
	InactivationTime   *time.Time
	FailureTime        *time.Time
}

func (ss *ServiceState) Pending(cancelFunc context.CancelFunc) ServiceState {
	now := time.Now()
	ss.Status = services_helpers.Pending
	ss.CancelFunc = &cancelFunc
	ss.ActiveTime = nil
	ss.PendingTime = &now
	ss.InitializationTime = nil
	return *ss
}

func (ss *ServiceState) Initializing() ServiceState {
	now := time.Now()
	ss.Status = services_helpers.Initializing
	ss.ActiveTime = nil
	ss.InitializationTime = &now
	return *ss
}

func (ss *ServiceState) Activate() ServiceState {
	now := time.Now()
	ss.Status = services_helpers.Active
	ss.ActiveTime = &now
	ss.PendingTime = nil
	ss.InitializationTime = nil
	ss.InactivationTime = nil
	ss.FailureTime = nil
	return *ss
}

func (ss *ServiceState) Cancel() ServiceState {
	if ss.CancelFunc != nil {
		log.Debug().Msgf("Cancel function called.")
		(*ss.CancelFunc)()
		ss.CancelFunc = nil
	}
	return *ss
}

func (ss *ServiceState) Inactivate() ServiceState {
	now := time.Now()
	ss.Cancel()
	ss.Status = services_helpers.Inactive
	ss.PendingTime = nil
	ss.InitializationTime = nil
	ss.InactivationTime = &now
	ss.FailureTime = nil
	ss.CancelFunc = nil
	return *ss
}

func (ss *ServiceState) Failure() ServiceState {
	now := time.Now()
	ss.Inactivate()
	ss.FailureTime = &now
	return *ss
}

type NodeConnectionDetails struct {
	Implementation         core.Implementation
	GRPCAddress            string
	TLSFileBytes           []byte
	MacaroonFileBytes      []byte
	CertificateFileBytes   []byte
	KeyFileBytes           []byte
	CaCertificateFileBytes []byte
	CustomSettings         core.NodeConnectionDetailCustomSettings
}

func ServiceCacheHandler(ch <-chan ServiceCache, ctx context.Context) {
	successTimes := make(map[nodeIdType]map[services_helpers.ImportType]time.Time)
	nodeConnectionDetailsCache := make(map[nodeIdType]NodeConnectionDetails)
	var torqCurrentStateCache SystemServiceState
	var torqDesiredStateCache SystemServiceState

	for {
		select {
		case <-ctx.Done():
			return
		case serviceCache := <-ch:
			torqCurrentStateCache, torqDesiredStateCache = handleServiceOperation(serviceCache,
				torqCurrentStateCache, torqDesiredStateCache, nodeConnectionDetailsCache, successTimes)
		}
	}
}

func handleServiceOperation(
	serviceCache ServiceCache,
	torqCurrentStateCache SystemServiceState,
	torqDesiredStateCache SystemServiceState,
	nodeConnectionDetailsCache map[nodeIdType]NodeConnectionDetails,
	successTimes map[nodeIdType]map[services_helpers.ImportType]time.Time) (SystemServiceState, SystemServiceState) {

	switch serviceCache.Type {
	case initStates:
		coreDesiredServiceState := make(map[services_helpers.ServiceType]ServiceState)
		for _, coreServiceType := range services_helpers.GetCoreServiceTypes() {
			coreDesiredServiceState[coreServiceType] = ServiceState{Status: services_helpers.Active}
		}
		torqDesiredStateCache = SystemServiceState{
			NoSubscriptions:   serviceCache.NoSubscriptions,
			CoreServiceStates: coreDesiredServiceState,
			NodeServiceStates: make(map[core.Implementation]map[nodeIdType]map[services_helpers.ServiceType]ServiceState),
		}
		for _, implementation := range core.GetImplementations() {
			torqDesiredStateCache.NodeServiceStates[implementation] = make(map[nodeIdType]map[services_helpers.ServiceType]ServiceState)
		}

		coreCurrentServiceState := make(map[services_helpers.ServiceType]ServiceState)
		for _, coreServiceType := range services_helpers.GetCoreServiceTypes() {
			coreCurrentServiceState[coreServiceType] = ServiceState{Status: services_helpers.Inactive}
		}
		torqCurrentStateCache = SystemServiceState{
			NoSubscriptions:   serviceCache.NoSubscriptions,
			CoreServiceStates: coreCurrentServiceState,
			NodeServiceStates: make(map[core.Implementation]map[nodeIdType]map[services_helpers.ServiceType]ServiceState),
		}
		for _, implementation := range core.GetImplementations() {
			torqCurrentStateCache.NodeServiceStates[implementation] = make(map[nodeIdType]map[services_helpers.ServiceType]ServiceState)
		}
	case readCurrentCoreServiceState:
		serviceCache.ServiceStateOut <- torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
	case readCurrentNodeServiceState:
		serviceCache.ServiceStateOut <- getNodeServiceStates(serviceCache, torqCurrentStateCache)[serviceCache.ServiceType]
	case readDesiredCoreServiceState:
		serviceCache.ServiceStateOut <- torqDesiredStateCache.CoreServiceStates[serviceCache.ServiceType]
	case readDesiredNodeServiceState:
		serviceCache.ServiceStateOut <- getNodeServiceStates(serviceCache, torqDesiredStateCache)[serviceCache.ServiceType]
	case readCoreServiceTime:
		var t *time.Time
		state, exist := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		if exist {
			switch serviceCache.ServiceStatus {
			case services_helpers.Pending:
				t = state.PendingTime
			case services_helpers.Initializing:
				t = state.InitializationTime
			case services_helpers.Active:
				t = state.ActiveTime
			case services_helpers.Inactive:
				t = state.InactivationTime
			}
		}
		serviceCache.TimeOut <- t
	case readNodeServiceTime:
		serviceCache.TimeOut <- getServiceTime(serviceCache, torqCurrentStateCache)
	case readLatestCoreFailedAttempt:
		var t *time.Time
		state, exist := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		if exist {
			t = state.FailureTime
		}
		serviceCache.TimeOut <- t
	case readLatestNodeFailedAttempt:
		serviceCache.TimeOut <- getFailureTime(serviceCache, torqCurrentStateCache)
	case readSuccessTimes:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
		}
		st, exists := successTimes[nodeIdType(serviceCache.NodeId)]
		if !exists {
			st = make(map[services_helpers.ImportType]time.Time)
		}
		serviceCache.SuccessTimesOut <- st
	case readNodeConnectionDetails:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
		}
		serviceCache.NodeConnectionDetailsOut <- nodeConnectionDetailsCache[nodeIdType(serviceCache.NodeId)]
	case readActiveState:
		s := false
		n := getNodeServiceStates(serviceCache, torqCurrentStateCache)
		state, exists := n[serviceCache.ServiceType]
		if exists {
			s = state.Status == services_helpers.Active
		}
		serviceCache.BoolOut <- s
	case readActiveStateChannelBalanceCache:
		s := false
		switch nodeConnectionDetailsCache[nodeIdType(serviceCache.NodeId)].Implementation {
		case core.LND:
			n, nodeExist := torqCurrentStateCache.NodeServiceStates[core.LND][nodeIdType(serviceCache.NodeId)]
			if nodeExist {
				s = true
				for _, serviceType := range services_helpers.GetLndServiceTypes() {
					if !serviceType.IsChannelBalanceCache() {
						continue
					}
					state, exists := n[serviceType]
					if !exists {
						s = false
						break
					}
					if state.Status != services_helpers.Active {
						s = false
						break
					}
				}
			}
		case core.CLN:
			n, nodeExist := torqCurrentStateCache.NodeServiceStates[core.CLN][nodeIdType(serviceCache.NodeId)]
			if nodeExist {
				s = true
				for _, serviceType := range services_helpers.GetClnServiceTypes() {
					if !serviceType.IsChannelBalanceCache() {
						continue
					}
					state, exists := n[serviceType]
					if !exists {
						s = false
						break
					}
					if state.Status != services_helpers.Active {
						s = false
						break
					}
				}
			}
		}
		serviceCache.BoolOut <- s
	case readLndNodeIds:
		serviceCache.IntsOut <- readNodeIds(core.LND, torqDesiredStateCache.NodeServiceStates[core.LND],
			torqCurrentStateCache.NodeServiceStates[core.LND],
			nodeConnectionDetailsCache)
	case readClnNodeIds:
		serviceCache.IntsOut <- readNodeIds(core.CLN, torqDesiredStateCache.NodeServiceStates[core.CLN],
			torqCurrentStateCache.NodeServiceStates[core.CLN],
			nodeConnectionDetailsCache)
	case writeDesiredCoreServiceState:
		state := torqDesiredStateCache.CoreServiceStates[serviceCache.ServiceType]
		state.Status = serviceCache.ServiceStatus
		torqDesiredStateCache.CoreServiceStates[serviceCache.ServiceType] = state
	case writeDesiredNodeServiceState:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		writeState(serviceCache, torqDesiredStateCache)
	case writeCurrentCoreServiceState:
		state := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		if state.Status == serviceCache.ServiceStatus && serviceCache.ServiceStatus == services_helpers.Initializing {
			torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Initializing()
		}
		if state.Status != serviceCache.ServiceStatus {
			switch serviceCache.ServiceStatus {
			case services_helpers.Pending:
				switch {
				case serviceCache.CancelFunc != nil:
					torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Pending(serviceCache.CancelFunc)
				case state.CancelFunc != nil:
					torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", serviceCache.CancelFunc)
				}
			case services_helpers.Active:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Activate()
			case services_helpers.Inactive:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Inactivate()
			case services_helpers.Initializing:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Initializing()
			}
		}
		if state.CancelFunc == nil && serviceCache.CancelFunc != nil {
			state.CancelFunc = &serviceCache.CancelFunc
			torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state
		}
	case writeCurrentNodeServiceState:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		n := getNodeServiceStates(serviceCache, torqCurrentStateCache)
		state := n[serviceCache.ServiceType]
		if state.Status == serviceCache.ServiceStatus && serviceCache.ServiceStatus == services_helpers.Initializing {
			n[serviceCache.ServiceType] = state.Initializing()
			setNodeServiceStates(serviceCache, torqCurrentStateCache, n)
		}
		if state.Status != serviceCache.ServiceStatus {
			switch serviceCache.ServiceStatus {
			case services_helpers.Pending:
				switch {
				case serviceCache.CancelFunc != nil:
					n[serviceCache.ServiceType] = state.Pending(serviceCache.CancelFunc)
				case state.CancelFunc != nil:
					n[serviceCache.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", serviceCache.CancelFunc)
				}
				setNodeServiceStates(serviceCache, torqCurrentStateCache, n)
			case services_helpers.Active:
				n[serviceCache.ServiceType] = state.Activate()
				setNodeServiceStates(serviceCache, torqCurrentStateCache, n)
			case services_helpers.Inactive:
				n[serviceCache.ServiceType] = state.Inactivate()
				setNodeServiceStates(serviceCache, torqCurrentStateCache, n)
			case services_helpers.Initializing:
				n[serviceCache.ServiceType] = state.Initializing()
				setNodeServiceStates(serviceCache, torqCurrentStateCache, n)
			}
		}
	case cancelCoreService:
		state := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Cancel()
	case cancelNodeService:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		n := getNodeServiceStates(serviceCache, torqCurrentStateCache)
		state := n[serviceCache.ServiceType]
		n[serviceCache.ServiceType] = state.Cancel()
		setNodeServiceStates(serviceCache, torqCurrentStateCache, n)
	case writeCurrentCoreServiceFailure:
		state := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Failure()
	case writeCurrentNodeServiceFailure:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		n := getNodeServiceStates(serviceCache, torqCurrentStateCache)
		state := n[serviceCache.ServiceType]
		n[serviceCache.ServiceType] = state.Failure()
		setNodeServiceStates(serviceCache, torqCurrentStateCache, n)
	case writeSuccessTimes:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		successTimes[nodeIdType(serviceCache.NodeId)] = serviceCache.SuccessTimes
	case writeNodeConnectionDetails:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		nodeConnectionDetailsCache[nodeIdType(serviceCache.NodeId)] = serviceCache.NodeConnectionDetails
	}
	return torqCurrentStateCache, torqDesiredStateCache
}

func setNodeServiceStates(serviceCache ServiceCache,
	stateCache SystemServiceState,
	nodeServiceStates map[services_helpers.ServiceType]ServiceState) {

	implementation := serviceCache.ServiceType.GetImplementation()
	if implementation != nil {
		stateCache.NodeServiceStates[*implementation][nodeIdType(serviceCache.NodeId)] = nodeServiceStates
	}
}

func getNodeServiceStates(serviceCache ServiceCache, stateCache SystemServiceState) map[services_helpers.ServiceType]ServiceState {
	n := make(map[services_helpers.ServiceType]ServiceState)
	var nodeExists bool
	implementation := serviceCache.ServiceType.GetImplementation()
	if implementation != nil {
		n, nodeExists = stateCache.NodeServiceStates[*implementation][nodeIdType(serviceCache.NodeId)]
		if !nodeExists {
			n = make(map[services_helpers.ServiceType]ServiceState)
		}
	}
	return n
}

func getFailureTime(serviceCache ServiceCache, stateCache SystemServiceState) *time.Time {
	var t *time.Time
	n := getNodeServiceStates(serviceCache, stateCache)
	state, exists := n[serviceCache.ServiceType]
	if exists {
		t = state.FailureTime
	}
	return t
}

func readNodeIds(implementation core.Implementation,
	desiredStateCache map[nodeIdType]map[services_helpers.ServiceType]ServiceState,
	currentStateCache map[nodeIdType]map[services_helpers.ServiceType]ServiceState,
	nodeConnectionDetails map[nodeIdType]NodeConnectionDetails) []int {
	var nodeIds []int
	for nodeId := range desiredStateCache {
		nodeIds = append(nodeIds, int(nodeId))
	}
	for nodeId := range currentStateCache {
		if slices.Contains(nodeIds, int(nodeId)) {
			continue
		}
		nodeIds = append(nodeIds, int(nodeId))
	}
	for nodeId, ncd := range nodeConnectionDetails {
		if slices.Contains(nodeIds, int(nodeId)) {
			continue
		}
		if ncd.Implementation == implementation {
			nodeIds = append(nodeIds, int(nodeId))
		}
	}
	return nodeIds
}

func writeState(serviceCache ServiceCache, stateCache SystemServiceState) {
	n := getNodeServiceStates(serviceCache, stateCache)
	state := n[serviceCache.ServiceType]
	state.Status = serviceCache.ServiceStatus
	n[serviceCache.ServiceType] = state
	setNodeServiceStates(serviceCache, stateCache, n)
}

func getServiceTime(serviceCache ServiceCache, stateCache SystemServiceState) *time.Time {
	var t *time.Time
	n := getNodeServiceStates(serviceCache, stateCache)
	state, exists := n[serviceCache.ServiceType]
	if exists {
		switch serviceCache.ServiceStatus {
		case services_helpers.Pending:
			t = state.PendingTime
		case services_helpers.Initializing:
			t = state.InitializationTime
		case services_helpers.Active:
			t = state.ActiveTime
		case services_helpers.Inactive:
			t = state.InactivationTime
		}
	}
	return t
}

func InitStates(noSubscriptions bool) {
	serviceCache := ServiceCache{
		NoSubscriptions: noSubscriptions,
		Type:            initStates,
	}
	ServicesCacheChannel <- serviceCache
}

func GetCurrentCoreServiceState(serviceType services_helpers.ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		Type:            readCurrentCoreServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetCurrentNodeServiceState(serviceType services_helpers.ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readCurrentNodeServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetDesiredCoreServiceState(serviceType services_helpers.ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		Type:            readDesiredCoreServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetDesiredNodeServiceState(serviceType services_helpers.ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readDesiredNodeServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetCoreFailedAttemptTime(serviceType services_helpers.ServiceType) *time.Time {
	responseChannel := make(chan *time.Time)
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: services_helpers.Inactive,
		Type:          readLatestCoreFailedAttempt,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetNodeFailedAttemptTime(serviceType services_helpers.ServiceType, nodeId int) *time.Time {
	responseChannel := make(chan *time.Time)
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        readLatestNodeFailedAttempt,
		TimeOut:     responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetCoreServiceTime(serviceType services_helpers.ServiceType, serviceStatus services_helpers.ServiceStatus) *time.Time {
	responseChannel := make(chan *time.Time)
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          readCoreServiceTime,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetNodeServiceTime(serviceType services_helpers.ServiceType, nodeId int, serviceStatus services_helpers.ServiceStatus) *time.Time {
	responseChannel := make(chan *time.Time)
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          readNodeServiceTime,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func IsLndServiceActive(nodeId int) bool {
	responseChannel := make(chan bool)
	serviceCache := ServiceCache{
		NodeId:      nodeId,
		ServiceType: services_helpers.LndServiceChannelEventStream,
		Type:        readActiveState,
		BoolOut:     responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func IsClnServiceActive(nodeId int) bool {
	responseChannel := make(chan bool)
	serviceCache := ServiceCache{
		NodeId:      nodeId,
		ServiceType: services_helpers.ClnServicePeersService,
		Type:        readActiveState,
		BoolOut:     responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func IsChannelBalanceCacheStreamActive(nodeId int) bool {
	responseChannel := make(chan bool)
	serviceCache := ServiceCache{
		NodeId:  nodeId,
		Type:    readActiveStateChannelBalanceCache,
		BoolOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func SetDesiredCoreServiceState(serviceType services_helpers.ServiceType, serviceStatus services_helpers.ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v.", serviceType.String(), serviceStatus.String())
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func SetDesiredNodeServiceState(serviceType services_helpers.ServiceType, nodeId int, serviceStatus services_helpers.ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v with nodeId: %v.", serviceType.String(), serviceStatus.String(),
		nodeId)

	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredNodeServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func InitRootService(cancelFunc context.CancelFunc) {
	serviceCache := ServiceCache{
		ServiceType:   services_helpers.RootService,
		CancelFunc:    cancelFunc,
		ServiceStatus: services_helpers.Inactive,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func InitCoreServiceState(serviceType services_helpers.ServiceType, cancelFunc context.CancelFunc) {
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		CancelFunc:    cancelFunc,
		ServiceStatus: services_helpers.Pending,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func InitNodeServiceState(serviceType services_helpers.ServiceType, nodeId int, cancelFunc context.CancelFunc) {
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		CancelFunc:    cancelFunc,
		ServiceStatus: services_helpers.Pending,
		Type:          writeCurrentNodeServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func SetPendingCoreServiceState(serviceType services_helpers.ServiceType) {
	setCoreServiceStatus(serviceType, services_helpers.Pending)
}

func SetPendingNodeServiceState(serviceType services_helpers.ServiceType, nodeId int) {
	setNodeServiceStatus(serviceType, nodeId, services_helpers.Pending)
}

func SetInitializingCoreServiceState(serviceType services_helpers.ServiceType) {
	setCoreServiceStatus(serviceType, services_helpers.Initializing)
}

func SetInitializingNodeServiceState(serviceType services_helpers.ServiceType, nodeId int) {
	setNodeServiceStatus(serviceType, nodeId, services_helpers.Initializing)
}

func SetActiveCoreServiceState(serviceType services_helpers.ServiceType) {
	setCoreServiceStatus(serviceType, services_helpers.Active)
}

func SetActiveNodeServiceState(serviceType services_helpers.ServiceType, nodeId int) {
	setNodeServiceStatus(serviceType, nodeId, services_helpers.Active)
}

func SetInactiveCoreServiceState(serviceType services_helpers.ServiceType) {
	setCoreServiceStatus(serviceType, services_helpers.Inactive)
}

func SetInactiveNodeServiceState(serviceType services_helpers.ServiceType, nodeId int) {
	setNodeServiceStatus(serviceType, nodeId, services_helpers.Inactive)
}

func CancelCoreService(serviceType services_helpers.ServiceType) {
	log.Debug().Msgf("%v cancellation requested", serviceType.String())
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		Type:        cancelCoreService,
	}
	ServicesCacheChannel <- serviceCache
}

func CancelNodeService(serviceType services_helpers.ServiceType, nodeId int) {
	log.Debug().Msgf("%v cancellation requested for nodeId: %v", serviceType.String(), nodeId)
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        cancelNodeService,
	}
	ServicesCacheChannel <- serviceCache

}

func SetFailedCoreServiceState(serviceType services_helpers.ServiceType) {
	inactive := services_helpers.Inactive
	log.Debug().Msgf("%v updating current state to %v (due to failure)", serviceType.String(), (&inactive).String())
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		Type:        writeCurrentCoreServiceFailure,
	}
	ServicesCacheChannel <- serviceCache
}

func SetFailedNodeServiceState(serviceType services_helpers.ServiceType, nodeId int) {
	inactive := services_helpers.Inactive
	log.Debug().Msgf("%v updating current state to %v (due to failure) for nodeId: %v", serviceType.String(),
		(&inactive).String(), nodeId)
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        writeCurrentNodeServiceFailure,
	}
	ServicesCacheChannel <- serviceCache
}

func setCoreServiceStatus(serviceType services_helpers.ServiceType, serviceStatus services_helpers.ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v", serviceType.String(), serviceStatus.String())
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func setNodeServiceStatus(serviceType services_helpers.ServiceType, nodeId int, serviceStatus services_helpers.ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v for nodeId: %v", serviceType.String(),
		serviceStatus.String(), nodeId)

	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentNodeServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func GetLndNodeIds() []int {
	responseChannel := make(chan []int)
	serviceCache := ServiceCache{
		Type:    readLndNodeIds,
		IntsOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetClnNodeIds() []int {
	responseChannel := make(chan []int)
	serviceCache := ServiceCache{
		Type:    readClnNodeIds,
		IntsOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func InactivateCoreService(ctx context.Context, serviceType services_helpers.ServiceType) bool {
	SetDesiredCoreServiceState(serviceType, services_helpers.Inactive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentCoreServiceState(serviceType)
	if state.Status == services_helpers.Inactive {
		return true
	}

	// Slow check
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			state = GetCurrentCoreServiceState(serviceType)
			if state.Status == services_helpers.Inactive {
				return true
			}
		}
	}
}

func ActivateCoreService(ctx context.Context, serviceType services_helpers.ServiceType) bool {
	SetDesiredCoreServiceState(serviceType, services_helpers.Active)

	// Fast check in case the state is already what we wanted
	state := GetCurrentCoreServiceState(serviceType)
	if state.Status != services_helpers.Inactive {
		return true
	}

	// Slow check
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			state = GetCurrentCoreServiceState(serviceType)
			if state.Status != services_helpers.Inactive {
				return true
			}
		}
	}
}

func InactivateNodeServiceState(ctx context.Context, serviceType services_helpers.ServiceType, nodeId int) bool {
	SetDesiredNodeServiceState(serviceType, nodeId, services_helpers.Inactive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentNodeServiceState(serviceType, nodeId)
	if state.Status == services_helpers.Inactive {
		return true
	}

	// Slow check
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			state = GetCurrentNodeServiceState(serviceType, nodeId)
			if state.Status == services_helpers.Inactive {
				return true
			}
		}
	}
}

func ActivateNodeServiceState(ctx context.Context, serviceType services_helpers.ServiceType, nodeId int) bool {
	SetDesiredNodeServiceState(serviceType, nodeId, services_helpers.Active)

	// Fast check in case the state is already what we wanted
	state := GetCurrentNodeServiceState(serviceType, nodeId)
	if state.Status != services_helpers.Inactive {
		return true
	}

	// Slow check
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			state = GetCurrentNodeServiceState(serviceType, nodeId)
			if state.Status != services_helpers.Inactive {
				return true
			}
		}
	}
}

func InactivateNodeService(ctx context.Context, nodeId int) bool {
	ncd := GetNodeConnectionDetails(nodeId)
	var serviceTypes []services_helpers.ServiceType
	switch ncd.Implementation {
	case core.LND:
		serviceTypes = services_helpers.GetLndServiceTypes()
	case core.CLN:
		serviceTypes = services_helpers.GetClnServiceTypes()
	}
	for _, serviceType := range serviceTypes {
		SetDesiredNodeServiceState(serviceType, nodeId, services_helpers.Inactive)
	}

	// Fast check in case the state is already what we wanted
	allGood := true
	for _, serviceType := range serviceTypes {
		state := GetCurrentNodeServiceState(serviceType, nodeId)
		if state.Status != services_helpers.Inactive {
			allGood = false
		}
	}
	if allGood {
		return true
	}

	// Slow check
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
recheck:
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			for _, serviceType := range serviceTypes {
				state := GetCurrentNodeServiceState(serviceType, nodeId)
				if state.Status != services_helpers.Inactive {
					continue recheck
				}
			}
			return true
		}
	}
}

func ActivateLndService(ctx context.Context,
	nodeId int,
	customSettings core.NodeConnectionDetailCustomSettings,
	pingSystem core.PingSystem) bool {

	ncd := GetNodeConnectionDetails(nodeId)
	var serviceTypes []services_helpers.ServiceType
	switch ncd.Implementation {
	case core.LND:
		serviceTypes = services_helpers.GetLndServiceTypes()
	case core.CLN:
		serviceTypes = services_helpers.GetClnServiceTypes()
	}

	var relavantServiceTypes []services_helpers.ServiceType
	for _, serviceType := range serviceTypes {
		switch serviceType {
		case services_helpers.LndServiceVectorService, services_helpers.LndServiceAmbossService,
			services_helpers.ClnServiceVectorService, services_helpers.ClnServiceAmbossService:
			if pingSystem&(*serviceType.GetPingSystem()) != 0 {
				relavantServiceTypes = append(relavantServiceTypes, serviceType)
			}
		case services_helpers.LndServiceTransactionStream,
			services_helpers.LndServiceHtlcEventStream,
			services_helpers.LndServiceForwardsService,
			services_helpers.LndServiceInvoiceStream,
			services_helpers.LndServicePaymentsService:
			active := false
			for _, cs := range serviceType.GetNodeConnectionDetailCustomSettings() {
				if customSettings&cs != 0 {
					active = true
					break
				}
			}
			if active {
				relavantServiceTypes = append(relavantServiceTypes, serviceType)
			}
		default:
			relavantServiceTypes = append(relavantServiceTypes, serviceType)
		}
	}
	for _, serviceType := range relavantServiceTypes {
		SetDesiredNodeServiceState(serviceType, nodeId, services_helpers.Active)
	}

	// Fast check in case the state is already what we wanted
	allGood := true
	for _, serviceType := range relavantServiceTypes {
		state := GetCurrentNodeServiceState(serviceType, nodeId)
		if state.Status == services_helpers.Inactive {
			allGood = false
		}
	}
	if allGood {
		return true
	}

	// Slow check
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
recheck:
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			for _, serviceType := range relavantServiceTypes {
				state := GetCurrentNodeServiceState(serviceType, nodeId)
				if state.Status == services_helpers.Inactive {
					continue recheck
				}
			}
			return true
		}
	}
}

func GetSuccessTimes(nodeId int) map[services_helpers.ImportType]time.Time {
	responseChannel := make(chan map[services_helpers.ImportType]time.Time)
	serviceCache := ServiceCache{
		NodeId:          nodeId,
		Type:            readSuccessTimes,
		SuccessTimesOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func SetSuccessTimes(nodeId int, successTimes map[services_helpers.ImportType]time.Time) {
	serviceCache := ServiceCache{
		NodeId:       nodeId,
		Type:         writeSuccessTimes,
		SuccessTimes: successTimes,
	}
	ServicesCacheChannel <- serviceCache
}

func GetNodeConnectionDetails(nodeId int) NodeConnectionDetails {
	nodeResponseChannel := make(chan NodeConnectionDetails)
	serviceCache := ServiceCache{
		NodeId:                   nodeId,
		Type:                     readNodeConnectionDetails,
		NodeConnectionDetailsOut: nodeResponseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-nodeResponseChannel
}

func SetNodeConnectionDetails(nodeId int, details NodeConnectionDetails) {
	serviceCache := ServiceCache{
		NodeId:                nodeId,
		Type:                  writeNodeConnectionDetails,
		NodeConnectionDetails: details,
	}
	ServicesCacheChannel <- serviceCache
}

func HasCustomSetting(nodeId int, customSetting core.NodeConnectionDetailCustomSettings) bool {
	ncd := GetNodeConnectionDetails(nodeId)
	return ncd.CustomSettings.HasNodeConnectionDetailCustomSettings(customSetting)
}
