// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import "time"

type Stats struct {
	TotalUsers   int
	ActiveToday  int
	ActiveWeek   int
	ActiveMonth  int
	TotalBalance float64
	AvgBalance   float64
	PaidTx       int
	PendingTx    int
	CanceledTx   int
	ExpiredTx    int
	TotalCharged float64
	OpenTickets  int
}

func (d *DB) GetStats() (*Stats, error) {
	s := &Stats{}

	if err := d.QueryRow("SELECT COUNT(*), COALESCE(SUM(balance_usd), 0) FROM users").Scan(&s.TotalUsers, &s.TotalBalance); err != nil {
		return nil, err
	}
	if s.TotalUsers > 0 {
		s.AvgBalance = s.TotalBalance / float64(s.TotalUsers)
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	rows, err := d.Query(`
		SELECT user_id, MAX(created_at) FROM transactions GROUP BY user_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid int64
		var maxDate string
		if err := rows.Scan(&uid, &maxDate); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, maxDate)
		if err != nil {
			continue
		}
		if t.After(todayStart) {
			s.ActiveToday++
		}
		if t.After(weekAgo) {
			s.ActiveWeek++
		}
		if t.After(monthAgo) {
			s.ActiveMonth++
		}
	}

	rows2, err := d.Query(`SELECT type, status, amount_usd FROM transactions`)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var typ, status string
		var amt float64
		if err := rows2.Scan(&typ, &status, &amt); err != nil {
			return nil, err
		}
		switch {
		case typ == "charge" && status == "paid":
			s.PaidTx++
			s.TotalCharged += amt
		case status == "pending":
			s.PendingTx++
		case status == "canceled":
			s.CanceledTx++
		case status == "expired":
			s.ExpiredTx++
		}
	}

	_ = d.QueryRow("SELECT COUNT(*) FROM tickets WHERE status != 'closed'").Scan(&s.OpenTickets)

	return s, nil
}
