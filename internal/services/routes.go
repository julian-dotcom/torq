package services

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterUnauthenticatedRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("status", func(c *gin.Context) { getServicesHandler(c, db) })
	r.GET("/:nodeId/lndStatus", func(c *gin.Context) { getLndServicesHandler(c, db) })
}

func getServicesHandler(c *gin.Context, db *sqlx.DB) {
	result := Services{}
	var bitcoinNetworks []commons.Network
	for _, torqServiceType := range commons.GetTorqServiceTypes() {
		torqService := commons.GetCurrentTorqServiceState(torqServiceType)
		if torqServiceType == commons.TorqService {
			result.MainService = TorqService{
				CommonService: CommonService{
					ServiceType:       torqServiceType,
					ServiceTypeString: torqServiceType.String(),
					Status:            torqService.Status,
					BootTime:          torqService.BootTime,
					StatusString:      torqService.Status.String(),
				},
			}
			continue
		}
		result.TorqServices = append(result.TorqServices, TorqService{
			CommonService: CommonService{
				ServiceType:       torqServiceType,
				ServiceTypeString: torqServiceType.String(),
				Status:            torqService.Status,
				BootTime:          torqService.BootTime,
				StatusString:      torqService.Status.String(),
			},
		})
	}
	for _, lndNodeId := range commons.GetLndNodeIds() {
		bitcoinNetwork := commons.GetNodeSettingsByNodeId(lndNodeId).Network
		for _, lndServiceType := range commons.GetTorqServiceTypes() {
			lndService := commons.GetCurrentLndServiceState(lndServiceType, lndNodeId)
			result.LndServices = append(result.LndServices, LndService{
				CommonService: CommonService{
					ServiceType:       lndServiceType,
					ServiceTypeString: lndServiceType.String(),
					Status:            lndService.Status,
					BootTime:          lndService.BootTime,
					StatusString:      lndService.Status.String(),
				},
				NodeId:         lndNodeId,
				BitcoinNetwork: bitcoinNetwork,
			})
		}
	}
	result.BitcoinNetworks = bitcoinNetworks
	result.Version = build.ExtendedVersion()
	c.JSON(http.StatusOK, result)
}

func getLndServicesHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}

	var lndServices []LndService
	bitcoinNetwork := commons.GetNodeSettingsByNodeId(nodeId).Network
	for _, lndServiceType := range commons.GetTorqServiceTypes() {
		lndService := commons.GetCurrentLndServiceState(lndServiceType, nodeId)
		lndServices = append(lndServices, LndService{
			CommonService: CommonService{
				ServiceType:       lndServiceType,
				ServiceTypeString: lndServiceType.String(),
				Status:            lndService.Status,
				BootTime:          lndService.BootTime,
				StatusString:      lndService.Status.String(),
			},
			NodeId:         nodeId,
			BitcoinNetwork: bitcoinNetwork,
		})
	}
	c.JSON(http.StatusOK, lndServices)
}
