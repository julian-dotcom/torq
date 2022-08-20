package torqsrv

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lncapital/torq/pkg/server_errors"
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

type wsError struct {
	Id    string `json:"id"`
	Type  string `json:"type"`
	Error string `json:"error"`
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
		conn.ReadJSON(&req)
		if err != nil {
			wsr := wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: "Could not parse request",
			}
			conn.WriteJSON(wsr)
			continue
		}
		if req.Id == "" {
			wsr := wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: "Id cannot be empty",
			}
			conn.WriteJSON(wsr)
			continue
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
			err := fmt.Errorf("Unknown type: %s", req.Type)
			wsr := wsError{
				Id:    req.Id,
				Type:  "Error",
				Error: err.Error(),
			}
			conn.WriteJSON(wsr)
		}

	}
}
