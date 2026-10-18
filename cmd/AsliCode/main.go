// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package main

import (
	"log"
	"os"

	"github.com/theasli/GamajBot/internal/bot"
	"github.com/theasli/GamajBot/internal/isenv"
)

func main() {
	// ساخت پوشه data اگه نیست
	_ = os.MkdirAll("data", 0755)

	cfg := isenv.Load()

	isBot, err := bot.New(cfg)
	if err != nil {
		log.Fatal("❌ خطا:", err)
	}

	log.Println("✅ AsliCode Bot is running...")
	isBot.Start()
}
