package settings

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/server_errors"
)

type settings struct {
	DefaultDateRange  string `json:"defaultDateRange" db:"default_date_range"`
	DefaultLanguage   string `json:"defaultLanguage" db:"default_language"`
	PreferredTimezone string `json:"preferredTimezone" db:"preferred_timezone"`
	WeekStartsOn      string `json:"weekStartsOn" db:"week_starts_on"`
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

func (connectionDetails *ConnectionDetails) AddPingSystem(pingSystem commons.PingSystem) {
	connectionDetails.PingSystem |= pingSystem
}
func (connectionDetails *ConnectionDetails) HasPingSystem(pingSystem commons.PingSystem) bool {
	return connectionDetails.PingSystem&pingSystem != 0
}
func (connectionDetails *ConnectionDetails) RemovePingSystem(pingSystem commons.PingSystem) {
	connectionDetails.PingSystem &= ^pingSystem
}

func (connectionDetails *ConnectionDetails) AddNodeConnectionDetailCustomSettings(customSettings commons.NodeConnectionDetailCustomSettings) {
	connectionDetails.CustomSettings |= customSettings
}
func (connectionDetails *ConnectionDetails) HasNodeConnectionDetailCustomSettings(customSettings commons.NodeConnectionDetailCustomSettings) bool {
	return connectionDetails.CustomSettings&customSettings != 0
}
func (connectionDetails *ConnectionDetails) RemoveNodeConnectionDetailCustomSettings(customSettings commons.NodeConnectionDetailCustomSettings) {
	connectionDetails.CustomSettings &= ^customSettings
}

func startServiceOrRestartWhenRunning(serviceChannel chan commons.ServiceChannelMessage,
	serviceType commons.ServiceType, nodeId int, active bool) bool {
	if active {
		enforcedServiceStatus := commons.Active
		resultChannel := make(chan commons.Status)
		serviceChannel <- commons.ServiceChannelMessage{
			NodeId:                nodeId,
			ServiceType:           serviceType,
			EnforcedServiceStatus: &enforcedServiceStatus,
			ServiceCommand:        commons.Kill,
			NoDelay:               true,
			Out:                   resultChannel,
		}
		switch <-resultChannel {
		case commons.Active:
			// THE RUNNING SERVICE WAS KILLED EnforcedServiceStatus is ACTIVE (subscription will attempt to start)
		case commons.Pending:
			// THE SERVICE FAILED TO BE KILLED BECAUSE OF A BOOT ATTEMPT THAT IS LOCKING THE SERVICE
			return false
		case commons.Inactive:
			// THE SERVICE WAS NOT RUNNING
			serviceChannel <- commons.ServiceChannelMessage{
				NodeId:                nodeId,
				ServiceType:           serviceType,
				EnforcedServiceStatus: &enforcedServiceStatus,
				ServiceCommand:        commons.Boot,
			}
		}
	} else {
		enforcedServiceStatus := commons.Inactive
		resultChannel := make(chan commons.Status)
		serviceChannel <- commons.ServiceChannelMessage{
			NodeId:                nodeId,
			ServiceType:           serviceType,
			EnforcedServiceStatus: &enforcedServiceStatus,
			ServiceCommand:        commons.Kill,
			NoDelay:               true,
			Out:                   resultChannel,
		}
		switch <-resultChannel {
		case commons.Active:
			// THE RUNNING SERVICE WAS KILLED AND EnforcedServiceStatus is INACTIVE (subscription will stay down)
		case commons.Pending:
			// THE SERVICE FAILED TO BE KILLED BECAUSE OF A BOOT ATTEMPT THAT IS LOCKING THE SERVICE
			return false
		case commons.Inactive:
			// THE SERVICE WAS NOT RUNNING
		}
	}
	return true
}

func RegisterSettingRoutes(r *gin.RouterGroup, db *sqlx.DB, serviceChannel chan commons.ServiceChannelMessage) {
	r.GET("", func(c *gin.Context) { getSettingsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateSettingsHandler(c, db) })
	r.GET("nodeConnectionDetails", func(c *gin.Context) { getAllNodeConnectionDetailsHandler(c, db) })
	r.GET("nodeConnectionDetails/:nodeId", func(c *gin.Context) { getNodeConnectionDetailsHandler(c, db) })
	r.POST("nodeConnectionDetails", func(c *gin.Context) { addNodeConnectionDetailsHandler(c, db, serviceChannel) })
	r.PUT("nodeConnectionDetails", func(c *gin.Context) { setNodeConnectionDetailsHandler(c, db, serviceChannel) })
	r.PUT("nodeConnectionDetails/:nodeId/:statusId", func(c *gin.Context) { setNodeConnectionDetailsStatusHandler(c, db, serviceChannel) })
	r.PUT("nodePingSystem/:nodeId/:pingSystem/:statusId", func(c *gin.Context) { setNodeConnectionDetailsPingSystemHandler(c, db, serviceChannel) })
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
	settings, err := getSettings(db)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func updateSettingsHandler(c *gin.Context, db *sqlx.DB) {
	var settings settings
	if err := c.BindJSON(&settings); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	err := updateSettings(db, settings)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func getAllNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB) {
	node, err := getAllNodeConnectionDetails(db, false)
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

func addNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB,
	serviceChannel chan commons.ServiceChannelMessage) {

	var ncd NodeConnectionDetails

	if err := c.Bind(&ncd); err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	if ncd.TLSFile == nil || ncd.MacaroonFile == nil || ncd.GRPCAddress == nil || *ncd.GRPCAddress == "" {
		server_errors.SendBadRequest(c, "All node details are required to add new node connection details")
		return
	}
	tlsDataFile, err := ncd.TLSFile.Open()
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	tlsCert, err := io.ReadAll(tlsDataFile)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	if len(tlsCert) == 0 {
		server_errors.SendBadRequest(c, "Can't check new gRPC details without TLS Cert")
		return
	}

	macaroonDataFile, err := ncd.MacaroonFile.Open()
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	macaroonFile, err := io.ReadAll(macaroonDataFile)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	if len(macaroonFile) == 0 {
		server_errors.SendBadRequest(c, "Can't check new gRPC details without Macaroon File")
		return
	}

	publicKey, chain, network, err := getInformationFromLndNode(*ncd.GRPCAddress, tlsCert, macaroonFile)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Obtaining publicKey/chain/network from gRPC (gRPC connection fails)")
		return
	}
	node, err := nodes.GetNodeByPublicKey(db, publicKey)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting existing node by public key")
		return
	}
	if node.PublicKey == publicKey && node.Chain == chain && node.Network == network {
		server_errors.SendUnprocessableEntity(c, "This node already exists")
		return
	}

	ncd, err = processTLS(ncd)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	ncd, err = processMacaroon(ncd)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	nodeId := commons.GetNodeIdByPublicKey(publicKey, chain, network)
	if nodeId == 0 {
		newNode := nodes.Node{
			PublicKey: publicKey,
			Network:   network,
			Chain:     chain,
		}
		nodeId, err = nodes.AddNodeWhenNew(db, newNode)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Adding node")
		}
	}
	ncd.NodeId = nodeId
	ncd.Status = commons.Active
	if strings.TrimSpace(ncd.Name) == "" {
		ncd.Name = fmt.Sprintf("Node_%v", ncd.NodeId)
	}
	ncd, err = addNodeConnectionDetails(db, ncd)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding node connection details")
		return
	}
	commons.SetTorqNode(nodeId, ncd.Status, publicKey, chain, network)

	if ncd.Status == commons.Active {
		serviceChannel <- commons.ServiceChannelMessage{
			NodeId:         nodeId,
			ServiceType:    commons.LndService,
			ServiceCommand: commons.Boot,
		}
	}

	c.JSON(http.StatusOK, ncd)
}

func setNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB,
	serviceChannel chan commons.ServiceChannelMessage) {

	var ncd NodeConnectionDetails
	if err := c.Bind(&ncd); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	if ncd.NodeId == 0 {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}
	if strings.TrimSpace(ncd.Name) == "" {
		ncd.Name = fmt.Sprintf("Node_%v", ncd.NodeId)
	}

	existingNcd, err := getNodeConnectionDetails(db, ncd.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Obtaining existing node connection details")
		return
	}
	existingNode, err := nodes.GetNodeById(db, ncd.NodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Obtaining existing node")
		return
	}

	if existingNcd.NodeId != 0 {
		ncd.MacaroonDataBytes = existingNcd.MacaroonDataBytes
		ncd.MacaroonFileName = existingNcd.MacaroonFileName
		ncd.TLSDataBytes = existingNcd.TLSDataBytes
		ncd.TLSFileName = existingNcd.TLSFileName
	}

	// if gRPC details have changed we need to check that the public keys (if existing) matches
	if existingNcd.GRPCAddress != ncd.GRPCAddress {
		var tlsCert []byte
		if ncd.TLSFile != nil {
			tlsDataFile, err := ncd.TLSFile.Open()
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Opening TLS file")
				return
			}
			tlsData, err := io.ReadAll(tlsDataFile)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Reading TLS file")
				return
			}
			tlsCert = tlsData
		}
		if len(tlsCert) == 0 && len(existingNcd.TLSDataBytes) != 0 {
			tlsCert = existingNcd.TLSDataBytes
		}
		if len(tlsCert) == 0 {
			server_errors.LogAndSendServerError(c, errors.New("Can't check new gRPC details without TLS Cert"))
			return
		}

		var macaroonFile []byte
		if ncd.MacaroonFile != nil {
			macaroonDataFile, err := ncd.MacaroonFile.Open()
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Opening Macaroon file")
				return
			}
			macaroonData, err := io.ReadAll(macaroonDataFile)
			if err != nil {
				server_errors.WrapLogAndSendServerError(c, err, "Reading Macaroon file")
				return
			}
			macaroonFile = macaroonData
		}
		if len(macaroonFile) == 0 && len(existingNcd.MacaroonDataBytes) != 0 {
			macaroonFile = existingNcd.MacaroonDataBytes
		}
		if len(macaroonFile) == 0 {
			server_errors.LogAndSendServerError(c, errors.New("Can't check new gRPC details without Macaroon File"))
			return
		}

		publicKey, chain, network, err := getInformationFromLndNode(*ncd.GRPCAddress, tlsCert, macaroonFile)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Obtaining publicKey/chain/network from gRPC (gRPC connection fails)")
			return
		}

		if existingNode.PublicKey != publicKey || existingNode.Chain != chain || existingNode.Network != network {
			server_errors.SendUnprocessableEntity(c, "PublicKey/chain/network does not match, create a new node instead of updating this one")
			return
		}
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

	nodeSettings := commons.GetNodeSettingsByNodeId(ncd.NodeId)
	if ncd.HasNotificationType(commons.Amboss) &&
		(nodeSettings.Chain != commons.Bitcoin && nodeSettings.Network != commons.MainNet) {
		server_errors.LogAndSendServerError(c, errors.New("Amboss Ping Service is only allowed on Bitcoin Mainnet."))
		return
	}
	if ncd.HasNotificationType(commons.Vector) &&
		(nodeSettings.Chain != commons.Bitcoin && nodeSettings.Network != commons.MainNet) {
		server_errors.LogAndSendServerError(c, errors.New("Vector Ping Service is only allowed on Bitcoin Mainnet."))
		return
	}
	commons.RunningServices[commons.LndService].SetIncludeIncomplete(ncd.NodeId, ncd.HasNodeConnectionDetailCustomSettings(commons.ImportFailedPayments))

	lndDone := startServiceOrRestartWhenRunning(serviceChannel, commons.LndService, ncd.NodeId, ncd.Status == commons.Active)
	ambossDone := startServiceOrRestartWhenRunning(serviceChannel, commons.AmbossService, ncd.NodeId, ncd.HasNotificationType(commons.Amboss))
	vectorDone := startServiceOrRestartWhenRunning(serviceChannel, commons.VectorService, ncd.NodeId, ncd.HasNotificationType(commons.Vector))
	if lndDone && ambossDone && vectorDone {
		ncd, err = SetNodeConnectionDetails(db, ncd)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Opening Macaroon file")
			return
		}
	} else {
		server_errors.LogAndSendServerError(c, errors.New("Service could not be stopped please try again."))
		return
	}

	c.JSON(http.StatusOK, ncd)
}

func setNodeConnectionDetailsStatusHandler(c *gin.Context, db *sqlx.DB,
	serviceChannel chan commons.ServiceChannelMessage) {

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

	done := startServiceOrRestartWhenRunning(serviceChannel, commons.LndService, nodeId, commons.Status(statusId) == commons.Active)
	if done {
		_, err := setNodeConnectionDetailsStatus(db, nodeId, commons.Status(statusId))
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	} else {
		server_errors.LogAndSendServerError(c, errors.New("Service could not be stopped please try again."))
		return
	}

	c.Status(http.StatusOK)
}

func setNodeConnectionDetailsPingSystemHandler(c *gin.Context, db *sqlx.DB,
	serviceChannel chan commons.ServiceChannelMessage) {

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

	var subscription commons.ServiceType
	if commons.PingSystem(pingSystem) == commons.Amboss {
		subscription = commons.AmbossService
	}
	if commons.PingSystem(pingSystem) == commons.Vector {
		subscription = commons.VectorService
	}

	done := startServiceOrRestartWhenRunning(serviceChannel, subscription, nodeId, commons.Status(statusId) == commons.Active)
	if done {
		_, err := setNodeConnectionDetailsPingSystemStatus(db, nodeId, commons.PingSystem(pingSystem), commons.Status(statusId))
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	} else {
		server_errors.LogAndSendServerError(c, errors.New("Service could not be stopped please try again."))
		return
	}

	c.Status(http.StatusOK)
}

func GetActiveNodesConnectionDetails(db *sqlx.DB) ([]ConnectionDetails, error) {
	activeNcds, err := getNodeConnectionDetailsByStatus(db, commons.Active)
	if err != nil {
		return []ConnectionDetails{}, errors.Wrap(err, "Getting active node connection details from db")
	}
	return processConnectionDetails(activeNcds), nil
}

func GetAmbossPingNodesConnectionDetails(db *sqlx.DB) ([]ConnectionDetails, error) {
	ncds, err := getPingConnectionDetails(db, commons.Amboss)
	if err != nil {
		return nil, errors.Wrap(err, "Getting node connection details for Amboss from db")
	}
	return processConnectionDetails(ncds), nil
}

func GetVectorPingNodesConnectionDetails(db *sqlx.DB) ([]ConnectionDetails, error) {
	ncds, err := getPingConnectionDetails(db, commons.Vector)
	if err != nil {
		return nil, errors.Wrap(err, "Getting node connection details for Vector from db")
	}
	return processConnectionDetails(ncds), nil
}

func processConnectionDetails(ncds []NodeConnectionDetails) []ConnectionDetails {
	var processedNodes []ConnectionDetails
	for _, ncd := range ncds {
		if ncd.GRPCAddress == nil || ncd.TLSDataBytes == nil || ncd.MacaroonDataBytes == nil {
			continue
		}
		processedNodes = append(processedNodes, ConnectionDetails{
			NodeId:            ncd.NodeId,
			GRPCAddress:       *ncd.GRPCAddress,
			TLSFileBytes:      ncd.TLSDataBytes,
			MacaroonFileBytes: ncd.MacaroonDataBytes,
			Name:              ncd.Name,
			PingSystem:        ncd.PingSystem,
			CustomSettings:    ncd.CustomSettings,
		})
	}
	return processedNodes
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

func processTLS(ncd NodeConnectionDetails) (NodeConnectionDetails, error) {
	if ncd.TLSFile != nil {
		ncd.TLSFileName = &ncd.TLSFile.Filename
		tlsDataFile, err := ncd.TLSFile.Open()
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Opening TLS file")
		}
		tlsData, err := io.ReadAll(tlsDataFile)
		if err != nil {
			return NodeConnectionDetails{}, errors.Wrap(err, "Reasing TLS file")
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
