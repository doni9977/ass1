package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type HTTPPaymentGateway struct {
	client  *http.Client
	baseURL string
}

func NewHTTPPaymentGateway(baseURL string) *HTTPPaymentGateway {
	return &HTTPPaymentGateway{
		client:  &http.Client{Timeout: 2 * time.Second},
		baseURL: baseURL,
	}
}

func (g *HTTPPaymentGateway) AuthorizePayment(orderID string, amount int64) (string, error) {
	payload := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	}
	body, _ := json.Marshal(payload)

	resp, err := g.client.Post(g.baseURL+"/payments", "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode >= 500 {
		return "", errors.New("service unavailable")
	}
	defer resp.Body.Close()

	var res map[string]string
	json.NewDecoder(resp.Body).Decode(&res)

	return res["status"], nil
}
