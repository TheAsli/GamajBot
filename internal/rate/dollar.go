// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package rate

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type DollarRate struct {
	Price     int64 // تومان
	UpdatedAt time.Time
}

var (
	cachedRate *DollarRate
	cacheMu    sync.RWMutex
)

// GetDollarRate — نرخ خام دلار (بدون سود)
func GetDollarRate(apiURL string, cacheMinutes int) (int64, error) {
	cacheMu.RLock()
	if cachedRate != nil && time.Since(cachedRate.UpdatedAt) < time.Duration(cacheMinutes)*time.Minute {
		price := cachedRate.Price
		cacheMu.RUnlock()
		return price, nil
	}
	cacheMu.RUnlock()

	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	// ریال → تومان
	price := int64(result.Price) / 10

	cacheMu.Lock()
	cachedRate = &DollarRate{
		Price:     price,
		UpdatedAt: time.Now(),
	}
	cacheMu.Unlock()

	return price, nil
}

// GetFinalRate — نرخ نهایی (با سود) — کاربر این رو می‌بینه
func GetFinalRate(apiURL string, cacheMinutes int, marginPercent float64) (int64, error) {
	baseRate, err := GetDollarRate(apiURL, cacheMinutes)
	if err != nil {
		return 0, err
	}
	finalRate := float64(baseRate) * (1 + marginPercent/100)
	return int64(finalRate), nil
}

// CalculateToman — تبدیل دلار به تومان با نرخ نهایی
func CalculateToman(dollarAmount float64, finalRate int64) int64 {
	return int64(float64(finalRate) * dollarAmount)
}

// CalculateUSD — تبدیل تومان به دلار با نرخ نهایی
func CalculateUSD(toman int64, finalRate int64) float64 {
	if finalRate == 0 {
		return 0
	}
	return float64(toman) / float64(finalRate)
}
