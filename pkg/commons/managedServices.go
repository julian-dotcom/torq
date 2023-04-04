package commons

import (
	"context"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

var ManagedServiceChannel = make(chan ManagedService) //nolint:gochecknoglobals

type ManagedServiceCacheOperationType uint

const (
	initStates ManagedServiceCacheOperationType = iota
	readCurrentTorqServiceState
	readCurrentLndServiceState
	readDesiredTorqServiceState
	readDesiredLndServiceState
	readTorqServiceTime
	readLndServiceTime
	readLatestTorqFailedAttempt
	readLatestLndFailedAttempt
	readSuccessTimes
	readNodeConnectionDetails
	readActiveState
	readActiveStateChannelBalanceCache
	writeCurrentTorqServiceState
	writeCurrentLndServiceState
	writeDesiredTorqServiceState
	writeDesiredLndServiceState
	writeCurrentTorqServiceFailure
	writeCurrentLndServiceFailure
	readLndNodeIds
	writeSuccessTimes
	writeNodeConnectionDetails
	cancelTorqService
	cancelLndService
)

type ManagedService struct {
	Type                               ManagedServiceCacheOperationType
	NoSubscriptions                    bool
	ServiceType                        ServiceType
	NodeId                             int
	ServiceStatus                      ServiceStatus
	CancelFunc                         context.CancelFunc
	SuccessTimes                       map[ImportType]time.Time
	LndNodeConnectionDetails           LndNodeConnectionDetails
	NodeConnectionDetailCustomSettings NodeConnectionDetailCustomSettings
	SystemServiceStateOut              chan<- SystemServiceState
	ServiceStateOut                    chan<- ServiceState
	TimeOut                            chan<- *time.Time
	BoolOut                            chan<- bool
	IntsOut                            chan<- []int
	SuccessTimesOut                    chan<- map[ImportType]time.Time
	LndNodeConnectionDetailsOut        chan<- LndNodeConnectionDetails
}

type SystemServiceState struct {
	NoSubscriptions          bool
	TorqServiceStates        map[ServiceType]ServiceState
	LndNodeServiceStates     map[int]map[ServiceType]ServiceState
	LndNodeConnectionDetails map[int]LndNodeConnectionDetails
}

type ServiceState struct {
	Status             ServiceStatus
	CancelFunc         *context.CancelFunc
	ActiveTime         *time.Time
	PendingTime        *time.Time
	InitializationTime *time.Time
	InactivationTime   *time.Time
	FailureTime        *time.Time
}

func (ss *ServiceState) Pending(cancelFunc context.CancelFunc) ServiceState {
	now := time.Now()
	ss.Status = ServicePending
	ss.CancelFunc = &cancelFunc
	ss.ActiveTime = nil
	ss.PendingTime = &now
	ss.InitializationTime = nil
	return *ss
}

func (ss *ServiceState) Initializing() ServiceState {
	now := time.Now()
	ss.Status = ServiceInitializing
	ss.ActiveTime = nil
	ss.InitializationTime = &now
	return *ss
}

func (ss *ServiceState) Activate() ServiceState {
	now := time.Now()
	ss.Status = ServiceActive
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
	ss.Status = ServiceInactive
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
	CustomSettings    NodeConnectionDetailCustomSettings
}

func ManagedServiceCache(ch <-chan ManagedService, ctx context.Context) {
	successTimes := make(map[int]map[ImportType]time.Time)
	torqNodeConnectionDetailsCache := make(map[int]LndNodeConnectionDetails)
	var torqCurrentStateCache SystemServiceState
	var torqDesiredStateCache SystemServiceState

	for {
		select {
		case <-ctx.Done():
			return
		case managedService := <-ch:
			torqCurrentStateCache, torqDesiredStateCache = processManagedService(managedService,
				torqCurrentStateCache, torqDesiredStateCache, torqNodeConnectionDetailsCache, successTimes)
		}
	}
}

func processManagedService(
	managedService ManagedService,
	torqCurrentStateCache SystemServiceState,
	torqDesiredStateCache SystemServiceState,
	torqNodeConnectionDetailsCache map[int]LndNodeConnectionDetails,
	successTimes map[int]map[ImportType]time.Time) (SystemServiceState, SystemServiceState) {

	switch managedService.Type {
	case initStates:
		torqDesiredServiceState := make(map[ServiceType]ServiceState)
		for _, torqServiceType := range GetTorqServiceTypes() {
			torqDesiredServiceState[torqServiceType] = ServiceState{Status: ServiceActive}
		}
		torqDesiredStateCache = SystemServiceState{
			NoSubscriptions:          managedService.NoSubscriptions,
			TorqServiceStates:        torqDesiredServiceState,
			LndNodeConnectionDetails: make(map[int]LndNodeConnectionDetails),
			LndNodeServiceStates:     make(map[int]map[ServiceType]ServiceState),
		}

		torqCurrentServiceState := make(map[ServiceType]ServiceState)
		for _, torqServiceType := range GetTorqServiceTypes() {
			torqCurrentServiceState[torqServiceType] = ServiceState{Status: ServiceInactive}
		}
		torqCurrentStateCache = SystemServiceState{
			NoSubscriptions:          managedService.NoSubscriptions,
			TorqServiceStates:        torqCurrentServiceState,
			LndNodeConnectionDetails: make(map[int]LndNodeConnectionDetails),
			LndNodeServiceStates:     make(map[int]map[ServiceType]ServiceState),
		}
	case readCurrentTorqServiceState:
		managedService.ServiceStateOut <- torqCurrentStateCache.TorqServiceStates[managedService.ServiceType]
	case readCurrentLndServiceState:
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if nodeExist {
			managedService.ServiceStateOut <- n[managedService.ServiceType]
			break
		}
		managedService.ServiceStateOut <- ServiceState{}
	case readDesiredTorqServiceState:
		managedService.ServiceStateOut <- torqDesiredStateCache.TorqServiceStates[managedService.ServiceType]
	case readDesiredLndServiceState:
		n, nodeExist := torqDesiredStateCache.LndNodeServiceStates[managedService.NodeId]
		if nodeExist {
			managedService.ServiceStateOut <- n[managedService.ServiceType]
			break
		}
		managedService.ServiceStateOut <- ServiceState{}
	case readTorqServiceTime:
		var t *time.Time
		state, exist := torqCurrentStateCache.TorqServiceStates[managedService.ServiceType]
		if exist {
			switch managedService.ServiceStatus {
			case ServicePending:
				t = state.PendingTime
			case ServiceInitializing:
				t = state.InitializationTime
			case ServiceActive:
				t = state.ActiveTime
			case ServiceInactive:
				t = state.InactivationTime
			}
		}
		managedService.TimeOut <- t
	case readLndServiceTime:
		var t *time.Time
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if nodeExist {
			state, exists := n[managedService.ServiceType]
			if exists {
				switch managedService.ServiceStatus {
				case ServicePending:
					t = state.PendingTime
				case ServiceInitializing:
					t = state.InitializationTime
				case ServiceActive:
					t = state.ActiveTime
				case ServiceInactive:
					t = state.InactivationTime
				}
			}
		}
		managedService.TimeOut <- t
	case readLatestTorqFailedAttempt:
		var t *time.Time
		state, exist := torqCurrentStateCache.TorqServiceStates[managedService.ServiceType]
		if exist {
			t = state.FailureTime
		}
		managedService.TimeOut <- t
	case readLatestLndFailedAttempt:
		var t *time.Time
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if nodeExist {
			state, exists := n[managedService.ServiceType]
			if exists {
				t = state.FailureTime
			}
		}
		managedService.TimeOut <- t
	case readSuccessTimes:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
		}
		st, exists := successTimes[managedService.NodeId]
		if !exists {
			st = make(map[ImportType]time.Time)
		}
		managedService.SuccessTimesOut <- st
	case readNodeConnectionDetails:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
		}
		managedService.LndNodeConnectionDetailsOut <- torqNodeConnectionDetailsCache[managedService.NodeId]
	case readActiveState:
		s := false
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if nodeExist {
			state, exists := n[managedService.ServiceType]
			if exists {
				s = state.Status == ServiceActive
			}
		}
		managedService.BoolOut <- s
	case readActiveStateChannelBalanceCache:
		s := false
		n, nodeExist := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if nodeExist {
			s = true
			for _, serviceType := range GetLndServiceTypes() {
				if !serviceType.IsChannelBalanceCache() {
					continue
				}
				state, exists := n[serviceType]
				if !exists {
					s = false
					break
				}
				if state.Status != ServiceActive {
					s = false
					break
				}
			}
		}
		managedService.BoolOut <- s
	case readLndNodeIds:
		var nodeIds []int
		for nodeId := range torqDesiredStateCache.LndNodeServiceStates {
			nodeIds = append(nodeIds, nodeId)
		}
		for nodeId := range torqDesiredStateCache.LndNodeConnectionDetails {
			if !slices.Contains(nodeIds, nodeId) {
				nodeIds = append(nodeIds, nodeId)
			}
		}
		for nodeId := range torqCurrentStateCache.LndNodeServiceStates {
			if !slices.Contains(nodeIds, nodeId) {
				nodeIds = append(nodeIds, nodeId)
			}
		}
		for nodeId := range torqCurrentStateCache.LndNodeConnectionDetails {
			if !slices.Contains(nodeIds, nodeId) {
				nodeIds = append(nodeIds, nodeId)
			}
		}
		managedService.IntsOut <- nodeIds
	case writeDesiredTorqServiceState:
		state := torqDesiredStateCache.TorqServiceStates[managedService.ServiceType]
		state.Status = managedService.ServiceStatus
		torqDesiredStateCache.TorqServiceStates[managedService.ServiceType] = state
	case writeDesiredLndServiceState:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
			break
		}
		n, nodeExists := torqDesiredStateCache.LndNodeServiceStates[managedService.NodeId]
		if !nodeExists {
			n = make(map[ServiceType]ServiceState)
		}
		state := n[managedService.ServiceType]
		state.Status = managedService.ServiceStatus
		n[managedService.ServiceType] = state
		torqDesiredStateCache.LndNodeServiceStates[managedService.NodeId] = n
	case writeCurrentTorqServiceState:
		state := torqCurrentStateCache.TorqServiceStates[managedService.ServiceType]
		if state.Status == managedService.ServiceStatus && managedService.ServiceStatus == ServiceInitializing {
			torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Initializing()
		}
		if state.Status != managedService.ServiceStatus {
			switch managedService.ServiceStatus {
			case ServicePending:
				switch {
				case managedService.CancelFunc != nil:
					torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Pending(managedService.CancelFunc)
				case state.CancelFunc != nil:
					torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", managedService.CancelFunc)
				}
			case ServiceActive:
				torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Activate()
			case ServiceInactive:
				torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Inactivate()
			case ServiceInitializing:
				torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Initializing()
			}
		}
		if state.CancelFunc == nil && managedService.CancelFunc != nil {
			state.CancelFunc = &managedService.CancelFunc
			torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state
		}
	case writeCurrentLndServiceState:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
			break
		}
		n, nodeExists := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if !nodeExists {
			n = make(map[ServiceType]ServiceState)
		}
		state := n[managedService.ServiceType]
		if state.Status == managedService.ServiceStatus && managedService.ServiceStatus == ServiceInitializing {
			n[managedService.ServiceType] = state.Initializing()
			torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId] = n
		}
		if state.Status != managedService.ServiceStatus {
			switch managedService.ServiceStatus {
			case ServicePending:
				switch {
				case managedService.CancelFunc != nil:
					n[managedService.ServiceType] = state.Pending(managedService.CancelFunc)
				case state.CancelFunc != nil:
					n[managedService.ServiceType] = state.Pending(*state.CancelFunc)
				default:
					log.Error().Msgf("No empty cancelFunc (%v) allowed", managedService.CancelFunc)
				}
				torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId] = n
			case ServiceActive:
				n[managedService.ServiceType] = state.Activate()
				torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId] = n
			case ServiceInactive:
				n[managedService.ServiceType] = state.Inactivate()
				torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId] = n
			case ServiceInitializing:
				n[managedService.ServiceType] = state.Initializing()
				torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId] = n
			}
		}
	case cancelTorqService:
		state := torqCurrentStateCache.TorqServiceStates[managedService.ServiceType]
		torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Cancel()
	case cancelLndService:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
			break
		}
		n, nodeExists := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if !nodeExists {
			n = make(map[ServiceType]ServiceState)
		}
		state := n[managedService.ServiceType]
		n[managedService.ServiceType] = state.Cancel()
		torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId] = n
	case writeCurrentTorqServiceFailure:
		state := torqCurrentStateCache.TorqServiceStates[managedService.ServiceType]
		torqCurrentStateCache.TorqServiceStates[managedService.ServiceType] = state.Failure()
	case writeCurrentLndServiceFailure:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
			break
		}
		n, nodeExists := torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId]
		if !nodeExists {
			n = make(map[ServiceType]ServiceState)
		}
		state := n[managedService.ServiceType]
		n[managedService.ServiceType] = state.Failure()
		torqCurrentStateCache.LndNodeServiceStates[managedService.NodeId] = n
	case writeSuccessTimes:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
			break
		}
		successTimes[managedService.NodeId] = managedService.SuccessTimes
	case writeNodeConnectionDetails:
		if managedService.NodeId == 0 {
			log.Error().Msgf("No empty nodeId (%v) allowed", managedService.NodeId)
			break
		}
		torqNodeConnectionDetailsCache[managedService.NodeId] = managedService.LndNodeConnectionDetails
	}
	return torqCurrentStateCache, torqDesiredStateCache
}

func InitStates(noSubscriptions bool) {
	managedService := ManagedService{
		NoSubscriptions: noSubscriptions,
		Type:            initStates,
	}
	ManagedServiceChannel <- managedService
}

func GetCurrentTorqServiceState(serviceType ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ManagedService{
		ServiceType:     serviceType,
		Type:            readCurrentTorqServiceState,
		ServiceStateOut: responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func GetCurrentLndServiceState(serviceType ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ManagedService{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readCurrentLndServiceState,
		ServiceStateOut: responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func GetDesiredTorqServiceState(serviceType ServiceType) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ManagedService{
		ServiceType:     serviceType,
		Type:            readDesiredTorqServiceState,
		ServiceStateOut: responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func GetDesiredLndServiceState(serviceType ServiceType, nodeId int) ServiceState {
	responseChannel := make(chan ServiceState)
	managedService := ManagedService{
		ServiceType:     serviceType,
		NodeId:          nodeId,
		Type:            readDesiredLndServiceState,
		ServiceStateOut: responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func GetTorqFailedAttemptTime(serviceType ServiceType) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ManagedService{
		ServiceType:   serviceType,
		ServiceStatus: ServiceInactive,
		Type:          readLatestTorqFailedAttempt,
		TimeOut:       responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func GetLndFailedAttemptTime(serviceType ServiceType, nodeId int) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ManagedService{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        readLatestLndFailedAttempt,
		TimeOut:     responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func GetTorqServiceTime(serviceType ServiceType, serviceStatus ServiceStatus) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ManagedService{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          readTorqServiceTime,
		TimeOut:       responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func GetLndServiceTime(serviceType ServiceType, nodeId int, serviceStatus ServiceStatus) *time.Time {
	responseChannel := make(chan *time.Time)
	managedService := ManagedService{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          readLndServiceTime,
		TimeOut:       responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func IsLndServiceActive(nodeId int) bool {
	responseChannel := make(chan bool)
	managedService := ManagedService{
		NodeId:      nodeId,
		ServiceType: LndServiceChannelEventStream,
		Type:        readActiveState,
		BoolOut:     responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func IsChannelBalanceCacheStreamActive(nodeId int) bool {
	responseChannel := make(chan bool)
	managedService := ManagedService{
		NodeId:  nodeId,
		Type:    readActiveStateChannelBalanceCache,
		BoolOut: responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func SetDesiredTorqServiceState(serviceType ServiceType, serviceStatus ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v.", serviceType.String(), serviceStatus.String())
	managedService := ManagedService{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredTorqServiceState,
	}
	ManagedServiceChannel <- managedService
}

func SetDesiredLndServiceState(serviceType ServiceType, nodeId int, serviceStatus ServiceStatus) {
	log.Info().Msgf("%v desired state is now %v with nodeId: %v.", serviceType.String(), serviceStatus.String(),
		nodeId)

	managedService := ManagedService{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeDesiredLndServiceState,
	}
	ManagedServiceChannel <- managedService
}

func InitTorqService(cancelFunc context.CancelFunc) {
	managedService := ManagedService{
		ServiceType:   TorqService,
		CancelFunc:    cancelFunc,
		ServiceStatus: ServiceInactive,
		Type:          writeCurrentTorqServiceState,
	}
	ManagedServiceChannel <- managedService
}

func InitTorqServiceState(serviceType ServiceType, cancelFunc context.CancelFunc) {
	managedService := ManagedService{
		ServiceType:   serviceType,
		CancelFunc:    cancelFunc,
		ServiceStatus: ServicePending,
		Type:          writeCurrentTorqServiceState,
	}
	ManagedServiceChannel <- managedService
}

func InitLndServiceState(serviceType ServiceType, nodeId int, cancelFunc context.CancelFunc) {
	managedService := ManagedService{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		CancelFunc:    cancelFunc,
		ServiceStatus: ServicePending,
		Type:          writeCurrentLndServiceState,
	}
	ManagedServiceChannel <- managedService
}

func SetPendingTorqServiceState(serviceType ServiceType) {
	setTorqServiceStatus(serviceType, ServicePending)
}

func SetPendingLndServiceState(serviceType ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, ServicePending)
}

func SetInitializingTorqServiceState(serviceType ServiceType) {
	setTorqServiceStatus(serviceType, ServiceInitializing)
}

func SetInitializingLndServiceState(serviceType ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, ServiceInitializing)
}

func SetActiveTorqServiceState(serviceType ServiceType) {
	setTorqServiceStatus(serviceType, ServiceActive)
}

func SetActiveLndServiceState(serviceType ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, ServiceActive)
}

func SetInactiveTorqServiceState(serviceType ServiceType) {
	setTorqServiceStatus(serviceType, ServiceInactive)
}

func SetInactiveLndServiceState(serviceType ServiceType, nodeId int) {
	setLndServiceStatus(serviceType, nodeId, ServiceInactive)
}

func CancelTorqService(serviceType ServiceType) {
	log.Debug().Msgf("%v cancellation requested", serviceType.String())
	managedService := ManagedService{
		ServiceType: serviceType,
		Type:        cancelTorqService,
	}
	ManagedServiceChannel <- managedService
}

func CancelLndService(serviceType ServiceType, nodeId int) {
	log.Debug().Msgf("%v cancellation requested for nodeId: %v", serviceType.String(), nodeId)
	managedService := ManagedService{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        cancelLndService,
	}
	ManagedServiceChannel <- managedService

}

func SetFailedTorqServiceState(serviceType ServiceType) {
	inactive := ServiceInactive
	log.Debug().Msgf("%v updating current state to %v (due to failure)", serviceType.String(), (&inactive).String())
	managedService := ManagedService{
		ServiceType: serviceType,
		Type:        writeCurrentTorqServiceFailure,
	}
	ManagedServiceChannel <- managedService
}

func SetFailedLndServiceState(serviceType ServiceType, nodeId int) {
	inactive := ServiceInactive
	log.Debug().Msgf("%v updating current state to %v (due to failure) for nodeId: %v", serviceType.String(),
		(&inactive).String(), nodeId)
	managedService := ManagedService{
		ServiceType: serviceType,
		NodeId:      nodeId,
		Type:        writeCurrentLndServiceFailure,
	}
	ManagedServiceChannel <- managedService
}

func setTorqServiceStatus(serviceType ServiceType, serviceStatus ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v", serviceType.String(), serviceStatus.String())
	managedService := ManagedService{
		ServiceType:   serviceType,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentTorqServiceState,
	}
	ManagedServiceChannel <- managedService
}

func setLndServiceStatus(serviceType ServiceType, nodeId int, serviceStatus ServiceStatus) {
	log.Debug().Msgf("%v updating current state to %v for nodeId: %v", serviceType.String(),
		serviceStatus.String(), nodeId)

	managedService := ManagedService{
		ServiceType:   serviceType,
		NodeId:        nodeId,
		ServiceStatus: serviceStatus,
		Type:          writeCurrentLndServiceState,
	}
	ManagedServiceChannel <- managedService
}

func GetLndNodeIds() []int {
	responseChannel := make(chan []int)
	managedService := ManagedService{
		Type:    readLndNodeIds,
		IntsOut: responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func InactivateTorqService(ctx context.Context, serviceType ServiceType) bool {
	SetDesiredTorqServiceState(serviceType, ServiceInactive)
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state := GetCurrentTorqServiceState(serviceType)
			if state.Status == ServiceInactive {
				return true
			}
		}
	}
}

func ActivateTorqService(ctx context.Context, serviceType ServiceType) bool {
	SetDesiredTorqServiceState(serviceType, ServiceActive)
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state := GetCurrentTorqServiceState(serviceType)
			if state.Status != ServiceInactive {
				return true
			}
		}
	}
}

func InactivateLndServiceState(ctx context.Context, serviceType ServiceType, nodeId int) bool {
	SetDesiredLndServiceState(serviceType, nodeId, ServiceInactive)
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state := GetCurrentLndServiceState(serviceType, nodeId)
			if state.Status == ServiceInactive {
				return true
			}
		}
	}
}

func ActivateLndServiceState(ctx context.Context, serviceType ServiceType, nodeId int) bool {
	SetDesiredLndServiceState(serviceType, nodeId, ServiceActive)
	ticker := clock.New().Tick(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			state := GetCurrentLndServiceState(serviceType, nodeId)
			if state.Status != ServiceInactive {
				return true
			}
		}
	}
}

func InactivateLndService(ctx context.Context, nodeId int) bool {
	for _, lndServiceType := range GetLndServiceTypes() {
		SetDesiredLndServiceState(lndServiceType, nodeId, ServiceInactive)
	}
	ticker := clock.New().Tick(1 * time.Second)
recheck:
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			for _, lndServiceType := range GetLndServiceTypes() {
				state := GetCurrentLndServiceState(lndServiceType, nodeId)
				if state.Status != ServiceInactive {
					continue recheck
				}
			}
			return true
		}
	}
}

func ActivateLndService(ctx context.Context,
	nodeId int,
	customSettings NodeConnectionDetailCustomSettings,
	pingSystem PingSystem) bool {

	var relavantServiceTypes []ServiceType
	for _, lndServiceType := range GetLndServiceTypes() {
		switch lndServiceType {
		case VectorService, AmbossService:
			if pingSystem&(*lndServiceType.GetPingSystem()) != 0 {
				relavantServiceTypes = append(relavantServiceTypes, lndServiceType)
			}
		case LndServiceTransactionStream,
			LndServiceHtlcEventStream,
			LndServiceForwardStream,
			LndServiceInvoiceStream,
			LndServicePaymentStream:
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
		SetDesiredLndServiceState(lndServiceType, nodeId, ServiceActive)
	}
	ticker := clock.New().Tick(1 * time.Second)
recheck:
	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker:
			for _, lndServiceType := range relavantServiceTypes {
				state := GetCurrentLndServiceState(lndServiceType, nodeId)
				if state.Status == ServiceInactive {
					continue recheck
				}
			}
			return true
		}
	}
}

func GetSuccessTimes(nodeId int) map[ImportType]time.Time {
	responseChannel := make(chan map[ImportType]time.Time)
	managedService := ManagedService{
		NodeId:          nodeId,
		Type:            readSuccessTimes,
		SuccessTimesOut: responseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-responseChannel
}

func SetSuccessTimes(nodeId int, successTimes map[ImportType]time.Time) {
	managedService := ManagedService{
		NodeId:       nodeId,
		Type:         writeSuccessTimes,
		SuccessTimes: successTimes,
	}
	ManagedServiceChannel <- managedService
}

func GetLndNodeConnectionDetails(nodeId int) LndNodeConnectionDetails {
	nodeResponseChannel := make(chan LndNodeConnectionDetails)
	managedService := ManagedService{
		NodeId:                      nodeId,
		Type:                        readNodeConnectionDetails,
		LndNodeConnectionDetailsOut: nodeResponseChannel,
	}
	ManagedServiceChannel <- managedService
	return <-nodeResponseChannel
}

func SetLndNodeConnectionDetails(nodeId int, details LndNodeConnectionDetails) {
	managedService := ManagedService{
		NodeId:                   nodeId,
		Type:                     writeNodeConnectionDetails,
		LndNodeConnectionDetails: details,
	}
	ManagedServiceChannel <- managedService
}

func HasCustomSetting(nodeId int, customSetting NodeConnectionDetailCustomSettings) bool {
	ncd := GetLndNodeConnectionDetails(nodeId)
	return ncd.CustomSettings.HasNodeConnectionDetailCustomSettings(customSetting)
}
