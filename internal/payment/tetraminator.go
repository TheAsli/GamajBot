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

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// ─── ایجاد فاکتور ───

type CreateInvoiceRequest struct {
	Price       int64  `json:"price"`
	CallbackURL string `json:"callback_url"`
}

type CreateInvoiceResponse struct {
	Status      bool   `json:"status"`
	Message     string `json:"message"`
	PayID       string `json:"pay_id"`
	PaymentLink string `json:"payment_link"`
}

func (c *Client) CreateInvoice(ctx context.Context, price int64, callbackURL string) (*CreateInvoiceResponse, error) {
	body, _ := json.Marshal(CreateInvoiceRequest{
		Price:       price,
		CallbackURL: callbackURL,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/invoice/create", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("خطای درگاه (%d): %s", resp.StatusCode, string(data))
	}

	var out CreateInvoiceResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if !out.Status {
		return nil, fmt.Errorf("درگاه خطا داد: %s", out.Message)
	}
	return &out, nil
}

// ─── استعلام تراکنش ───

type InquiryResponse struct {
	Status        bool   `json:"status"`
	PaymentStatus string `json:"payment_status"` // "paid" یا ...
	PayID         string `json:"pay_id"`
	Amount        int64  `json:"amount"`
}

func (c *Client) Inquiry(ctx context.Context, payID string) (*InquiryResponse, error) {
	url := fmt.Sprintf("%s/payment/inquiry/%s", c.baseURL, payID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("خطای استعلام (%d): %s", resp.StatusCode, string(data))
	}

	var out InquiryResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
