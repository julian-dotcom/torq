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
}

func RegisterSettingRoutes(r *gin.RouterGroup, db *sqlx.DB, restartLNDSub func() error) {
	r.GET("", func(c *gin.Context) { getSettingsHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateSettingsHandler(c, db) })
	r.GET("nodeConnectionDetails", func(c *gin.Context) { getAllNodeConnectionDetailsHandler(c, db) })
	r.GET("nodeConnectionDetails/:nodeId", func(c *gin.Context) { getNodeConnectionDetailsHandler(c, db) })
	r.POST("nodeConnectionDetails", func(c *gin.Context) { addNodeConnectionDetailsHandler(c, db, restartLNDSub) })
	r.PUT("nodeConnectionDetails", func(c *gin.Context) { setNodeConnectionDetailsHandler(c, db, restartLNDSub) })
	r.PUT("nodeConnectionDetails/:nodeId/:status", func(c *gin.Context) { setNodeConnectionDetailsStatusHandler(c, db, restartLNDSub) })
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
	localNode, err := getAllNodeConnectionDetails(db, false)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, localNode)
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

func addNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB, restartLNDSub func() error) {
	var ncd nodeConnectionDetails

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
		server_errors.SendBadRequest(c, "Can't check new GRPC details without TLS Cert")
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
		server_errors.SendBadRequest(c, "Can't check new GRPC details without Macaroon File")
		return
	}

	publicKey, chain, network, err := getInformationFromLndNode(*ncd.GRPCAddress, tlsCert, macaroonFile)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting public key from node")
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

	nodeId := commons.GetNodeIdFromPublicKey(publicKey, chain, network)
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
	ncd, err = addNodeConnectionDetails(db, ncd)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding node connection details")
		return
	}
	if strings.TrimSpace(ncd.Name) == "" {
		ncd.Name = fmt.Sprintf("Node_%v", ncd.NodeId)
		err := setNodeConnectionDetailsName(db, ncd.NodeId, ncd.Name)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	}

	go func() {
		if err := restartLNDSub(); err != nil {
			log.Warn().Msg("Already restarting subscriptions, discarding restart request")
		}
	}()

	c.JSON(http.StatusOK, ncd)
}

func setNodeConnectionDetailsHandler(c *gin.Context, db *sqlx.DB, restartLNDSub func() error) {
	var ncd nodeConnectionDetails
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

	// if GRPC details have changed we need to check that the public keys (if existing) matches
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
			server_errors.LogAndSendServerError(c, errors.New("Can't check new GRPC details without TLS Cert"))
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
			server_errors.LogAndSendServerError(c, errors.New("Can't check new GRPC details without Macaroon File"))
			return
		}

		publicKey, chain, network, err := getInformationFromLndNode(*ncd.GRPCAddress, tlsCert, macaroonFile)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err, "Obtaining publicKey/chain/network from grpc")
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
	ncd, err = setNodeConnectionDetails(db, ncd)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Opening Macaroon file")
		return
	}

	if restartLND(c, restartLNDSub) {
		return
	}

	c.JSON(http.StatusOK, ncd)
}

func setNodeConnectionDetailsStatusHandler(c *gin.Context, db *sqlx.DB, restartLNDSub func() error) {
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

	err = setNodeConnectionDetailsStatus(db, nodeId, commons.Status(statusId))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	if restartLND(c, restartLNDSub) {
		return
	}

	c.Status(http.StatusOK)
}

func GetActiveNodesConnectionDetails(db *sqlx.DB) ([]ConnectionDetails, error) {
	activeNcds, err := getNodeConnectionDetailsByStatus(db, commons.Active)
	if err != nil {
		return []ConnectionDetails{}, errors.Wrap(err, "Getting active node connection details from db")
	}

	var activeNodes []ConnectionDetails
	for _, ncd := range activeNcds {
		if ncd.GRPCAddress == nil || ncd.TLSDataBytes == nil || ncd.MacaroonDataBytes == nil {
			continue
		}
		activeNodes = append(activeNodes, ConnectionDetails{
			NodeId:            ncd.NodeId,
			GRPCAddress:       *ncd.GRPCAddress,
			TLSFileBytes:      ncd.TLSDataBytes,
			MacaroonFileBytes: ncd.MacaroonDataBytes,
			Name:              ncd.Name,
		})
	}

	return activeNodes, nil
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
			log.Debug().Err(err).Msg("Failed to close grpc connection.")
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
		network = commons.MainNet
	case "signet":
		network = commons.MainNet
	case "regtest":
		network = commons.MainNet
	default:
		return "", 0, 0, errors.Wrapf(err, "Obtaining network from LND %v", info.Chains[0].Network)
	}
	return info.IdentityPubkey, chain, network, nil
}

func processTLS(ncd nodeConnectionDetails) (nodeConnectionDetails, error) {
	if ncd.TLSFile != nil {
		ncd.TLSFileName = &ncd.TLSFile.Filename
		tlsDataFile, err := ncd.TLSFile.Open()
		if err != nil {
			return ncd, err
		}
		tlsData, err := io.ReadAll(tlsDataFile)
		if err != nil {
			return ncd, err
		}
		ncd.TLSDataBytes = tlsData
	}
	return ncd, nil
}

func processMacaroon(ncd nodeConnectionDetails) (nodeConnectionDetails, error) {
	if ncd.MacaroonFile != nil {
		ncd.MacaroonFileName = &ncd.MacaroonFile.Filename
		macaroonDataFile, err := ncd.MacaroonFile.Open()
		if err != nil {
			return ncd, err
		}
		macaroonData, err := io.ReadAll(macaroonDataFile)
		if err != nil {
			return ncd, err
		}
		ncd.MacaroonDataBytes = macaroonData
	}
	return ncd, nil
}

func restartLND(c *gin.Context, restartLNDSub func() error) bool {
	maxTries := 30
	attempts := 0
	for {
		attempts++
		if attempts > maxTries {
			server_errors.LogAndSendServerError(c, errors.New("Failed to restart node subscriptions"))
			return true
		}
		if err := restartLNDSub(); err != nil {
			log.Warn().Msg("Already restarting subscriptions, retrying")
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	return false
}
