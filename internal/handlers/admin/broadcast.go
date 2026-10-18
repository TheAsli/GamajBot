// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"strings"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

// BroadcastHandler — پیام همگانی
func BroadcastHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		SetAwaiting(c.Sender().ID, "broadcast")
		return c.Send("📢 متن پیام رو بفرست:\n\nبرای لغو /cancel رو بزن.")
	}
}

// BroadcastInputHandler — هندلر ورودی پیام همگانی
func BroadcastInputHandler(b *AsliBot.Bot, c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	user := c.Sender()
	if db.GetAdminRole(user.ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		return nil
	}
	if getAwaiting(user.ID) != "broadcast" {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	SetAwaiting(user.ID, "")

	users := db.GetAllUsers()
	count := 0
	for _, u := range users {
		_, err := b.Send(AsliBot.ChatID(u.ID), text)
		if err == nil {
			count++
		}
	}

	return c.Send(fmt.Sprintf("✅ پیام به %d کاربر ارسال شد.", count))
}
