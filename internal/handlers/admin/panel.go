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

func PanelHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role == "" {
			return c.Send("⛔ دسترسی نداری.")
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		switch role {
		case database.RoleSuperAdmin:
			rows = append(rows, inline.Row(
				inline.Data("📊 آمار", "adm_stats"),
				inline.Data("👥 کاربران", "adm_users"),
			))
			rows = append(rows, inline.Row(
				inline.Data("💰 مالی", "adm_finance"),
				inline.Data("🎫 تیکت‌ها", "adm_tickets"),
			))
			rows = append(rows, inline.Row(
				inline.Data("📦 سایر محصولات", "adm_products"),
				inline.Data("🧾 سفارش‌ها", "adm_orders"),
			))
			rows = append(rows, inline.Row(
				inline.Data("📈 گزارش فروش", "adm_report_all"),
				inline.Data("🖥 پنل‌های 3x-ui", "adm_xui_panels"),
			))
			rows = append(rows, inline.Row(
				inline.Data("👮 ادمین‌ها", "adm_admins"),
				inline.Data("🏢 دپارتمان‌ها", "adm_depts"),
			))
			rows = append(rows, inline.Row(
				inline.Data("⚙️ تنظیمات", "adm_settings"),
				inline.Data("📢 پیام همگانی", "adm_broadcast"),
			))
			rows = append(rows, inline.Row(
				inline.Data("📝 لاگ گروه", "adm_log_setup"),
			))

		case database.RoleFinance:
			rows = append(rows, inline.Row(
				inline.Data("📊 آمار", "adm_stats"),
				inline.Data("💰 مالی", "adm_finance"),
			))
			rows = append(rows, inline.Row(
				inline.Data("👥 کاربران", "adm_users"),
				inline.Data("📦 سایر محصولات", "adm_products"),
			))
			rows = append(rows, inline.Row(
				inline.Data("🧾 سفارش‌ها", "adm_orders"),
			))
			rows = append(rows, inline.Row(
				inline.Data("📈 گزارش فروش", "adm_report_all"),
			))
			rows = append(rows, inline.Row(inline.Data("⚙️ تنظیمات مالی", "adm_settings")))

		case database.RoleSupport:
			rows = append(rows, inline.Row(
				inline.Data("📊 آمار", "adm_stats"),
				inline.Data("🎫 تیکت‌ها", "adm_tickets"),
			))
			rows = append(rows, inline.Row(
				inline.Data("👥 کاربران", "adm_users"),
				inline.Data("📦 سایر محصولات", "adm_products"),
			))
			rows = append(rows, inline.Row(
				inline.Data("🧾 سفارش‌ها", "adm_orders"),
				inline.Data("💰 مالی", "adm_finance"),
			))

		case database.RoleViewer:
			rows = append(rows, inline.Row(inline.Data("📊 آمار", "adm_stats")))
		}

		inline.Inline(rows...)
		return c.Send(fmt.Sprintf(
			"🎛️ پنل مدیریت\n\n👤 نقش شما: <b>%s</b>\n\nیه گزینه رو انتخاب کن:",
			roleLabel(role),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}
