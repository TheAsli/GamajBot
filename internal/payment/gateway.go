// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package payment

import (
	"context"
	"time"
)

type GatewayType string

const (
	GatewayTetraminator GatewayType = "tetraminator"
	GatewayCubePay      GatewayType = "cubepay"
	GatewayAban         GatewayType = "abangateway"
	GatewayManualCard   GatewayType = "manual_card"
)

type Invoice struct {
	ID          string
	PayID       string
	PaymentLink string
	AmountRial  int64
	AmountToman int64
	ExpiresAt   time.Time
}

type InvoiceStatus string

const (
	StatusPending   InvoiceStatus = "pending"
	StatusPaid      InvoiceStatus = "paid"
	StatusExpired   InvoiceStatus = "expired"
	StatusCancelled InvoiceStatus = "cancelled"
)

type Gateway interface {
	Type() GatewayType
	CreateInvoice(ctx context.Context, orderID string, amountRial int64, callbackURL string) (*Invoice, error)
	Inquiry(ctx context.Context, invoiceID string) (InvoiceStatus, error)
	Verify(ctx context.Context, invoiceID string) error
	Cancel(ctx context.Context, invoiceID string) error
}
