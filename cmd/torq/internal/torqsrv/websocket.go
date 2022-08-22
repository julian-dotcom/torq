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

func processWsReq(db *sqlx.DB, c *gin.Context, conn *websocket.Conn, req wsRequest) {
	if req.Type == "ping" {
		err := conn.WriteJSON(Pong{Message: "pong"})
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
		}
		return
	}

	// Validate request
	if req.Id == "" {
		wsr := wsError{
			Id:    req.Id,
			Type:  "Error",
			Error: "Id cannot be empty",
		}
		err := conn.WriteJSON(wsr)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
		}
		return
	}

	switch req.Type {
	case "newPayment":
		fmt.Println("newPayment")

		if req.NewPaymentRequest == nil {
			wsr := wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: "newPaymentRequest cannot be empty",
			}
			err := conn.WriteJSON(wsr)
			if err != nil {
				server_errors.LogAndSendServerError(c, err)
			}
			break
		}
		// Process a valid payment request
		payments.SendNewPayment(conn, db, c, *req.NewPaymentRequest)
		break
	default:
		err := fmt.Errorf("Unknown request type: %s", req.Type)
		wsr := wsError{
			Id:    req.Id,
			Type:  "Error",
			Error: err.Error(),
		}
		err = conn.WriteJSON(wsr)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
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
			go processWsReq(db, c, conn, req)
			continue
		default:
			wsr := wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: "Could not parse request, please check your JSON",
			}
			conn.WriteJSON(wsr)
			continue
		}

	}
}
