package commons

import (
	"context"
	"sync"
	"time"
)

var ServiceChannel = make(chan ServiceChannelMessage) //nolint:gochecknoglobals

type Services struct {
	ServiceType ServiceType
	// serviceStatus: Active=Service is running normal, Inactive=Service has been cancelled, Pending=Service is booting
	serviceStatus map[int]ServiceStatus
	mu            sync.RWMutex
	runningList   map[int]context.CancelFunc
	// bootLock entry guards against running restart code whilst it's already running
	bootLock   map[int]*sync.Mutex
	bootTime   map[int]time.Time
	cancelTime map[int]time.Time
	// enforcedServiceStatus entry is a one time status enforcement for a service
	enforcedServiceStatus map[int]*ServiceStatus
	// noDelay entry is a one time no delay enforcement for a service
	noDelay map[int]bool

	// streamStatus ONLY FOR serviceType=LndSubscription
	streamStatus                       map[int]map[SubscriptionStream]ServiceStatus
	streamBootTime                     map[int]map[SubscriptionStream]time.Time
	streamInitializationPingTime       map[int]map[SubscriptionStream]time.Time
	nodeConnectionDetailCustomSettings map[int]NodeConnectionDetailCustomSettings
}

var RunningServices map[ServiceType]*Services //nolint:gochecknoglobals

func (rs *Services) AddSubscription(nodeId int, cancelFunc context.CancelFunc) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	rs.runningList[nodeId] = cancelFunc
	rs.serviceStatus[nodeId] = ServicePending

	if rs.ServiceType == LndService {
		setStreamStatuses(nodeId, rs, ServiceInactive)
	}
	return previousStatus
}

func (rs *Services) RemoveSubscription(nodeId int) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	_, exists := rs.runningList[nodeId]
	if exists {
		delete(rs.runningList, nodeId)
	}
	rs.serviceStatus[nodeId] = ServiceInactive

	if rs.ServiceType == LndService {
		setStreamStatuses(nodeId, rs, ServiceInactive)
	}
	return previousStatus
}

func (rs *Services) Cancel(
	nodeId int,
	enforcedServiceStatus *ServiceStatus,
	noDelay bool) (previousStatus ServiceStatus, currentStatus ServiceStatus) {

	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus = rs.serviceStatus[nodeId]
	_, exists := rs.runningList[nodeId]
	if exists {
		_, exists = rs.bootLock[nodeId]
		if exists && MutexLocked(rs.bootLock[nodeId]) {
			return previousStatus, ServicePending
		}
		rs.noDelay[nodeId] = noDelay
		rs.enforcedServiceStatus[nodeId] = enforcedServiceStatus
		rs.runningList[nodeId]()
		delete(rs.runningList, nodeId)
		rs.cancelTime[nodeId] = time.Now().UTC()
		rs.serviceStatus[nodeId] = ServiceInactive
		if rs.ServiceType == LndService {
			setStreamStatuses(nodeId, rs, ServiceInactive)
		}
		return previousStatus, ServiceActive
	}
	return previousStatus, ServiceInactive
}

func (rs *Services) GetEnforcedServiceStatusCheck(nodeId int) *ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	_, exists := rs.enforcedServiceStatus[nodeId]
	if exists {
		enforcedServiceStatus := rs.enforcedServiceStatus[nodeId]
		delete(rs.enforcedServiceStatus, nodeId)
		return enforcedServiceStatus
	}
	return nil
}

func (rs *Services) IsNoDelay(nodeId int) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	_, exists := rs.noDelay[nodeId]
	if exists {
		noDelay := rs.noDelay[nodeId]
		delete(rs.noDelay, nodeId)
		return noDelay
	}
	return false
}

// GetStatus return the status of the Service but in case of ServiceType = LndSubscription then streamStatuses are also verified.
// If one of the streamStatuses is not active then this function return Pending
func (rs *Services) GetStatus(nodeId int) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists {
		return ServiceInactive
	}
	if serviceStatus != ServiceActive {
		return serviceStatus
	}
	if rs.ServiceType == LndService {
		var streamStatus *ServiceStatus
		for stream, status := range rs.streamStatus[nodeId] {
			existingStatus := status
			if !isStreamOperationalForService(stream, status) {
				if streamStatus == nil {
					streamStatus = &existingStatus
				}
				if *streamStatus != status {
					return ServicePending
				}
			}
		}
		if streamStatus != nil {
			return *streamStatus
		}
	}
	return ServiceActive
}

func (rs *Services) GetNodeIds() []int {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.serviceStatus == nil {
		return []int{}
	}
	var nodeIds []int
	for nodeId := range rs.serviceStatus {
		nodeIds = append(nodeIds, nodeId)
	}
	return nodeIds
}

func (rs *Services) GetActiveNodeIds() []int {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.serviceStatus == nil {
		return []int{}
	}
	var nodeIds []int
node:
	for nodeId, serviceStatus := range rs.serviceStatus {
		if serviceStatus != ServiceActive {
			continue
		}
		if rs.ServiceType == LndService {
			for stream, status := range rs.streamStatus[nodeId] {
				if !isStreamOperationalForService(stream, status) {
					continue node
				}
			}
		}
		nodeIds = append(nodeIds, nodeId)
	}
	return nodeIds
}

// GetStreamStatus when the status of the LND Service is active then streamStatus will be returned.
func (rs *Services) GetStreamStatus(nodeId int, stream SubscriptionStream) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists {
		return ServiceInactive
	}
	switch serviceStatus {
	case ServiceBootRequested, ServiceBootRequestedWithDelay:
		return ServiceInactive
	case ServiceActive:
		streamStatus, streamStatusExists := rs.streamStatus[nodeId][stream]
		if streamStatusExists {
			return streamStatus
		}
	default:
		return serviceStatus
	}
	return ServiceInactive
}

// GetStreamBootTime when the status of the LND Service is active and streamStatus is active then bootTime will have a value
func (rs *Services) GetStreamBootTime(nodeId int, stream SubscriptionStream) *time.Time {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists || serviceStatus != ServiceActive {
		return nil
	}
	streamStatus, exists := rs.streamStatus[nodeId][stream]
	if !exists || streamStatus != ServiceActive {
		return nil
	}
	bootTime, exists := rs.streamBootTime[nodeId][stream]
	if !exists {
		return nil
	}
	return &bootTime
}

func (rs *Services) GetStreamInitializationPingTime(nodeId int, stream SubscriptionStream) *time.Time {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	_, exists := rs.streamInitializationPingTime[nodeId]
	if !exists {
		return nil
	}
	initializationPingTime, exists := rs.streamInitializationPingTime[nodeId][stream]
	if !exists {
		return nil
	}
	return &initializationPingTime
}

func (rs *Services) GetChannelBalanceCacheStreamStatus(nodeId int) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists {
		return ServiceInactive
	}
	if serviceStatus != ServiceActive {
		return serviceStatus
	}
	var streamStatus *ServiceStatus
	for _, stream := range SubscriptionStreams {
		if stream.IsChannelBalanceCache() {
			status := rs.streamStatus[nodeId][stream]
			if !isStreamOperationalForService(stream, status) {
				if streamStatus == nil {
					streamStatus = &status
				}
				if *streamStatus != status {
					return ServicePending
				}
			}
		}
	}
	if streamStatus != nil {
		return *streamStatus
	}
	return ServiceActive
}

// GetBootTime When the service is active it will return it's boot time
func (rs *Services) GetBootTime(nodeId int) *time.Time {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists || serviceStatus != ServiceActive {
		return nil
	}
	bootTime, exists := rs.bootTime[nodeId]
	if exists {
		return &bootTime
	}
	return nil
}

func (rs *Services) GetBootLock(nodeId int) *sync.Mutex {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	lock := rs.bootLock[nodeId]
	if lock == nil {
		lock = &sync.Mutex{}
		rs.bootLock[nodeId] = lock
	}
	return lock
}

func (rs *Services) HasCustomSetting(nodeId int, customSetting NodeConnectionDetailCustomSettings) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	nodeCustomSettings, exists := rs.nodeConnectionDetailCustomSettings[nodeId]
	if exists {
		return nodeCustomSettings&customSetting != 0
	}
	return false
}

func (rs *Services) SetNodeConnectionDetailCustomSettings(nodeId int, customSetting NodeConnectionDetailCustomSettings) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	rs.nodeConnectionDetailCustomSettings[nodeId] = customSetting
}

func (rs *Services) Initialising(nodeId int) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	rs.serviceStatus[nodeId] = ServiceInitializing
	return previousStatus
}

func (rs *Services) Booted(nodeId int, bootLock *sync.Mutex) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if bootLock != nil {
		bootLock.Unlock()
	}
	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	rs.bootTime[nodeId] = time.Now().UTC()
	rs.serviceStatus[nodeId] = ServiceActive

	if rs.ServiceType == LndService {
		setStreamStatuses(nodeId, rs, ServicePending)
	}
	return previousStatus
}

func (rs *Services) SetBootRequestedStatus(nodeId int, delayed bool) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	status := ServiceBootRequested
	if delayed {
		status = ServiceBootRequestedWithDelay
	}

	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	if previousStatus == status {
		return previousStatus
	}
	rs.serviceStatus[nodeId] = status
	return previousStatus
}

func (rs *Services) SetStreamStatus(nodeId int, stream SubscriptionStream, status ServiceStatus) ServiceStatus {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	if status == ServiceInitializing {
		rs.streamInitializationPingTime[nodeId][stream] = time.Now().UTC()
	}
	previousStatus := rs.streamStatus[nodeId][stream]
	if previousStatus == status {
		return previousStatus
	}
	rs.streamStatus[nodeId][stream] = status
	if status == ServiceActive {
		rs.streamBootTime[nodeId][stream] = time.Now().UTC()
	}
	return previousStatus
}

func isStreamOperationalForService(stream SubscriptionStream, status ServiceStatus) bool {
	// stream == InFlightPaymentStream: We don't care about the status of the InFlightPaymentStream
	// status == Active: The stream is healthy
	// status == Deleted: The stream is not wanted
	return stream == InFlightPaymentStream || status == ServiceActive || status == ServiceDeleted
}

func initServiceMaps(rs *Services, nodeId int) {
	if rs.runningList == nil {
		rs.runningList = make(map[int]context.CancelFunc)
		rs.serviceStatus = make(map[int]ServiceStatus)
		rs.bootLock = make(map[int]*sync.Mutex)
		rs.bootTime = make(map[int]time.Time)
		rs.cancelTime = make(map[int]time.Time)
		rs.enforcedServiceStatus = make(map[int]*ServiceStatus)
		rs.noDelay = make(map[int]bool)
		rs.streamStatus = make(map[int]map[SubscriptionStream]ServiceStatus)
		rs.streamBootTime = make(map[int]map[SubscriptionStream]time.Time)
		rs.streamInitializationPingTime = make(map[int]map[SubscriptionStream]time.Time)
		rs.nodeConnectionDetailCustomSettings = make(map[int]NodeConnectionDetailCustomSettings)
	}
	_, exists := rs.streamStatus[nodeId]
	if !exists {
		rs.streamStatus[nodeId] = make(map[SubscriptionStream]ServiceStatus)
	}
	_, exists = rs.streamBootTime[nodeId]
	if !exists {
		rs.streamBootTime[nodeId] = make(map[SubscriptionStream]time.Time)
	}
	_, exists = rs.streamInitializationPingTime[nodeId]
	if !exists {
		rs.streamInitializationPingTime[nodeId] = make(map[SubscriptionStream]time.Time)
	}
}

func setStreamStatuses(nodeId int, rs *Services, status ServiceStatus) {
	nodeStreamStatus := make(map[SubscriptionStream]ServiceStatus)
	for _, ss := range SubscriptionStreams {
		nodeStreamStatus[ss] = status
	}
	rs.streamStatus[nodeId] = nodeStreamStatus
}
