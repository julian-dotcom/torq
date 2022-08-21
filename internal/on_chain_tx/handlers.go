package on_chain_tx

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

type NewAddressRequestBody struct {
	//AddressType has to be one of:
	//- `p2wkh`: Pay to witness key hash (`WITNESS_PUBKEY_HASH` = 0)
	//- `np2wkh`: Pay to nested witness key hash (`NESTED_PUBKEY_HASH` = 1)
	//- `p2tr`: Pay to taproot pubkey (`TAPROOT_PUBKEY` = 4)
	//WITNESS_PUBKEY_HASH = 0;
	//NESTED_PUBKEY_HASH = 1;
	//UNUSED_WITNESS_PUBKEY_HASH = 2;
	//UNUSED_NESTED_PUBKEY_HASH = 3;
	//TAPROOT_PUBKEY = 4;
	//UNUSED_TAPROOT_PUBKEY = 5;
	Type int32
	//The name of the account to generate a new address for. If empty, the default wallet account is used.
	Account string
}

type newAddressResponse struct {
	Address     string
	AddressType string
}

type sendCoinsRequestBody struct {
	Addr        string
	Amount      int64
	TargetConf  int32
	SatPerVbyte *uint64
}

type sendCoinsResponse struct {
	Address string
	Amount  int64
	Txid    string
}

func getOnChainTxsHandler(c *gin.Context, db *sqlx.DB) {

	// Filter parser with whitelisted columns
	var filter sq.Sqlizer
	filterParam := c.Query("filter")
	var err error
	if filterParam != "" {
		filter, err = qp.ParseFilterParam(filterParam, []string{
			"date",
			"dest_addresses",
			"dest_addresses_count",
			"amount",
			"total_fees",
			"label",
			"lnd_tx_type_label",
			"lnd_short_chan_id",
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
				"dest_addresses",
				"dest_addresses_count",
				"amount",
				"total_fees",
				"label",
				"lnd_tx_type_label",
				"lnd_short_chan_id",
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

	r, total, err := getOnChainTxs(db, filter, sort, limit, offset)
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

// "tx_hash", "raw_tx_hex",
//func getOnChainTxHandler(c *gin.Context, db *sqlx.DB) {
//
//	r, err := getOnChainTxDetails(db, c.Param("identifier"))
//	switch err.(type) {
//	case nil:
//		break
//	case ErrOnChainTxNotFound:
//		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error(), "Identifier": c.Param("identifier")})
//		return
//	default:
//		server_errors.LogAndSendServerError(c, err)
//		return
//	}
//
//	c.JSON(http.StatusOK, r)
//}

func newAddressHandler(c *gin.Context, db *sqlx.DB) {
	var requestBody NewAddressRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("JSON binding the request body")
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	addressType := requestBody.Type
	account := requestBody.Account

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

	//TODO
	//
	resp, err := newAddress(client, addressType, account)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Creating new address")
		return
	}

	newAddressResp := newAddressResponse{
		Address:     resp,
		AddressType: lnrpc.AddressType(addressType).String(),
	}

	c.JSON(http.StatusOK, newAddressResp)
}

func sendCoinsHandler(c *gin.Context, db *sqlx.DB) {
	var requestBody sendCoinsRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		log.Error().Msgf("JSON binding the request body")
		server_errors.WrapLogAndSendServerError(c, err, "JSON binding the request body")
		return
	}

	address := requestBody.Addr
	amount := requestBody.Amount
	targetConf := requestBody.TargetConf
	satPerVbyte := requestBody.SatPerVbyte

	if address == "" {
		log.Error().Msgf("Address value not valid")
		server_errors.LogAndSendServerError(c, errors.New("Invoice value and/or expiry time not valid"))
		return
	}

	if targetConf != 0 && satPerVbyte != nil {
		log.Error().Msgf("Only one fee model accepted")
		server_errors.LogAndSendServerError(c, errors.New("Only one fee model accepted"))
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

	resp, err := sendCoins(client, address, amount, targetConf, satPerVbyte)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Sending on-chain payment")
		return
	}

	sendCoinsResp := sendCoinsResponse{
		Address: resp,
		Amount:  amount,
		Txid:    resp,
	}
	c.JSON(http.StatusOK, sendCoinsResp)
}

func RegisterOnChainTxsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getOnChainTxsHandler(c, db) })
	//r.GET(":identifier", func(c *gin.Context) { getOnChainTxHandler(c, db) })
	r.POST("newaddress", func(c *gin.Context) { newAddressHandler(c, db) })
	r.POST("sendcoins", func(c *gin.Context) { sendCoinsHandler(c, db) })
}
