package automation

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/rebalances"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

const rebalanceQueueTickerSeconds = 10
const rebalanceMaximumConcurrency = 100
const rebalanceRouteFailedHopAllowedDeltaPerMille = 10
const rebalanceRebalanceDelayMilliseconds = 2_000
const rebalanceTimeoutSeconds = 2 * 60 * 60
const rebalanceRunnerTimeoutSeconds = 1 * 60 * 60
const rebalanceRoutesTimeoutSeconds = 1 * 60
const rebalancePayTimeoutSeconds = 10 * 60
const rebalanceMinimumDeltaSeconds = 10 * 60
const rebalancePreviousResultTimeoutMinutes = 1 * 24 * 60
const rebalancePreviousSuccessResultTimeoutMinutes = 5

type Rebalancer struct {
	NodeId          int
	RebalanceId     int
	Status          commons.Status
	CreatedOn       time.Time
	ScheduleTarget  time.Time
	UpdateOn        time.Time
	GlobalCtx       context.Context
	RebalanceCtx    context.Context
	RebalanceCancel context.CancelFunc
	Runners         map[int]*RebalanceRunner
	Request         commons.RebalanceRequest
}

type RebalanceRunner struct {
	RebalanceId       int
	OutgoingChannelId int
	IncomingChannelId int
	Invoices          map[uint64]*lnrpc.AddInvoiceResponse
	// FailedHops map[hopSourcePublicKey_hopDestinationPublicKey]amountMsat
	FailedHops  map[string]uint64
	FailedPairs []*lnrpc.NodePair
	Status      commons.Status
	Ctx         context.Context
	Cancel      context.CancelFunc
}

func (runner *RebalanceRunner) addFailedHop(hopSourcePublicKey string, hopDestinationPublicKey string, amountMsat uint64) {
	runner.FailedHops[hopSourcePublicKey+"_"+hopDestinationPublicKey] = amountMsat
}

func (runner *RebalanceRunner) isFailedHop(hopSourcePublicKey string, hopDestinationPublicKey string, amountMsat uint64) bool {
	failedHopAmountMsat, exists := runner.FailedHops[hopSourcePublicKey+"_"+hopDestinationPublicKey]
	return exists &&
		commons.GetDeltaPerMille(failedHopAmountMsat, amountMsat) <
			rebalanceRouteFailedHopAllowedDeltaPerMille
}

func RebalanceServiceStart(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer) {

	defer log.Info().Msgf("RebalanceService terminated for nodeId: %v", nodeId)

	client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	go initiateDelayedRebalancers(ctx, db, client, router)

	wg := sync.WaitGroup{}
	listener := broadcaster.SubscribeRebalanceRequest()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range ctx.Done() {
			broadcaster.CancelSubscriptionRebalanceRequest(listener)
			return
		}
	}()
	go func() {
		for requests := range listener {
			if requests.NodeId != nodeId {
				continue
			}
			processRebalanceRequests(ctx, db, requests, nodeId)
		}
	}()
	wg.Wait()
}

func initiateDelayedRebalancers(ctx context.Context, db *sqlx.DB,
	client lnrpc.LightningClient, router routerrpc.RouterClient) {

	ticker := clock.New().Tick(rebalanceQueueTickerSeconds * time.Second)
	pending := commons.Pending
	active := commons.Active

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			activeRebalancers := getRebalancers(&active)
			log.Debug().Msgf("Active rebalancers: %v/%v", len(activeRebalancers), rebalanceMaximumConcurrency)
			if len(activeRebalancers) > rebalanceMaximumConcurrency {
				continue
			}

			pendingRebalancers := getRebalancers(&pending)
			if len(pendingRebalancers) > 0 {
				sort.Slice(pendingRebalancers, func(i, j int) bool {
					return pendingRebalancers[i].ScheduleTarget.Before(pendingRebalancers[j].ScheduleTarget)
				})

				var pendingRebalancer *Rebalancer
				i := 0
				for {
					pendingRebalancer = pendingRebalancers[i]
					if pendingRebalancer.RebalanceCtx.Err() == nil {
						break
					}
					removeRebalancer(pendingRebalancer)
					runningFor := time.Since(pendingRebalancer.CreatedOn).Round(1 * time.Second)
					if pendingRebalancer.Request.IncomingChannelId != 0 {
						log.Debug().Msgf("Rebalancer timed out after %s for Origin: %v, OriginId: %v, Incoming Channel: %v",
							runningFor, pendingRebalancer.Request.Origin, pendingRebalancer.Request.OriginId, pendingRebalancer.Request.IncomingChannelId)
					}
					if pendingRebalancer.Request.OutgoingChannelId != 0 {
						log.Debug().Msgf("Rebalancer timed out after %s for Origin: %v, OriginId: %v, Outgoing Channel: %v",
							runningFor, pendingRebalancer.Request.Origin, pendingRebalancer.Request.OriginId, pendingRebalancer.Request.OutgoingChannelId)
					}
					i++
				}

				if pendingRebalancer.ScheduleTarget.Before(time.Now()) {
					go pendingRebalancer.start(db, client, router,
						rebalanceRunnerTimeoutSeconds,
						rebalanceRoutesTimeoutSeconds,
						rebalancePayTimeoutSeconds)
				}
			}
		}
	}
}

func processRebalanceRequests(ctx context.Context, db *sqlx.DB, requests commons.RebalanceRequests, nodeId int) {
	var incoming bool
	var outgoing bool
	for _, request := range requests.Requests {
		if request.IncomingChannelId != 0 {
			incoming = true
		} else {
			outgoing = true
		}
	}
	if incoming && outgoing {
		err := errors.New(fmt.Sprintf(
			"Rebalance request's ignored because focus was both incoming and outgoing, which is impossible for nodeId: %v", nodeId))
		sendError(requests, err)
		return
	}
	responses := make(map[int]commons.RebalanceResponse)
	for _, request := range requests.Requests {
		response := validateRebalanceRequest(request)
		if response != nil {
			log.Debug().Msgf("Rebalance request ignored due to validation issues: %v", response)
			if incoming {
				responses[request.IncomingChannelId] = *response
			} else {
				responses[request.OutgoingChannelId] = *response
			}
		}
	}

	for _, request := range requests.Requests {
		if incoming {
			_, exists := responses[request.IncomingChannelId]
			if exists {
				continue
			}
		} else {
			_, exists := responses[request.OutgoingChannelId]
			if exists {
				continue
			}
		}
		response := updateExistingRebalanceRequest(db, request)
		if response != nil {
			if incoming {
				responses[request.IncomingChannelId] = *response
			} else {
				responses[request.OutgoingChannelId] = *response
			}
		}
	}

	var channelIdsToCheck []int
	for _, request := range requests.Requests {
		if incoming {
			_, exists := responses[request.IncomingChannelId]
			if exists {
				continue
			}
		} else {
			_, exists := responses[request.OutgoingChannelId]
			if exists {
				continue
			}
		}
		for _, channelId := range request.ChannelIds {
			if !slices.Contains(channelIdsToCheck, channelId) {
				channelIdsToCheck = append(channelIdsToCheck, channelId)
			}
		}
	}

	pending := commons.Pending
	pendingRebalancers := getRebalancers(&pending)
	timeout := time.Duration(-1 * rebalancePreviousResultTimeoutMinutes)

	var badChannelIds []int
	if incoming {
		sqlString := `
			SELECT incoming_channel_id
			FROM rebalance_log
			WHERE incoming_channel_id = ANY($1) AND created_on >= $2
			ORDER BY created_on DESC;`
		err := db.Select(&badChannelIds, sqlString, pq.Array(channelIdsToCheck), time.Now().Add(timeout*time.Minute))
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				sendError(requests, errors.Wrapf(err, "Getting rebalance results with incomingChannelIds: %v", channelIdsToCheck))
				return
			}
		}
	} else {
		sqlString := `
			SELECT outgoing_channel_id
			FROM rebalance_log
			WHERE outgoing_channel_id = ANY($1) AND created_on >= $2
			ORDER BY created_on DESC;`
		err := db.Select(&badChannelIds, sqlString, pq.Array(channelIdsToCheck), time.Now().Add(timeout*time.Minute))
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				sendError(requests, errors.Wrapf(err, "Getting rebalance results with outgoingChannelIds: %v", channelIdsToCheck))
				return
			}
		}
	}

	if incoming {
		for _, request := range requests.Requests {
			_, exists := responses[request.IncomingChannelId]
			if exists {
				continue
			}

			var channelIds []int
			for _, outgoingChannelId := range request.ChannelIds {
				if !slices.Contains(badChannelIds, outgoingChannelId) {
					channelIds = append(channelIds, outgoingChannelId)
				}
			}

			var finalChannelIds []int
			//check pending rebalances for opposite direction requests
			for _, outgoingChannelId := range channelIds {
				if !hasPendingRebalance(pendingRebalancers, outgoingChannelId, request.IncomingChannelId) {
					finalChannelIds = append(finalChannelIds, outgoingChannelId)
				}
			}

			if len(finalChannelIds) == 0 {
				responses[request.IncomingChannelId] = commons.RebalanceResponse{
					Request: request,
					CommunicationResponse: commons.CommunicationResponse{
						Status: commons.Inactive,
						Error:  "No channelIds found after filtering based on historic records",
					},
				}
				continue
			}

			createdOn := time.Now().UTC()
			rebalancer := &Rebalancer{
				NodeId:    nodeId,
				CreatedOn: createdOn,
				// Previous rebalance cleanup delay
				ScheduleTarget: createdOn.Add(rebalanceRebalanceDelayMilliseconds * time.Millisecond),
				UpdateOn:       createdOn,
				GlobalCtx:      ctx,
				Runners:        make(map[int]*RebalanceRunner),
				Request:        request,
				Status:         commons.Pending,
			}
			rebalancerCtx, rebalancerCancel := context.WithTimeout(rebalancer.GlobalCtx,
				time.Second*time.Duration(rebalanceTimeoutSeconds))
			rebalancer.RebalanceCtx = rebalancerCtx
			rebalancer.RebalanceCancel = rebalancerCancel
			if !addRebalancer(rebalancer) {
				rebalanceResponse := commons.RebalanceResponse{
					Request: request,
					CommunicationResponse: commons.CommunicationResponse{
						Status: commons.Active,
					},
				}
				rebalanceResponse.Message = fmt.Sprintf(
					"IncomingChannelId: %v already has a running rebalancer for origin: %v, originId: %v with reference number: %v",
					rebalancer.Request.IncomingChannelId, rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OriginReference)
				responses[request.IncomingChannelId] = rebalanceResponse
				continue
			}
			responses[request.IncomingChannelId] = commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Active,
				},
			}
		}
	} else {
		for _, request := range requests.Requests {
			_, exists := responses[request.OutgoingChannelId]
			if exists {
				continue
			}

			var channelIds []int
			for _, incomingChannelId := range request.ChannelIds {
				if !slices.Contains(badChannelIds, incomingChannelId) {
					channelIds = append(channelIds, incomingChannelId)
				}
			}

			var finalChannelIds []int
			//check pending rebalances for opposite direction requests
			for _, incomingChannelId := range channelIds {
				if !hasPendingRebalance(pendingRebalancers, request.OutgoingChannelId, incomingChannelId) {
					finalChannelIds = append(finalChannelIds, incomingChannelId)
				}
			}

			if len(finalChannelIds) == 0 {
				responses[request.OutgoingChannelId] = commons.RebalanceResponse{
					Request: request,
					CommunicationResponse: commons.CommunicationResponse{
						Status: commons.Inactive,
						Error:  "No channelIds found after filtering based on historic records",
					},
				}
				continue
			}

			createdOn := time.Now().UTC()
			rebalancer := &Rebalancer{
				NodeId:    nodeId,
				CreatedOn: createdOn,
				// Previous rebalance cleanup delay
				ScheduleTarget: createdOn.Add(rebalanceRebalanceDelayMilliseconds * time.Millisecond),
				UpdateOn:       createdOn,
				GlobalCtx:      ctx,
				Runners:        make(map[int]*RebalanceRunner),
				Request:        request,
				Status:         commons.Pending,
			}
			rebalancerCtx, rebalancerCancel := context.WithTimeout(rebalancer.GlobalCtx,
				time.Second*time.Duration(rebalanceTimeoutSeconds))
			rebalancer.RebalanceCtx = rebalancerCtx
			rebalancer.RebalanceCancel = rebalancerCancel
			if !addRebalancer(rebalancer) {
				rebalanceResponse := commons.RebalanceResponse{
					Request: request,
					CommunicationResponse: commons.CommunicationResponse{
						Status: commons.Active,
					},
				}
				rebalanceResponse.Message = fmt.Sprintf(
					"OutgoingChannelId: %v already has a running rebalancer for origin: %v, originId: %v with reference number: %v",
					rebalancer.Request.OutgoingChannelId, rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OriginReference)
				responses[request.OutgoingChannelId] = rebalanceResponse
				continue
			}
			responses[request.OutgoingChannelId] = commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Active,
				},
			}
		}
	}
	if requests.ResponseChannel != nil {
		var res []commons.RebalanceResponse
		for _, resp := range responses {
			res = append(res, resp)
		}
		requests.ResponseChannel <- res
	}
}

func sendError(requests commons.RebalanceRequests, err error) {
	log.Error().Err(err).Msg("processRebalanceRequests failed")
	if requests.ResponseChannel != nil {
		requests.ResponseChannel <- []commons.RebalanceResponse{{
			Request:               commons.RebalanceRequest{},
			CommunicationResponse: commons.CommunicationResponse{},
		}}
	}
}

func hasPendingRebalance(pendingRebalancers []*Rebalancer, incomingChannelId int, outgoingChannelId int) bool {
	for _, rebalancer := range pendingRebalancers {
		if rebalancer.Request.IncomingChannelId == incomingChannelId {
			for _, rebalanceOutgoingChannelId := range rebalancer.Request.ChannelIds {
				if rebalanceOutgoingChannelId == outgoingChannelId {
					return true
				}
			}
		}
		if rebalancer.Request.OutgoingChannelId == outgoingChannelId {
			for _, rebalanceIncomingChannelId := range rebalancer.Request.ChannelIds {
				if rebalanceIncomingChannelId == incomingChannelId {
					return true
				}
			}
		}
	}
	return false
}

func (rebalancer *Rebalancer) start(
	db *sqlx.DB,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	runnerTimeout int,
	routesTimeout int,
	payTimeout int) {

	log.Debug().Msgf("Rebalance initiated for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
		rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
	if rebalancer.Request.IncomingChannelId != 0 {
		incomingChannel := commons.GetChannelSettingByChannelId(rebalancer.Request.IncomingChannelId)
		if incomingChannel.Capacity == 0 || incomingChannel.Status != commons.Open {
			log.Error().Msgf("IncomingChannelId is invalid for origin: %v, originReference: %v and incomingChannelId: %v",
				rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId)
			removeRebalancer(rebalancer)
			rebalancer.RebalanceCancel()
			return
		}
	}
	if rebalancer.Request.OutgoingChannelId != 0 {
		outgoingChannel := commons.GetChannelSettingByChannelId(rebalancer.Request.OutgoingChannelId)
		if outgoingChannel.Capacity == 0 || outgoingChannel.Status != commons.Open {
			log.Error().Msgf("OutgoingChannelId is invalid for origin: %v, originReference: %v and outgoingChannelId: %v",
				rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.OutgoingChannelId)
			removeRebalancer(rebalancer)
			rebalancer.RebalanceCancel()
			return
		}
	}

	rebalancer.Status = commons.Active
	latestResult, err := rebalances.GetLatestResultByOrigin(db, rebalancer.Request.Origin, rebalancer.Request.OriginId,
		rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId, commons.Active,
		rebalancePreviousSuccessResultTimeoutMinutes)
	if err != nil {
		log.Error().Err(err).Msgf("Obtaining latest result for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
			rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
	}
	previousSuccess := rebalancer.convertPreviousSuccess(latestResult)

	err = AddRebalanceAndChannels(db, rebalancer)
	if err != nil {
		log.Error().Err(err).Msgf("Storing rebalance for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
			rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
		return
	}

	if previousSuccess.Hops != "" {
		log.Debug().Msgf("Previous success found for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
			rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
		runnerCtx, runnerCancel := context.WithTimeout(rebalancer.RebalanceCtx, time.Second*time.Duration(runnerTimeout))
		defer runnerCancel()
		previousSuccessRunner := &RebalanceRunner{
			RebalanceId:       rebalancer.RebalanceId,
			OutgoingChannelId: previousSuccess.OutgoingChannelId,
			IncomingChannelId: previousSuccess.IncomingChannelId,
			Invoices:          make(map[uint64]*lnrpc.AddInvoiceResponse),
			FailedHops:        make(map[string]uint64),
			Status:            commons.Active,
			Ctx:               runnerCtx,
			Cancel:            runnerCancel,
		}
		if rebalancer.Request.IncomingChannelId != 0 {
			// When incoming channel is provided then the runners loop over the outgoing channels
			rebalancer.Runners[previousSuccess.OutgoingChannelId] = previousSuccessRunner
		} else {
			rebalancer.Runners[previousSuccess.IncomingChannelId] = previousSuccessRunner
		}
		result := rebalances.RebalanceResult{
			Status:            commons.Initializing,
			RebalanceId:       rebalancer.RebalanceId,
			CreatedOn:         time.Now().UTC(),
			IncomingChannelId: previousSuccessRunner.IncomingChannelId,
			OutgoingChannelId: previousSuccessRunner.OutgoingChannelId,
		}
		result = rebalancer.startRunner(db, client, router, previousSuccessRunner, routesTimeout, payTimeout, result)
		if result.Status == commons.Active {
			log.Debug().Msgf("Previous success successfully reused for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
				rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
			removeRebalancer(rebalancer)
			rebalancer.RebalanceCancel()
		}
		rebalancer.processResult(db, result)
		if result.Status == commons.Active {
			return
		}
		log.Debug().Msgf("Previous success reuse failed for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
			rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
	}
	for i := 0; i < rebalancer.Request.MaximumConcurrency; i++ {
		go rebalancer.createRunner(db, client, router, runnerTimeout, routesTimeout, payTimeout)
	}
}

func (rebalancer *Rebalancer) convertPreviousSuccess(previousSuccess rebalances.RebalanceResult) rebalances.RebalanceResult {
	if rebalancer.Request.OutgoingChannelId != 0 && !slices.Contains(rebalancer.Request.ChannelIds, previousSuccess.IncomingChannelId) {
		log.Debug().Msgf("Previous success ignored as it's not available anymore for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
			rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
		return rebalances.RebalanceResult{}
	}
	if rebalancer.Request.IncomingChannelId != 0 && !slices.Contains(rebalancer.Request.ChannelIds, previousSuccess.OutgoingChannelId) {
		log.Debug().Msgf("Previous success ignored as it's not available anymore for origin: %v, originReference: %v, incomingChannelId: %v, outgoingChannelId: %v",
			rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId)
		return rebalances.RebalanceResult{}
	}
	return previousSuccess
}

func (rebalancer *Rebalancer) createRunner(
	db *sqlx.DB,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	runnerTimeout int,
	routesTimeout int,
	payTimeout int) {

	if rebalancer.Status == commons.Inactive {
		return
	}

	result := rebalances.RebalanceResult{
		Status:            commons.Initializing,
		RebalanceId:       rebalancer.RebalanceId,
		CreatedOn:         time.Now().UTC(),
		IncomingChannelId: rebalancer.Request.IncomingChannelId,
		OutgoingChannelId: rebalancer.Request.OutgoingChannelId,
	}
	channelId := rebalancer.getPendingChannelId()
	if channelId == 0 {
		for _, runner := range rebalancer.Runners {
			if runner.Status == commons.Active {
				return
			}
		}
		removeRebalancer(rebalancer)
		runningFor := time.Since(rebalancer.ScheduleTarget).Round(1 * time.Second)
		if rebalancer.Request.IncomingChannelId != 0 {
			log.Debug().Msgf("Pending Outgoing ChannelIds got exhausted for Origin: %v, OriginId: %v, IncomingChannelId: %v (%s)",
				rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.IncomingChannelId, runningFor)
		}
		if rebalancer.Request.OutgoingChannelId != 0 {
			log.Debug().Msgf("Pending Incoming ChannelIds got exhausted for Origin: %v, OriginId: %v, OutgoingChannelId: %v (%s)",
				rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OutgoingChannelId, runningFor)
		}
		rebalancer.ScheduleTarget = time.Now().UTC()
		if runningFor.Seconds() < rebalanceMinimumDeltaSeconds {
			rebalancer.ScheduleTarget = rebalancer.ScheduleTarget.Add(rebalanceMinimumDeltaSeconds*time.Second - runningFor)
		}
		rebalancer.Runners = make(map[int]*RebalanceRunner)
		rebalancer.Status = commons.Pending
		if !addRebalancer(rebalancer) {
			if rebalancer.Request.IncomingChannelId != 0 {
				log.Error().Msgf("Failed to reschedule the incoming rebalancer for Origin: %v, OriginId: %v (%v)",
					rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.IncomingChannelId)
			}
			if rebalancer.Request.OutgoingChannelId != 0 {
				log.Error().Msgf("Failed to reschedule the outgoing rebalancer for Origin: %v, OriginId: %v (%v)",
					rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OutgoingChannelId)
			}
		}
		return
	}

	runnerCtx, runnerCancel := context.WithTimeout(rebalancer.RebalanceCtx, time.Second*time.Duration(runnerTimeout))
	defer runnerCancel()
	runner := rebalancer.addRunner(channelId, runnerCtx, runnerCancel)

	result.IncomingChannelId = runner.IncomingChannelId
	result.OutgoingChannelId = runner.OutgoingChannelId

	result = rebalancer.startRunner(db, client, router, runner, routesTimeout, payTimeout, result)
	if result.Status == commons.Active {
		removeRebalancer(rebalancer)
		runningFor := time.Since(rebalancer.ScheduleTarget).Round(1 * time.Second)
		msg := fmt.Sprintf("Successfully rebalanced after %s %vmsats @ %vmsats (%v ppm) using incomingChannelId: %v, outgoingChannelId: %v",
			runningFor, result.TotalAmountMsat, result.TotalFeeMsat,
			((result.TotalFeeMsat*1_000_000)/result.TotalAmountMsat)+1, // + 1 for rounding error
			result.IncomingChannelId, result.OutgoingChannelId)
		log.Debug().Msgf("%v for Origin: %v, OriginId: %v (Hops: %v)", msg, rebalancer.Request.Origin, rebalancer.Request.OriginId, result.Hops)
		rebalancer.Status = commons.Inactive
		return
	}
	runner.Cancel()
	runner.Status = commons.Inactive

	rebalancer.createRunner(db, client, router, runnerTimeout, routesTimeout, payTimeout)
}

func (rebalancer *Rebalancer) startRunner(
	db *sqlx.DB,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	runner *RebalanceRunner,
	routesTimeout int,
	payTimeout int,
	result rebalances.RebalanceResult) rebalances.RebalanceResult {

	routesCtx, routesCancel := context.WithTimeout(runner.Ctx, time.Second*time.Duration(routesTimeout))
	defer routesCancel()
	routes, err := runner.getRoutes(routesCtx, client, router, rebalancer.NodeId,
		rebalancer.Request.AmountMsat, rebalancer.Request.MaximumCostMsat)
	if err != nil {
		result.Status = commons.Inactive
		result.Error = err.Error()
		if routesCtx.Err() == context.DeadlineExceeded {
			result.Error = routesCtx.Err().Error()
		}
		rebalancer.processResult(db, result)
	}
	routesCancel()

	for _, route := range routes {
		payCtx, payCancel := context.WithTimeout(runner.Ctx, time.Second*time.Duration(payTimeout))
		result = runner.pay(payCtx, client, router, rebalancer.Request.AmountMsat, route)
		payCancel()
		if payCtx.Err() == context.DeadlineExceeded {
			result.Error = payCtx.Err().Error()
		}
		rebalancer.processResult(db, result)
		if result.Status == commons.Active {
			return result
		}
	}

	if result.Status == commons.Pending {
		result = rebalancer.startRunner(db, client, router, runner, routesTimeout, payTimeout, result)
	}
	return result
}

// TODO FIXME make channel selection smarter instead of at random...
func (rebalancer *Rebalancer) getPendingChannelId() int {
outer:
	for _, channelId := range rebalancer.Request.ChannelIds {
		for existingChannelId := range rebalancer.Runners {
			if existingChannelId == channelId {
				continue outer
			}
		}
		channelSettings := commons.GetChannelSettingByChannelId(channelId)
		if channelSettings.Capacity == 0 || channelSettings.Status != commons.Open {
			continue outer
		}
		return channelId
	}
	return 0
}

func (rebalancer *Rebalancer) addRunner(channelId int, runnerCtx context.Context, runnerCancel context.CancelFunc) *RebalanceRunner {
	runner := RebalanceRunner{
		RebalanceId: rebalancer.RebalanceId,
		Invoices:    make(map[uint64]*lnrpc.AddInvoiceResponse),
		FailedHops:  make(map[string]uint64),
		Ctx:         runnerCtx,
		Cancel:      runnerCancel,
		Status:      commons.Active,
	}

	if rebalancer.Request.IncomingChannelId != 0 {
		runner.IncomingChannelId = rebalancer.Request.IncomingChannelId
		runner.OutgoingChannelId = channelId
	} else {
		runner.IncomingChannelId = channelId
		runner.OutgoingChannelId = rebalancer.Request.OutgoingChannelId
	}

	rebalancer.Runners[channelId] = &runner
	return &runner
}

func (rebalancer *Rebalancer) processResult(db *sqlx.DB, result rebalances.RebalanceResult) {
	result.UpdateOn = time.Now().UTC()
	err := rebalances.AddRebalanceResult(db, result)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to add rebalance log entry for rebalanceId: %v (ref: %v)",
			rebalancer.RebalanceId, rebalancer.Request.OriginReference)
	}
}

func (runner *RebalanceRunner) getRoutes(
	ctx context.Context,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	nodeId int,
	amountMsat uint64,
	fixedFeeMsat uint64) ([]*lnrpc.Route, error) {

	outgoingChannel := commons.GetChannelSettingByChannelId(runner.OutgoingChannelId)
	incomingChannel := commons.GetChannelSettingByChannelId(runner.IncomingChannelId)
	var remoteNode commons.ManagedNodeSettings
	if outgoingChannel.FirstNodeId == nodeId {
		remoteNode = commons.GetNodeSettingsByNodeId(incomingChannel.SecondNodeId)
	} else {
		remoteNode = commons.GetNodeSettingsByNodeId(incomingChannel.FirstNodeId)
	}
	remoteNodePublicKey, err := hex.DecodeString(remoteNode.PublicKey)
	if err != nil {
		return nil, errors.Wrapf(err, "Decoding public key for outgoing nodeId: %v", outgoingChannel.SecondNodeId)
	}

	routes, err := client.QueryRoutes(ctx, &lnrpc.QueryRoutesRequest{
		PubKey:            commons.GetNodeSettingsByNodeId(nodeId).PublicKey,
		OutgoingChanId:    outgoingChannel.LndShortChannelId,
		LastHopPubkey:     remoteNodePublicKey,
		AmtMsat:           int64(amountMsat),
		UseMissionControl: true,
		FeeLimit:          &lnrpc.FeeLimit{Limit: &lnrpc.FeeLimit_FixedMsat{FixedMsat: int64(fixedFeeMsat)}},
		IgnoredPairs:      runner.FailedPairs,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "QueryRoutes for outgoing nodeId: %v, publicKey: %v", outgoingChannel.FirstNodeId, remoteNode.PublicKey)
	}

	var result []*lnrpc.Route
	for i := range routes.Routes {
		if runner.validateRoute(nodeId, routes.Routes[i]) {
			result = append(result, routes.Routes[i])
		}
	}
	if len(result) == 0 {
		return runner.getRoutes(ctx, client, router, nodeId, amountMsat, fixedFeeMsat)
	}
	return result, nil
}

func (runner *RebalanceRunner) pay(
	ctx context.Context,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	amountMsat uint64,
	route *lnrpc.Route) rebalances.RebalanceResult {

	rebalanceResult := rebalances.RebalanceResult{
		OutgoingChannelId: runner.OutgoingChannelId,
		IncomingChannelId: runner.IncomingChannelId,
		RebalanceId:       runner.RebalanceId,
		CreatedOn:         time.Now().UTC(),
		Status:            commons.Inactive,
	}

	invoice, err := runner.createInvoice(ctx, client, amountMsat)
	if err != nil {
		rebalanceResult.Error = err.Error()
		return rebalanceResult
	}
	lastHop := route.Hops[len(route.Hops)-1]
	lastHop.MppRecord = &lnrpc.MPPRecord{
		PaymentAddr:  invoice.PaymentAddr,
		TotalAmtMsat: int64(amountMsat),
	}

	result, err := router.SendToRouteV2(ctx,
		&routerrpc.SendToRouteRequest{
			PaymentHash: invoice.RHash,
			Route:       route,
		})
	if result != nil && result.Route != nil {
		rebalanceResult.TotalFeeMsat = uint64(result.Route.TotalFeesMsat)
		rebalanceResult.TotalTimeLock = result.Route.TotalTimeLock
		rebalanceResult.TotalAmountMsat = uint64(result.Route.TotalAmtMsat)
	}
	if err != nil {
		rebalanceResult.Error = err.Error()
		return rebalanceResult
	}
	if result.Status == lnrpc.HTLCAttempt_FAILED {
		rebalanceResult.Status = commons.Inactive
		if result.Failure.FailureSourceIndex >= uint32(len(route.Hops)) {
			rebalanceResult.Error = fmt.Sprintf("%s unknown hop index: %d. Maximum hop index: %d",
				result.Failure.Code.String(), result.Failure.FailureSourceIndex, len(route.Hops))
			return rebalanceResult
		}
		if result.Failure.FailureSourceIndex == 0 {
			rebalanceResult.Error = fmt.Sprintf("%s unknown hop index %d. Minimum hop index is greater than 0",
				result.Failure.Code.String(), result.Failure.FailureSourceIndex)
			return rebalanceResult
		}
		prevHop := route.Hops[result.Failure.FailureSourceIndex-1]
		failedHop := route.Hops[result.Failure.FailureSourceIndex]
		if result.Failure.Code == lnrpc.Failure_TEMPORARY_CHANNEL_FAILURE {
			rebalanceResult.Status = commons.Pending
			runner.addFailedHop(prevHop.PubKey, failedHop.PubKey, uint64(prevHop.AmtToForwardMsat))
		}
		rebalanceResult.Error = fmt.Sprintf("error: %s occured at hop index %d (%v -> %v)",
			result.Failure.Code.String(), result.Failure.FailureSourceIndex, prevHop.PubKey, failedHop.PubKey)
		return rebalanceResult
	}
	delete(runner.Invoices, amountMsat)
	rebalanceResult.Status = commons.Active
	if result != nil && result.Route != nil {
		hopsJsonByteArray, err := json.Marshal(result.Route.Hops)
		if err != nil {
			log.Error().Err(err).Msgf("Marshalling the route hops for rebalancerId: %v", runner.RebalanceId)
			return rebalanceResult
		}
		rebalanceResult.Hops = string(hopsJsonByteArray)
	}
	return rebalanceResult
}

func (runner *RebalanceRunner) validateRoute(nodeId int, route *lnrpc.Route) bool {
	previousHopPublicKey := commons.GetNodeSettingsByNodeId(nodeId).PublicKey
	for _, h := range route.Hops {
		if runner.isFailedHop(previousHopPublicKey, h.PubKey, uint64(h.AmtToForwardMsat)) {
			from, err := hex.DecodeString(previousHopPublicKey)
			if err != nil {
				return false
			}
			to, err := hex.DecodeString(h.PubKey)
			if err != nil {
				return false
			}
			runner.FailedPairs = append(runner.FailedPairs, &lnrpc.NodePair{From: from, To: to})
			return false
		}
		previousHopPublicKey = h.PubKey
	}
	return true
}

func (runner *RebalanceRunner) createInvoice(
	ctx context.Context,
	client lnrpc.LightningClient,
	amountMsat uint64) (*lnrpc.AddInvoiceResponse, error) {

	invoice, exists := runner.Invoices[amountMsat]
	if exists {
		return invoice, nil
	}
	invoice, err := client.AddInvoice(ctx, &lnrpc.Invoice{ValueMsat: int64(amountMsat),
		Memo:   "Rebalance attempt",
		Expiry: int64(rebalanceTimeoutSeconds)})
	if err != nil {
		return nil, errors.Wrapf(err, "AddInvoice for %v msat", amountMsat)
	}
	runner.Invoices[amountMsat] = invoice
	return invoice, nil
}

func validateRebalanceRequest(request commons.RebalanceRequest) *commons.RebalanceResponse {
	if request.IncomingChannelId == 0 && request.OutgoingChannelId == 0 {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "IncomingChannelId and OutgoingChannelId are 0",
			},
		}
	}
	if request.IncomingChannelId != 0 && request.OutgoingChannelId != 0 {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "IncomingChannelId and OutgoingChannelId are populated",
			},
		}
	}

	if request.IncomingChannelId != 0 {
		incomingChannel := commons.GetChannelSettingByChannelId(request.IncomingChannelId)
		if incomingChannel.Capacity == 0 || incomingChannel.Status != commons.Open {
			return &commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  "IncomingChannelId is invalid",
				},
			}
		}
		if slices.Contains(request.ChannelIds, request.IncomingChannelId) {
			return &commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  fmt.Sprintf("ChannelIds also contain IncomingChannelId: %v (%v)", request.IncomingChannelId, request.ChannelIds),
				},
			}
		}
	}
	if request.OutgoingChannelId != 0 {
		outgoingChannel := commons.GetChannelSettingByChannelId(request.OutgoingChannelId)
		if outgoingChannel.Capacity == 0 || outgoingChannel.Status != commons.Open {
			return &commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  "OutgoingChannelId is invalid",
				},
			}
		}
		if slices.Contains(request.ChannelIds, request.OutgoingChannelId) {
			return &commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  fmt.Sprintf("ChannelIds also contain OutgoingChannelId: %v (%v)", request.OutgoingChannelId, request.ChannelIds),
				},
			}
		}
	}

	response := verifyNotZeroInt(request, int64(request.MaximumConcurrency), "MaximumConcurrency")
	if response != nil {
		return response
	}
	response = verifyNotZeroInt(request, int64(request.OriginId), "OriginId")
	if response != nil {
		return response
	}
	response = verifyNotZeroUint(request, request.AmountMsat, "AmountMsat")
	if response != nil {
		return response
	}
	response = verifyNotZeroUint(request, request.MaximumCostMsat, "MaximumCostMsat")
	if response != nil {
		return response
	}
	if len(request.ChannelIds) == 0 {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "ChannelIds are not specified",
			},
		}
	}

	for _, channelId := range request.ChannelIds {
		channelSettings := commons.GetChannelSettingByChannelId(channelId)
		if channelSettings.Capacity == 0 || channelSettings.Status != commons.Open {
			return &commons.RebalanceResponse{
				Request: request,
				CommunicationResponse: commons.CommunicationResponse{
					Status: commons.Inactive,
					Error:  "ChannelIds contain an invalid channelId",
				},
			}
		}
	}
	return nil
}

func verifyNotZeroUint(request commons.RebalanceRequest, value uint64, label string) *commons.RebalanceResponse {
	if value == 0 {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  label + " is 0",
			},
		}
	}
	return nil
}

func verifyNotZeroInt(request commons.RebalanceRequest, value int64, label string) *commons.RebalanceResponse {
	if value == 0 {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  label + " is 0",
			},
		}
	}
	return nil
}

func updateExistingRebalanceRequest(db *sqlx.DB, request commons.RebalanceRequest) *commons.RebalanceResponse {
	rebalancer := getRebalancer(request.Origin, request.OriginId, request.IncomingChannelId, request.OutgoingChannelId)
	if rebalancer == nil {
		return nil
	}
	if rebalancer.Request.IncomingChannelId != request.IncomingChannelId || rebalancer.Request.OutgoingChannelId != request.OutgoingChannelId {
		removeRebalancer(rebalancer)
		rebalancer.RebalanceCancel()
		return nil
	}
	var err error
	if rebalancer.Request.AmountMsat != request.AmountMsat {
		err = setRebalancer(db, request, rebalancer)
	} else if rebalancer.Request.MaximumCostMsat != request.MaximumCostMsat {
		err = setRebalancer(db, request, rebalancer)
	} else if rebalancer.Request.MaximumConcurrency != request.MaximumConcurrency {
		err = setRebalancer(db, request, rebalancer)
	} else if len(rebalancer.Request.ChannelIds) != len(request.ChannelIds) {
		err = setRebalancer(db, request, rebalancer)
	} else {
		for _, channelId := range rebalancer.Request.ChannelIds {
			if !slices.Contains(request.ChannelIds, channelId) {
				err = setRebalancer(db, request, rebalancer)
				break
			}
		}
	}
	if err != nil {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error: fmt.Sprintf(
					"(%v) for IncomingChannelId: %v, OutgoingChannelId: %v that already has a running rebalancer for origin: %v with originId: %v (ref: %v)",
					err.Error(), rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId,
					rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OriginReference),
			},
		}
	}
	rebalanceResponse := &commons.RebalanceResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status: commons.Active,
		},
	}
	if rebalancer.Request.IncomingChannelId != 0 {
		rebalanceResponse.Message = fmt.Sprintf(
			"Updated existing rebalancer for origin: %v with originId: %v and IncomingChannelId: %v (ref: %v)",
			rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.IncomingChannelId, rebalancer.Request.OriginReference)
	}
	if rebalancer.Request.OutgoingChannelId != 0 {
		rebalanceResponse.Message = fmt.Sprintf(
			"Updated existing rebalancer for origin: %v with originId: %v and OutgoingChannelId: %v (ref: %v)",
			rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OutgoingChannelId, rebalancer.Request.OriginReference)
	}
	return rebalanceResponse
}

func setRebalancer(db *sqlx.DB, request commons.RebalanceRequest, rebalancer *Rebalancer) error {
	rebalancer.UpdateOn = time.Now().UTC()
	rebalancer.Request = request
	if rebalancer.RebalanceId != 0 {
		// If RebalanceId == 0 then it was not stored yet.
		err := rebalances.SetRebalanceAndChannels(db, rebalancer.Request.OriginReference, rebalancer.Request.AmountMsat,
			rebalancer.Request.MaximumConcurrency, rebalancer.Request.MaximumCostMsat, rebalancer.UpdateOn,
			rebalancer.RebalanceId, rebalancer.Request.ChannelIds)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to add rebalance log entry for rebalanceId: %v", rebalancer.RebalanceId)
			return errors.Wrapf(err,
				"Updating the database with the new rebalance settings for origin: %v with originId: %v (ref: %v)",
				rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OriginReference)
		}
	}
	return nil
}

func AddRebalanceAndChannels(db *sqlx.DB, rebalancer *Rebalancer) error {
	var incomingChannelId *int
	if rebalancer.Request.IncomingChannelId != 0 {
		incomingChannelId = &rebalancer.Request.IncomingChannelId
	}
	var outgoingChannelId *int
	if rebalancer.Request.OutgoingChannelId != 0 {
		outgoingChannelId = &rebalancer.Request.OutgoingChannelId
	}
	dbRebalancer := rebalances.Rebalance{
		OutgoingChannelId:  outgoingChannelId,
		IncomingChannelId:  incomingChannelId,
		Status:             rebalancer.Status,
		Origin:             rebalancer.Request.Origin,
		OriginId:           rebalancer.Request.OriginId,
		OriginReference:    rebalancer.Request.OriginReference,
		AmountMsat:         rebalancer.Request.AmountMsat,
		MaximumConcurrency: rebalancer.Request.MaximumConcurrency,
		MaximumCostMsat:    rebalancer.Request.MaximumCostMsat,
		ScheduleTarget:     rebalancer.ScheduleTarget,
		CreatedOn:          rebalancer.CreatedOn,
		UpdateOn:           rebalancer.UpdateOn,
	}
	var err error
	rebalancer.RebalanceId, err = rebalances.AddRebalanceAndChannels(db, dbRebalancer, rebalancer.Request.ChannelIds)
	if err != nil {
		return errors.Wrap(err, "Storing rebalance information")
	}
	return nil
}
