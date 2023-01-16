package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
)

func RegisterUnauthenticatedRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("status", func(c *gin.Context) { getServicesHandler(c, db) })
}

func getServicesHandler(c *gin.Context, db *sqlx.DB) {
	result := Services{}
	torqService := commons.RunningServices[commons.TorqService]
	result.TorqService = TorqService{
		Service: Service{
			Status:   torqService.GetStatus(commons.TorqDummyNodeId),
			BootTime: torqService.GetBootTime(commons.TorqDummyNodeId),
		},
		Version: build.ExtendedVersion(),
	}
	lndService := commons.RunningServices[commons.LndService]
	for _, torqNodeId := range commons.GetAllActiveTorqNodeIds(nil, nil) {
		result.LndServices = append(result.LndServices, LndService{
			Service: Service{
				Status:   commons.RunningServices[commons.LndService].GetStatus(torqNodeId),
				BootTime: lndService.GetBootTime(torqNodeId),
			},
			NodeId:                        torqNodeId,
			TransactionStreamStatus:       commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.TransactionStream),
			TransactionStreamBootTime:     commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.TransactionStream),
			HtlcEventStreamStatus:         commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.HtlcEventStream),
			HtlcEventStreamBootTime:       commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.HtlcEventStream),
			ChannelEventStreamStatus:      commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.ChannelEventStream),
			ChannelEventStreamBootTime:    commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.ChannelEventStream),
			GraphEventStreamStatus:        commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.GraphEventStream),
			GraphEventStreamBootTime:      commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.GraphEventStream),
			ForwardStreamStatus:           commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.ForwardStream),
			ForwardStreamBootTime:         commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.ForwardStream),
			InvoiceStreamStatus:           commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.InvoiceStream),
			InvoiceStreamBootTime:         commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.InvoiceStream),
			PaymentStreamStatus:           commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.PaymentStream),
			PaymentStreamBootTime:         commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.PaymentStream),
			InFlightPaymentStreamStatus:   commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.InFlightPaymentStream),
			InFlightPaymentStreamBootTime: commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.InFlightPaymentStream),
			PeerEventStreamStatus:         commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, commons.PeerEventStream),
			PeerEventStreamBootTime:       commons.RunningServices[commons.LndService].GetStreamBootTime(torqNodeId, commons.PeerEventStream),
		})
	}
	vectorNodeIds, err := settings.GetPingSystemNodeIds(db, commons.Vector)
	if err == nil {
		for _, vectorNodeId := range vectorNodeIds {
			result.VectorServices = append(result.VectorServices, VectorService{
				Service: Service{
					Status:   commons.RunningServices[commons.VectorService].GetStatus(vectorNodeId),
					BootTime: commons.RunningServices[commons.VectorService].GetBootTime(vectorNodeId),
				},
				NodeId: vectorNodeId,
			})
		}
	} else {
		log.Info().Err(err).Msgf("Failed to obtain Vector ping systems maybe the database is not ready yet?")
	}
	ambossNodeIds, err := settings.GetPingSystemNodeIds(db, commons.Amboss)
	if err == nil {
		for _, ambossNodeId := range ambossNodeIds {
			result.AmbossServices = append(result.AmbossServices, AmbossService{
				Service: Service{
					Status:   commons.RunningServices[commons.AmbossService].GetStatus(ambossNodeId),
					BootTime: commons.RunningServices[commons.AmbossService].GetBootTime(ambossNodeId),
				},
				NodeId: ambossNodeId,
			})
		}
	} else {
		log.Info().Err(err).Msgf("Failed to obtain Amboss ping systems maybe the database is not ready yet?")
	}
	c.JSON(http.StatusOK, result)
}
