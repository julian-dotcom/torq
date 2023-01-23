package torqsrv

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/on_chain_tx"
	"github.com/lncapital/torq/internal/payments"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

type wsRequest struct {
	RequestId           string                       `json:"requestId"`
	Type                string                       `json:"type"`
	NewPaymentRequest   *commons.NewPaymentRequest   `json:"newPaymentRequest"`
	PayOnChainRequest   *commons.PayOnChainRequest   `json:"payOnChainRequest"`
	OpenChannelRequest  *commons.OpenChannelRequest  `json:"openChannelRequest"`
	CloseChannelRequest *commons.CloseChannelRequest `json:"closeChannelRequest"`
	Password            *string                      `json:"password"`
	NewAddressRequest   *commons.NewAddressRequest   `json:"newAddressRequest"`
}

type Pong struct {
	Message string `json:"message"`
}

type AuthSuccess struct {
	AuthSuccess bool `json:"authSuccess"`
}

type wsError struct {
	RequestId string `json:"id"`
	Type      string `json:"type"`
	Error     string `json:"error"`
}

func processWsReq(db *sqlx.DB, c *gin.Context, eventChannel, webSocketChannel chan interface{}, req wsRequest) {
	if req.Type == "ping" {
		webSocketChannel <- Pong{Message: "pong"}
		return
	}

	if req.RequestId == "" {
		sendError(fmt.Errorf("unknown requestId for type: %s", req.Type), req, webSocketChannel)
		return
	}

	switch req.Type {
	case "newPayment":
		if req.NewPaymentRequest == nil {
			sendError(fmt.Errorf("unknown NewPaymentRequest for type: %s", req.Type), req, webSocketChannel)
			break
		}
		sendError(payments.SendNewPayment(eventChannel, db, c, *req.NewPaymentRequest, req.RequestId), req, webSocketChannel)
	case "newAddress":
		if req.NewAddressRequest == nil {
			sendError(fmt.Errorf("unknown NewAddressRequest for type: %s", req.Type), req, webSocketChannel)
			break
		}
		sendError(on_chain_tx.NewAddress(eventChannel, db, c, *req.NewAddressRequest, req.RequestId), req, webSocketChannel)
	case "closeChannel":
		if req.CloseChannelRequest == nil {
			sendError(fmt.Errorf("unknown CloseChannelRequest for type: %s", req.Type), req, webSocketChannel)
			break
		}
		sendError(channels.CloseChannel(eventChannel, db, c, *req.CloseChannelRequest, req.RequestId), req, webSocketChannel)
	case "openChannel":
		if req.OpenChannelRequest == nil {
			sendError(fmt.Errorf("unknown OpenChannelRequest for type: %s", req.Type), req, webSocketChannel)
			break
		}
		sendError(channels.OpenChannel(eventChannel, db, *req.OpenChannelRequest, req.RequestId), req, webSocketChannel)
	default:
		sendError(fmt.Errorf("unknown request type: %s", req.Type), req, webSocketChannel)
	}
}

func WebsocketHandler(c *gin.Context, db *sqlx.DB, eventChannel chan interface{}, broadcaster broadcast.BroadcastServer) error {
	var wsUpgrade = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if r.Header.Get("Origin") == "http://localhost:3000" {
				return true
			}

			origin := r.Header["Origin"]
			if len(origin) == 0 {
				return true
			}
			u, err := url.Parse(origin[0])
			if err != nil {
				return false
			}
			return equalASCIIFold(u.Host, r.Host)
		},
	}

	conn, err := wsUpgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return errors.Wrap(err, "WebSocket upgrade.")
	}
	defer conn.Close()

	webSocketChannel := make(chan interface{})
	done := make(chan struct{})
	go processWebsocketRequests(c, conn, done, db, eventChannel, webSocketChannel)
	go processBroadcasterEvents(done, broadcaster, webSocketChannel)

	for {
		select {
		case <-done:
			return errors.New("WebSocket Terminated.")
		case data := <-webSocketChannel:
			err := conn.WriteJSON(data)
			if err != nil {
				log.Error().Err(err).Msg("Writing JSON to WebSocket failure.")
				return errors.New("Writing JSON to WebSocket failure.")
			}
		}
	}
}
func processWebsocketRequests(c *gin.Context, conn *websocket.Conn, done chan struct{}, db *sqlx.DB, eventChannel chan interface{},
	webSocketChannel chan interface{}) {
	defer close(done)
	for {
		req := wsRequest{}
		err := conn.ReadJSON(&req)
		switch err.(type) {
		case *websocket.CloseError:
			log.Debug().Err(err).Msg("WebSocket Close Error.")
			return
		case *websocket.HandshakeError:
			log.Debug().Err(err).Msg("WebSocket Handshake Error.")
			return
		case nil:
			go processWsReq(db, c, eventChannel, webSocketChannel, req)
		default:
			wsr := wsError{
				RequestId: req.RequestId,
				Type:      "Error",
				Error:     "Could not parse request, please check that your JSON is correctly formated.",
			}
			webSocketChannel <- wsr
		}
	}
}

func processBroadcasterEvents(done chan struct{}, broadcaster broadcast.BroadcastServer,
	webSocketChannel chan interface{}) {

	listener := broadcaster.SubscribeWebSocketResponse()
	for event := range listener {
		select {
		case <-done:
			broadcaster.CancelSubscriptionWebSocketResponse(listener)
			return
		default:
		}
		if openChannelEvent, ok := event.(commons.OpenChannelResponse); ok {
			webSocketChannel <- openChannelEvent
		} else if closeChannelEvent, ok := event.(commons.CloseChannelResponse); ok {
			webSocketChannel <- closeChannelEvent
		} else if newAddressEvent, ok := event.(commons.NewAddressResponse); ok {
			webSocketChannel <- newAddressEvent
		} else if newPaymentEvent, ok := event.(commons.NewPaymentResponse); ok {
			webSocketChannel <- newPaymentEvent
		}
	}
}

func sendError(err error, req wsRequest, webSocketChannel chan interface{}) {
	if err != nil {
		webSocketChannel <- wsError{
			RequestId: req.RequestId,
			Type:      "Error",
			Error:     err.Error(),
		}
	}
}
