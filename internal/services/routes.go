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
		desiredState := commons.GetDesiredTorqServiceState(torqServiceType)
		if desiredState.Status != torqService.Status {
			result.ServiceMismatches = append(result.ServiceMismatches, ServiceMismatch{
				ServiceType:         torqServiceType,
				ServiceTypeString:   torqServiceType.String(),
				Status:              torqService.Status,
				StatusString:        torqService.Status.String(),
				DesiredStatus:       desiredState.Status,
				DesiredStatusString: desiredState.Status.String(),
				FailureTime:         commons.GetTorqFailedAttemptTime(torqServiceType),
			})
		}
		if torqServiceType == commons.TorqService {
			result.MainService = TorqService{
				CommonService: CommonService{
					ServiceType:       torqServiceType,
					ServiceTypeString: torqServiceType.String(),
					Status:            torqService.Status,
					BootTime:          torqService.ActiveTime,
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
				BootTime:          torqService.ActiveTime,
				StatusString:      torqService.Status.String(),
			},
		})
	}
	for _, lndNodeId := range commons.GetLndNodeIds() {
		nodeId := lndNodeId
		bitcoinNetwork := commons.GetNodeSettingsByNodeId(nodeId).Network
		for _, lndServiceType := range commons.GetLndServiceTypes() {
			lndService := commons.GetCurrentLndServiceState(lndServiceType, nodeId)
			desiredState := commons.GetDesiredLndServiceState(lndServiceType, nodeId)
			if desiredState.Status != lndService.Status {
				result.ServiceMismatches = append(result.ServiceMismatches, ServiceMismatch{
					ServiceType:         lndServiceType,
					ServiceTypeString:   lndServiceType.String(),
					Status:              lndService.Status,
					StatusString:        lndService.Status.String(),
					DesiredStatus:       desiredState.Status,
					DesiredStatusString: desiredState.Status.String(),
					NodeId:              &nodeId,
					BitcoinNetwork:      &bitcoinNetwork,
					FailureTime:         commons.GetLndFailedAttemptTime(lndServiceType, nodeId),
				})
			}
			result.LndServices = append(result.LndServices, LndService{
				CommonService: CommonService{
					ServiceType:       lndServiceType,
					ServiceTypeString: lndServiceType.String(),
					Status:            lndService.Status,
					BootTime:          lndService.ActiveTime,
					StatusString:      lndService.Status.String(),
				},
				NodeId:         nodeId,
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
				BootTime:          lndService.ActiveTime,
				StatusString:      lndService.Status.String(),
			},
			NodeId:         nodeId,
			BitcoinNetwork: bitcoinNetwork,
		})
	}
	c.JSON(http.StatusOK, lndServices)
}
