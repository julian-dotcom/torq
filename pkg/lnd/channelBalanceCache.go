package lnd

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

func ChannelBalanceCacheMaintenance(ctx context.Context, nodeSettings commons.ManagedNodeSettings,
	broadcaster broadcast.BroadcastServer, eventChannel chan interface{}, serviceEventChannel chan commons.ServiceEvent) {

	var err error
	serviceStatus := commons.Inactive
	bootStrapping := true

	// Create the default ticker used to fetch forwards at a set interval
	c := clock.New()
	ticker := c.Tick(60 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
		case event := <-broadcaster.Subscribe():
		}
	}
}
