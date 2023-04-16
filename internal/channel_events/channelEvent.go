package channel_events

import (
	"time"

	"github.com/lncapital/torq/proto/lnrpc"
)

type ChannelEvent struct {
	EventTime time.Time                           `json:"eventTime" db:"event_time"`
	EventType lnrpc.ChannelEventUpdate_UpdateType `json:"eventType" db:"event_type"`
	ChannelId int                                 `json:"channelId" db:"channel_id"`
	Imported  bool                                `json:"imported" db:"imported"`
	Event     string                              `json:"event" db:"event"`
	NodeId    int                                 `json:"nodeId" db:"node_id"`
}
