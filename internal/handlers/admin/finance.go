// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

func FinancePanelHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		inline := &AsliBot.ReplyMarkup{}
		rows := []AsliBot.Row{
			inline.Row(inline.Data("🚪 درگاه‌ها", "adm_gateways")),
			inline.Row(inline.Data("💰 تراکنش‌ها", "adm_tx")),
			inline.Row(inline.Data("💳 شارژ دستی", "adm_manual_charge")),
			inline.Row(inline.Data("➖ کسر موجودی", "adm_deduct")),
			inline.Row(inline.Data("🚫 بلاک کاربر", "adm_ban"), inline.Data("✅ آنبلاک", "adm_unban")),
		}
		inline.Inline(rows...)

		return c.Send("💰 پنل مالی\n\nیکی از گزینه‌ها رو انتخاب کن:", inline)
	}
}
