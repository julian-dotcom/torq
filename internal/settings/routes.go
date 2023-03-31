package settings

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
)

type settings struct {
	SettingsId        int        `json:"settingsId" db:"settings_id"`
	DefaultDateRange  string     `json:"defaultDateRange" db:"default_date_range"`
	DefaultLanguage   string     `json:"defaultLanguage" db:"default_language"`
	PreferredTimezone string     `json:"preferredTimezone" db:"preferred_timezone"`
	WeekStartsOn      string     `json:"weekStartsOn" db:"week_starts_on"`
	TorqUuid          string     `json:"torqUuid" db:"torq_uuid"`
	MixpanelOptOut    bool       `json:"mixpanelOptOut" db:"mixpanel_opt_out"`
	CreatedOn         time.Time  `json:"createdOn" db:"created_on"`
	UpdateOn          *time.Time `json:"updatedOn" db:"updated_on"`
}

type timeZone struct {
	Name string `json:"name" db:"name"`
}

type ConnectionDetails struct {
	NodeId            int
	Name              string
	GRPCAddress       string
	TLSFileBytes      []byte
	MacaroonFileBytes []byte
	Status            commons.Status
	PingSystem        commons.PingSystem
	CustomSettings    commons.NodeConnectionDetailCustomSettings
}

func setAllLndServices(nodeId int,
	lndActive bool,
	customSettings commons.NodeConnectionDetailCustomSettings,
	pingSystem commons.PingSystem) bool {

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	success := commons.InactivateLndService(ctxWithTimeout, nodeId)
	if success && lndActive {
		success = commons.ActivateLndService(ctxWithTimeout, nodeId, customSettings, pingSystem)
	}
	return success
}

func setService(serviceType commons.ServiceType, nodeId int, active bool) bool {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	success := commons.InactivateLndServiceState(ctxWithTimeout, serviceType, nodeId)
	if success && active {
		success = commons.ActivateLndServiceState(ctxWithTimeout, serviceType, nodeId)
	}
	return success
}

func RegisterSettingRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getSettingsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateSettingsHandler(c, db) })
	r.GET("nodeConnectionDetails", func(c *gin.Context) { getAllNodeConnectionDetailsHandler(c, db) })
	r.GET("nodeConnectionDetails/:nodeId", func(c *gin.Context) { getNodeConnectionDetailsHandler(c, db) })
	r.POST("nodeConnectionDetails", func(c *gin.Context) { addNodeConnectionDetailsHandler(c, db) })
	r.PUT("nodeConnectionDetails", func(c *gin.Context) { setNodeConnectionDetailsHandler(c, db) })
	r.PUT("nodeConnectionDetails/:nodeId/:statusId", func(c *gin.Context) {
		setNodeConnectionDetailsStatusHandler(c, db)
	})
	r.PUT("nodePingSystem/:nodeId/:pingSystem/:statusId", func(c *gin.Context) {
		setNodeConnectionDetailsPingSystemHandler(c, db)
	})
	r.PUT("nodeCustomSetting/:nodeId/:customSetting/:statusId", func(c *gin.Context) {
		setNodeConnectionDetailsCustomSettingHandler(c, db)
	})
}
func RegisterUnauthenticatedRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("timezones", func(c *gin.Context) { getTimeZonesHandler(c, db) })
}

func getTimeZonesHandler(c *gin.Context, db *sqlx.DB) {
	timeZones, err := getTimeZones(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, timeZones)
}
func getSettingsHandler(c *gin.Context, db *sqlx.DB) {
	setts, err := getSettings(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, setts)
}

func updateSettingsHandler(c *gin.Context, db *sqlx.DB) {

	var setts settings
	if err := c.BindJSON(&setts); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	err := updateSettings(db, setts)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, setts)
}

func getAllNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	node, err := GetAllNodeConnectionDetails(db, false)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, node)
}

func getNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	ncd, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, ncd)
}

func addNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	var ncd NodeConnectionDetails
	var err error
	existingNcd := NodeConnectionDetails{}

	if err = c.Bind(&ncd); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	ncd, err = fixBindFailures(c, ncd)
	if err != nil {
		server_errors.SendBadRequest(c, err.Error())
		return
	}

	if ncd.TLSFile == nil || ncd.MacaroonFile == nil || ncd.GRPCAddress == nil || *ncd.GRPCAddress == "" {
		server_errors.SendBadRequest(c, "All node details are required to add new node connection details")
		return
	}

	ncd, err = processTLS(ncd)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	if len(ncd.TLSDataBytes) == 0 {
		server_errors.SendBadRequest(c, "Can't check new gRPC details without TLS Cert")
		return
	}

	ncd, err = processMacaroon(ncd)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	if len(ncd.MacaroonDataBytes) == 0 {
		server_errors.SendBadRequest(c, "Can't check new gRPC details without Macaroon File")
		return
	}

	publicKey, chain, network, err := getInformationFromLndNode(*ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
	if err != nil {
		server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get info from LND Node"), "LNDConnect", nil)
		return
	}

	nodeId := commons.GetNodeIdByPublicKey(publicKey, chain, network)
	if nodeId == 0 {
		nodeId, err = AddNodeWhenNew(db, publicKey, chain, network)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Adding node")
			return
		}
	}
	ncd.NodeId = nodeId
	existingNcd, err = getNodeConnectionDetails(db, ncd.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting existing node connection details by nodeId")
		return
	}
	if existingNcd.NodeId == ncd.NodeId && existingNcd.Status != commons.Deleted {
		server_errors.SendUnprocessableEntity(c, "This node already exists")
		return
	}
	if existingNcd.Status == commons.Deleted {
		ncd.CreateOn = existingNcd.CreateOn
		if strings.TrimSpace(ncd.Name) == "" {
			ncd.Name = existingNcd.Name
		}
	}

	canSignMessages, err := getSignMessagesPermissionFromLndNode(*ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Verify sign messages macaroon ability from gRPC")
		return
	}

	if !canSignMessages {
		ncd.PingSystem = 0
	}

	if strings.TrimSpace(ncd.Name) == "" {
		ncd.Name = fmt.Sprintf("Node_%v", ncd.NodeId)
	}
	ncd.Status = commons.Active
	if existingNcd.NodeId == ncd.NodeId {
		ncd, err = SetNodeConnectionDetails(db, ncd)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Updating and reactivating deleted node connection details")
			return
		}
	} else {
		ncd, err = addNodeConnectionDetails(db, ncd)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Adding node connection details")
			return
		}
	}
	commons.SetTorqNode(ncd.NodeId, ncd.Name, ncd.Status, publicKey, chain, network)

	if ncd.Status == commons.Active {
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		success := commons.ActivateLndService(ctxWithTimeout, nodeId, ncd.CustomSettings, ncd.PingSystem)
		if !success {
			server_errors.WrapLogAndSendServerError(c, err, "Service activation failed.")
			return
		}
	}

	c.JSON(http.StatusOK, ncd)
}

func setNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	var ncd NodeConnectionDetails
	var err error
	if err = c.Bind(&ncd); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	ncd, err = fixBindFailures(c, ncd)
	if err != nil {
		server_errors.SendBadRequest(c, err.Error())
		return
	}

	if ncd.NodeId == 0 {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	if ncd.GRPCAddress == nil || *ncd.GRPCAddress == "" {
		c.JSON(http.StatusBadRequest, server_errors.SingleFieldErrorCode("grpcAddress", "missingField", map[string]string{"field": "GRPC Address"}))
		return
	}
	if strings.TrimSpace(ncd.Name) == "" {
		ncd.Name = fmt.Sprintf("Node_%v", ncd.NodeId)
	}

	ncd, err = processTLS(ncd)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Processing TLS file")
		return
	}
	ncd, err = processMacaroon(ncd)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Processing Macaroon file")
		return
	}

	existingNcd, err := getNodeConnectionDetails(db, ncd.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Obtaining existing node connection details")
		return
	}

	lndDetailsUpdate := false
	if len(ncd.MacaroonDataBytes) > 0 || len(ncd.TLSDataBytes) > 0 || *ncd.GRPCAddress != *existingNcd.GRPCAddress {
		lndDetailsUpdate = true
	}

	if len(ncd.TLSDataBytes) == 0 && len(existingNcd.TLSDataBytes) != 0 {
		ncd.TLSDataBytes = existingNcd.TLSDataBytes
		ncd.TLSFileName = existingNcd.TLSFileName
	}
	if len(ncd.TLSDataBytes) == 0 {
		server_errors.LogAndSendServerError(c, errors.New("Can't check new gRPC details without TLS Cert"))
		return
	}

	if len(ncd.MacaroonDataBytes) == 0 && len(existingNcd.MacaroonDataBytes) != 0 {
		ncd.MacaroonDataBytes = existingNcd.MacaroonDataBytes
		ncd.MacaroonFileName = existingNcd.MacaroonFileName
	}
	if len(ncd.MacaroonDataBytes) == 0 {
		server_errors.LogAndSendServerError(c, errors.New("Can't check new gRPC details without Macaroon File"))
		return
	}

	if lndDetailsUpdate {
		publicKey, chain, network, err := getInformationFromLndNode(*ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
		if err != nil {
			server_errors.LogAndSendServerErrorCode(c, errors.Wrap(err, "Get info from LND Node"), "LNDConnect", nil)
			return
		}

		existingPublicKey, existingChain, existingNetwork, err := GetNodeDetailsById(db, ncd.NodeId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Obtaining existing node")
			return
		}
		if existingPublicKey != publicKey || existingChain != chain || existingNetwork != network {
			server_errors.SendUnprocessableEntity(c, "PublicKey/chain/network does not match, create a new node instead of updating this one")
			return
		}

		canSignMessages, err := getSignMessagesPermissionFromLndNode(*ncd.GRPCAddress, ncd.TLSDataBytes, ncd.MacaroonDataBytes)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Verify sign messages macaroon ability from gRPC")
			return
		}

		if !canSignMessages {
			ncd.PingSystem = 0
		}
	}

	nodeSettings := commons.GetNodeSettingsByNodeId(ncd.NodeId)
	if ncd.PingSystem.HasPingSystem(commons.Amboss) &&
		(nodeSettings.Chain != commons.Bitcoin && nodeSettings.Network != commons.MainNet) {
		server_errors.LogAndSendServerError(c, errors.New("Amboss Ping Service is only allowed on Bitcoin Mainnet."))
		return
	}
	if ncd.PingSystem.HasPingSystem(commons.Vector) &&
		(nodeSettings.Chain != commons.Bitcoin && nodeSettings.Network != commons.MainNet) {
		server_errors.LogAndSendServerError(c, errors.New("Vector Ping Service is only allowed on Bitcoin Mainnet."))
		return
	}

	ncd, err = SetNodeConnectionDetails(db, ncd)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Updating connection details")
		return
	}
	commons.SetTorqNode(ncd.NodeId, ncd.Name, ncd.Status,
		nodeSettings.PublicKey, nodeSettings.Chain, nodeSettings.Network)

	success := setAllLndServices(ncd.NodeId, ncd.Status == commons.Active, ncd.CustomSettings, ncd.PingSystem)
	if !success {
		server_errors.LogAndSendServerError(c, errors.New("Some services did not start correctly"))
		return
	}

	c.JSON(http.StatusOK, ncd)
}

func fixBindFailures(c *gin.Context, ncd NodeConnectionDetails) (NodeConnectionDetails, error) {
	// TODO c.Bind cannot process status?
	statusId, err := strconv.Atoi(c.Request.Form.Get("status"))
	if err != nil {
		return NodeConnectionDetails{}, errors.New("Failed to find/parse status in the request.")
	}
	if statusId > int(commons.Archived) {
		return NodeConnectionDetails{}, errors.New("Failed to parse status in the request.")
	}
	ncd.Status = commons.Status(statusId)

	// TODO c.Bind cannot process pingSystem?
	pingSystem, err := strconv.Atoi(c.Request.Form.Get("pingSystem"))
	if err != nil {
		return NodeConnectionDetails{}, errors.New("Failed to find/parse pingSystem in the request.")
	}
	if pingSystem > commons.PingSystemMax {
		return NodeConnectionDetails{}, errors.New("Failed to parse pingSystem in the request.")
	}
	ncd.PingSystem = commons.PingSystem(pingSystem)

	// TODO c.Bind cannot process customSettings?
	customSettings, err := strconv.Atoi(c.Request.Form.Get("customSettings"))
	if err != nil {
		return NodeConnectionDetails{}, errors.New("Failed to find/parse customSettings in the request.")
	}
	if customSettings > commons.NodeConnectionDetailCustomSettingsMax {
		return NodeConnectionDetails{}, errors.New("Failed to parse customSettings in the request.")
	}
	ncd.CustomSettings = commons.NodeConnectionDetailCustomSettings(customSettings)
	return ncd, nil
}

func setNodeConnectionDetailsStatusHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	statusId, err := strconv.Atoi(c.Param("statusId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}

	_, err = setNodeConnectionDetailsStatus(db, nodeId, commons.Status(statusId))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	switch commons.Status(statusId) {
	case commons.Deleted:
		node := commons.GetNodeSettingsByNodeId(nodeId)
		commons.RemoveManagedNodeFromCache(node)
		commons.RemoveManagedChannelStatesFromCache(node.NodeId)
	default:
		nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
		var name string
		if nodeSettings.Name != nil {
			name = *nodeSettings.Name
		}
		commons.SetTorqNode(nodeId, name, commons.Status(statusId),
			nodeSettings.PublicKey, nodeSettings.Chain, nodeSettings.Network)
	}

	if commons.Status(statusId) != commons.Active {
		success := setAllLndServices(nodeId, false, 0, 0)
		if !success {
			server_errors.LogAndSendServerError(c, errors.New("Some services did not start correctly"))
			return
		}
		c.Status(http.StatusOK)
		return
	}

	ncd, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	success := setAllLndServices(nodeId, true, ncd.CustomSettings, ncd.PingSystem)
	if !success {
		server_errors.LogAndSendServerError(c, errors.New("Some services did not start correctly"))
		return
	}
	c.Status(http.StatusOK)
}

func setNodeConnectionDetailsPingSystemHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	pingSystem, err := strconv.Atoi(c.Param("pingSystem"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse pingSystem in the request.")
		return
	}
	if pingSystem > commons.PingSystemMax {
		server_errors.SendBadRequest(c, "Failed to parse pingSystem in the request.")
		return
	}
	statusId, err := strconv.Atoi(c.Param("statusId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}
	nodeSettings := commons.GetNodeSettingsByNodeId(nodeId)
	if commons.Status(statusId) == commons.Active &&
		(nodeSettings.Chain != commons.Bitcoin || nodeSettings.Network != commons.MainNet) {
		server_errors.SendBadRequest(c, "Ping Services are only allowed on Bitcoin Mainnet.")
		return
	}

	var serviceType commons.ServiceType
	switch commons.PingSystem(pingSystem) {
	case commons.Amboss:
		serviceType = commons.AmbossService
	case commons.Vector:
		serviceType = commons.VectorService
	}

	_, err = setNodeConnectionDetailsPingSystemStatus(db, nodeId, commons.PingSystem(pingSystem), commons.Status(statusId))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	done := setService(serviceType, nodeId, commons.Status(statusId) == commons.Active)
	if !done {
		server_errors.LogAndSendServerError(c, errors.New("Service failed please try again."))
		return
	}

	c.Status(http.StatusOK)
}

func setNodeConnectionDetailsCustomSettingHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	customSetting, err := strconv.Atoi(c.Param("customSetting"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse customSetting in the request.")
		return
	}
	if customSetting > commons.NodeConnectionDetailCustomSettingsMax {
		server_errors.SendBadRequest(c, "Failed to parse customSetting in the request.")
		return
	}
	statusId, err := strconv.Atoi(c.Param("statusId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}

	var serviceType commons.ServiceType
	switch commons.NodeConnectionDetailCustomSettings(customSetting) {
	case commons.ImportFailedPayments, commons.ImportPayments:
		serviceType = commons.LndServicePaymentStream
	case commons.ImportHtlcEvents:
		serviceType = commons.LndServiceHtlcEventStream
	case commons.ImportPeerEvents:
		serviceType = commons.LndServicePeerEventStream
	case commons.ImportTransactions:
		serviceType = commons.LndServiceTransactionStream
	case commons.ImportInvoices:
		serviceType = commons.LndServiceInvoiceStream
	case commons.ImportForwards, commons.ImportHistoricForwards:
		serviceType = commons.LndServiceForwardStream
	}

	_, err = setNodeConnectionDetailsCustomSettingStatus(db, nodeId,
		commons.NodeConnectionDetailCustomSettings(customSetting), commons.Status(statusId))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	done := setService(serviceType, nodeId, commons.Status(statusId) == commons.Active)
	if !done {
		server_errors.LogAndSendServerError(c, errors.New("Service failed please try again."))
		return
	}

	c.Status(http.StatusOK)
}

// GetConnectionDetailsById will still fetch details even if node is disabled or deleted
func GetConnectionDetailsById(db *sqlx.DB, nodeId int) (ConnectionDetails, error) {
	ncd, err := getNodeConnectionDetails(db, nodeId)
	if err != nil {
		return ConnectionDetails{}, errors.Wrapf(err, "Getting node connection details from db for nodeId: %v", nodeId)
	}
	cd := ConnectionDetails{
		NodeId:            ncd.NodeId,
		TLSFileBytes:      ncd.TLSDataBytes,
		MacaroonFileBytes: ncd.MacaroonDataBytes,
		Name:              ncd.Name,
		Status:            ncd.Status,
		PingSystem:        ncd.PingSystem,
		CustomSettings:    ncd.CustomSettings,
	}
	if ncd.GRPCAddress != nil {
		cd.GRPCAddress = *ncd.GRPCAddress
	}
	return cd, nil
}

func getInformationFromLndNode(grpcAddress string, tlsCert []byte, macaroonFile []byte) (
	string, commons.Chain, commons.Network, error) {
	conn, err := lnd_connect.Connect(grpcAddress, tlsCert, macaroonFile)
	if err != nil {
		return "", 0, 0, errors.Wrap(err,
			"Can't connect to node to verify public key, check all details including TLS Cert and Macaroon")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to close gRPC connection.")
		}
	}(conn)

	client := lnrpc.NewLightningClient(conn)
	ctx := context.Background()
	info, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return "", 0, 0, errors.Wrap(err, "Obtaining information from LND")
	}
	if len(info.Chains) != 1 {
		return "", 0, 0, errors.Wrapf(err, "Obtaining chains from LND %v", info.Chains)
	}

	var chain commons.Chain
	switch info.Chains[0].Chain {
	case "bitcoin":
		chain = commons.Bitcoin
	case "litecoin":
		chain = commons.Litecoin
	default:
		return "", 0, 0, errors.Wrapf(err, "Obtaining chain from LND %v", info.Chains[0].Chain)
	}

	var network commons.Network
	switch info.Chains[0].Network {
	case "mainnet":
		network = commons.MainNet
	case "testnet":
		network = commons.TestNet
	case "signet":
		network = commons.SigNet
	case "simnet":
		network = commons.SimNet
	case "regtest":
		network = commons.RegTest
	default:
		return "", 0, 0, errors.Wrapf(err, "Obtaining network from LND %v", info.Chains[0].Network)
	}
	return info.IdentityPubkey, chain, network, nil
}

func getSignMessagesPermissionFromLndNode(grpcAddress string, tlsCert []byte, macaroonFile []byte) (bool, error) {
	conn, err := lnd_connect.Connect(grpcAddress, tlsCert, macaroonFile)
	if err != nil {
		return false, errors.Wrap(err,
			"Can't connect to node to verify public key, check all details including TLS Cert and Macaroon")
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to close gRPC connection.")
		}
	}(conn)

	client := lnrpc.NewLightningClient(conn)
	ctx := context.Background()

	signMsgReq := lnrpc.SignMessageRequest{
		Msg: []byte("test"),
	}
	signMsgResp, err := client.SignMessage(ctx, &signMsgReq)
	if err != nil {
		return false, nil
	}
	return signMsgResp.Signature != "", nil
}

func processTLS(ncd NodeConnectionDetails) (NodeConnectionDetails, error) {
	if ncd.TLSFile != nil {
		ncd.TLSFileName = &ncd.TLSFile.Filename
		tlsDataFile, err := ncd.TLSFile.Open()
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Opening TLS file")
		}
		tlsData, err := io.ReadAll(tlsDataFile)
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Reading TLS file")
		}
		ncd.TLSDataBytes = tlsData
	}
	return ncd, nil
}

func processMacaroon(ncd NodeConnectionDetails) (NodeConnectionDetails, error) {
	if ncd.MacaroonFile != nil {
		ncd.MacaroonFileName = &ncd.MacaroonFile.Filename
		macaroonDataFile, err := ncd.MacaroonFile.Open()
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Opening macaroon file")
		}
		macaroonData, err := io.ReadAll(macaroonDataFile)
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Reading macaroon file")
		}
		ncd.MacaroonDataBytes = macaroonData
	}
	return ncd, nil
}
