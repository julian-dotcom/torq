package commons

import (
	"sync"
	"time"
)

type Services struct {
	ServiceType ServiceType
	// serviceStatus: Active=Service is running normal, Inactive=Service has been cancelled, Pending=Service is booting
	serviceStatus map[int]Status
	mu            sync.RWMutex
	runningList   map[int]func()
	// bootLock entry guards against running restart code whilst it's already running
	bootLock   map[int]*sync.Mutex
	bootTime   map[int]time.Time
	cancelTime map[int]time.Time
	// enforcedServiceStatus entry is a one time status enforcement for a service
	enforcedServiceStatus map[int]*Status
	// noDelay entry is a one time no delay enforcement for a service
	noDelay map[int]bool

	// streamStatus ONLY FOR serviceType=LndSubscription
	streamStatus   map[int]map[SubscriptionStream]Status
	streamBootTime map[int]map[SubscriptionStream]time.Time
}

var RunningServices map[ServiceType]*Services //nolint:gochecknoglobals

func (rs *Services) AddSubscription(nodeId int, cancelFunc func(), serviceEventChannel chan ServiceEvent) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	rs.runningList[nodeId] = cancelFunc
	rs.serviceStatus[nodeId] = Pending

	if rs.ServiceType == LndService {
		setStreamStatuses(nodeId, rs, Inactive)
	}

	sendServiceEvent(nodeId, serviceEventChannel, previousStatus, rs.serviceStatus[nodeId], rs.ServiceType, nil)
}

func (rs *Services) RemoveSubscription(nodeId int, serviceEventChannel chan ServiceEvent) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	_, exists := rs.runningList[nodeId]
	if exists {
		delete(rs.runningList, nodeId)
		rs.serviceStatus[nodeId] = Inactive
	}

	if rs.ServiceType == LndService {
		setStreamStatuses(nodeId, rs, Inactive)
	}

	sendServiceEvent(nodeId, serviceEventChannel, previousStatus, rs.serviceStatus[nodeId], rs.ServiceType, nil)
}

func (rs *Services) Cancel(nodeId int, enforcedServiceStatus *Status, noDelay bool, serviceEventChannel chan ServiceEvent) Status {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	_, exists := rs.runningList[nodeId]
	if exists {
		_, exists = rs.bootLock[nodeId]
		if exists && MutexLocked(rs.bootLock[nodeId]) {
			return Pending
		} else {
			rs.noDelay[nodeId] = noDelay
			rs.enforcedServiceStatus[nodeId] = enforcedServiceStatus
			rs.runningList[nodeId]()
			delete(rs.runningList, nodeId)
			rs.cancelTime[nodeId] = time.Now().UTC()
			rs.serviceStatus[nodeId] = Inactive
			setStreamStatuses(nodeId, rs, Inactive)
			sendServiceEvent(nodeId, serviceEventChannel, previousStatus, rs.serviceStatus[nodeId], rs.ServiceType, nil)
			return Active
		}
	}
	return Inactive
}

func (rs *Services) GetEnforcedServiceStatusCheck(nodeId int) *Status {
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
func (rs *Services) GetStatus(nodeId int) Status {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists {
		return Inactive
	}
	if serviceStatus != Active {
		return serviceStatus
	}
	if rs.ServiceType == LndService {
		for _, status := range rs.streamStatus[nodeId] {
			if status != Active {
				return Pending
			}
		}
	}
	return Active
}

// GetStreamStatus when the status of the LND Service is active then streamStatus will be returned.
func (rs *Services) GetStreamStatus(nodeId int, stream SubscriptionStream) Status {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists {
		return Inactive
	}
	if serviceStatus != Active {
		return serviceStatus
	}
	streamStatus, exists := rs.streamStatus[nodeId][stream]
	if exists {
		return streamStatus
	}
	return Inactive
}

// GetStreamBootTime when the status of the LND Service is active and streamStatus is active then bootTime will have a value
func (rs *Services) GetStreamBootTime(nodeId int, stream SubscriptionStream) *time.Time {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists || serviceStatus != Active {
		return nil
	}
	streamStatus, exists := rs.streamStatus[nodeId][stream]
	if !exists || streamStatus != Active {
		return nil
	}
	bootTime, exists := rs.streamBootTime[nodeId][stream]
	if !exists {
		return nil
	}
	return &bootTime
}

func (rs *Services) GetCombinedStatus(nodeId int, streams ...SubscriptionStream) Status {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists {
		return Inactive
	}
	if serviceStatus != Active {
		return serviceStatus
	}
	for _, stream := range streams {
		if rs.streamStatus[nodeId][stream] != Active {
			return Pending
		}
	}
	return Active
}

// GetBootTime When the service is active it will return it's boot time
func (rs *Services) GetBootTime(nodeId int) *time.Time {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	serviceStatus, exists := rs.serviceStatus[nodeId]
	if !exists || serviceStatus != Active {
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

func (rs *Services) Booted(nodeId int, bootLock *sync.Mutex, serviceEventChannel chan ServiceEvent) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	bootLock.Unlock()
	initServiceMaps(rs, nodeId)
	previousStatus := rs.serviceStatus[nodeId]
	rs.bootTime[nodeId] = time.Now().UTC()
	rs.serviceStatus[nodeId] = Active

	if rs.ServiceType == LndService {
		setStreamStatuses(nodeId, rs, Pending)
	}
	sendServiceEvent(nodeId, serviceEventChannel, previousStatus, rs.serviceStatus[nodeId], rs.ServiceType, nil)
}

func (rs *Services) SetStreamStatus(nodeId int, stream SubscriptionStream, status Status, serviceEventChannel chan ServiceEvent) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs, nodeId)
	previousStatus := rs.streamStatus[nodeId][stream]
	if previousStatus == status {
		return
	}
	rs.streamStatus[nodeId][stream] = status
	if status == Active {
		rs.streamBootTime[nodeId][stream] = time.Now().UTC()
	}
	sendServiceEvent(nodeId, serviceEventChannel, previousStatus, rs.streamStatus[nodeId][stream], rs.ServiceType, &stream)
}

func initServiceMaps(rs *Services, nodeId int) {
	if rs.runningList == nil {
		rs.runningList = make(map[int]func())
		rs.serviceStatus = make(map[int]Status)
		rs.bootLock = make(map[int]*sync.Mutex)
		rs.bootTime = make(map[int]time.Time)
		rs.cancelTime = make(map[int]time.Time)
		rs.enforcedServiceStatus = make(map[int]*Status)
		rs.noDelay = make(map[int]bool)
		rs.streamStatus = make(map[int]map[SubscriptionStream]Status)
		rs.streamStatus[nodeId] = make(map[SubscriptionStream]Status, 0)
		rs.streamBootTime = make(map[int]map[SubscriptionStream]time.Time)
		rs.streamBootTime[nodeId] = make(map[SubscriptionStream]time.Time, 0)
	}
}

func sendServiceEvent(nodeId int, serviceEventChannel chan ServiceEvent, previousStatus Status, status Status,
	serviceType ServiceType, subscriptionStream *SubscriptionStream) {
	if previousStatus != status {
		if serviceEventChannel != nil {
			serviceEventChannel <- ServiceEvent{
				EventData: EventData{
					EventTime: time.Now().UTC(),
					NodeId:    nodeId,
				},
				Type:               serviceType,
				SubscriptionStream: subscriptionStream,
				Status:             status,
				PreviousStatus:     previousStatus,
			}
		}
	}
}

func setStreamStatuses(nodeId int, rs *Services, status Status) {
	nodeStreamStatus := make(map[SubscriptionStream]Status)
	nodeStreamStatus[TransactionStream] = status
	nodeStreamStatus[HtlcEventStream] = status
	nodeStreamStatus[ChannelEventStream] = status
	nodeStreamStatus[GraphEventStream] = status
	nodeStreamStatus[ForwardStream] = status
	nodeStreamStatus[InvoiceStream] = status
	nodeStreamStatus[PaymentStream] = status
	nodeStreamStatus[InFlightPaymentStream] = status
	nodeStreamStatus[PeerEventStream] = status
	rs.streamStatus[nodeId] = nodeStreamStatus
}
