// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package rate

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/theasli/GamajBot/internal/database"
)

const currencyAPI = "https://nerkh.jahankhahan.shop/data/live.json"

type liveResponse struct {
	Rates map[string]float64 `json:"rates"`
}

var (
	currencyMu   sync.Mutex
	lastFetch    time.Time
	fetchTimeout = 5 * time.Minute
)

func FetchRates() error {
	currencyMu.Lock()
	defer currencyMu.Unlock()

	if time.Since(lastFetch) < fetchTimeout {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(currencyAPI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data liveResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	dollar := int64(data.Rates["dollar"])
	tether := int64(data.Rates["tether"])
	if dollar == 0 {
		return fmt.Errorf("نرخ دلار دریافت نشد")
	}
	if tether == 0 {
		tether = dollar
	}

	database.SetCachedRates(dollar, tether)
	lastFetch = time.Now()
	return nil
}

func GetRateFor(currency database.CurrencyType) (int64, error) {
	rates := database.GetCachedRates()
	if rates == nil {
		if err := FetchRates(); err != nil {
			return 0, err
		}
		rates = database.GetCachedRates()
	}
	if rates == nil {
		return 0, fmt.Errorf("نرخ در دسترس نیست")
	}

	switch currency {
	case database.CurrencyToman:
		return 1, nil
	case database.CurrencyRial:
		return 1, nil
	case database.CurrencyUSD:
		return rates.DollarToman, nil
	case database.CurrencyUSDT:
		return rates.TetherToman, nil
	}
	return 0, fmt.Errorf("واحد پول نامعتبر")
}

func ConvertToDisplay(amountUSD float64, currency database.CurrencyType, marginPercent float64) string {
	rate, err := GetRateFor(currency)
	if err != nil {
		return fmt.Sprintf("%.2f دلار", amountUSD)
	}

	if currency == database.CurrencyUSD || currency == database.CurrencyUSDT {
		value := amountUSD
		if marginPercent > 0 {
			value = value * (1 + marginPercent/100)
		}
		return fmt.Sprintf("%.2f %s", value, currency)
	}

	valueToman := amountUSD * float64(rate)
	if marginPercent > 0 {
		valueToman = valueToman * (1 + marginPercent/100)
	}

	if currency == database.CurrencyRial {
		return fmt.Sprintf("%s ریال", formatMoney(int64(math.Round(valueToman*10))))
	}
	return fmt.Sprintf("%s تومان", formatMoney(int64(math.Round(valueToman))))
}

func ConvertFromDisplay(value float64, currency database.CurrencyType, marginPercent float64) float64 {
	rate, err := GetRateFor(currency)
	if err != nil {
		return value
	}

	if currency == database.CurrencyUSD || currency == database.CurrencyUSDT {
		if marginPercent > 0 {
			return value / (1 + marginPercent/100)
		}
		return value
	}

	if currency == database.CurrencyRial {
		value = value / 10
	}

	toman := value
	if marginPercent > 0 {
		toman = toman / (1 + marginPercent/100)
	}
	if rate == 0 {
		return 0
	}
	return toman / float64(rate)
}

func Display(amountUSD float64, currency database.CurrencyType) string {
	return ConvertToDisplay(amountUSD, currency, 0)
}

func CurrencySymbol(c database.CurrencyType) string {
	switch c {
	case database.CurrencyUSD:
		return "$"
	case database.CurrencyUSDT:
		return "USDT"
	case database.CurrencyToman:
		return "تومان"
	case database.CurrencyRial:
		return "ریال"
	}
	return "$"
}

func formatMoney(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	var out []byte
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(ch))
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}
