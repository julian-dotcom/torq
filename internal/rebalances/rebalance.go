package rebalances

import (
	"time"

	"github.com/lncapital/torq/pkg/core"
)

type Rebalance struct {
	RebalanceId        int                         `json:"rebalanceId" db:"rebalance_id"`
	OutgoingChannelId  *int                        `json:"outgoingChannelId" db:"outgoing_channel_id"`
	IncomingChannelId  *int                        `json:"incomingChannelId" db:"incoming_channel_id"`
	Status             core.Status                 `json:"status" db:"status"`
	Origin             core.RebalanceRequestOrigin `json:"origin" db:"origin"`
	OriginId           int                         `json:"originId" db:"origin_id"`
	OriginReference    string                      `json:"originReference" db:"origin_reference"`
	AmountMsat         uint64                      `json:"amountMsat" db:"amount_msat"`
	MaximumConcurrency int                         `json:"maximumConcurrency" db:"maximum_concurrency"`
	MaximumCostMsat    uint64                      `json:"maximumCostMsat" db:"maximum_costmsat"`
	ScheduleTarget     time.Time                   `json:"scheduleTarget" db:"schedule_target"`
	CreatedOn          time.Time                   `json:"createdOn" db:"created_on"`
	UpdateOn           time.Time                   `json:"updatedOn" db:"updated_on"`
}

type RebalanceChannel struct {
	RebalanceChannelId int         `json:"rebalanceChannelId" db:"rebalance_channel_id"`
	ChannelId          int         `json:"channelId" db:"channel_id"`
	Status             core.Status `json:"status" db:"status"`
	RebalanceId        int         `json:"rebalanceId" db:"rebalance_id"`
	CreatedOn          time.Time   `json:"createdOn" db:"created_on"`
	UpdateOn           time.Time   `json:"updatedOn" db:"updated_on"`
}

type RebalanceResult struct {
	RebalanceLogId    int         `json:"rebalanceLogId" db:"rebalance_log_id"`
	IncomingChannelId int         `json:"incomingChannelId" db:"incoming_channel_id"`
	OutgoingChannelId int         `json:"outgoingChannelId" db:"outgoing_channel_id"`
	Hops              string      `json:"hops" db:"hops"`
	Status            core.Status `json:"status" db:"status"`
	TotalTimeLock     uint32      `json:"totalTimeLock" db:"total_time_lock"`
	TotalFeeMsat      uint64      `json:"totalFeeMsat" db:"total_fee_msat"`
	TotalAmountMsat   uint64      `json:"totalAmountMsat" db:"total_amount_msat"`
	Error             string      `json:"error" db:"error"`
	RebalanceId       int         `json:"rebalanceId" db:"rebalance_id"`
	CreatedOn         time.Time   `json:"createdOn" db:"created_on"`
	UpdateOn          time.Time   `json:"updatedOn" db:"updated_on"`
}
