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

const abanBaseURL = "https://abangateway.ir/api/v1"

type AbanClient struct {
	apiKey string
	http   *http.Client
}

func NewAbanClient(apiKey string) *AbanClient {
	return &AbanClient{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *AbanClient) Type() GatewayType {
	return GatewayAban
}

type abanCreateRequest struct {
	AmountRial  int64  `json:"amount_rial"`
	OrderID     string `json:"order_id"`
	CallbackURL string `json:"callback_url,omitempty"`
	Description string `json:"description,omitempty"`
}

type abanCreateResponse struct {
	InvoiceID   string `json:"invoice_id"`
	PaymentLink string `json:"payment_link"`
	AmountRial  int64  `json:"amount_rial"`
	Status      string `json:"status"`
}

func (c *AbanClient) CreateInvoice(ctx context.Context, orderID string, amountRial int64, callbackURL string) (*Invoice, error) {
	body, _ := json.Marshal(abanCreateRequest{
		AmountRial:  amountRial,
		OrderID:     orderID,
		CallbackURL: callbackURL,
		Description: "شارژ حساب",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", abanBaseURL+"/invoices", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("AbanGateway: %s", string(data))
	}

	var out abanCreateResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("پاسخ Aban نامعتبر: %s", string(data))
	}

	return &Invoice{
		ID:          out.InvoiceID,
		PayID:       out.InvoiceID,
		PaymentLink: out.PaymentLink,
		AmountRial:  out.AmountRial,
		AmountToman: out.AmountRial / 10,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}, nil
}

type abanInquiryResponse struct {
	Status string `json:"status"`
}

func (c *AbanClient) Inquiry(ctx context.Context, invoiceID string) (InvoiceStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", abanBaseURL+"/invoices/"+invoiceID, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var out abanInquiryResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("پاسخ استعلام نامعتبر: %s", string(data))
	}
	return InvoiceStatus(out.Status), nil
}

type abanVerifyResponse struct {
	Verified bool `json:"verified"`
}

func (c *AbanClient) Verify(ctx context.Context, invoiceID string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", abanBaseURL+"/invoices/"+invoiceID+"/verify", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("قبلاً تأیید شده (already_verified)")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("خطا در verify: %s", string(data))
	}

	var out abanVerifyResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return fmt.Errorf("پاسخ verify نامعتبر: %s", string(data))
	}
	if !out.Verified {
		return fmt.Errorf("verify موفق نبود")
	}
	return nil
}

func (c *AbanClient) Cancel(ctx context.Context, invoiceID string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", abanBaseURL+"/invoices/"+invoiceID+"/cancel", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("فاکتور قابل لغو نیست")
	}
	return nil
}
