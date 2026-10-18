// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import "fmt"

type PaymentGatewaySettings struct {
	TetraminatorEnabled bool
	CubePayEnabled      bool
	ManualCardEnabled   bool
	AbanEnabled         bool

	ManualCardNumber    string
	ManualCardHolder    string
	CubePayToken        string
	TetraminatorAPIKey  string
	TetraminatorBaseURL string
	AbanAPIKey          string
	AbanWebhookSecret   string
}

func (d *DB) GetPaymentGateways() (*PaymentGatewaySettings, error) {
	rows, err := d.Query("SELECT key, value FROM payment_gateways")
	if err != nil {
		return nil, err
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

	return &PaymentGatewaySettings{
		TetraminatorEnabled: m["tetraminator_enabled"] == "true",
		CubePayEnabled:      m["cubepay_enabled"] == "true",
		ManualCardEnabled:   m["manual_card_enabled"] == "true",
		AbanEnabled:         m["aban_enabled"] == "true",
		ManualCardNumber:    m["manual_card_number"],
		ManualCardHolder:    m["manual_card_holder"],
		CubePayToken:        m["cubepay_token"],
		TetraminatorAPIKey:  m["tetraminator_api_key"],
		TetraminatorBaseURL: m["tetraminator_base_url"],
		AbanAPIKey:          m["aban_api_key"],
		AbanWebhookSecret:   m["aban_webhook_secret"],
	}, nil
}

func (d *DB) UpdatePaymentGateways(s *PaymentGatewaySettings) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	pairs := map[string]string{
		"tetraminator_enabled":  boolToStr(s.TetraminatorEnabled),
		"cubepay_enabled":       boolToStr(s.CubePayEnabled),
		"manual_card_enabled":   boolToStr(s.ManualCardEnabled),
		"aban_enabled":          boolToStr(s.AbanEnabled),
		"manual_card_number":    s.ManualCardNumber,
		"manual_card_holder":    s.ManualCardHolder,
		"cubepay_token":         s.CubePayToken,
		"tetraminator_api_key":  s.TetraminatorAPIKey,
		"tetraminator_base_url": s.TetraminatorBaseURL,
		"aban_api_key":          s.AbanAPIKey,
		"aban_webhook_secret":   s.AbanWebhookSecret,
	}
	for k, v := range pairs {
		_, err := tx.Exec(
			"INSERT INTO payment_gateways (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
			k, v,
		)
		if err != nil {
			return fmt.Errorf("آپدیت درگاه %s: %w", k, err)
		}
	}
	return tx.Commit()
}

func (d *DB) SetPaymentGatewayField(key string, value interface{}) error {
	var strVal string
	switch v := value.(type) {
	case bool:
		strVal = boolToStr(v)
	case string:
		strVal = v
	default:
		strVal = fmt.Sprintf("%v", v)
	}
	_, err := d.Exec(
		"INSERT INTO payment_gateways (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, strVal,
	)
	return err
}
