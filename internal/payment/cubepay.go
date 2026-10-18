// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const cubePayBaseURL = "https://cubevps.ir/smspay"

type CubePayClient struct {
	apiToken string
	http     *http.Client
}

func NewCubePayClient(apiToken string) *CubePayClient {
	return &CubePayClient{
		apiToken: apiToken,
		http:     &http.Client{Timeout: 20 * time.Second},
	}
}

type cubePayCreateRequest struct {
	Amount      int64  `json:"amount"` // ریال
	OrderID     string `json:"order_id"`
	CallbackURL string `json:"callback_url"`
	Description string `json:"description,omitempty"`
	TTLMinutes  int    `json:"ttl_minutes,omitempty"`
}

type CubePayCreateResponse struct {
	Success        bool   `json:"success"`
	Authority      string `json:"authority"`
	PaymentLink    string `json:"payment_link"`
	PayAmount      int64  `json:"pay_amount"`
	PayAmountToman int64  `json:"pay_amount_toman"`
	IsTest         bool   `json:"is_test"`
	Message        string `json:"message,omitempty"`
}

func (c *CubePayClient) CreatePayment(ctx context.Context, orderID string, amountToman int64, callbackURL, description string) (*CubePayCreateResponse, error) {
	body, _ := json.Marshal(cubePayCreateRequest{
		Amount:      amountToman * 10, // تومان → ریال
		OrderID:     orderID,
		CallbackURL: callbackURL,
		Description: description,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", cubePayBaseURL+"/api/v1/create-payment.php", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var out CubePayCreateResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("خطا در پاسخ CubePay: %s", string(data))
	}
	if !out.Success {
		msg := out.Message
		if msg == "" {
			msg = string(data)
		}
		return nil, fmt.Errorf("CubePay: %s", msg)
	}
	return &out, nil
}

type CubePayVerifyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

func (c *CubePayClient) VerifyPayment(ctx context.Context, authority string) (*CubePayVerifyResponse, error) {
	body, _ := json.Marshal(map[string]string{"authority": authority})

	req, err := http.NewRequestWithContext(ctx, "POST", cubePayBaseURL+"/api/v1/verify-payment.php", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var out CubePayVerifyResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("خطا در پاسخ verify: %s", string(data))
	}
	return &out, nil
}

type CubePayStatusResponse struct {
	Status string `json:"status"` // pending | paid | expired | failed
}

func (c *CubePayClient) CheckStatus(ctx context.Context, authority string) (*CubePayStatusResponse, error) {
	url := fmt.Sprintf("%s/status.php?authority=%s&known_status=pending", cubePayBaseURL, authority)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var out CubePayStatusResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("خطا در پاسخ status: %s", string(data))
	}
	return &out, nil
}
