// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrOrderNotFound = errors.New("سفارش پیدا نشد")

type OrderStatus string

const (
	OrderPending       OrderStatus = "pending"
	OrderDelivered     OrderStatus = "delivered"
	OrderRejected      OrderStatus = "rejected"
	OrderCodeRequested OrderStatus = "code_requested"
	OrderCodeSent      OrderStatus = "code_sent"
)

type Order struct {
	ID              string
	UserID          int64
	UserName        string
	ProductID       int64
	ProductName     string
	CategoryID      int64
	PriceUSD        float64
	Status          OrderStatus
	ServiceInfo     string
	DeliveryData    map[string]string
	RejectNote      string
	CodeValue       string
	ClientEmail     string
	ClientLinks     []string
	ActionRequests  []ActionRequest
	CreatedAt       string
	DeliveredAt     string
	CodeRequestedAt string
}

func (d *DB) CreateOrder(userID int64, userName string, productID int64) (*Order, float64, error) {
	tx, err := d.Begin()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var p Product
	var available, active int
	var inboundJSON string
	err = tx.QueryRow(
		`SELECT id, category_id, name, description, price_usd, available, delivery_type, active, created_at, product_type, xui_panel_id, xui_inbound_ids, traffic_gb, duration_days
		 FROM products WHERE id = ?`, productID,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.PriceUSD, &available, &p.DeliveryType, &active, &p.CreatedAt, &p.ProductType, &p.XUIPanelID, &inboundJSON, &p.TrafficGB, &p.DurationDays)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrProductNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	if active != 1 || available != 1 {
		return nil, 0, ErrProductUnavailable
	}

	var bal float64
	err = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&bal)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrUserNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	if bal < p.PriceUSD {
		return nil, 0, ErrInsufficientBalance
	}

	if _, err := tx.Exec("UPDATE users SET balance_usd = balance_usd - ? WHERE id = ?", p.PriceUSD, userID); err != nil {
		return nil, 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO transactions (user_id, amount_usd, type, status, description, created_at)
		 VALUES (?, ?, 'purchase', 'paid', ?, ?)`,
		userID, -p.PriceUSD, "خرید "+p.Name, time.Now().Format(time.RFC3339),
	); err != nil {
		return nil, 0, err
	}

	order := &Order{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		UserID:      userID,
		UserName:    userName,
		ProductID:   p.ID,
		ProductName: p.Name,
		CategoryID:  p.CategoryID,
		PriceUSD:    p.PriceUSD,
		Status:      OrderPending,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	if _, err := tx.Exec(
		`INSERT INTO orders (id, user_id, user_name, product_id, product_name, category_id, price_usd, status, service_info, delivery_data, reject_note, code_value, client_email, client_links, action_requests, created_at, delivered_at, code_requested_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', '', '{}', '', '', '', '[]', '[]', ?, '', '')`,
		order.ID, userID, userName, p.ID, p.Name, p.CategoryID, p.PriceUSD, order.CreatedAt,
	); err != nil {
		return nil, 0, err
	}

	var newBal float64
	_ = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&newBal)

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return order, newBal, nil
}

func (d *DB) CreateSubscriptionOrder(userID int64, userName string, priceUSD float64, productName string) (*Order, float64, error) {
	tx, err := d.Begin()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var bal float64
	err = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&bal)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrUserNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	if bal < priceUSD {
		return nil, 0, ErrInsufficientBalance
	}

	if _, err := tx.Exec("UPDATE users SET balance_usd = balance_usd - ? WHERE id = ?", priceUSD, userID); err != nil {
		return nil, 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO transactions (user_id, amount_usd, type, status, description, created_at)
		 VALUES (?, ?, 'purchase', 'paid', ?, ?)`,
		userID, -priceUSD, "خرید "+productName, time.Now().Format(time.RFC3339),
	); err != nil {
		return nil, 0, err
	}

	order := &Order{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		UserID:      userID,
		UserName:    userName,
		ProductID:   0,
		ProductName: productName,
		PriceUSD:    priceUSD,
		Status:      OrderPending,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	if _, err := tx.Exec(
		`INSERT INTO orders (id, user_id, user_name, product_id, product_name, category_id, price_usd, status, service_info, delivery_data, reject_note, code_value, client_email, client_links, action_requests, created_at, delivered_at, code_requested_at)
		 VALUES (?, ?, ?, 0, ?, 0, ?, 'pending', '', '{}', '', '', '', '[]', '[]', ?, '', '')`,
		order.ID, userID, userName, productName, priceUSD, order.CreatedAt,
	); err != nil {
		return nil, 0, err
	}

	var newBal float64
	_ = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&newBal)

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return order, newBal, nil
}

func (d *DB) GetOrder(id string) (*Order, error) {
	var o Order
	var linksJSON, deliveryJSON, actionsJSON string
	err := d.QueryRow(
		`SELECT id, user_id, user_name, product_id, product_name, category_id, price_usd, status, service_info, delivery_data, reject_note, code_value, client_email, client_links, action_requests, created_at, delivered_at, code_requested_at
		 FROM orders WHERE id = ?`, id,
	).Scan(&o.ID, &o.UserID, &o.UserName, &o.ProductID, &o.ProductName, &o.CategoryID, &o.PriceUSD, &o.Status, &o.ServiceInfo, &deliveryJSON, &o.RejectNote, &o.CodeValue, &o.ClientEmail, &linksJSON, &actionsJSON, &o.CreatedAt, &o.DeliveredAt, &o.CodeRequestedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(linksJSON), &o.ClientLinks)
	o.DeliveryData = map[string]string{}
	_ = json.Unmarshal([]byte(deliveryJSON), &o.DeliveryData)
	_ = json.Unmarshal([]byte(actionsJSON), &o.ActionRequests)
	return &o, nil
}

func (d *DB) GetOrdersByUser(userID int64) ([]*Order, error) {
	rows, err := d.Query(
		`SELECT id, user_id, user_name, product_id, product_name, category_id, price_usd, status, service_info, delivery_data, reject_note, code_value, client_email, client_links, action_requests, created_at, delivered_at, code_requested_at
		 FROM orders WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Order
	for rows.Next() {
		var o Order
		var linksJSON, deliveryJSON, actionsJSON string
		if err := rows.Scan(&o.ID, &o.UserID, &o.UserName, &o.ProductID, &o.ProductName, &o.CategoryID, &o.PriceUSD, &o.Status, &o.ServiceInfo, &deliveryJSON, &o.RejectNote, &o.CodeValue, &o.ClientEmail, &linksJSON, &actionsJSON, &o.CreatedAt, &o.DeliveredAt, &o.CodeRequestedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(linksJSON), &o.ClientLinks)
		o.DeliveryData = map[string]string{}
		_ = json.Unmarshal([]byte(deliveryJSON), &o.DeliveryData)
		_ = json.Unmarshal([]byte(actionsJSON), &o.ActionRequests)
		out = append(out, &o)
	}
	return out, nil
}

func (d *DB) GetAllOrders() ([]*Order, error) {
	rows, err := d.Query(
		`SELECT id, user_id, user_name, product_id, product_name, category_id, price_usd, status, service_info, delivery_data, reject_note, code_value, client_email, client_links, action_requests, created_at, delivered_at, code_requested_at
		 FROM orders ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Order
	for rows.Next() {
		var o Order
		var linksJSON, deliveryJSON, actionsJSON string
		if err := rows.Scan(&o.ID, &o.UserID, &o.UserName, &o.ProductID, &o.ProductName, &o.CategoryID, &o.PriceUSD, &o.Status, &o.ServiceInfo, &deliveryJSON, &o.RejectNote, &o.CodeValue, &o.ClientEmail, &linksJSON, &actionsJSON, &o.CreatedAt, &o.DeliveredAt, &o.CodeRequestedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(linksJSON), &o.ClientLinks)
		o.DeliveryData = map[string]string{}
		_ = json.Unmarshal([]byte(deliveryJSON), &o.DeliveryData)
		_ = json.Unmarshal([]byte(actionsJSON), &o.ActionRequests)
		out = append(out, &o)
	}
	return out, nil
}

func (d *DB) UpdateOrder(o *Order) error {
	linksJSON, _ := json.Marshal(o.ClientLinks)
	deliveryJSON, _ := json.Marshal(o.DeliveryData)
	actionsJSON, _ := json.Marshal(o.ActionRequests)
	_, err := d.Exec(
		`UPDATE orders SET status = ?, service_info = ?, delivery_data = ?, reject_note = ?, code_value = ?, client_email = ?, client_links = ?, action_requests = ?, delivered_at = ?, code_requested_at = ?
		 WHERE id = ?`,
		o.Status, o.ServiceInfo, string(deliveryJSON), o.RejectNote, o.CodeValue, o.ClientEmail, string(linksJSON), string(actionsJSON), o.DeliveredAt, o.CodeRequestedAt, o.ID,
	)
	return err
}

func (d *DB) DeliverOrder(id string, serviceInfo string) (*Order, error) {
	o, err := d.GetOrder(id)
	if err != nil {
		return nil, err
	}
	if o.Status != OrderPending {
		return nil, errors.New("این سفارش قبلاً بررسی شده")
	}
	o.Status = OrderDelivered
	o.ServiceInfo = serviceInfo
	o.DeliveredAt = time.Now().Format(time.RFC3339)
	if err := d.UpdateOrder(o); err != nil {
		return nil, err
	}
	return o, nil
}

func (d *DB) RejectOrder(id string, note string) (*Order, float64, error) {
	tx, err := d.Begin()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var o Order
	var linksJSON, deliveryJSON, actionsJSON string
	err = tx.QueryRow(
		`SELECT id, user_id, user_name, product_id, product_name, category_id, price_usd, status, service_info, delivery_data, reject_note, code_value, client_email, client_links, action_requests, created_at, delivered_at, code_requested_at
		 FROM orders WHERE id = ?`, id,
	).Scan(&o.ID, &o.UserID, &o.UserName, &o.ProductID, &o.ProductName, &o.CategoryID, &o.PriceUSD, &o.Status, &o.ServiceInfo, &deliveryJSON, &o.RejectNote, &o.CodeValue, &o.ClientEmail, &linksJSON, &actionsJSON, &o.CreatedAt, &o.DeliveredAt, &o.CodeRequestedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrOrderNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	if o.Status != OrderPending {
		return nil, 0, errors.New("این سفارش قبلاً بررسی شده")
	}

	o.Status = OrderRejected
	o.RejectNote = note

	if _, err := tx.Exec("UPDATE users SET balance_usd = balance_usd + ? WHERE id = ?", o.PriceUSD, o.UserID); err != nil {
		return nil, 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO transactions (user_id, amount_usd, type, status, description, created_at)
		 VALUES (?, ?, 'refund', 'paid', ?, ?)`,
		o.UserID, o.PriceUSD, "بازگشت وجه سفارش رد شده", time.Now().Format(time.RFC3339),
	); err != nil {
		return nil, 0, err
	}
	if _, err := tx.Exec(
		"UPDATE orders SET status = 'rejected', reject_note = ? WHERE id = ?", note, id,
	); err != nil {
		return nil, 0, err
	}

	var newBal float64
	_ = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", o.UserID).Scan(&newBal)

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return &o, newBal, nil
}

func (d *DB) RequestOrderCode(id string) (*Order, error) {
	o, err := d.GetOrder(id)
	if err != nil {
		return nil, err
	}
	if o.Status != OrderDelivered && o.Status != OrderCodeSent {
		return nil, errors.New("این سفارش در وضعیت دریافت کد نیست")
	}
	o.Status = OrderCodeRequested
	o.CodeRequestedAt = time.Now().Format(time.RFC3339)
	if err := d.UpdateOrder(o); err != nil {
		return nil, err
	}
	return o, nil
}

func (d *DB) SendOrderCode(id string, code string) (*Order, error) {
	o, err := d.GetOrder(id)
	if err != nil {
		return nil, err
	}
	if o.Status != OrderCodeRequested {
		return nil, errors.New("درخواست کدی برای این سفارش وجود ندارد")
	}
	o.Status = OrderCodeSent
	o.CodeValue = code
	if err := d.UpdateOrder(o); err != nil {
		return nil, err
	}
	return o, nil
}

func (d *DB) CreatePlanOrder(userID int64, userName string, planName string, priceUSD float64, trafficGB int64, durationDays int, limitIP int, panelID int64, inboundIDs []int64) (*Order, float64, error) {
	tx, err := d.Begin()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var bal float64
	err = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&bal)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrUserNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	if bal < priceUSD {
		return nil, 0, ErrInsufficientBalance
	}

	if _, err := tx.Exec("UPDATE users SET balance_usd = balance_usd - ? WHERE id = ?", priceUSD, userID); err != nil {
		return nil, 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO transactions (user_id, amount_usd, type, status, description, created_at)
		 VALUES (?, ?, 'purchase', 'paid', ?, ?)`,
		userID, -priceUSD, "خرید "+planName, time.Now().Format(time.RFC3339),
	); err != nil {
		return nil, 0, err
	}

	order := &Order{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		UserID:      userID,
		UserName:    userName,
		ProductID:   0,
		ProductName: planName,
		PriceUSD:    priceUSD,
		Status:      OrderPending,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	if _, err := tx.Exec(
		`INSERT INTO orders (id, user_id, user_name, product_id, product_name, category_id, price_usd, status, service_info, delivery_data, reject_note, code_value, client_email, client_links, action_requests, created_at, delivered_at, code_requested_at)
		 VALUES (?, ?, ?, 0, ?, 0, ?, 'pending', '', '{}', '', '', '', '[]', '[]', ?, '', '')`,
		order.ID, userID, userName, planName, priceUSD, order.CreatedAt,
	); err != nil {
		return nil, 0, err
	}

	var newBal float64
	_ = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&newBal)

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return order, newBal, nil
}
