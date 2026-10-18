// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package config

type Config struct {
	BOT_TOKEN      string
	ASLICODE_TEXT  string
	ASLICODE_TEXTS string
	DEBUG          bool
	ADMIN_ID       int64

	TETRAMINATOR_API_KEY  string
	TETRAMINATOR_BASE_URL string

	CHARGE_MIN   int64
	CHARGE_MAX   int64
	QuickAmounts []int64

	DollarRateAPI    string
	DollarRateMargin float64
	DollarRateCache  int

	ForceJoinEnabled    bool
	ForceJoinChannel    string
	ForceJoinChannelURL string

	WEBHOOK_PORT     string
	WEBHOOK_BASE_URL string
}
