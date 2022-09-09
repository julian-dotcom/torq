package invoices

import (
	"context"
	"encoding/hex"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

type feature struct {
	Name       string `json:"name"`
	IsKnown    bool   `json:"isKnown"`
	IsRequired bool   `json:"isRequired"`
}

type featureMap map[uint32]feature

type hopHint struct {
	LNDShortChannelId uint64 `json:"chanId"`
	NodeId            string `json:"nodeId"`
	FeeBase           uint32 `json:"feeBase"`
	CltvExpiryDelta   uint32 `json:"cltvExpiryDelta"`
	FeeProportional   uint32 `json:"feeProportionalMillionths"`
}

type routeHint struct {
	HopHints []hopHint `json:"hopHints"`
}

type DecodedInvoice struct {
	PaymentRequest    string      `json:"paymentRequest"`
	DestinationPubKey string      `json:"destinationPubKey"`
	RHash             string      `json:"rHash"`
	Memo              string      `json:"memo"`
	ValueMsat         int64       `json:"value"`
	PaymentAddr       string      `json:"paymentAddr"`
	FallbackAddr      string      `json:"fallbackAddr"`
	Expiry            int64       `json:"expiry"`
	CltvExpiry        int64       `json:"cltvExpiry"`
	Private           bool        `json:"private"`
	Features          featureMap  `json:"features"`
	RouteHints        []routeHint `json:"routeHints"`
}

func constructRouteHints(routeHints []*lnrpc.RouteHint) []routeHint {

	var r []routeHint

	for _, rh := range routeHints {
		var hopHints []hopHint
		for _, hh := range rh.HopHints {
			hopHints = append(hopHints, hopHint{
				LNDShortChannelId: hh.ChanId,
				NodeId:            hh.NodeId,
				FeeBase:           hh.FeeBaseMsat,
				CltvExpiryDelta:   hh.CltvExpiryDelta,
				FeeProportional:   hh.FeeProportionalMillionths,
			})
		}
		r = append(r, routeHint{
			HopHints: hopHints,
		})
	}

	return r
}

func constructFeatureMap(features map[uint32]*lnrpc.Feature) featureMap {

	f := featureMap{}
	for n, v := range features {
		f[n] = feature{
			Name:       v.Name,
			IsKnown:    v.IsKnown,
			IsRequired: v.IsRequired,
		}
	}

	return f
}

func constructDecodedInvoice(decodedInvoice *lnrpc.PayReq) *DecodedInvoice {
	return &DecodedInvoice{
		DestinationPubKey: decodedInvoice.Destination,
		RHash:             decodedInvoice.PaymentHash,
		Memo:              decodedInvoice.Description,
		ValueMsat:         decodedInvoice.NumMsat,
		FallbackAddr:      decodedInvoice.FallbackAddr,
		Expiry:            decodedInvoice.Expiry,
		CltvExpiry:        decodedInvoice.CltvExpiry,
		RouteHints:        constructRouteHints(decodedInvoice.RouteHints),
		Features:          constructFeatureMap(decodedInvoice.Features),
		PaymentAddr:       hex.EncodeToString(decodedInvoice.PaymentAddr),
	}
}

// decodeInvoice Decode a lightning invoice
func decodeInvoice(db *sqlx.DB, invoice string) (*DecodedInvoice, error) {
	log.Info().Msgf("Decoding invoice: %s", invoice)
	// Get lnd client
	connectionDetails, err := settings.GetConnectionDetails(db)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting node connection details from the db: %s", err.Error())
		return nil, errors.New("Error getting node connection details from the db")
	}

	// TODO: change to select which local node
	conn, err := lnd_connect.Connect(
		connectionDetails[0].GRPCAddress,
		connectionDetails[0].TLSFileBytes,
		connectionDetails[0].MacaroonFileBytes)
	if err != nil {
		log.Error().Err(err).Msgf("can't connect to LND: %s", err.Error())
		return nil, errors.Newf("can't connect to LND")
	}
	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)
	// Decode invoice
	ctx := context.Background()
	decodedInvoice, err := client.DecodePayReq(ctx, &lnrpc.PayReqString{
		PayReq: invoice,
	})
	if err != nil {
		return nil, err
	}

	return constructDecodedInvoice(decodedInvoice), nil
}

func decodeInvoiceHandler(c *gin.Context, db *sqlx.DB) {
	invoice := c.Query("invoice")

	di, err := decodeInvoice(db, invoice)

	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "could not decode invoice")
		return
	}

	c.JSON(http.StatusOK, di)
}
