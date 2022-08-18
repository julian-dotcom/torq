package payments

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	qp "github.com/lncapital/torq/internal/query_parser"
	"github.com/lncapital/torq/internal/settings"
	ah "github.com/lncapital/torq/pkg/api_helpers"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

type NewPaymentRequestBody struct {
	Dest        []byte
	Amt         int64
	AmtMSat     int64
	PaymentHash []byte
	Invoice     string
	TimeOutSecs int32
}

func getPaymentsHandler(c *gin.Context, db *sqlx.DB) {

	// Filter parser with whitelisted columns
	var filter sq.Sqlizer
	filterParam := c.Query("filter")
	var err error
	if filterParam != "" {
		filter, err = qp.ParseFilterParam(filterParam, []string{
			"date",
			"destination_pub_key",
			"status",
			"value",
			"fee",
			"ppm",
			"failure_reason",
			"is_rebalance",
			"is_mpp",
			"count_successful_attempts",
			"count_failed_attempts",
			"seconds_in_flight",
			"payment_hash",
			"payment_preimage",
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var sort []string
	sortParam := c.Query("order")
	if sortParam != "" {
		// Order parser with whitelisted columns
		sort, err = qp.ParseOrderParams(
			sortParam,
			[]string{
				"date",
				"status",
				"value",
				"fee",
				"ppm",
				"failure_reason",
				"count_successful_attempts",
				"count_failed_attempts",
				"seconds_in_flight",
			})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var limit uint64
	if c.Query("limit") != "" {
		limit, err = strconv.ParseUint(c.Query("limit"), 10, 64)
		switch err.(type) {
		case nil:
			break
		case *strconv.NumError:
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Limit must be a positive number"})
			return
		default:
			server_errors.LogAndSendServerError(c, err)
		}
		if limit == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Limit must be a at least 1"})
			return
		}
	}

	var offset uint64
	if c.Query("offset") != "" {
		offset, err = strconv.ParseUint(c.Query("offset"), 10, 64)
		switch err.(type) {
		case nil:
			break
		case *strconv.NumError:
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Offset must be a positive number"})
			return
		default:
			server_errors.LogAndSendServerError(c, err)
		}
	}

	r, total, err := getPayments(db, filter, sort, limit, offset)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, ah.ApiResponse{
		Data: r, Pagination: ah.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}})
}

func getPaymentHandler(c *gin.Context, db *sqlx.DB) {

	r, err := getPaymentDetails(db, c.Param("identifier"))
	switch err.(type) {
	case nil:
		break
	case ErrPaymentNotFound:
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error(), "Identifier": c.Param("identifier")})
		return
	default:
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

//newPaymentHandler - new payment handler
//A new payment can be made either by providing an invoice or by providing:
// - dest - the identity pubkey of the payment recipient
// - amt(number of satoshis) or amt_msat(number of millisatoshis)
// - amt and amt_msat are mutually exclusive
// - payments hash - the hash to use within the payment's HTLC
//Timeout seconds is mandatory for both ways
func newPaymentHandler(c *gin.Context, db *sqlx.DB) {
	connectionDetails, err := settings.GetConnectionDetails(db)
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Connecting to LND")
	}
	defer conn.Close()
	client := routerrpc.NewRouterClient(conn)

	var requestBody NewPaymentRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	dest := requestBody.Dest
	amt := requestBody.Amt
	amtMSat := requestBody.AmtMSat
	paymentHash := requestBody.PaymentHash
	invoice := requestBody.Invoice
	timeOutSecs := requestBody.TimeOutSecs

	if len(invoice) == 0 {
		if len(dest) == 0 && (amt == 0 || amtMSat == 0) && len(paymentHash) == 0 {
			server_errors.LogAndSendServerError(c, errors.New("Payment destination missing"))
			return
		}
	}

	if timeOutSecs == 0 {
		server_errors.LogAndSendServerError(c, errors.New("timeout_seconds must be specified"))
		return
	}

	log.Debug().Msgf("Invoice: %v", invoice)

	resp, err := SendNewPayment(dest, amt, amtMSat, paymentHash, invoice, timeOutSecs, client)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Sending payment")
		return
	}

	c.JSON(http.StatusOK, resp)
}

func RegisterPaymentsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getPaymentsHandler(c, db) })
	r.GET(":identifier", func(c *gin.Context) { getPaymentHandler(c, db) })
	r.POST("newpayment", func(c *gin.Context) { newPaymentHandler(c, db) })
}
