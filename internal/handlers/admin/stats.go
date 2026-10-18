// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

// StatsHandler — آمار ربات
func StatsHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		s := db.GetStats()

		msg := fmt.Sprintf(
			"📊 آمار ربات\n\n"+
				"👥 کاربران:\n"+
				"   کل: %d\n"+
				"   امروز: %d\n"+
				"   هفته: %d\n"+
				"   ماه: %d\n\n"+
				"💵 موجودی:\n"+
				"   مجموع: <b>%.2f دلار</b>\n"+
				"   میانگین: <b>%.2f دلار</b>\n\n"+
				"📈 تراکنش‌ها:\n"+
				"   ✅ موفق: %d\n"+
				"   ⏳ در انتظار: %d\n"+
				"   ❌ لغو: %d\n"+
				"   ⌛ منقضی: %d\n\n"+
				"💰 مجموع شارژ موفق: <b>%.2f دلار</b>\n\n"+
				"🎫 تیکت‌های باز: <b>%d</b>",
			s.TotalUsers,
			s.ActiveToday, s.ActiveWeek, s.ActiveMonth,
			s.TotalBalance, s.AvgBalance,
			s.PaidTx, s.PendingTx, s.CanceledTx, s.ExpiredTx,
			s.TotalCharged,
			s.OpenTickets,
		)

		return c.Send(msg, &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		})
	}
}
