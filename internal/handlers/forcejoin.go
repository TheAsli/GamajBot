// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"log"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/menus"
	"github.com/theasli/GamajBot/internal/middleware"

	AsliBot "gopkg.in/telebot.v4"
)

func ForceJoinCheckHandler(b *AsliBot.Bot, cfg *config.Config) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		user := c.Sender()
		log.Printf("🔘 عضو شدم: %d", user.ID)

		if cfg.ADMIN_ID != 0 && user.ID == cfg.ADMIN_ID {
			_ = c.Delete()
			return c.Send("✅ تأیید شد!", menus.MainMenu)
		}

		if !middleware.IsMember(b, cfg.ForceJoinChannel, user) {
			return c.Respond(&AsliBot.CallbackResponse{
				Text: "❌ هنوز عضو نشدی!",
			})
		}

		_ = c.Delete()
		return c.Send("✅ تأیید شد!", menus.MainMenu)
	}
}
