package services

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterUnauthenticatedRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("status", func(c *gin.Context) { getServicesHandler(c, db) })
	r.GET("/:nodeId/lndStatus", func(c *gin.Context) { getLndServicesHandler(c, db) })
}

func getServicesHandler(c *gin.Context, db *sqlx.DB) {
	result := Services{}
	var bitcoinNetworks []core.Network
	for _, coreServiceType := range services_helpers.GetCoreServiceTypes() {
		torqService := cache.GetCurrentCoreServiceState(coreServiceType)
		desiredState := cache.GetDesiredCoreServiceState(coreServiceType)
		if desiredState.Status != torqService.Status {
			result.ServiceMismatches = append(result.ServiceMismatches, ServiceMismatch{
				ServiceType:         coreServiceType,
				ServiceTypeString:   coreServiceType.String(),
				Status:              torqService.Status,
				StatusString:        torqService.Status.String(),
				DesiredStatus:       desiredState.Status,
				DesiredStatusString: desiredState.Status.String(),
				FailureTime:         cache.GetCoreFailedAttemptTime(coreServiceType),
			})
		}
		if coreServiceType == services_helpers.RootService {
			result.MainService = CoreService{
				CommonService: CommonService{
					ServiceType:       coreServiceType,
					ServiceTypeString: coreServiceType.String(),
					Status:            torqService.Status,
					BootTime:          torqService.ActiveTime,
					StatusString:      torqService.Status.String(),
				},
			}
			continue
		}
		result.TorqServices = append(result.TorqServices, CoreService{
			CommonService: CommonService{
				ServiceType:       coreServiceType,
				ServiceTypeString: coreServiceType.String(),
				Status:            torqService.Status,
				BootTime:          torqService.ActiveTime,
				StatusString:      torqService.Status.String(),
			},
		})
	}
	for _, lndNodeId := range cache.GetLndNodeIds() {
		nodeId := lndNodeId
		bitcoinNetwork := cache.GetNodeSettingsByNodeId(nodeId).Network
		for _, lndServiceType := range services_helpers.GetLndServiceTypes() {
			lndService := cache.GetCurrentNodeServiceState(lndServiceType, nodeId)
			desiredState := cache.GetDesiredNodeServiceState(lndServiceType, nodeId)
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
					FailureTime:         cache.GetNodeFailedAttemptTime(lndServiceType, nodeId),
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
	bitcoinNetwork := cache.GetNodeSettingsByNodeId(nodeId).Network
	for _, lndServiceType := range services_helpers.GetCoreServiceTypes() {
		lndService := cache.GetCurrentNodeServiceState(lndServiceType, nodeId)
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
