// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import "time"

type SalesStats struct {
	TotalOrders     int
	PendingOrders   int
	DeliveredOrders int
	RejectedOrders  int
	TotalRevenue    float64
	RefundedAmount  float64
	NetRevenue      float64
}

type ProductSales struct {
	ProductID    int64
	ProductName  string
	OrderCount   int
	Delivered    int
	TotalRevenue float64
}

func (d *DB) GetSalesStats(since time.Time) (*SalesStats, error) {
	s := &SalesStats{}

	q := `SELECT status, price_usd FROM orders`
	args := []interface{}{}
	if !since.IsZero() {
		q += " WHERE created_at >= ?"
		args = append(args, since.Format(time.RFC3339))
	}

	rows, err := d.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var price float64
		if err := rows.Scan(&status, &price); err != nil {
			return nil, err
		}
		s.TotalOrders++
		switch OrderStatus(status) {
		case OrderPending, OrderCodeRequested:
			s.PendingOrders++
		case OrderDelivered, OrderCodeSent:
			s.DeliveredOrders++
			s.TotalRevenue += price
		case OrderRejected:
			s.RejectedOrders++
			s.RefundedAmount += price
		}
	}
	s.NetRevenue = s.TotalRevenue
	return s, nil
}

func (d *DB) GetTopProducts(since time.Time, limit int) ([]ProductSales, error) {
	q := `
		SELECT product_id, product_name,
		       COUNT(*) AS order_count,
		       SUM(CASE WHEN status IN ('delivered', 'code_sent') THEN 1 ELSE 0 END) AS delivered_count,
		       COALESCE(SUM(CASE WHEN status IN ('delivered', 'code_sent') THEN price_usd ELSE 0 END), 0) AS total_revenue
		FROM orders`
	args := []interface{}{}
	if !since.IsZero() {
		q += " WHERE created_at >= ? AND status != 'rejected'"
		args = append(args, since.Format(time.RFC3339))
	} else {
		q += " WHERE status != 'rejected'"
	}
	q += " GROUP BY product_id, product_name ORDER BY total_revenue DESC"
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := d.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProductSales
	for rows.Next() {
		var p ProductSales
		if err := rows.Scan(&p.ProductID, &p.ProductName, &p.OrderCount, &p.Delivered, &p.TotalRevenue); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
