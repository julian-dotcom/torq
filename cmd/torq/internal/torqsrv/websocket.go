package torqsrv

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/internal/payments"
	"github.com/lncapital/torq/pkg/server_errors"
)

type wsRequest struct {
	Id                string                      `json:"id"`
	Type              string                      `json:"type"`
	NewPaymentRequest *payments.NewPaymentRequest `json:"newPaymentRequest"`
}

type Pong struct {
	Message string `json:"message"`
}

type wsError struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	Error string `json:"error"`
}

func processWsReq(db *sqlx.DB, c *gin.Context, wChan chan interface{}, req wsRequest) {
	if req.Type == "ping" {
		wChan <- Pong{Message: "pong"}
		return
	}

	// Validate request
	if req.Id == "" {
		wChan <- wsError{
			Id:    req.Id,
			Type:  "Error",
			Error: "Id cannot be empty",
		}
		return
	}

	switch req.Type {
	case "newPayment":
		if req.NewPaymentRequest == nil {
			wChan <- wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: "newPaymentRequest cannot be empty",
			}
			break
		}
		// Process a valid payment request
		err := payments.SendNewPayment(wChan, db, c, *req.NewPaymentRequest)
		if err != nil {
			wChan <- wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: err.Error(),
			}
		}
		break
	default:
		err := fmt.Errorf("Unknown request type: %s", req.Type)
		wChan <- wsError{
			Id:    req.Id,
			Type:  "Error",
			Error: err.Error(),
		}
	}
}

func WebsocketHandler(c *gin.Context, db *sqlx.DB) {
	conn, err := wsUpgrad.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	defer conn.Close()

	wc := make(chan interface{})

	go func(c *gin.Context) {
		for {
			err := conn.WriteJSON(<-wc)
			if err != nil {
				server_errors.LogAndSendServerError(c, err)
			}
		}
	}(c)

	for {
		req := wsRequest{}
		err := conn.ReadJSON(&req)
		switch err.(type) {
		case *websocket.CloseError:
			conn.Close()
			return
		case *websocket.HandshakeError:
			server_errors.LogAndSendServerError(c, err)
			return
		case nil:
			go processWsReq(db, c, wc, req)
			continue
		default:
			wsr := wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: fmt.Sprintf("Could not parse request, please check that your JSON is correctly formated"),
			}
			wc <- wsr
			continue
		}

	}
}
