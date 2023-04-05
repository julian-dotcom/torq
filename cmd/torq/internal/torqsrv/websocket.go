package torqsrv

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/pkg/server_errors"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/payments"
)

type wsRequest struct {
	RequestId           string                        `json:"requestId"`
	Type                string                        `json:"type"`
	NewPaymentRequest   *core.NewPaymentRequest       `json:"newPaymentRequest"`
	PayOnChainRequest   *core.PayOnChainRequest       `json:"payOnChainRequest"`
	CloseChannelRequest *channels.CloseChannelRequest `json:"closeChannelRequest"`
	Password            *string                       `json:"password"`
}

type Pong struct {
	Message string `json:"message"`
}

type AuthSuccess struct {
	AuthSuccess bool `json:"authSuccess"`
}

type wsError struct {
	RequestId string                    `json:"id"`
	Type      string                    `json:"type"`
	Error     server_errors.ServerError `json:"error"`
}

func processWsReq(db *sqlx.DB,
	webSocketResponseChannel chan<- interface{},
	req wsRequest) {

	if req.Type == "ping" {
		webSocketResponseChannel <- Pong{Message: "pong"}
		return
	}

	if req.RequestId == "" {
		sendError(fmt.Errorf("unknown requestId for type: %s", req.Type), req, webSocketResponseChannel)
		return
	}

	switch req.Type {
	case "newPayment":
		if req.NewPaymentRequest == nil {
			sendError(fmt.Errorf("unknown NewPaymentRequest for type: %s", req.Type), req, webSocketResponseChannel)
			break
		}
		sendError(payments.SendNewPayment(webSocketResponseChannel, db, *req.NewPaymentRequest, req.RequestId), req, webSocketResponseChannel)
	default:
		sendError(fmt.Errorf("unknown request type: %s", req.Type), req, webSocketResponseChannel)
	}
}

func WebsocketHandler(c *gin.Context, db *sqlx.DB) error {
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
	webSocketResponseChannel := make(chan interface{})
	done := make(chan struct{})

	conn, err := wsUpgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return errors.Wrap(err, "WebSocket upgrade.")
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("WebSocket close failure.")
		}
	}(conn)

	go processWebsocketRequests(conn, db, done, webSocketResponseChannel)

	for {
		select {
		case <-done:
			return errors.New("WebSocket Terminated.")
		case data := <-webSocketResponseChannel:
			err := conn.WriteJSON(data)
			if err != nil {
				log.Error().Err(err).Msg("Writing JSON to WebSocket failure.")
				return errors.New("Writing JSON to WebSocket failure.")
			}
		}
	}
}

func processWebsocketRequests(conn *websocket.Conn,
	db *sqlx.DB,
	done chan<- struct{},
	webSocketResponseChannel chan<- interface{}) {

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
			go processWsReq(db, webSocketResponseChannel, req)
		default:
			serverError := server_errors.SingleServerError("Could not parse request, please check that your JSON is correctly formated.")
			wsr := wsError{
				RequestId: req.RequestId,
				Type:      "Error",
				Error:     *serverError,
			}
			webSocketResponseChannel <- wsr
		}
	}
}

func sendError(err error, req wsRequest, webSocketResponseChannel chan<- interface{}) {
	if err != nil {
		serverError := server_errors.SingleServerError(err.Error())
		webSocketResponseChannel <- wsError{
			RequestId: req.RequestId,
			Type:      "Error",
			Error:     *serverError,
		}
	}
}
