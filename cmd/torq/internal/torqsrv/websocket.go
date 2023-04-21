package torqsrv

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/lightning_helpers"
	"github.com/lncapital/torq/pkg/server_errors"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"

	"github.com/rs/zerolog/log"
)

type wsRequest struct {
	Type              string                               `json:"type"`
	NewPaymentRequest *lightning_helpers.NewPaymentRequest `json:"newPaymentRequest"`
}

type Pong struct {
	Message string `json:"message"`
}

type wsError struct {
	Type  string                    `json:"type"`
	Error server_errors.ServerError `json:"error"`
}

func processWsReq(webSocketResponseChannel chan<- interface{}, req wsRequest) {
	switch req.Type {
	case "ping":
		webSocketResponseChannel <- Pong{Message: "pong"}
		return
	case "newPayment":
		if req.NewPaymentRequest == nil {
			sendError(fmt.Errorf("unknown NewPaymentRequest for type: %s", req.Type), req, webSocketResponseChannel)
			break
		}
		req.NewPaymentRequest.ProgressReportChannel = webSocketResponseChannel
		_, err := lightning.NewPayment(*req.NewPaymentRequest)
		if err != nil {
			sendError(err, req, webSocketResponseChannel)
		}
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
			go processWsReq(webSocketResponseChannel, req)
		default:
			serverError := server_errors.SingleServerError("Could not parse request, please check that your JSON is correctly formated.")
			wsr := wsError{
				Type:  "Error",
				Error: *serverError,
			}
			webSocketResponseChannel <- wsr
		}
	}
}

func sendError(err error, req wsRequest, webSocketResponseChannel chan<- interface{}) {
	if err != nil {
		serverError := server_errors.SingleServerError(err.Error())
		webSocketResponseChannel <- wsError{
			Type:  "Error",
			Error: *serverError,
		}
	}
}
