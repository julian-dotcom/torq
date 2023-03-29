package services

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/internal/settings"
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
	result.TorqService = TorqService{
		CommonService: CommonService{
			Status:   commons.RunningServices[commons.TorqService].GetStatus(commons.TorqDummyNodeId),
			BootTime: commons.RunningServices[commons.TorqService].GetBootTime(commons.TorqDummyNodeId),
		},
		Version: build.ExtendedVersion(),
	}
	result.TorqService.StatusString = result.TorqService.Status.String()
	torqNodeIds := commons.GetAllActiveTorqNodeIds(nil, nil)
	for _, torqNodeId := range torqNodeIds {
		nodeSettings := commons.GetNodeSettingsByNodeId(torqNodeId)
		if nodeSettings.Chain == commons.Bitcoin && !slices.Contains(bitcoinNetworks, nodeSettings.Network) {
			bitcoinNetworks = append(bitcoinNetworks, nodeSettings.Network)
		}
		lndService := LndService{
			CommonService: CommonService{
				Status:   commons.RunningServices[commons.LndService].GetStatus(torqNodeId),
				BootTime: commons.RunningServices[commons.LndService].GetBootTime(torqNodeId),
			},
			NodeId:         torqNodeId,
			BitcoinNetwork: nodeSettings.Network,
		}
		lndService.StatusString = lndService.Status.String()
		for _, subscriptionStream := range commons.SubscriptionStreams {
			lndServiceStream := Stream{
				CommonService: CommonService{
					Status:   commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, subscriptionStream),
					BootTime: commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, subscriptionStream),
				},
				NodeId:     torqNodeId,
				Type:       subscriptionStream,
				TypeString: subscriptionStream.String(),
			}
			lndServiceStream.StatusString = lndServiceStream.Status.String()
			lndService.StreamStatus = append(lndService.StreamStatus, lndServiceStream)
		}
		result.LndServices = append(result.LndServices, lndService)
	}
	vectorNodeIds, err := settings.GetPingSystemNodeIds(db, settings.Vector)
	if err == nil {
		for _, vectorNodeId := range vectorNodeIds {
			vectorService := VectorService{
				CommonService: CommonService{
					Status:   commons.RunningServices[commons.VectorService].GetStatus(vectorNodeId),
					BootTime: commons.RunningServices[commons.VectorService].GetBootTime(vectorNodeId),
				},
				NodeId: vectorNodeId,
			}
			vectorService.StatusString = vectorService.Status.String()
			result.VectorServices = append(result.VectorServices, vectorService)
		}
	} else {
		log.Info().Err(err).Msgf("Failed to obtain Vector ping systems maybe the database is not ready yet?")
	}
	ambossNodeIds, err := settings.GetPingSystemNodeIds(db, settings.Amboss)
	if err == nil {
		for _, ambossNodeId := range ambossNodeIds {
			ambossService := AmbossService{
				CommonService: CommonService{
					Status:   commons.RunningServices[commons.AmbossService].GetStatus(ambossNodeId),
					BootTime: commons.RunningServices[commons.AmbossService].GetBootTime(ambossNodeId),
				},
				NodeId: ambossNodeId,
			}
			ambossService.StatusString = ambossService.Status.String()
			result.AmbossServices = append(result.AmbossServices, ambossService)
		}
	} else {
		log.Info().Err(err).Msgf("Failed to obtain Amboss ping systems maybe the database is not ready yet?")
	}
	for svc := range commons.RunningServices {
		switch svc {
		case commons.TorqService,
			commons.LndService,
			commons.VectorService,
			commons.AmbossService:
			// Already done
		case commons.AutomationService,
			commons.MaintenanceService,
			commons.CronService:
			service := Service{
				CommonService: CommonService{
					Status:   commons.RunningServices[svc].GetStatus(commons.TorqDummyNodeId),
					BootTime: commons.RunningServices[svc].GetBootTime(commons.TorqDummyNodeId),
				},
				Type:       svc,
				TypeString: svc.String(),
			}
			service.StatusString = service.Status.String()
			result.Services = append(result.Services, service)
		default:
			for index, torqNodeId := range torqNodeIds {
				service := Service{
					CommonService: CommonService{
						Status:   commons.RunningServices[svc].GetStatus(torqNodeId),
						BootTime: commons.RunningServices[svc].GetBootTime(torqNodeId),
					},
					Type:       svc,
					TypeString: svc.String(),
					NodeId:     &torqNodeIds[index],
				}
				service.StatusString = service.Status.String()
				result.Services = append(result.Services, service)
			}
		}
	}
	result.BitcoinNetworks = bitcoinNetworks
	c.JSON(http.StatusOK, result)
}

func getLndServicesHandler(c *gin.Context, db *sqlx.DB) {
	nodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse nodeId in the request.")
		return
	}

	var lndService LndService
	torqNodeIds := commons.GetAllActiveTorqNodeIds(nil, nil)
	for _, torqNodeId := range torqNodeIds {
		if torqNodeId == nodeId {
			lndService = LndService{
				CommonService: CommonService{
					Status:   commons.RunningServices[commons.LndService].GetStatus(torqNodeId),
					BootTime: commons.RunningServices[commons.LndService].GetBootTime(torqNodeId),
				},
				NodeId: torqNodeId,
			}
			lndService.StatusString = lndService.Status.String()
			for _, subscriptionStream := range commons.SubscriptionStreams {
				lndServiceStream := Stream{
					CommonService: CommonService{
						Status:   commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, subscriptionStream),
						BootTime: commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, subscriptionStream),
					},
					NodeId:     torqNodeId,
					Type:       subscriptionStream,
					TypeString: subscriptionStream.String(),
				}
				lndServiceStream.StatusString = lndServiceStream.Status.String()
				lndService.StreamStatus = append(lndService.StreamStatus, lndServiceStream)
			}
		}
	}
	c.JSON(http.StatusOK, lndService)
}
