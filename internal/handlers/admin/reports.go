// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"strings"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/rate"

	AsliBot "gopkg.in/telebot.v4"
)

func ReportsHandler(cfg *config.Config, db *database.DB, period string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		_ = c.Respond()

		since := periodToSince(period)
		stats := db.GetSalesStats(since)
		cur := db.GetCurrency()

		var periodLabel string
		switch period {
		case "today":
			periodLabel = "📅 امروز"
		case "week":
			periodLabel = "📅 هفته اخیر"
		case "month":
			periodLabel = "📅 ماه اخیر"
		case "all":
			periodLabel = "📅 از ابتدا"
		default:
			period = "all"
			periodLabel = "📅 از ابتدا"
		}

		msg := fmt.Sprintf(
			"📊 <b>گزارش فروش</b> — %s\n\n"+
				"📦 <b>سفارش‌ها:</b>\n"+
				"   کل: <b>%d</b>\n"+
				"   ⏳ در انتظار: %d\n"+
				"   ✅ تحویل شده: %d\n"+
				"   ❌ رد شده: %d\n\n"+
				"💰 <b>درآمد:</b>\n"+
				"   درآمد کل: <b>%s</b>\n"+
				"   بازگشتی: %s\n"+
				"   خالص: <b>%s</b>",
			periodLabel,
			stats.TotalOrders,
			stats.PendingOrders,
			stats.DeliveredOrders,
			stats.RejectedOrders,
			rate.Display(stats.TotalRevenue, cur),
			rate.Display(stats.RefundedAmount, cur),
			rate.Display(stats.NetRevenue, cur),
		)

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data(periodBtn("امروز", period, "today"), "adm_report_today"),
				inline.Data(periodBtn("هفته", period, "week"), "adm_report_week"),
			),
			inline.Row(
				inline.Data(periodBtn("ماه", period, "month"), "adm_report_month"),
				inline.Data(periodBtn("کل", period, "all"), "adm_report_all"),
			),
			inline.Row(
				inline.Data("🏆 پرفروش‌ترین محصولات", "adm_report_top_"+period),
			),
			inline.Row(
				inline.Data("⬅️ برگشت", "adm_settings"),
			),
		)

		return c.Send(msg, &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func TopProductsHandler(cfg *config.Config, db *database.DB, period string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		_ = c.Respond()

		since := periodToSince(period)
		top := db.GetTopProducts(since, 10)
		cur := db.GetCurrency()

		if len(top) == 0 {
			return c.Send("🏆 هیچ فروشی در این بازه ثبت نشده.")
		}

		var sb strings.Builder
		sb.WriteString("🏆 <b>پرفروش‌ترین محصولات</b>\n\n")

		for i, p := range top {
			sb.WriteString(fmt.Sprintf(
				"%d. <b>%s</b>\n"+
					"   📦 %d سفارش | ✅ %d تحویل\n"+
					"   💰 %s\n\n",
				i+1, p.ProductName,
				p.OrderCount, p.Delivered,
				rate.Display(p.TotalRevenue, cur),
			))
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("⬅️ برگشت به گزارش", "adm_report_"+period),
			),
		)

		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func periodToSince(period string) time.Time {
	now := time.Now()
	switch period {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "all":
		return time.Time{}
	}
	return time.Time{}
}

func periodBtn(label, current, target string) string {
	if current == target {
		return "• " + label + " •"
	}
	return label
}
