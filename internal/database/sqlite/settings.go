// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Settings struct {
	DollarRateMargin float64
	ChargeMin        int64
	ChargeMax        int64
	QuickAmounts     []int64
	MaintenanceMode  bool
	WelcomeText      string
	ForceJoinText    string
	ForceJoinChannel string
	ForceJoinURL     string
	SupportEnabled   bool
	Currency         string
	LogGroupChatID   int64
}

func (d *DB) GetSettings() (*Settings, error) {
	rows, err := d.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, fmt.Errorf("دریافت تنظیمات: %w", err)
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}

	s := &Settings{
		DollarRateMargin: 5.0,
		ChargeMin:        50000,
		ChargeMax:        10000000,
		QuickAmounts:     []int64{1, 5, 10, 20},
		MaintenanceMode:  false,
		WelcomeText:      "سلام %s 👋\nبه ربات خوش اومدی!",
		ForceJoinText:    "⚠️ برای استفاده از ربات، باید اول توی کانال ما عضو بشی:",
		SupportEnabled:   true,
		Currency:         "TOMAN",
	}

	if v, ok := m["dollar_rate_margin"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			s.DollarRateMargin = f
		}
	}
	if v, ok := m["charge_min"]; ok {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			s.ChargeMin = i
		}
	}
	if v, ok := m["charge_max"]; ok {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			s.ChargeMax = i
		}
	}
	if v, ok := m["quick_amounts"]; ok {
		var arr []int64
		if json.Unmarshal([]byte(v), &arr) == nil {
			s.QuickAmounts = arr
		}
	}
	if v, ok := m["maintenance_mode"]; ok {
		s.MaintenanceMode = v == "true"
	}
	if v, ok := m["welcome_text"]; ok {
		s.WelcomeText = v
	}
	if v, ok := m["force_join_text"]; ok {
		s.ForceJoinText = v
	}
	if v, ok := m["force_join_channel"]; ok {
		s.ForceJoinChannel = v
	}
	if v, ok := m["force_join_url"]; ok {
		s.ForceJoinURL = v
	}
	if v, ok := m["support_enabled"]; ok {
		s.SupportEnabled = v == "true"
	}
	if v, ok := m["currency"]; ok && v != "" {
		s.Currency = v
	}
	if v, ok := m["log_group_chat_id"]; ok {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			s.LogGroupChatID = i
		}
	}
	return s, nil
}

func (d *DB) UpdateSettings(s *Settings) error {
	quickJSON, _ := json.Marshal(s.QuickAmounts)

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	pairs := map[string]string{
		"dollar_rate_margin": fmt.Sprintf("%f", s.DollarRateMargin),
		"charge_min":         fmt.Sprintf("%d", s.ChargeMin),
		"charge_max":         fmt.Sprintf("%d", s.ChargeMax),
		"quick_amounts":      string(quickJSON),
		"maintenance_mode":   boolToStr(s.MaintenanceMode),
		"welcome_text":       s.WelcomeText,
		"force_join_text":    s.ForceJoinText,
		"force_join_channel": s.ForceJoinChannel,
		"force_join_url":     s.ForceJoinURL,
		"support_enabled":    boolToStr(s.SupportEnabled),
		"currency":           s.Currency,
		"log_group_chat_id":  fmt.Sprintf("%d", s.LogGroupChatID),
	}
	for k, v := range pairs {
		_, err := tx.Exec(
			"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
			k, v,
		)
		if err != nil {
			return fmt.Errorf("آپدیت تنظیم %s: %w", k, err)
		}
	}
	return tx.Commit()
}

func (d *DB) SetSetting(key, value string) error {
	_, err := d.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value,
	)
	return err
}
