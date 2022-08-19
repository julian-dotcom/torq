package invoices

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	qp "github.com/lncapital/torq/internal/query_parser"
	"github.com/lncapital/torq/internal/settings"
	ah "github.com/lncapital/torq/pkg/api_helpers"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

type NewInvoiceRequestBody struct {
	Memo      string
	ValueMsat int64
	Expiry    int64
	Amp       bool
}

type newInvoiceResponse struct {
	PaymentRequest string
	ValueMsat      int64
	Expiry         int64
}

func getInvoicesHandler(c *gin.Context, db *sqlx.DB) {

	// Filter parser with whitelisted columns
	var filter sq.Sqlizer
	filterParam := c.Query("filter")
	var err error
	if filterParam != "" {
		filter, err = qp.ParseFilterParam(filterParam, []string{
			"add_index",
			"creation_date",
			"settle_date",
			"settle_index",
			"payment_request",
			"destination_pub_key",
			"r_hash",
			"r_preimage",
			"memo",
			"value",
			"amt_paid",
			"invoice_state",
			"is_rebalance",
			"is_keysend",
			"is_amp",
			"payment_addr",
			"fallback_addr",
			"updated_on",
			"expiry",
			"cltv_expiry",
			"private",
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
				"creation_date",
				"settle_date",
				"add_index",
				"settle_index",
				"memo",
				"value",
				"amt_paid",
				"invoice_state",
				"is_rebalance",
				"is_keysend",
				"is_amp",
				"updated_on",
				"expiry",
				"private",
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

	r, total, err := getInvoices(db, filter, sort, limit, offset)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, ah.ApiResponse{
		Data: r,
		Pagination: ah.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}})
}

func getInvoiceHandler(c *gin.Context, db *sqlx.DB) {

	r, err := getInvoiceDetails(db, c.Param("identifier"))
	switch err.(type) {
	case nil:
		break
	case ErrInvoiceNotFound:
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error(), "Identifier": c.Param("identifier")})
		return
	default:
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

func newInvoiceHandler(c *gin.Context, db *sqlx.DB) {

	var requestBody NewInvoiceRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("JSON binding the request body")
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	memo := requestBody.Memo
	valueMsat := requestBody.ValueMsat
	//Not mandatory. Default value is 3600 seconds
	expiry := requestBody.Expiry
	amp := requestBody.Amp

	//log.Debug().Msgf("Memo: %v, value: %v, expiry: %v, amp: %v", memo, valueMsat, expiry, amp)

	if valueMsat <= 0 {
		log.Error().Msgf("Invoice value not valid")
		server_errors.LogAndSendServerError(c, errors.New("Invoice value and/or expiry time not valid"))
		return
	}

	connectionDetails, err := settings.GetConnectionDetails(db)
	conn, err := lnd_connect.Connect(
		connectionDetails.GRPCAddress,
		connectionDetails.TLSFileBytes,
		connectionDetails.MacaroonFileBytes)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Connecting to LND")
	}

	defer conn.Close()

	client := lnrpc.NewLightningClient(conn)

	resp, err := newInvoice(client, memo, valueMsat, expiry, amp)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Creating new invoice")
		return
	}

	//If expiry time not provided,default value is 3600 seconds(1 hour)
	if expiry == 0 {
		expiry = 3600
	}

	invoiceResp := newInvoiceResponse{
		PaymentRequest: resp,
		ValueMsat:      valueMsat,
		Expiry:         expiry,
	}
	c.JSON(http.StatusOK, invoiceResp)
}

func RegisterInvoicesRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getInvoicesHandler(c, db) })
	r.GET(":identifier", func(c *gin.Context) { getInvoiceHandler(c, db) })
	r.POST("newinvoice", func(c *gin.Context) { newInvoiceHandler(c, db) })
}
