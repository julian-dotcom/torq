package torqsrv

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/on_chain_tx"
	"github.com/lncapital/torq/internal/payments"
	"github.com/lncapital/torq/pkg/broadcast"

	"github.com/rs/zerolog/log"
)

type wsRequest struct {
	ReqId               string                         `json:"reqId"`
	Type                string                         `json:"type"`
	NewPaymentRequest   *payments.NewPaymentRequest    `json:"newPaymentRequest"`
	PayOnChainRequest   *on_chain_tx.PayOnChainRequest `json:"payOnChainRequest"`
	OpenChannelRequest  *channels.OpenChannelRequest   `json:"openChannelRequest"`
	CloseChannelRequest *channels.CloseChannelRequest  `json:"closeChannelRequest"`
	Password            *string                        `json:"password"`
	NewAddressRequest   *on_chain_tx.NewAddressRequest `json:"newAddressRequest"`
}

type Pong struct {
	Message string `json:"message"`
}

type AuthSuccess struct {
	AuthSuccess bool `json:"authSuccess"`
}

type wsError struct {
	ReqId string `json:"id"`
	Type  string `json:"type"`
	Error string `json:"error"`
}

func processWsReq(db *sqlx.DB, c *gin.Context, eventChannel, webSocketChannel chan interface{}, req wsRequest) {
	if req.Type == "ping" {
		webSocketChannel <- Pong{Message: "pong"}
		return
	}

	// Validate request
	if req.ReqId == "" {
		webSocketChannel <- wsError{
			ReqId: req.ReqId,
			Type:  "Error",
			Error: "ReqId cannot be empty",
		}
		return
	}

	switch req.Type {
	case "newPayment":
		if req.NewPaymentRequest == nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: "newPaymentRequest cannot be empty",
			}
			break
		}
		// Process a valid payment request
		err := payments.SendNewPayment(eventChannel, db, c, *req.NewPaymentRequest, req.ReqId)
		if err != nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: err.Error(),
			}
		}
	case "newAddress":
		if req.NewAddressRequest == nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: "newAddressRequest cannot be empty",
			}
			break
		}
		// Process a valid payment request
		err := on_chain_tx.NewAddress(eventChannel, db, c, *req.NewAddressRequest, req.ReqId)
		if err != nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: err.Error(),
			}
		}
	case "closeChannel":
		if req.CloseChannelRequest == nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: "Close Channel request cannot be empty",
			}
			break
		}
		// Process a valid payment request
		err := channels.CloseChannel(eventChannel, db, c, *req.CloseChannelRequest, req.ReqId)
		if err != nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: err.Error(),
			}
		}
	case "openChannel":
		if req.OpenChannelRequest == nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: "OpenChannelRequest cannot be empty",
			}
			break
		}
		err := channels.OpenChannel(db, eventChannel, *req.OpenChannelRequest, req.ReqId)
		if err != nil {
			webSocketChannel <- wsError{
				ReqId: req.ReqId,
				Type:  "Error",
				Error: err.Error(),
			}
		}
	default:
		err := fmt.Errorf("Unknown request type: %s", req.Type)
		webSocketChannel <- wsError{
			ReqId: req.ReqId,
			Type:  "Error",
			Error: err.Error(),
		}
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
	go func() {
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
					ReqId: req.ReqId,
					Type:  "Error",
					Error: "Could not parse request, please check that your JSON is correctly formated.",
				}
				webSocketChannel <- wsr
			}
		}
	}()

	go func() {
		for event := range broadcaster.Subscribe() {
			select {
			case <-done:
				return
			default:
				// TODO FIXME FILTER OUT ONLY THE EVENTS THE USER ACTUALLY WANTS!!!
				if openChannelEvent, ok := event.(channels.OpenChannelResponse); ok {
					webSocketChannel <- openChannelEvent
				} else if closeChannelEvent, ok := event.(channels.CloseChannelResponse); ok {
					webSocketChannel <- closeChannelEvent
				} else if newAddressEvent, ok := event.(on_chain_tx.NewAddressResponse); ok {
					webSocketChannel <- newAddressEvent
				} else if newPaymentEvent, ok := event.(payments.NewPaymentResponse); ok {
					webSocketChannel <- newPaymentEvent
				}
			}
		}
	}()

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
