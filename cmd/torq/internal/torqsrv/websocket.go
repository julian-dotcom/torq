package torqsrv

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lncapital/torq/pkg/server_errors"
	"github.com/rs/zerolog/log"
)

type wsRequest struct {
	Id                 string `json:"id"`
	Type               string `json:"type"`
	*NewPaymentRequest `json:"newPaymentRequest"`
}

type NewPaymentRequest struct {
	Amount float64 `json:"amount"`
}

type NewPaymentResponse struct {
	Id     string  `json:"id"`
	Amount float64 `json:"amount"`
}

type Pong struct {
	Message string `json:"message"`
}

type wsError struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	Error string `json:"error"`
}

func processWsReq(conn *websocket.Conn, req wsRequest) {
	if req.Type == "ping" {
		conn.WriteJSON(Pong{Message: "pong"})
		return
	}

	// Validate request
	if req.Id == "" {
		wsr := wsError{
			Id:    req.Id,
			Type:  "Error",
			Error: "Id cannot be empty",
		}
		conn.WriteJSON(wsr)
		return
	}

	switch req.Type {
	case "newPayment":
		if req.NewPaymentRequest == nil {
			wsr := wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: "newPaymentRequest cannot be empty",
			}
			conn.WriteJSON(wsr)
			break
		}

		wsr := NewPaymentResponse{
			Id:     req.Id,
			Amount: req.NewPaymentRequest.Amount,
		}
		conn.WriteJSON(wsr)
		break
	default:
		err := fmt.Errorf("Unknown request type: %s", req.Type)
		wsr := wsError{
			Id:    req.Id,
			Type:  "Error",
			Error: err.Error(),
		}
		conn.WriteJSON(wsr)
	}
}

func WebsocketHandler(c *gin.Context) {
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
			log.Error().Msgf("%v", err)
			return
		case nil:
			processWsReq(conn, req)
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
