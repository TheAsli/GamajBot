// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package isenv

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/theasli/GamajBot/internal/config"
)

func Load() *config.Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  فایل .env پیدا نشد")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ BOT_TOKEN تنظیم نشده")
	}

	return &config.Config{
		BOT_TOKEN:      token,
		ASLICODE_TEXT:  os.Getenv("ASLICODE_TEXT"),
		ASLICODE_TEXTS: os.Getenv("ASLICODE_TEXTS"),
		DEBUG:          os.Getenv("DEBUG") == "true",
		ADMIN_ID:       parseInt(os.Getenv("ADMIN_ID")),

		TETRAMINATOR_API_KEY:  os.Getenv("TETRAMINATOR_API_KEY"),
		TETRAMINATOR_BASE_URL: os.Getenv("TETRAMINATOR_BASE_URL"),

		CHARGE_MIN:   parseInt(os.Getenv("CHARGE_MIN")),
		CHARGE_MAX:   parseInt(os.Getenv("CHARGE_MAX")),
		QuickAmounts: parseAmounts(os.Getenv("QUICK_AMOUNTS")),

		DollarRateAPI:    os.Getenv("DOLLAR_RATE_API"),
		DollarRateMargin: parseFloat(os.Getenv("DOLLAR_RATE_MARGIN")),
		DollarRateCache:  int(parseInt(os.Getenv("DOLLAR_RATE_CACHE"))),

		ForceJoinEnabled:    os.Getenv("FORCE_JOIN_ENABLED") == "true",
		ForceJoinChannel:    os.Getenv("FORCE_JOIN_CHANNEL"),
		ForceJoinChannelURL: os.Getenv("FORCE_JOIN_CHANNEL_URL"),

		WEBHOOK_PORT:     os.Getenv("WEBHOOK_PORT"),
		WEBHOOK_BASE_URL: os.Getenv("WEBHOOK_BASE_URL"),
	}
}

func parseInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func parseFloat(s string) float64 {
	n, _ := strconv.ParseFloat(s, 64)
	return n
}

func parseAmounts(s string) []int64 {
	var out []int64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.ParseInt(part, 10, 64)
		if err == nil && n > 0 {
			out = append(out, n)
		}
	}
	return out
}
