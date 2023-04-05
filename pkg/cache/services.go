package cache

import (
	"context"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/pkg/core"
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
	ServiceType                        core.ServiceType
	NodeId                             int
	ServiceStatus                      core.ServiceStatus
	CancelFunc                         context.CancelFunc
	SuccessTimes                       map[core.ImportType]time.Time
	LndNodeConnectionDetails           LndNodeConnectionDetails
	NodeConnectionDetailCustomSettings core.NodeConnectionDetailCustomSettings
	SystemServiceStateOut              chan<- SystemServiceState
	ServiceStateOut                    chan<- ServiceState
	TimeOut                            chan<- *time.Time
	BoolOut                            chan<- bool
	IntsOut                            chan<- []int
	SuccessTimesOut                    chan<- map[core.ImportType]time.Time
	LndNodeConnectionDetailsOut        chan<- LndNodeConnectionDetails
}

type SystemServiceState struct {
	NoSubscriptions          bool
	CoreServiceStates        map[core.ServiceType]ServiceState
	LndNodeServiceStates     map[nodeId]map[core.ServiceType]ServiceState
	LndNodeConnectionDetails map[nodeId]LndNodeConnectionDetails
}

type ServiceState struct {
	Status             core.ServiceStatus
	CancelFunc         *context.CancelFunc
	ActiveTime         *time.Time
	PendingTime        *time.Time
	InitializationTime *time.Time
	InactivationTime   *time.Time
	FailureTime        *time.Time
}

func (ss *ServiceState) Pending(cancelFunc context.CancelFunc) ServiceState {
	now := time.Now()
	ss.Status = core.ServicePending
	ss.CancelFunc = &cancelFunc
	ss.ActiveTime = nil
	ss.PendingTime = &now
	ss.InitializationTime = nil
	return *ss
}

func (ss *ServiceState) Initializing() ServiceState {
	now := time.Now()
	ss.Status = core.ServiceInitializing
	ss.ActiveTime = nil
	ss.InitializationTime = &now
	return *ss
}

func (ss *ServiceState) Activate() ServiceState {
	now := time.Now()
	ss.Status = core.ServiceActive
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
	ss.Status = core.ServiceInactive
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
	CustomSettings    core.NodeConnectionDetailCustomSettings
}

func ServiceCacheHandler(ch <-chan ServiceCache, ctx context.Context) {
	successTimes := make(map[nodeId]map[core.ImportType]time.Time)
	torqNodeConnectionDetailsCache := make(map[nodeId]LndNodeConnectionDetails)
	var torqCurrentStateCache SystemServiceState
	var torqDesiredStateCache SystemServiceState

	for {
		select {
		case <-ctx.Done():
			return
		case serviceCache := <-ch:
			torqCurrentStateCache, torqDesiredStateCache = handleServiceOperation(serviceCache,
				torqCurrentStateCache, torqDesiredStateCache, torqNodeConnectionDetailsCache, successTimes)
		}
	}
}

func handleServiceOperation(
	serviceCache ServiceCache,
	torqCurrentStateCache SystemServiceState,
	torqDesiredStateCache SystemServiceState,
	lndNodeConnectionDetailsCache map[nodeId]LndNodeConnectionDetails,
	successTimes map[nodeId]map[core.ImportType]time.Time) (SystemServiceState, SystemServiceState) {

	switch serviceCache.Type {
	case initStates:
		coreDesiredServiceState := make(map[core.ServiceType]ServiceState)
		for _, coreServiceType := range core.GetCoreServiceTypes() {
			coreDesiredServiceState[coreServiceType] = ServiceState{Status: core.ServiceActive}
		}
		torqDesiredStateCache = SystemServiceState{
			NoSubscriptions:          serviceCache.NoSubscriptions,
			CoreServiceStates:        coreDesiredServiceState,
			LndNodeConnectionDetails: make(map[nodeId]LndNodeConnectionDetails),
			LndNodeServiceStates:     make(map[nodeId]map[core.ServiceType]ServiceState),
		}

		coreCurrentServiceState := make(map[core.ServiceType]ServiceState)
		for _, coreServiceType := range core.GetCoreServiceTypes() {
			coreCurrentServiceState[coreServiceType] = ServiceState{Status: core.ServiceInactive}
		}
		torqCurrentStateCache = SystemServiceState{
			NoSubscriptions:          serviceCache.NoSubscriptions,
			CoreServiceStates:        coreCurrentServiceState,
			LndNodeConnectionDetails: make(map[nodeId]LndNodeConnectionDetails),
			LndNodeServiceStates:     make(map[nodeId]map[core.ServiceType]ServiceState),
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
			case core.ServicePending:
				t = state.PendingTime
			case core.ServiceInitializing:
				t = state.InitializationTime
			case core.ServiceActive:
				t = state.ActiveTime
			case core.ServiceInactive:
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
				case core.ServicePending:
					t = state.PendingTime
				case core.ServiceInitializing:
					t = state.InitializationTime
				case core.ServiceActive:
					t = state.ActiveTime
				case core.ServiceInactive:
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
			st = make(map[core.ImportType]time.Time)
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
				s = state.Status == core.ServiceActive
			}
		}
		serviceCache.BoolOut <- s
	case readActiveStateChannelBalanceCache:
		s := false
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)]
		if nodeExist {
			s = true
			for _, serviceType := range core.GetLndServiceTypes() {
				if !serviceType.IsChannelBalanceCache() {
					continue
				}
				state, exists := n[serviceType]
				if !exists {
					s = false
					break
				}
				if state.Status != core.ServiceActive {
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
			n = make(map[core.ServiceType]ServiceState)
		}
		state := n[serviceCache.ServiceType]
		state.Status = serviceCache.ServiceStatus
		n[serviceCache.ServiceType] = state
		torqDesiredStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
	case writeCurrentCoreServiceState:
		state := torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType]
		if state.Status == serviceCache.ServiceStatus && serviceCache.ServiceStatus == core.ServiceInitializing {
			torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Initializing()
		}
		if state.Status != serviceCache.ServiceStatus {
			switch serviceCache.ServiceStatus {
			case core.ServicePending:
				switch {
				case serviceCache.CancelFunc != nil:
					torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Pending(serviceCache.CancelFunc)
				case state.CancelFunc != nil:
					torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", serviceCache.CancelFunc)
				}
			case core.ServiceActive:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Activate()
			case core.ServiceInactive:
				torqCurrentStateCache.CoreServiceStates[serviceCache.ServiceType] = state.Inactivate()
			case core.ServiceInitializing:
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
			n = make(map[core.ServiceType]ServiceState)
		}
		state := n[serviceCache.ServiceType]
		if state.Status == serviceCache.ServiceStatus && serviceCache.ServiceStatus == core.ServiceInitializing {
			n[serviceCache.ServiceType] = state.Initializing()
			torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
		}
		if state.Status != serviceCache.ServiceStatus {
			switch serviceCache.ServiceStatus {
			case core.ServicePending:
				switch {
				case serviceCache.CancelFunc != nil:
					n[serviceCache.ServiceType] = state.Pending(serviceCache.CancelFunc)
				case state.CancelFunc != nil:
					n[serviceCache.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", serviceCache.CancelFunc)
				}
				torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
			case core.ServiceActive:
				n[serviceCache.ServiceType] = state.Activate()
				torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
			case core.ServiceInactive:
				n[serviceCache.ServiceType] = state.Inactivate()
				torqCurrentStateCache.LndNodeServiceStates[nodeId(serviceCache.NodeId)] = n
			case core.ServiceInitializing:
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
			n = make(map[core.ServiceType]ServiceState)
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
			n = make(map[core.ServiceType]ServiceState)
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
	serviceCache := ServiceCache{
		NoSubscriptions: noSubscriptions,
		Type:            initStates,
	}
	ServicesCacheChannel <- serviceCache
}

func GetCurrentCoreServiceState(serviceType core.ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		Type:            readCurrentCoreServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetCurrentLndServiceState(serviceType core.ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readCurrentLndServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetDesiredCoreServiceState(serviceType core.ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		Type:            readDesiredCoreServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetDesiredLndServiceState(serviceType core.ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	serviceCache := ServiceCache{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readDesiredLndServiceState,
		ServiceStateOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetCoreFailedAttemptTime(serviceType core.ServiceType) *time.Time {
	responseChannel := make(chan *time.Time)
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: core.ServiceInactive,
		Type:          readLatestCoreFailedAttempt,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetLndFailedAttemptTime(serviceType core.ServiceType, nodeId int) *time.Time {
	responseChannel := make(chan *time.Time)
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        readLatestLndFailedAttempt,
		TimeOut:     responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func GetCoreServiceTime(serviceType core.ServiceType, serviceStatus core.ServiceStatus) *time.Time {
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

func GetLndServiceTime(serviceType core.ServiceType, nodeId int, serviceStatus core.ServiceStatus) *time.Time {
	responseChannel := make(chan *time.Time)
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          readLndServiceTime,
		TimeOut:       responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func IsLndServiceActive(nodeId int) bool {
	responseChannel := make(chan bool)
	serviceCache := ServiceCache{
		NodeId:      nodeId,
		ServiceType: core.LndServiceChannelEventStream,
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

func SetDesiredCoreServiceState(serviceType core.ServiceType, serviceStatus core.ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v.", serviceType.String(), serviceStatus.String())
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func SetDesiredLndServiceState(serviceType core.ServiceType, nodeId int, serviceStatus core.ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v with nodeId: %v.", serviceType.String(), serviceStatus.String(),
		nodeId)

	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredLndServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func InitRootService(cancelFunc context.CancelFunc) {
	serviceCache := ServiceCache{
		ServiceType:   core.RootService,
		CancelFunc:    cancelFunc,
		ServiceStatus: core.ServiceInactive,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func InitCoreServiceState(serviceType core.ServiceType, cancelFunc context.CancelFunc) {
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		CancelFunc:    cancelFunc,
		ServiceStatus: core.ServicePending,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func InitLndServiceState(serviceType core.ServiceType, nodeId int, cancelFunc context.CancelFunc) {
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		CancelFunc:    cancelFunc,
		ServiceStatus: core.ServicePending,
		Type:          writeCurrentLndServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func SetPendingCoreServiceState(serviceType core.ServiceType) {
	setCoreServiceStatus(serviceType, core.ServicePending)
}

func SetPendingLndServiceState(serviceType core.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, core.ServicePending)
}

func SetInitializingCoreServiceState(serviceType core.ServiceType) {
	setCoreServiceStatus(serviceType, core.ServiceInitializing)
}

func SetInitializingLndServiceState(serviceType core.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, core.ServiceInitializing)
}

func SetActiveCoreServiceState(serviceType core.ServiceType) {
	setCoreServiceStatus(serviceType, core.ServiceActive)
}

func SetActiveLndServiceState(serviceType core.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, core.ServiceActive)
}

func SetInactiveCoreServiceState(serviceType core.ServiceType) {
	setCoreServiceStatus(serviceType, core.ServiceInactive)
}

func SetInactiveLndServiceState(serviceType core.ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, core.ServiceInactive)
}

func CancelCoreService(serviceType core.ServiceType) {
	log.Debug().Msgf("%v cancellation requested", serviceType.String())
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		Type:        cancelCoreService,
	}
	ServicesCacheChannel <- serviceCache
}

func CancelLndService(serviceType core.ServiceType, nodeId int) {
	log.Debug().Msgf("%v cancellation requested for nodeId: %v", serviceType.String(), nodeId)
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        cancelLndService,
	}
	ServicesCacheChannel <- serviceCache

}

func SetFailedCoreServiceState(serviceType core.ServiceType) {
	inactive := core.ServiceInactive
	log.Debug().Msgf("%v updating current state to %v (due to failure)", serviceType.String(), (&inactive).String())
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		Type:        writeCurrentCoreServiceFailure,
	}
	ServicesCacheChannel <- serviceCache
}

func SetFailedLndServiceState(serviceType core.ServiceType, nodeId int) {
	inactive := core.ServiceInactive
	log.Debug().Msgf("%v updating current state to %v (due to failure) for nodeId: %v", serviceType.String(),
		(&inactive).String(), nodeId)
	serviceCache := ServiceCache{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        writeCurrentLndServiceFailure,
	}
	ServicesCacheChannel <- serviceCache
}

func setCoreServiceStatus(serviceType core.ServiceType, serviceStatus core.ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v", serviceType.String(), serviceStatus.String())
	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentCoreServiceState,
	}
	ServicesCacheChannel <- serviceCache
}

func setLndServiceStatus(serviceType core.ServiceType, nodeId int, serviceStatus core.ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v for nodeId: %v", serviceType.String(),
		serviceStatus.String(), nodeId)

	serviceCache := ServiceCache{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentLndServiceState,
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

func InactivateCoreService(ctx context.Context, serviceType core.ServiceType) bool {
	SetDesiredCoreServiceState(serviceType, core.ServiceInactive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentCoreServiceState(serviceType)
	if state.Status == core.ServiceInactive {
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
			if state.Status == core.ServiceInactive {
				return true
			}
		}
	}
}

func ActivateCoreService(ctx context.Context, serviceType core.ServiceType) bool {
	SetDesiredCoreServiceState(serviceType, core.ServiceActive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentCoreServiceState(serviceType)
	if state.Status != core.ServiceInactive {
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
			if state.Status != core.ServiceInactive {
				return true
			}
		}
	}
}

func InactivateLndServiceState(ctx context.Context, serviceType core.ServiceType, nodeId int) bool {
	SetDesiredLndServiceState(serviceType, nodeId, core.ServiceInactive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentLndServiceState(serviceType, nodeId)
	if state.Status == core.ServiceInactive {
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
			if state.Status == core.ServiceInactive {
				return true
			}
		}
	}
}

func ActivateLndServiceState(ctx context.Context, serviceType core.ServiceType, nodeId int) bool {
	SetDesiredLndServiceState(serviceType, nodeId, core.ServiceActive)

	// Fast check in case the state is already what we wanted
	state := GetCurrentLndServiceState(serviceType, nodeId)
	if state.Status != core.ServiceInactive {
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
			if state.Status != core.ServiceInactive {
				return true
			}
		}
	}
}

func InactivateLndService(ctx context.Context, nodeId int) bool {
	for _, lndServiceType := range core.GetLndServiceTypes() {
		SetDesiredLndServiceState(lndServiceType, nodeId, core.ServiceInactive)
	}

	// Fast check in case the state is already what we wanted
	allGood := true
	for _, lndServiceType := range core.GetLndServiceTypes() {
		state := GetCurrentLndServiceState(lndServiceType, nodeId)
		if state.Status != core.ServiceInactive {
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
			for _, lndServiceType := range core.GetLndServiceTypes() {
				state := GetCurrentLndServiceState(lndServiceType, nodeId)
				if state.Status != core.ServiceInactive {
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

	var relavantServiceTypes []core.ServiceType
	for _, lndServiceType := range core.GetLndServiceTypes() {
		switch lndServiceType {
		case core.VectorService, core.AmbossService:
			if pingSystem&(*lndServiceType.GetPingSystem()) != 0 {
				relavantServiceTypes = append(relavantServiceTypes, lndServiceType)
			}
		case core.LndServiceTransactionStream,
			core.LndServiceHtlcEventStream,
			core.LndServiceForwardStream,
			core.LndServiceInvoiceStream,
			core.LndServicePaymentStream:
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
		SetDesiredLndServiceState(lndServiceType, nodeId, core.ServiceActive)
	}

	// Fast check in case the state is already what we wanted
	allGood := true
	for _, lndServiceType := range relavantServiceTypes {
		state := GetCurrentLndServiceState(lndServiceType, nodeId)
		if state.Status == core.ServiceInactive {
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
				if state.Status == core.ServiceInactive {
					continue recheck
				}
			}
			return true
		}
	}
}

func GetSuccessTimes(nodeId int) map[core.ImportType]time.Time {
	responseChannel := make(chan map[core.ImportType]time.Time)
	serviceCache := ServiceCache{
		NodeId:          nodeId,
		Type:            readSuccessTimes,
		SuccessTimesOut: responseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-responseChannel
}

func SetSuccessTimes(nodeId int, successTimes map[core.ImportType]time.Time) {
	serviceCache := ServiceCache{
		NodeId:       nodeId,
		Type:         writeSuccessTimes,
		SuccessTimes: successTimes,
	}
	ServicesCacheChannel <- serviceCache
}

func GetLndNodeConnectionDetails(nodeId int) LndNodeConnectionDetails {
	nodeResponseChannel := make(chan LndNodeConnectionDetails)
	serviceCache := ServiceCache{
		NodeId:                      nodeId,
		Type:                        readNodeConnectionDetails,
		LndNodeConnectionDetailsOut: nodeResponseChannel,
	}
	ServicesCacheChannel <- serviceCache
	return <-nodeResponseChannel
}

func SetLndNodeConnectionDetails(nodeId int, details LndNodeConnectionDetails) {
	serviceCache := ServiceCache{
		NodeId:                   nodeId,
		Type:                     writeNodeConnectionDetails,
		LndNodeConnectionDetails: details,
	}
	ServicesCacheChannel <- serviceCache
}

func HasCustomSetting(nodeId int, customSetting core.NodeConnectionDetailCustomSettings) bool {
	ncd := GetLndNodeConnectionDetails(nodeId)
	return ncd.CustomSettings.HasNodeConnectionDetailCustomSettings(customSetting)
}
