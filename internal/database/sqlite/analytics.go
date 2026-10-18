// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"fmt"
	"time"
)

type DailyRevenue struct {
	Date    string
	Revenue float64
	Orders  int
}

type MonthlyComparison struct {
	CurrentMonthRevenue  float64
	PreviousMonthRevenue float64
	CurrentMonthOrders   int
	PreviousMonthOrders  int
	GrowthPercent        float64
}

func (d *DB) GetDailyRevenue(days int) ([]DailyRevenue, error) {
	if days <= 0 {
		days = 7
	}
	now := time.Now()
	start := now.AddDate(0, 0, -days)

	rows, err := d.Query(
		`SELECT substr(created_at, 1, 10) as day, price_usd
		 FROM orders
		 WHERE status IN ('delivered', 'code_sent') AND created_at >= ?`,
		start.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byDay := make(map[string]*DailyRevenue)
	for rows.Next() {
		var day string
		var price float64
		if err := rows.Scan(&day, &price); err != nil {
			return nil, err
		}
		if byDay[day] == nil {
			byDay[day] = &DailyRevenue{Date: day}
		}
		byDay[day].Revenue += price
		byDay[day].Orders++
	}

	out := make([]DailyRevenue, 0, days)
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		if byDay[day] != nil {
			out = append(out, *byDay[day])
		} else {
			out = append(out, DailyRevenue{Date: day})
		}
	}
	return out, nil
}

func (d *DB) GetMonthlyComparison() (*MonthlyComparison, error) {
	now := time.Now()
	currentStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	previousStart := currentStart.AddDate(0, -1, 0)

	m := &MonthlyComparison{}

	rows, err := d.Query(
		`SELECT price_usd, created_at FROM orders WHERE status IN ('delivered', 'code_sent')`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var price float64
		var createdAt string
		if err := rows.Scan(&price, &createdAt); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			continue
		}
		if t.After(currentStart) {
			m.CurrentMonthRevenue += price
			m.CurrentMonthOrders++
		} else if t.After(previousStart) && t.Before(currentStart) {
			m.PreviousMonthRevenue += price
			m.PreviousMonthOrders++
		}
	}

	if m.PreviousMonthRevenue > 0 {
		m.GrowthPercent = ((m.CurrentMonthRevenue - m.PreviousMonthRevenue) / m.PreviousMonthRevenue) * 100
	} else if m.CurrentMonthRevenue > 0 {
		m.GrowthPercent = 100
	}

	return m, nil
}

func (d *DB) GetRevenueByGateway(days int) (map[string]float64, error) {
	if days <= 0 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -days)

	rows, err := d.Query(
		`SELECT gateway, amount_usd FROM transactions
		 WHERE type = 'charge' AND status = 'paid' AND created_at >= ?`,
		start.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var gateway string
		var amount float64
		if err := rows.Scan(&gateway, &amount); err != nil {
			return nil, err
		}
		if gateway == "" {
			gateway = "unknown"
		}
		out[gateway] += amount
	}
	return out, nil
}

var _ = fmt.Sprintf
