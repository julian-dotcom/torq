package cache

import (
	"context"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/pkg/commons"
)

var ServicesCacheChannel = make(chan ServiceCache) //nolint:gochecknoglobals

type ServiceCacheOperationType uint
type nodeId int

const (
	initStates ServiceCacheOperationType = iota
	readCurrentCoreServiceState
	readCurrentLndServiceState
	readDesiredCoreServiceState
	readDesiredLndServiceState
	readCoreServiceTime
	readLndServiceTime
	readLatestCoreFailedAttempt
	readLatestLndFailedAttempt
	readSuccessTimes
	readNodeConnectionDetails
	readActiveState
	readActiveStateChannelBalanceCache
	writeCurrentCoreServiceState
	writeCurrentLndServiceState
	writeDesiredCoreServiceState
	writeDesiredLndServiceState
	writeCurrentCoreServiceFailure
	writeCurrentLndServiceFailure
	readLndNodeIds
	writeSuccessTimes
	writeNodeConnectionDetails
	cancelCoreService
	cancelLndService
)

type ServiceCache struct {
	Type                               ServiceCacheOperationType
	NoSubscriptions                    bool
	ServiceType                        commons.ServiceType
	NodeId                             int
	ServiceStatus                      commons.ServiceStatus
	CancelFunc                         context.CancelFunc
	SuccessTimes                       map[commons.ImportType]time.Time
	LndNodeConnectionDetails           LndNodeConnectionDetails
	NodeConnectionDetailCustomSettings commons.NodeConnectionDetailCustomSettings
	SystemServiceStateOut              chan<- SystemServiceState
	ServiceStateOut                    chan<- ServiceState
	TimeOut                            chan<- *time.Time
	BoolOut                            chan<- bool
	IntsOut                            chan<- []int
	SuccessTimesOut                    chan<- map[commons.ImportType]time.Time
	LndNodeConnectionDetailsOut        chan<- LndNodeConnectionDetails
}

type SystemServiceState struct {
	NoSubscriptions          bool
	CoreServiceStates        map[commons.ServiceType]ServiceState
	LndNodeServiceStates     map[nodeId]map[commons.ServiceType]ServiceState
	LndNodeConnectionDetails map[nodeId]LndNodeConnectionDetails
}

type ServiceState struct {
	Status             commons.ServiceStatus
	CancelFunc         *context.CancelFunc
	ActiveTime         *time.Time
	PendingTime        *time.Time
	InitializationTime *time.Time
	InactivationTime   *time.Time
	FailureTime        *time.Time
}

func (ss *ServiceState) Pending(cancelFunc context.CancelFunc) ServiceState {
	now := time.Now()
	ss.Status = commons.ServicePending
	ss.CancelFunc = &cancelFunc
	ss.ActiveTime = nil
	ss.PendingTime = &now
	ss.InitializationTime = nil
	return *ss
}

func (ss *ServiceState) Initializing() ServiceState {
	now := time.Now()
	ss.Status = commons.ServiceInitializing
	ss.ActiveTime = nil
	ss.InitializationTime = &now
	return *ss
}

func (ss *ServiceState) Activate() ServiceState {
	now := time.Now()
	ss.Status = commons.ServiceActive
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
	ss.Status = commons.ServiceInactive
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

type LndNodeConnectionDetails struct {
	GRPCAddress       string
	TLSFileBytes      []byte
	MacaroonFileBytes []byte
	CustomSettings    commons.NodeConnectionDetailCustomSettings
}

func ServiceCacheHandler(ch <-chan ServiceCache, ctx context.Context) {
	successTimes := make(map[nodeId]map[commons.ImportType]time.Time)
	torqNodeConnectionDetailsCache := make(map[nodeId]LndNodeConnectionDetails)
	var torqCurrentStateCache SystemServiceState
	var torqDesiredStateCache SystemServiceState

	for {
		select {
		case <-ctx.Done():
			return
		case managedService := <-ch:
			torqCurrentStateCache, torqDesiredStateCache = handleServiceOperation(managedService,
				torqCurrentStateCache, torqDesiredStateCache, torqNodeConnectionDetailsCache, successTimes)
		}
	}
}

func handleServiceOperation(
	serviceCache ServiceCache,
	torqCurrentStateCache SystemServiceState,
	torqDesiredStateCache SystemServiceState,
	lndNodeConnectionDetailsCache map[nodeId]LndNodeConnectionDetails,
	successTimes map[nodeId]map[commons.ImportType]time.Time) (SystemServiceState, SystemServiceState) {

	switch serviceCache.Type {
	case initStates:
		coreDesiredServiceState := make(map[commons.ServiceType]ServiceState)
		for _, coreServiceType := range commons.GetCoreServiceTypes() {
			coreDesiredServiceState[coreServiceType] = ServiceState{Status: commons.ServiceActive}
		}
		torqDesiredStateCache = SystemServiceState{
			NoSubscriptions:          serviceCache.NoSubscriptions,
			CoreServiceStates:        coreDesiredServiceState,
			LndNodeConnectionDetails: make(map[nodeId]LndNodeConnectionDetails),
			LndNodeServiceStates:     make(map[nodeId]map[commons.ServiceType]ServiceState),
		}

		coreCurrentServiceState := make(map[commons.ServiceType]ServiceState)
		for _, coreServiceType := range commons.GetCoreServiceTypes() {
			coreCurrentServiceState[coreServiceType] = ServiceState{Status: commons.ServiceInactive}
		}
		torqCurrentStateCache = SystemServiceState{
			NoSubscriptions:          serviceCache.NoSubscriptions,
			CoreServiceStates:        coreCurrentServiceState,
			LndNodeConnectionDetails: make(map[nodeId]LndNodeConnectionDetails),
			LndNodeServiceStates:     make(map[nodeId]map[commons.ServiceType]ServiceState),
		}
	case readCurrentCoreServiceState:
		serviceCache.ServiceStateOut <- torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
	case readCurrentLndServiceState:
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if nodeExist {
			serviceCache.ServiceStateOut <- n[serviceCache.ServiceType]
			break
		}
		serviceCache.ServiceStateOut <- ServiceState{}
	case readDesiredCoreServiceState:
		serviceCache.ServiceStateOut <- torqDesiredStateCache.CoreServiceStates[serviceCache.ServiceType]
	case readDesiredLndServiceState:
		n, nodeExist := torqDesiredStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if nodeExist {
			serviceCache.ServiceStateOut <- n[serviceCache.ServiceType]
			break
		}
		serviceCache.ServiceStateOut <- ServiceState{}
	case readCoreServiceTime:
		var t *time.Time
		state, exist := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		if exist {
			switch serviceCache.ServiceStatus {
			case commons.ServicePending:
				t = state.PendingTime
			case commons.ServiceInitializing:
				t = state.InitializationTime
			case commons.ServiceActive:
				t = state.ActiveTime
			case commons.ServiceInactive:
				t = state.InactivationTime
			}
		}
		serviceCache.TimeOut <- t
	case readLndServiceTime:
		var t *time.Time
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if nodeExist {
			state, exists := n[serviceCache.ServiceType]
			if exists {
				switch serviceCache.ServiceStatus {
				case commons.ServicePending:
					t = state.PendingTime
				case commons.ServiceInitializing:
					t = state.InitializationTime
				case commons.ServiceActive:
					t = state.ActiveTime
				case commons.ServiceInactive:
					t = state.InactivationTime
				}
			}
		}
		serviceCache.TimeOut <- t
	case readLatestCoreFailedAttempt:
		var t *time.Time
		state, exist := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		if exist {
			t = state.FailureTime
		}
		serviceCache.TimeOut <- t
	case readLatestLndFailedAttempt:
		var t *time.Time
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if nodeExist {
			state, exists := n[serviceCache.ServiceType]
			if exists {
				t = state.FailureTime
			}
		}
		serviceCache.TimeOut <- t
	case readSuccessTimes:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
		}
		st, exists := successTimes[nodeId(serviceCache.NodeId)]
		if !exists {
			st = make(map[commons.ImportType]time.Time)
		}
		serviceCache.SuccessTimesOut <- st
	case readNodeConnectionDetails:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
		}
		serviceCache.LndNodeConnectionDetailsOut <- lndNodeConnectionDetailsCache[nodeId(serviceCache.NodeId)]
	case readActiveState:
		s := false
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if nodeExist {
			state, exists := n[serviceCache.ServiceType]
			if exists {
				s = state.Status == commons.ServiceActive
			}
		}
		serviceCache.BoolOut <- s
	case readActiveStateChannelBalanceCache:
		s := false
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if nodeExist {
			s = true
			for _, serviceType := range commons.GetLndServiceTypes() {
				if !serviceType.IsChannelBalanceCache() {
					continue
				}
				state, exists := n[serviceType]
				if !exists {
					s = false
					break
				}
				if state.Status != commons.ServiceActive {
					s = false
					break
				}
			}
		}
		serviceCache.BoolOut <- s
	case readLndNodeIds:
		var nodeIds []int
		for nId := range torqDesiredStateCache.LndNodeServiceStates {
			nodeIds = append(nodeIds, int(nId))
		}
		for nId := range torqDesiredStateCache.LndNodeConnectionDetails {
			if !slices.Contains(nodeIds, int(nId)) {
				nodeIds = append(nodeIds, int(nId))
			}
		}
		for nId := range torqCurrentStateCache.LndNodeServiceStates {
			if !slices.Contains(nodeIds, int(nId)) {
				nodeIds = append(nodeIds, int(nId))
			}
		}
		for nId := range torqCurrentStateCache.LndNodeConnectionDetails {
			if !slices.Contains(nodeIds, int(nId)) {
				nodeIds = append(nodeIds, int(nId))
			}
		}
		serviceCache.IntsOut <- nodeIds
	case writeDesiredCoreServiceState:
		state := torqDesiredStateCache.CoreServiceStates[serviceCache.ServiceType]
		state.Status = serviceCache.ServiceStatus
		torqDesiredStateCache.CoreServiceStates[serviceCache.ServiceType] = state
	case writeDesiredLndServiceState:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		n, nodeExists := torqDesiredStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if !nodeExists {
			n = make(map[commons.ServiceType]ServiceState)
		}
		state := n[serviceCache.ServiceType]
		state.Status = serviceCache.ServiceStatus
		n[serviceCache.ServiceType] = state
		torqDesiredStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
	case writeCurrentCoreServiceState:
		state := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		if state.Status == serviceCache.ServiceStatus && serviceCache.ServiceStatus == commons.ServiceInitializing {
			torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Initializing()
		}
		if state.Status != serviceCache.ServiceStatus {
			switch serviceCache.ServiceStatus {
			case commons.ServicePending:
				switch {
				case serviceCache.CancelFunc != nil:
					torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Pending(serviceCache.CancelFunc)
				case state.CancelFunc != nil:
					torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", serviceCache.CancelFunc)
				}
			case commons.ServiceActive:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Activate()
			case commons.ServiceInactive:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Inactivate()
			case commons.ServiceInitializing:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Initializing()
			}
		}
		if state.CancelFunc == nil && serviceCache.CancelFunc != nil {
			state.CancelFunc = &serviceCache.CancelFunc
			torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state
		}
	case writeCurrentLndServiceState:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		n, nodeExists := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if !nodeExists {
			n = make(map[commons.ServiceType]ServiceState)
		}
		state := n[serviceCache.ServiceType]
		if state.Status == serviceCache.ServiceStatus && serviceCache.ServiceStatus == commons.ServiceInitializing {
			n[serviceCache.ServiceType] = state.Initializing()
			torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
		}
		if state.Status != serviceCache.ServiceStatus {
			switch serviceCache.ServiceStatus {
			case commons.ServicePending:
				switch {
				case serviceCache.CancelFunc != nil:
					n[serviceCache.ServiceType] = state.Pending(serviceCache.CancelFunc)
				case state.CancelFunc != nil:
					n[serviceCache.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", serviceCache.CancelFunc)
				}
				torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
			case commons.ServiceActive:
				n[serviceCache.ServiceType] = state.Activate()
				torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
			case commons.ServiceInactive:
				n[serviceCache.ServiceType] = state.Inactivate()
				torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
			case commons.ServiceInitializing:
				n[serviceCache.ServiceType] = state.Initializing()
				torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
			}
		}
	case cancelCoreService:
		state := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Cancel()
	case cancelLndService:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		n, nodeExists := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if !nodeExists {
			n = make(map[commons.ServiceType]ServiceState)
		}
		state := n[serviceCache.ServiceType]
		n[serviceCache.ServiceType] = state.Cancel()
		torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
	case writeCurrentCoreServiceFailure:
		state := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Failure()
	case writeCurrentLndServiceFailure:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		n, nodeExists := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if !nodeExists {
			n = make(map[commons.ServiceType]ServiceState)
		}
		state := n[serviceCache.ServiceType]
		n[serviceCache.ServiceType] = state.Failure()
		torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
	case writeSuccessTimes:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		successTimes[nodeId(serviceCache.NodeId)] = serviceCache.SuccessTimes
	case writeNodeConnectionDetails:
		if serviceCache.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", serviceCache.NodeId)
			break
		}
		lndNodeConnectionDetailsCache[nodeId(serviceCache.NodeId)] = serviceCache.LndNodeConnectionDetails
	}
	return torqCurrentStateCache, torqDesiredStateCache
}

func InitStates(noSubscriptions bool) {
	managedService := ServiceCache{
		NoSubscriptions: noSubscriptions,
		Type:            initStates,
	}
	ServicesCacheChannel <- managedService
}

func GetCurrentCoreServiceState(serviceType commons.ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ServiceCache{
		ServiceType:     serviceType,
		Type:            readCurrentCoreServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func GetCurrentLndServiceState(serviceType commons.ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ServiceCache{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readCurrentLndServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func GetDesiredCoreServiceState(serviceType commons.ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ServiceCache{
		ServiceType:     serviceType,
		Type:            readDesiredCoreServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func GetDesiredLndServiceState(serviceType commons.ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ServiceCache{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readDesiredLndServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func GetCoreFailedAttemptTime(serviceType commons.ServiceType) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: commons.ServiceInactive,
		Type:          readLatestCoreFailedAttempt,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func GetLndFailedAttemptTime(serviceType commons.ServiceType, nodeId int) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        readLatestLndFailedAttempt,
		TimeOut:     responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func GetCoreServiceTime(serviceType commons.ServiceType, serviceStatus commons.ServiceStatus) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          readCoreServiceTime,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func GetLndServiceTime(serviceType commons.ServiceType, nodeId int, serviceStatus commons.ServiceStatus) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          readLndServiceTime,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func IsLndServiceActive(nodeId int) bool {
	responseChannel := make(chan bool)
	managedService := ServiceCache{
		NodeId:      nodeId,
		ServiceType: commons.LndServiceChannelEventStream,
		Type:        readActiveState,
		BoolOut:     responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func IsChannelBalanceCacheStreamActive(nodeId int) bool {
	responseChannel := make(chan bool)
	managedService := ServiceCache{
		NodeId:  nodeId,
		Type:    readActiveStateChannelBalanceCache,
		BoolOut: responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func SetDesiredCoreServiceState(serviceType commons.ServiceType, serviceStatus commons.ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v.", serviceType.String(), serviceStatus.String())
	managedService := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredCoreServiceState,
	}
	ServicesCacheChannel <- managedService
}

func SetDesiredLndServiceState(serviceType commons.ServiceType, nodeId int, serviceStatus commons.ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v with nodeId: %v.", serviceType.String(), serviceStatus.String(),
		nodeId)

	managedService := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredLndServiceState,
	}
	ServicesCacheChannel <- managedService
}

func InitRootService(cancelFunc context.CancelFunc) {
	managedService := ServiceCache{
		ServiceType:   commons.RootService,
		CancelFunc:    cancelFunc,
		ServiceStatus: commons.ServiceInactive,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- managedService
}

func InitCoreServiceState(serviceType commons.ServiceType, cancelFunc context.CancelFunc) {
	managedService := ServiceCache{
		ServiceType:   serviceType,
		CancelFunc:    cancelFunc,
		ServiceStatus: commons.ServicePending,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- managedService
}

func InitLndServiceState(serviceType commons.ServiceType, nodeId int, cancelFunc context.CancelFunc) {
	managedService := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		CancelFunc:    cancelFunc,
		ServiceStatus: commons.ServicePending,
		Type:          writeCurrentLndServiceState,
	}
	ServicesCacheChannel <- managedService
}

func SetPendingCoreServiceState(serviceType commons.ServiceType) {
	setCoreServiceStatus(serviceType, commons.ServicePending)
}

func SetPendingLndServiceState(serviceType commons.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, commons.ServicePending)
}

func SetInitializingCoreServiceState(serviceType commons.ServiceType) {
	setCoreServiceStatus(serviceType, commons.ServiceInitializing)
}

func SetInitializingLndServiceState(serviceType commons.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, commons.ServiceInitializing)
}

func SetActiveCoreServiceState(serviceType commons.ServiceType) {
	setCoreServiceStatus(serviceType, commons.ServiceActive)
}

func SetActiveLndServiceState(serviceType commons.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, commons.ServiceActive)
}

func SetInactiveCoreServiceState(serviceType commons.ServiceType) {
	setCoreServiceStatus(serviceType, commons.ServiceInactive)
}

func SetInactiveLndServiceState(serviceType commons.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, commons.ServiceInactive)
}

func CancelCoreService(serviceType commons.ServiceType) {
	log.Debug().Msgf("%v cancellation requested", serviceType.String())
	managedService := ServiceCache{
		ServiceType: serviceType,
		Type:        cancelCoreService,
	}
	ServicesCacheChannel <- managedService
}

func CancelLndService(serviceType commons.ServiceType, nodeId int) {
	log.Debug().Msgf("%v cancellation requested for nodeId: %v", serviceType.String(), nodeId)
	managedService := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        cancelLndService,
	}
	ServicesCacheChannel <- managedService

}

func SetFailedCoreServiceState(serviceType commons.ServiceType) {
	inactive := commons.ServiceInactive
	log.Debug().Msgf("%v updating current state to %v (due to failure)", serviceType.String(), (&inactive).String())
	managedService := ServiceCache{
		ServiceType: serviceType,
		Type:        writeCurrentCoreServiceFailure,
	}
	ServicesCacheChannel <- managedService
}

func SetFailedLndServiceState(serviceType commons.ServiceType, nodeId int) {
	inactive := commons.ServiceInactive
	log.Debug().Msgf("%v updating current state to %v (due to failure) for nodeId: %v", serviceType.String(),
		(&inactive).String(), nodeId)
	managedService := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        writeCurrentLndServiceFailure,
	}
	ServicesCacheChannel <- managedService
}

func setCoreServiceStatus(serviceType commons.ServiceType, serviceStatus commons.ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v", serviceType.String(), serviceStatus.String())
	managedService := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- managedService
}

func setLndServiceStatus(serviceType commons.ServiceType, nodeId int, serviceStatus commons.ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v for nodeId: %v", serviceType.String(),
		serviceStatus.String(), nodeId)

	managedService := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentLndServiceState,
	}
	ServicesCacheChannel <- managedService
}

func GetLndNodeIds() []int {
	responseChannel := make(chan []int)
	managedService := ServiceCache{
		Type:    readLndNodeIds,
		IntsOut: responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func InactivateCoreService(ctx context.Context, serviceType commons.ServiceType) bool {
	SetDesiredCoreServiceState(serviceType, commons.ServiceInactive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentCoreServiceState(serviceType)
	if state.Status == commons.ServiceInactive {
		return true
	}

	// Slow check
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state = GetCurrentCoreServiceState(serviceType)
			if state.Status == commons.ServiceInactive {
				return true
			}
		}
	}
}

func ActivateCoreService(ctx context.Context, serviceType commons.ServiceType) bool {
	SetDesiredCoreServiceState(serviceType, commons.ServiceActive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentCoreServiceState(serviceType)
	if state.Status != commons.ServiceInactive {
		return true
	}

	// Slow check
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state = GetCurrentCoreServiceState(serviceType)
			if state.Status != commons.ServiceInactive {
				return true
			}
		}
	}
}

func InactivateLndServiceState(ctx context.Context, serviceType commons.ServiceType, nodeId int) bool {
	SetDesiredLndServiceState(serviceType, nodeId, commons.ServiceInactive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentLndServiceState(serviceType, nodeId)
	if state.Status == commons.ServiceInactive {
		return true
	}

	// Slow check
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state = GetCurrentLndServiceState(serviceType, nodeId)
			if state.Status == commons.ServiceInactive {
				return true
			}
		}
	}
}

func ActivateLndServiceState(ctx context.Context, serviceType commons.ServiceType, nodeId int) bool {
	SetDesiredLndServiceState(serviceType, nodeId, commons.ServiceActive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentLndServiceState(serviceType, nodeId)
	if state.Status != commons.ServiceInactive {
		return true
	}

	// Slow check
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state = GetCurrentLndServiceState(serviceType, nodeId)
			if state.Status != commons.ServiceInactive {
				return true
			}
		}
	}
}

func InactivateLndService(ctx context.Context, nodeId int) bool {
	for _, lndServiceType := range commons.GetLndServiceTypes() {
		SetDesiredLndServiceState(lndServiceType, nodeId, commons.ServiceInactive)
	}

	// Fast check in case the state is already what we wanted
	allGood := true
	for _, lndServiceType := range commons.GetLndServiceTypes() {
		state := GetCurrentLndServiceState(lndServiceType, nodeId)
		if state.Status != commons.ServiceInactive {
			allGood = false
		}
	}
	if allGood {
		return true
	}

	// Slow check
	ticker := clock.New().Tick(1 * time.Second)
recheck:
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			for _, lndServiceType := range commons.GetLndServiceTypes() {
				state := GetCurrentLndServiceState(lndServiceType, nodeId)
				if state.Status != commons.ServiceInactive {
					continue recheck
				}
			}
			return true
		}
	}
}

func ActivateLndService(ctx context.Context,
	nodeId int,
	customSettings commons.NodeConnectionDetailCustomSettings,
	pingSystem commons.PingSystem) bool {

	var relavantServiceTypes []commons.ServiceType
	for _, lndServiceType := range commons.GetLndServiceTypes() {
		switch lndServiceType {
		case commons.VectorService, commons.AmbossService:
			if pingSystem&(*lndServiceType.GetPingSystem()) != 0 {
				relavantServiceTypes = append(relavantServiceTypes, lndServiceType)
			}
		case commons.LndServiceTransactionStream,
			commons.LndServiceHtlcEventStream,
			commons.LndServiceForwardStream,
			commons.LndServiceInvoiceStream,
			commons.LndServicePaymentStream:
			active := false
			for _, cs := range lndServiceType.GetNodeConnectionDetailCustomSettings() {
				if customSettings&cs != 0 {
					active = true
					break
				}
			}
			if active {
				relavantServiceTypes = append(relavantServiceTypes, lndServiceType)
			}
		default:
			relavantServiceTypes = append(relavantServiceTypes, lndServiceType)
		}
	}
	for _, lndServiceType := range relavantServiceTypes {
		SetDesiredLndServiceState(lndServiceType, nodeId, commons.ServiceActive)
	}

	// Fast check in case the state is already what we wanted
	allGood := true
	for _, lndServiceType := range relavantServiceTypes {
		state := GetCurrentLndServiceState(lndServiceType, nodeId)
		if state.Status == commons.ServiceInactive {
			allGood = false
		}
	}
	if allGood {
		return true
	}

	// Slow check
	ticker := clock.New().Tick(1 * time.Second)
recheck:
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			for _, lndServiceType := range relavantServiceTypes {
				state := GetCurrentLndServiceState(lndServiceType, nodeId)
				if state.Status == commons.ServiceInactive {
					continue recheck
				}
			}
			return true
		}
	}
}

func GetSuccessTimes(nodeId int) map[commons.ImportType]time.Time {
	responseChannel := make(chan map[commons.ImportType]time.Time)
	managedService := ServiceCache{
		NodeId:          nodeId,
		Type:            readSuccessTimes,
		SuccessTimesOut: responseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-responseChannel
}

func SetSuccessTimes(nodeId int, successTimes map[commons.ImportType]time.Time) {
	managedService := ServiceCache{
		NodeId:       nodeId,
		Type:         writeSuccessTimes,
		SuccessTimes: successTimes,
	}
	ServicesCacheChannel <- managedService
}

func GetLndNodeConnectionDetails(nodeId int) LndNodeConnectionDetails {
	nodeResponseChannel := make(chan LndNodeConnectionDetails)
	managedService := ServiceCache{
		NodeId:                      nodeId,
		Type:                        readNodeConnectionDetails,
		LndNodeConnectionDetailsOut: nodeResponseChannel,
	}
	ServicesCacheChannel <- managedService
	return <-nodeResponseChannel
}

func SetLndNodeConnectionDetails(nodeId int, details LndNodeConnectionDetails) {
	managedService := ServiceCache{
		NodeId:                   nodeId,
		Type:                     writeNodeConnectionDetails,
		LndNodeConnectionDetails: details,
	}
	ServicesCacheChannel <- managedService
}

func HasCustomSetting(nodeId int, customSetting commons.NodeConnectionDetailCustomSettings) bool {
	ncd := GetLndNodeConnectionDetails(nodeId)
	return ncd.CustomSettings.HasNodeConnectionDetailCustomSettings(customSetting)
}
