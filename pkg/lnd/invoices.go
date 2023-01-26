package lnd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/commons"
)

type invoicesClient interface {
	SubscribeInvoices(ctx context.Context, in *lnrpc.InvoiceSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error)
	ListInvoices(ctx context.Context, in *lnrpc.ListInvoiceRequest,
		opts ...grpc.CallOption) (*lnrpc.ListInvoiceResponse, error)
}

type Invoice struct {

	/*
	   An optional memo to attach along with the invoice. Used for record keeping
	   purposes for the invoice's creator, and will also be set in the description
	   field of the encoded payment request if the description_hash field is not
	   being used.
	*/
	Memo string `db:"memo" json:"memo"`

	/*
	   The hex-encoded preimage (32 byte) which will allow settling an incoming
	   HTLC payable to this preimage.
	*/
	RPreimage string `db:"r_preimage" json:"r_preimage"`

	/*
	   The hash of the preimage.
	*/
	RHash string `db:"r_hash" json:"r_hash"`

	// The value of the invoice
	ValueMsat int64 `db:"value_msat" json:"value_msat"`

	// When this invoice was created
	CreationDate time.Time `db:"creation_date" json:"creation_date"`

	// When this invoice was settled
	SettleDate time.Time `db:"settle_date" json:"settle_date"`

	/*
	   A bare-bones invoice for a payment within the Lightning Network. With the
	   details of the invoice, the sender has all the data necessary to send a
	   payment to the recipient.
	*/
	PaymentRequest string `db:"payment_request" json:"payment_request"`

	/*
	   A bare-bones invoice for a payment within the Lightning Network. With the
	   details of the invoice, the sender has all the data necessary to send a
	   payment to the recipient.
	*/
	Destination string `db:"destination_pub_key" json:"destination_pub_key"`

	/*
	   Hash (SHA-256) of a description of the payment. Used if the description of
	   payment (memo) is too long to naturally fit within the description field
	   of an encoded payment request.
	*/
	DescriptionHash []byte `db:"description_hash" json:"description_hash"`

	// Payment request expiry time in seconds. Default is 3600 (1 hour).
	Expiry int64 `db:"expiry" json:"expiry"`

	// Fallback on-chain address.
	FallbackAddr string `db:"fallback_addr" json:"fallback_addr"`

	// Delta to use for the time-lock of the CLTV extended to the final hop.
	CltvExpiry uint64 `db:"cltv_expiry" json:"cltv_expiry"`

	/*
	   Route hints that can each be individually used to assist in reaching the
	   invoice's destination.
	*/
	//repeated RouteHint route_hints = 14;
	RouteHints []byte `db:"route_hints" json:"route_hints"`

	// Whether this invoice should include routing hints for private channels.
	Private bool `db:"private" json:"private"`

	/*
	   The "add" index of this invoice. Each newly created invoice will increment
	   this index making it monotonically increasing. Callers to the
	   SubscribeInvoices call can use this to instantly get notified of all added
	   invoices with an add_index greater than this one.
	*/
	AddIndex uint64 `db:"add_index" json:"add_index"`

	/*
	   The "settle" index of this invoice. Each newly settled invoice will
	   increment this index making it monotonically increasing. Callers to the
	   SubscribeInvoices call can use this to instantly get notified of all
	   settled invoices with an settle_index greater than this one.
	*/
	SettleIndex uint64 `db:"settle_index" json:"settle_index"`

	/*
	   The amount that was accepted for this invoice, in satoshis. This will ONLY
	   be set if this invoice has been settled. We provide this field as if the
	   invoice was created with a zero value, then we need to record what amount
	   was ultimately accepted. Additionally, it's possible that the sender paid
	   MORE that was specified in the original invoice. So we'll record that here
	   as well.
	*/
	AmtPaidSat int64 `db:"amt_paid_sat" json:"amt_paid_sat"`

	/*
	   The amount that was accepted for this invoice, in millisatoshis. This will
	   ONLY be set if this invoice has been settled. We provide this field as if
	   the invoice was created with a zero value, then we need to record what
	   amount was ultimately accepted. Additionally, it's possible that the sender
	   paid MORE that was specified in the original invoice. So we'll record that
	   here as well.
	*/
	AmtPaidMsat int64 `db:"amt_paid_msat" json:"amt_paid_msat"`

	InvoiceState string `db:"invoice_state" json:"invoice_state"`
	//OPEN = 0;
	//SETTLED = 1;
	//CANCELED = 2;
	//ACCEPTED = 3;

	// List of HTLCs paying to this invoice [EXPERIMENTAL].
	Htlcs []byte `db:"htlcs" json:"htlcs"`
	//repeated InvoiceHTLC htlcs = 22;

	// List of features advertised on the invoice.
	//map<uint32, Feature> features = 24;
	// features []*lnrpc.Feature
	Features []byte `db:"features" json:"features"`

	/*
	   Indicates if this invoice was a spontaneous payment that arrived via keysend
	   [EXPERIMENTAL].
	*/
	IsKeysend bool `db:"is_keysend" json:"is_keysend"`

	/*
	   The payment address of this invoice. This value will be used in MPP
	   payments, and also for newer invoices that always require the MPP payload
	   for added end-to-end security.
	*/
	PaymentAddr string `db:"payment_addr" json:"payment_addr"`

	/*
	   Signals whether this is an AMP invoice.
	*/
	IsAmp bool `db:"is_amp" json:"is_amp"`

	/*
	   [EXPERIMENTAL]:
	   Maps a 32-byte hex-encoded set ID to the sub-invoice AMP state for the
	   given set ID. This field is always populated for AMP invoices, and can be
	   used alongside LookupInvoice to obtain the HTLC information related to a
	   given sub-invoice.
	*/
	//map<string, AMPInvoiceState> amp_invoice_state = 28;
	AmpInvoiceState   []byte    `db:"amp_invoice_state" json:"amp_invoice_state"`
	DestinationNodeId *int      `db:"destination_node_id" json:"destinationNodeId"`
	NodeId            int       `db:"node_id" json:"nodeId"`
	ChannelId         *int      `db:"channel_id" json:"channelId"`
	CreatedOn         time.Time `db:"created_on" json:"created_on"`
	UpdatedOn         time.Time `db:"updated_on" json:"updated_on"`
}

func fetchLastInvoiceIndexes(db *sqlx.DB, nodeId int) (addIndex uint64, settleIndex uint64, err error) {
	// index starts at 1
	sqlLatest := `select coalesce(max(add_index),1), coalesce(max(settle_index),1) from invoice where node_id = $1;`

	row := db.QueryRow(sqlLatest, nodeId)
	err = row.Scan(&addIndex, &settleIndex)

	if err != nil {
		log.Error().Msgf("getting max invoice indexes: %v", err)
		return 0, 0, errors.Wrap(err, "getting max invoice indexes")
	}

	return addIndex, settleIndex, nil
}

func SubscribeAndStoreInvoices(ctx context.Context, client invoicesClient, db *sqlx.DB,
	nodeSettings commons.ManagedNodeSettings,
	invoiceEventChannel chan commons.InvoiceEvent,
	serviceEventChannel chan commons.ServiceEvent) {

	defer log.Info().Msgf("SubscribeAndStoreInvoices terminated for nodeId: %v", nodeSettings.NodeId)

	var serviceStatus commons.Status
	bootStrapping := true
	subscriptionStream := commons.InvoiceStream
	importCounter := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Get the latest settle and add index to prevent duplicate entries.
		addIndex, _, err := fetchLastInvoiceIndexes(db, nodeSettings.NodeId)
		if err != nil {
			serviceStatus = processError(serviceStatus, serviceEventChannel, nodeSettings, subscriptionStream, err)
			continue
		}

		listInvoiceResponse, err := client.ListInvoices(ctx, &lnrpc.ListInvoiceRequest{
			NumMaxInvoices: commons.STREAM_LND_MAX_INVOICES,
			IndexOffset:    addIndex,
		})
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = processError(serviceStatus, serviceEventChannel, nodeSettings, subscriptionStream, err)
			continue
		}

		if bootStrapping {
			importCounter = importCounter + len(listInvoiceResponse.Invoices)
			if len(listInvoiceResponse.Invoices) >= commons.STREAM_LND_MAX_INVOICES {
				log.Info().Msgf("Still running bulk import of invoices (%v)", importCounter)
			}
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Initializing, serviceStatus)
		} else {
			serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
		}
		for _, invoice := range listInvoiceResponse.Invoices {
			processInvoice(invoice, nodeSettings, db, invoiceEventChannel, bootStrapping)
		}
		if bootStrapping && len(listInvoiceResponse.Invoices) < commons.STREAM_LND_MAX_INVOICES {
			bootStrapping = false
			log.Info().Msgf("Bulk import of invoices done (%v)", importCounter)
			break
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Get the latest settle and add index to prevent duplicate entries.
		addIndex, settleIndex, err := fetchLastInvoiceIndexes(db, nodeSettings.NodeId)
		if err != nil {
			serviceStatus = processError(serviceStatus, serviceEventChannel, nodeSettings, subscriptionStream, err)
			continue
		}

		streamCtx, cancel := context.WithCancel(ctx)
		stream, err := client.SubscribeInvoices(streamCtx, &lnrpc.InvoiceSubscription{
			AddIndex:    addIndex,
			SettleIndex: settleIndex,
		})
		if err != nil {
			cancel()
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = processError(serviceStatus, serviceEventChannel, nodeSettings, subscriptionStream, err)
			continue
		}

		serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Active, serviceStatus)
		invoice, err := stream.Recv()
		if err != nil {
			cancel()
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			serviceStatus = processError(serviceStatus, serviceEventChannel, nodeSettings, subscriptionStream, err)
			continue
		}
		processInvoice(invoice, nodeSettings, db, invoiceEventChannel, bootStrapping)
		cancel()
	}
}

func processInvoice(invoice *lnrpc.Invoice, nodeSettings commons.ManagedNodeSettings, db *sqlx.DB, invoiceEventChannel chan commons.InvoiceEvent, bootStrapping bool) {
	invoiceEvent := commons.InvoiceEvent{
		EventData: commons.EventData{
			EventTime: time.Now().UTC(),
			NodeId:    nodeSettings.NodeId,
		},
	}

	var destinationPublicKey = ""
	var destinationNodeId *int
	// if empty payment request invoice is likely keysend
	if invoice.PaymentRequest != "" {
		// Check the running nodes network. Currently we assume we are running on Bitcoin mainnet
		nodeNetwork := getNodeNetwork(invoice.PaymentRequest)

		inva, err := zpay32.Decode(invoice.PaymentRequest, nodeNetwork)
		if err != nil {
			log.Error().Msgf("Subscribe and store invoices - decode payment request: %v", err)
		} else {
			destinationPublicKey = fmt.Sprintf("%x", inva.Destination.SerializeCompressed())
			destinationNodeIdValue := commons.GetNodeIdByPublicKey(destinationPublicKey, nodeSettings.Chain, nodeSettings.Network)
			destinationNodeId = &destinationNodeIdValue
			invoiceEvent.DestinationNodeId = destinationNodeId
		}
	}

	err := insertInvoice(db, invoice, destinationPublicKey, destinationNodeId, nodeSettings.NodeId, invoiceEvent, invoiceEventChannel, bootStrapping)
	if err != nil {
		log.Error().Err(err).Msg("Storing invoice failed")
	}
}

func processError(serviceStatus commons.Status, serviceEventChannel chan commons.ServiceEvent, nodeSettings commons.ManagedNodeSettings,
	subscriptionStream commons.SubscriptionStream, err error) commons.Status {

	serviceStatus = SendStreamEvent(serviceEventChannel, nodeSettings.NodeId, subscriptionStream, commons.Pending, serviceStatus)
	log.Error().Err(err).Msgf("Failed to obtain last know invoice, will retry in %v seconds", commons.STREAM_ERROR_SLEEP_SECONDS)
	time.Sleep(commons.STREAM_ERROR_SLEEP_SECONDS * time.Second)
	return serviceStatus
}

// getNodeNetwork
// Obtained from invoice.PaymentRequest
// MainNetParams           bc
// RegressionNetParams     bcrt
// SigNetParams            tbs
// TestNet3Params          tb
// SimNetParams            sb
// Example: invoice.PaymentRequest = lnbcrt500u1p3vmd6upp5y7ndr6dmyehql..."
//   - First two characters should be "ln"
//   - Next 2+2 characters determine the network
//   - Here the network is RegressionNetParams - bcrt
//
// This values come from chaincfg.<Params>.Bech32HRPSegwit
func getNodeNetwork(pmntReq string) *chaincfg.Params {
	nodeNetworkPrefix := pmntReq[2:4]
	nodeNetworkSuffix := ""

	switch {
	case nodeNetworkPrefix == "bc":
		nodeNetworkSuffix = pmntReq[4:6]
		if nodeNetworkSuffix == "rt" {
			return &chaincfg.RegressionNetParams
		} else {
			return &chaincfg.MainNetParams
		}
	case nodeNetworkPrefix == "tb":
		nodeNetworkSuffix = pmntReq[4:5]
		if nodeNetworkSuffix == "s" {
			return &chaincfg.SigNetParams
		} else {
			return &chaincfg.TestNet3Params
		}
	case nodeNetworkPrefix == "sb":
		return &chaincfg.SimNetParams
	default:
		return &chaincfg.MainNetParams
	}
}

func insertInvoice(db *sqlx.DB, invoice *lnrpc.Invoice, destination string, destinationNodeId *int, nodeId int,
	invoiceEvent commons.InvoiceEvent, invoiceEventChannel chan commons.InvoiceEvent, bootStrapping bool) error {

	rhJson, err := json.Marshal(invoice.RouteHints)
	if err != nil {
		log.Error().Msgf("insert invoice: json marshal route hints: %v", err)
		return errors.Wrapf(err, "insert invoice: json marshal route hints")
	}

	htlcJson, err := json.Marshal(invoice.Htlcs)
	if err != nil {
		log.Error().Msgf("insert invoice - json marshal htlcs: %v", err)
		return errors.Wrapf(err, "insert invoice: json marshal htlcs")
	}
	var channelId *int
	if len(invoice.Htlcs) > 0 {
		channelId = getChannelIdByLndShortChannelId(invoice.Htlcs[len(invoice.Htlcs)-1].ChanId)
	}

	featuresJson, err := json.Marshal(invoice.Features)
	if err != nil {
		log.Error().Msgf("insert invoice - json marshal features: %v", err)
		return errors.Wrapf(err, "insert invoice: json marshal features")
	}

	aisJson, err := json.Marshal(invoice.AmpInvoiceState)
	if err != nil {
		log.Error().Msgf("")
		return errors.Wrapf(err, "insert invoice: amp invoice state")
	}

	i := Invoice{
		Memo:              invoice.Memo,
		RPreimage:         hex.EncodeToString(invoice.RPreimage),
		RHash:             hex.EncodeToString(invoice.RHash),
		ValueMsat:         invoice.ValueMsat,
		CreationDate:      time.Unix(invoice.CreationDate, 0).UTC(),
		SettleDate:        time.Unix(invoice.SettleDate, 0).UTC(),
		PaymentRequest:    invoice.PaymentRequest,
		Destination:       destination,
		DescriptionHash:   invoice.DescriptionHash,
		Expiry:            invoice.Expiry,
		FallbackAddr:      invoice.FallbackAddr,
		CltvExpiry:        invoice.CltvExpiry,
		RouteHints:        rhJson,
		Private:           false,
		AddIndex:          invoice.AddIndex,
		SettleIndex:       invoice.SettleIndex,
		AmtPaidSat:        invoice.AmtPaidSat,
		AmtPaidMsat:       invoice.AmtPaidMsat,
		InvoiceState:      invoice.State.String(), // ,
		Htlcs:             htlcJson,
		Features:          featuresJson,
		IsKeysend:         invoice.IsKeysend,
		PaymentAddr:       hex.EncodeToString(invoice.PaymentAddr),
		IsAmp:             invoice.IsAmp,
		AmpInvoiceState:   aisJson,
		DestinationNodeId: destinationNodeId,
		NodeId:            nodeId,
		ChannelId:         channelId,
		CreatedOn:         time.Now().UTC(),
		UpdatedOn:         time.Now().UTC(),
	}

	var sqlInvoice = `INSERT INTO invoice (
    memo,
    r_preimage,
    r_hash,
    value_msat,
    creation_date,
    settle_date,
    payment_request,
    destination_pub_key,
    description_hash,
    expiry,
    fallback_addr,
    cltv_expiry,
    route_hints,
    private,
    add_index,
    settle_index,
    amt_paid_msat,
    /*
    The state the invoice is in.
        OPEN = 0;
        SETTLED = 1;
        CANCELED = 2;
        ACCEPTED = 3;
    */
    invoice_state,
    htlcs,
    features,
    is_keysend,
    payment_addr,
    is_amp,
    amp_invoice_state,
    destination_node_id,
    node_id,
    channel_id,
    created_on,
    updated_on
) VALUES(
	:memo,
    :r_preimage,
    :r_hash,
    :value_msat,
    :creation_date,
    :settle_date,
    :payment_request,
	:destination_pub_key,
    :description_hash,
    :expiry,
    :fallback_addr,
    :cltv_expiry,
    :route_hints,
    :private,
    :add_index,
    :settle_index,
    :amt_paid_msat,
    :invoice_state,
    :htlcs,
    :features,
    :is_keysend,
    :payment_addr,
    :is_amp,
    :amp_invoice_state,
	:destination_node_id,
	:node_id,
    :channel_id,
    :created_on,
    :updated_on
);`

	_, err = db.NamedExec(sqlInvoice, i)

	if err != nil {
		log.Error().Msgf("insert invoice: %v", err)
		return errors.Wrapf(err, "insert invoice")
	}

	if invoiceEventChannel != nil && !bootStrapping {
		invoiceEvent.AddIndex = invoice.AddIndex
		invoiceEvent.ValueMSat = uint64(invoice.ValueMsat)
		invoiceEvent.State = invoice.GetState()
		// Add other info for settled and accepted states
		//	Invoice_OPEN     = 0
		//	Invoice_SETTLED  = 1
		//	Invoice_CANCELED = 2
		//	Invoice_ACCEPTED = 3
		if invoice.State == 1 || invoice.State == 3 {
			invoiceEvent.AmountPaidMsat = uint64(invoice.AmtPaidMsat)
			invoiceEvent.SettledDate = time.Unix(invoice.SettleDate, 0)
		}
		if channelId != nil {
			invoiceEvent.ChannelId = *channelId
		}
		invoiceEventChannel <- invoiceEvent
	}
	return nil
}
