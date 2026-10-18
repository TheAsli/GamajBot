// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import "sync"

type CurrencyType string

const (
	CurrencyUSD   CurrencyType = "USD"
	CurrencyUSDT  CurrencyType = "USDT"
	CurrencyToman CurrencyType = "TOMAN"
	CurrencyRial  CurrencyType = "RIAL"
)

type CurrencyRates struct {
	DollarToman int64
	TetherToman int64
}

var (
	cachedRates *CurrencyRates
	ratesMu     sync.RWMutex
)

func (d *DB) GetCurrency() CurrencyType {
	var v string
	err := d.QueryRow("SELECT value FROM settings WHERE key = 'currency'").Scan(&v)
	if err != nil || v == "" {
		return CurrencyToman
	}
	return CurrencyType(v)
}

func (d *DB) SetCurrency(c CurrencyType) error {
	return d.SetSetting("currency", string(c))
}

func SetCachedRates(dollar, tether int64) {
	ratesMu.Lock()
	defer ratesMu.Unlock()
	cachedRates = &CurrencyRates{DollarToman: dollar, TetherToman: tether}
}

func GetCachedRates() *CurrencyRates {
	ratesMu.RLock()
	defer ratesMu.RUnlock()
	return cachedRates
}
