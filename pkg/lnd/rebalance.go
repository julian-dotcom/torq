package lnd

import (
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/rebalances"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type Rebalancer struct {
	NodeId          int
	RebalanceId     int
	Status          commons.Status
	CreatedOn       time.Time
	UpdateOn        time.Time
	GlobalCtx       context.Context
	RebalanceCtx    context.Context
	RebalanceCancel context.CancelFunc
	Runners         map[int]*RebalanceRunner
	Request         commons.RebalanceRequest
}

func (rebalancer *Rebalancer) getRebalance() rebalances.Rebalance {
	rebalance := rebalances.Rebalance{
		Origin:             rebalancer.Request.Origin,
		OriginId:           rebalancer.Request.OriginId,
		OriginReference:    rebalancer.Request.OriginReference,
		Status:             rebalancer.Status,
		AmountMsat:         rebalancer.Request.AmountMsat,
		MaximumCostMsat:    rebalancer.Request.MaximumCostMsat,
		MaximumConcurrency: rebalancer.Request.MaximumConcurrency,
		CreatedOn:          rebalancer.CreatedOn,
		UpdateOn:           rebalancer.UpdateOn,
	}
	if rebalancer.Request.IncomingChannelId != 0 {
		rebalance.IncomingChannelId = &rebalancer.Request.IncomingChannelId
	}
	if rebalancer.Request.OutgoingChannelId != 0 {
		rebalance.OutgoingChannelId = &rebalancer.Request.OutgoingChannelId
	}
	return rebalance
}

func (rebalancer *Rebalancer) getRebalanceChannels() []rebalances.RebalanceChannel {
	var channels []rebalances.RebalanceChannel
	for _, channelId := range rebalancer.Request.ChannelIds {
		channels = append(channels, rebalances.RebalanceChannel{
			ChannelId:   channelId,
			Status:      commons.Active,
			RebalanceId: rebalancer.RebalanceId,
			CreatedOn:   rebalancer.CreatedOn,
			UpdateOn:    rebalancer.UpdateOn,
		})
	}
	return channels
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
		getDeltaPerMille(failedHopAmountMsat, amountMsat) <
			commons.REBALANCE_ROUTE_FAILED_HOP_ALLOWED_DELTA_PER_MILLE
}

type RebalanceResult struct {
	RebalanceId       int            `json:"rebalanceId"`
	OutgoingChannelId int            `json:"outgoingChannelId"`
	IncomingChannelId int            `json:"incomingChannelId"`
	Status            commons.Status `json:"status"`
	Hops              string         `json:"hops"`
	TotalTimeLock     uint32         `json:"total_time_lock"`
	TotalFeeMsat      uint64         `json:"total_fee_msat"`
	TotalAmountMsat   uint64         `json:"total_amount_msat"`
	Error             string         `json:"error"`
	CreatedOn         time.Time      `json:"createdOn"`
	UpdateOn          time.Time      `json:"updateOn"`

	Route *lnrpc.Route `json:"-"`
}

func (result RebalanceResult) getLog() rebalances.RebalanceLog {
	return rebalances.RebalanceLog{
		Error:             result.Error,
		TotalTimeLock:     result.TotalTimeLock,
		TotalFeeMsat:      result.TotalFeeMsat,
		Status:            result.Status,
		IncomingChannelId: result.IncomingChannelId,
		OutgoingChannelId: result.OutgoingChannelId,
		RebalanceId:       result.RebalanceId,
		TotalAmountMsat:   result.TotalAmountMsat,
		Hops:              result.Hops,
		CreatedOn:         result.CreatedOn,
		UpdateOn:          result.UpdateOn,
	}
}

func RebalanceService(ctx context.Context, conn *grpc.ClientConn, db *sqlx.DB, nodeId int,
	broadcaster broadcast.BroadcastServer) {

	client := lnrpc.NewLightningClient(conn)
	router := routerrpc.NewRouterClient(conn)

	go initiateDelayedRebalancers(ctx, db, client, router)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		listener := broadcaster.Subscribe()
		for event := range listener {
			select {
			case <-ctx.Done():
				broadcaster.CancelSubscription(listener)
				return
			default:
			}
			if request, ok := event.(commons.RebalanceRequest); ok {
				if request.NodeId != nodeId {
					continue
				}
				if request.RequestTime == nil {
					now := time.Now().UTC()
					request.RequestTime = &now
				}
				// Previous rebalance cleanup delay
				time.Sleep(time.Millisecond * commons.REBALANCE_REBALANCE_DELAY_MILLISECONDS)
				processRebalanceRequest(ctx, db, request, nodeId)
			}
		}
	}
}

func initiateDelayedRebalancers(ctx context.Context, db *sqlx.DB,
	client lnrpc.LightningClient, router routerrpc.RouterClient) {

	ticker := clock.New().Tick(commons.REBALANCE_QUEUE_TICKER_SECONDS * time.Second)
	pending := commons.Pending
	active := commons.Active

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			activeRebalancers := getRebalancers(&active)
			if len(activeRebalancers) > commons.REBALANCE_MAXIMUM_CONCURRENCY {
				continue
			}

			pendingRebalancers := getRebalancers(&pending)
			if len(pendingRebalancers) > 0 {
				sort.Slice(pendingRebalancers, func(i, j int) bool {
					return pendingRebalancers[i].CreatedOn.Before(pendingRebalancers[j].CreatedOn)
				})

				if pendingRebalancers[0].CreatedOn.Before(time.Now()) {
					go pendingRebalancers[0].start(db, client, router,
						commons.REBALANCE_RUNNER_TIMEOUT_SECONDS,
						commons.REBALANCE_ROUTES_TIMEOUT_SECONDS,
						commons.REBALANCE_ROUTE_TIMEOUT_SECONDS)
				}
			}
		}
	}
}

func processRebalanceRequest(ctx context.Context, db *sqlx.DB, request commons.RebalanceRequest, nodeId int) {
	response := validateRebalanceRequest(request)
	if response != nil {
		sendResponse(request, *response)
		return
	}

	response = updateExistingRebalanceRequest(db, request)
	if response != nil {
		sendResponse(request, *response)
		return
	}

	// TODO CHECK if other direction ran before in the last 5 minutes
	// to prevent back-and-forth

	createdOn := time.Now().UTC()

	latestResult := getLatestResult(request.Origin, request.OriginId, request.IncomingChannelId, request.OutgoingChannelId, nil)
	if latestResult.RebalanceId != 0 {
		runningFor := request.RequestTime.Sub(latestResult.UpdateOn)
		if runningFor.Seconds() < commons.REBALANCE_MINIMUM_DELTA_SECONDS {
			sleepTime := commons.REBALANCE_MINIMUM_DELTA_SECONDS*time.Second - runningFor
			createdOn.Add(sleepTime)
		}
	}

	rebalancer := &Rebalancer{
		NodeId:    nodeId,
		CreatedOn: createdOn,
		UpdateOn:  createdOn,
		GlobalCtx: ctx,
		Runners:   make(map[int]*RebalanceRunner),
		Request:   request,
		Status:    commons.Pending,
	}
	rebalancerCtx, rebalancerCancel := context.WithTimeout(rebalancer.GlobalCtx,
		time.Second*time.Duration(commons.REBALANCE_TIMEOUT_SECONDS))
	rebalancer.RebalanceCtx = rebalancerCtx
	rebalancer.RebalanceCancel = rebalancerCancel
	if !addRebalancer(rebalancer) {
		response = &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error: fmt.Sprintf(
					"IncomingChannelId: %v already has a running rebalancer for origin: %v with reference number: %v",
					rebalancer.Request.IncomingChannelId, rebalancer.Request.Origin, rebalancer.Request.OriginReference),
			},
		}
		sendResponse(request, *response)
		return
	}
}

func (rebalancer *Rebalancer) start(
	db *sqlx.DB,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	runnerTimeout int,
	routesTimeout int,
	routeTimeout int) {

	active := commons.Active
	rebalancer.Status = commons.Active
	previousSuccess := getLatestResult(rebalancer.Request.Origin, rebalancer.Request.OriginId,
		rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId, &active)
	if time.Since(previousSuccess.UpdateOn).Seconds() > commons.REBALANCE_SUCCESS_TIMEOUT_SECONDS {
		previousSuccess = RebalanceResult{}
	}

	rebalanceId, err := rebalances.AddRebalanceAndChannels(db, (*rebalancer).getRebalance(), (*rebalancer).getRebalanceChannels())
	if err != nil {
		log.Error().Err(err).Msgf("Storing rebalance for origin: %v, originReference: %v and incomingChannelId: %v",
			rebalancer.Request.Origin, rebalancer.Request.OriginReference, rebalancer.Request.IncomingChannelId)
		response := &commons.RebalanceResponse{
			Request: rebalancer.Request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error:  "AddRebalancer was not completed",
			},
		}
		sendResponse(rebalancer.Request, *response)
		rebalancer.RebalanceCancel()
		return
	}
	rebalancer.RebalanceId = rebalanceId

	if previousSuccess.Hops != "" && previousSuccess.Route != nil {
		dummyRunner := &RebalanceRunner{
			RebalanceId:       rebalancer.RebalanceId,
			OutgoingChannelId: previousSuccess.OutgoingChannelId,
			IncomingChannelId: previousSuccess.IncomingChannelId,
			Invoices:          make(map[uint64]*lnrpc.AddInvoiceResponse),
			FailedHops:        make(map[string]uint64),
			Status:            commons.Active,
		}
		if rebalancer.Request.IncomingChannelId != 0 {
			// When incoming channel is provided then the runners loop over the outgoing channels
			rebalancer.Runners[previousSuccess.OutgoingChannelId] = dummyRunner
		} else {
			rebalancer.Runners[previousSuccess.IncomingChannelId] = dummyRunner
		}
		result := rebalancer.rerunPreviousSuccess(client, router, dummyRunner, previousSuccess.Route, routeTimeout)
		if result.Status == commons.Active {
			removeRebalancer(rebalancer)
		}
		rebalancer.processResult(db, result)
		if result.Status == commons.Active {
			return
		}
		if rebalancer.Request.IncomingChannelId != 0 {
			// When incoming channel is provided then the runners loop over the outgoing channels
			delete(rebalancer.Runners, previousSuccess.OutgoingChannelId)
		} else {
			delete(rebalancer.Runners, previousSuccess.IncomingChannelId)
		}
	}
	for i := 0; i < rebalancer.Request.MaximumConcurrency; i++ {
		go rebalancer.createRunner(db, client, router, runnerTimeout, routesTimeout, routeTimeout)
	}
}

func (rebalancer *Rebalancer) createRunner(
	db *sqlx.DB,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	runnerTimeout int,
	routesTimeout int,
	routeTimeout int) {

	result := RebalanceResult{
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
		rebalancer.RebalanceCancel()
		runningFor := time.Since(rebalancer.CreatedOn).Round(1 * time.Second)
		log.Info().Msgf("Pending ChannelId got exhausted for Origin: %v, OriginId: %v (%s)",
			rebalancer.Request.Origin, rebalancer.Request.OriginId, runningFor)
		if runningFor.Seconds() < commons.REBALANCE_MINIMUM_DELTA_SECONDS {
			sleepTime := commons.REBALANCE_MINIMUM_DELTA_SECONDS*time.Second - runningFor
			rebalancer.CreatedOn = time.Now().UTC()
			rebalancer.CreatedOn = rebalancer.CreatedOn.Add(sleepTime)
		}
		rebalancer.Status = commons.Pending
		if !addRebalancer(rebalancer) {
			log.Error().Msgf("Failed to reschedule the rebalancer for Origin: %v, OriginId: %v",
				rebalancer.Request.Origin, rebalancer.Request.OriginId)
		}
		return
	}

	runnerCtx, runnerCancel := context.WithTimeout(rebalancer.RebalanceCtx, time.Second*time.Duration(runnerTimeout))
	defer runnerCancel()
	runner := rebalancer.addRunner(channelId, runnerCtx, runnerCancel)

	result.IncomingChannelId = runner.IncomingChannelId
	result.OutgoingChannelId = runner.OutgoingChannelId

	rebalancer.startRunner(db, client, router, runner, routesTimeout, routeTimeout, result)

	runner.Cancel()
	runner.Status = commons.Inactive

	rebalancer.createRunner(db, client, router, runnerTimeout, routesTimeout, routeTimeout)
}

func (rebalancer *Rebalancer) startRunner(
	db *sqlx.DB,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	runner *RebalanceRunner,
	routesTimeout int,
	routeTimeout int,
	result RebalanceResult) {

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
		routeCtx, routeCancel := context.WithTimeout(runner.Ctx, time.Second*time.Duration(routeTimeout))
		result = runner.pay(routeCtx, client, router, rebalancer.Request.AmountMsat, route)
		routeCancel()
		if routeCtx.Err() == context.DeadlineExceeded {
			result.Error = routeCtx.Err().Error()
		}
		rebalancer.processResult(db, result)
		if result.Status == commons.Active {
			break
		}
	}

	if result.Status == commons.Pending {
		rebalancer.startRunner(db, client, router, runner, routesTimeout, routeTimeout, result)
	}
}

func (rebalancer *Rebalancer) getPendingChannelId() int {
outer:
	for _, channelId := range rebalancer.Request.ChannelIds {
		for existingChannelId := range rebalancer.Runners {
			if existingChannelId == channelId {
				continue outer
			}
		}
		return channelId
	}
	return 0
}

func (rebalancer *Rebalancer) rerunPreviousSuccess(
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	runner *RebalanceRunner,
	route *lnrpc.Route,
	routeTimeout int) RebalanceResult {

	routeCtx, routeCancel := context.WithTimeout(rebalancer.RebalanceCtx, time.Second*time.Duration(routeTimeout))
	defer routeCancel()
	return runner.pay(routeCtx, client, router, rebalancer.Request.AmountMsat, route)
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

func (rebalancer *Rebalancer) processResult(db *sqlx.DB, result RebalanceResult) {
	result.UpdateOn = time.Now().UTC()
	addRebalanceResult(rebalancer.Request.Origin, rebalancer.Request.OriginId,
		rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId, result)
	err := rebalances.AddRebalanceLog(db, result.getLog())
	if err != nil {
		log.Error().Err(err).Msgf("Failed to add rebalance log entry for rebalanceId: %v", rebalancer.RebalanceId)
	}
}

func (runner *RebalanceRunner) getRoutes(
	ctx context.Context,
	client lnrpc.LightningClient,
	router routerrpc.RouterClient,
	nodeId int,
	amountMsat uint64,
	fixedFeeMsat uint64) ([]*lnrpc.Route, error) {

	var err error

	outgoingChannel := commons.GetChannelSettingByChannelId(runner.OutgoingChannelId)
	var remoteNode commons.ManagedNodeSettings
	if outgoingChannel.FirstNodeId == nodeId {
		remoteNode = commons.GetNodeSettingsByNodeId(outgoingChannel.SecondNodeId)
	} else {
		remoteNode = commons.GetNodeSettingsByNodeId(outgoingChannel.FirstNodeId)
	}
	remoteNodePublicKey, err := hex.DecodeString(remoteNode.PublicKey)
	if err != nil {
		return nil, errors.Wrapf(err, "Decoding public key for outgoing nodeId: %v", outgoingChannel.SecondNodeId)
	}

	var routes *lnrpc.QueryRoutesResponse
	routes, err = client.QueryRoutes(ctx, &lnrpc.QueryRoutesRequest{
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
	route *lnrpc.Route) RebalanceResult {

	rebalanceResult := RebalanceResult{
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
	if result.Route != nil {
		rebalanceResult.Route = result.Route
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
	invoice, err := client.AddInvoice(ctx, &lnrpc.Invoice{Value: int64(amountMsat),
		Memo:   "Rebalance attempt",
		Expiry: int64(commons.REBALANCE_TIMEOUT_SECONDS)})
	if err != nil {
		return nil, errors.Wrapf(err, "AddInvoice for %v msat", amountMsat)
	}
	runner.Invoices[amountMsat] = invoice
	return invoice, nil
}

func sendResponse(request commons.RebalanceRequest, response commons.RebalanceResponse) {
	if request.ResponseChannel != nil {
		request.ResponseChannel <- response
	}
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
	rebalancer := getRebalancer(request.Origin, request.OriginId)
	if rebalancer == nil {
		return nil
	}
	if request.RequestTime != nil && rebalancer.UpdateOn.After(*request.RequestTime) {
		return &commons.RebalanceResponse{
			Request: request,
			CommunicationResponse: commons.CommunicationResponse{
				Status: commons.Inactive,
				Error: fmt.Sprintf(
					"IncomingChannelId: %v, OutgoingChannelId: %v already has a more recent running rebalancer for origin: %v with originId: %v (ref: %v)",
					rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId, rebalancer.Request.Origin, rebalancer.Request.OriginId,
					rebalancer.Request.OriginReference),
			},
		}
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
	return &commons.RebalanceResponse{
		Request: request,
		CommunicationResponse: commons.CommunicationResponse{
			Status: commons.Inactive,
			Error: fmt.Sprintf(
				"IncomingChannelId: %v, OutgoingChannelId: %v already has a running rebalancer for origin: %v with originId: %v (ref: %v)",
				rebalancer.Request.IncomingChannelId, rebalancer.Request.OutgoingChannelId,
				rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OriginReference),
		},
	}
}

func setRebalancer(db *sqlx.DB, request commons.RebalanceRequest, rebalancer *Rebalancer) error {
	rebalancer.UpdateOn = time.Now().UTC()
	rebalancer.Request = request
	err := rebalances.SetRebalanceAndChannels(db, rebalancer.getRebalance(), rebalancer.getRebalanceChannels())
	if err != nil {
		log.Error().Err(err).Msgf("Failed to add rebalance log entry for rebalanceId: %v", rebalancer.RebalanceId)
	}
	return errors.Wrapf(err,
		"Updating the database with the new rebalance settings for origin: %v with originId: %v (ref: %v)",
		rebalancer.Request.Origin, rebalancer.Request.OriginId, rebalancer.Request.OriginReference)
}
