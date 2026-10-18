// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type SubscriptionType string

const (
	SubTypeFixed   SubscriptionType = "fixed"
	SubTypeDynamic SubscriptionType = "dynamic"
)

type SubscriptionPlan struct {
	ID            int64
	Type          SubscriptionType
	Name          string
	Description   string
	PriceUSD      float64
	TrafficGB     int64
	DurationDays  int
	LimitIP       int
	XUIPanelID    int64
	XUIInboundIDs []int64
	Active        bool
	CreatedAt     string
}

type DynamicPrice struct {
	ID            int64
	TrafficGB     int64
	DurationDays  int
	LimitIP       int
	PriceUSD      float64
	XUIPanelID    int64
	XUIInboundIDs []int64
}

func (d *DB) CreateSubscriptionPlan(p *SubscriptionPlan) (*SubscriptionPlan, error) {
	now := time.Now().Format(time.RFC3339)
	inboundJSON, _ := json.Marshal(p.XUIInboundIDs)
	res, err := d.Exec(
		`INSERT INTO subscription_plans (type, name, description, price_usd, traffic_gb, duration_days, limit_ip, xui_panel_id, xui_inbound_ids, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Type, p.Name, p.Description, p.PriceUSD, p.TrafficGB, p.DurationDays, p.LimitIP,
		p.XUIPanelID, string(inboundJSON), boolToInt(p.Active), now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	p.ID = id
	p.CreatedAt = now
	return p, nil
}

func (d *DB) GetSubscriptionPlan(id int64) (*SubscriptionPlan, error) {
	var p SubscriptionPlan
	var active int
	var inboundJSON string
	err := d.QueryRow(
		`SELECT id, type, name, description, price_usd, traffic_gb, duration_days, limit_ip, xui_panel_id, xui_inbound_ids, active, created_at
		 FROM subscription_plans WHERE id = ?`, id,
	).Scan(&p.ID, &p.Type, &p.Name, &p.Description, &p.PriceUSD, &p.TrafficGB, &p.DurationDays, &p.LimitIP, &p.XUIPanelID, &inboundJSON, &active, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("پلن پیدا نشد")
	}
	if err != nil {
		return nil, err
	}
	p.Active = active == 1
	_ = json.Unmarshal([]byte(inboundJSON), &p.XUIInboundIDs)
	return &p, nil
}

func (d *DB) GetActiveSubscriptionPlans() ([]*SubscriptionPlan, error) {
	rows, err := d.Query(
		`SELECT id, type, name, description, price_usd, traffic_gb, duration_days, limit_ip, xui_panel_id, xui_inbound_ids, active, created_at
		 FROM subscription_plans WHERE active = 1 ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*SubscriptionPlan
	for rows.Next() {
		var p SubscriptionPlan
		var active int
		var inboundJSON string
		if err := rows.Scan(&p.ID, &p.Type, &p.Name, &p.Description, &p.PriceUSD, &p.TrafficGB, &p.DurationDays, &p.LimitIP, &p.XUIPanelID, &inboundJSON, &active, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Active = active == 1
		_ = json.Unmarshal([]byte(inboundJSON), &p.XUIInboundIDs)
		out = append(out, &p)
	}
	return out, nil
}

func (d *DB) GetAllSubscriptionPlans() ([]*SubscriptionPlan, error) {
	rows, err := d.Query(
		`SELECT id, type, name, description, price_usd, traffic_gb, duration_days, limit_ip, xui_panel_id, xui_inbound_ids, active, created_at
		 FROM subscription_plans ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*SubscriptionPlan
	for rows.Next() {
		var p SubscriptionPlan
		var active int
		var inboundJSON string
		if err := rows.Scan(&p.ID, &p.Type, &p.Name, &p.Description, &p.PriceUSD, &p.TrafficGB, &p.DurationDays, &p.LimitIP, &p.XUIPanelID, &inboundJSON, &active, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Active = active == 1
		_ = json.Unmarshal([]byte(inboundJSON), &p.XUIInboundIDs)
		out = append(out, &p)
	}
	return out, nil
}

func (d *DB) DeleteSubscriptionPlan(id int64) error {
	_, err := d.Exec("DELETE FROM subscription_plans WHERE id = ?", id)
	return err
}

func (d *DB) ToggleSubscriptionPlan(id int64) error {
	_, err := d.Exec("UPDATE subscription_plans SET active = CASE WHEN active = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}

func (d *DB) CreateDynamicPrice(dp *DynamicPrice) (*DynamicPrice, error) {
	inboundJSON, _ := json.Marshal(dp.XUIInboundIDs)
	res, err := d.Exec(
		`INSERT INTO dynamic_prices (traffic_gb, duration_days, limit_ip, price_usd, xui_panel_id, xui_inbound_ids)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		dp.TrafficGB, dp.DurationDays, dp.LimitIP, dp.PriceUSD, dp.XUIPanelID, string(inboundJSON),
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	dp.ID = id
	return dp, nil
}

func (d *DB) GetDynamicPrices() ([]*DynamicPrice, error) {
	rows, err := d.Query(`SELECT id, traffic_gb, duration_days, limit_ip, price_usd, xui_panel_id, xui_inbound_ids FROM dynamic_prices ORDER BY traffic_gb, duration_days`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*DynamicPrice
	for rows.Next() {
		var dp DynamicPrice
		var inboundJSON string
		if err := rows.Scan(&dp.ID, &dp.TrafficGB, &dp.DurationDays, &dp.LimitIP, &dp.PriceUSD, &dp.XUIPanelID, &inboundJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(inboundJSON), &dp.XUIInboundIDs)
		out = append(out, &dp)
	}
	return out, nil
}

func (d *DB) FindDynamicPrice(trafficGB int64, durationDays int, limitIP int) (*DynamicPrice, error) {
	var dp DynamicPrice
	var inboundJSON string
	err := d.QueryRow(
		`SELECT id, traffic_gb, duration_days, limit_ip, price_usd, xui_panel_id, xui_inbound_ids FROM dynamic_prices
		 WHERE traffic_gb = ? AND duration_days = ? AND limit_ip = ?`,
		trafficGB, durationDays, limitIP,
	).Scan(&dp.ID, &dp.TrafficGB, &dp.DurationDays, &dp.LimitIP, &dp.PriceUSD, &dp.XUIPanelID, &inboundJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("قیمتی برای این ترکیب تعریف نشده")
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(inboundJSON), &dp.XUIInboundIDs)
	return &dp, nil
}

func (d *DB) GetDynamicPrice(id int64) (*DynamicPrice, error) {
	var dp DynamicPrice
	var inboundJSON string
	err := d.QueryRow(
		`SELECT id, traffic_gb, duration_days, limit_ip, price_usd, xui_panel_id, xui_inbound_ids FROM dynamic_prices WHERE id = ?`, id,
	).Scan(&dp.ID, &dp.TrafficGB, &dp.DurationDays, &dp.LimitIP, &dp.PriceUSD, &dp.XUIPanelID, &inboundJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("قیمت پیدا نشد")
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(inboundJSON), &dp.XUIInboundIDs)
	return &dp, nil
}

func (d *DB) DeleteDynamicPrice(id int64) error {
	_, err := d.Exec("DELETE FROM dynamic_prices WHERE id = ?", id)
	return err
}
