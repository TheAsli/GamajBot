// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/rate"

	AsliBot "gopkg.in/telebot.v4"
)

func AnalyticsHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		cur := db.GetCurrency()
		daily, _ := db.GetDailyRevenue(7)
		monthly, _ := db.GetMonthlyComparison()
		byGateway, _ := db.GetRevenueByGateway(30)

		var sb strings.Builder
		sb.WriteString("📈 <b>آمار پیشرفته</b>\n\n")

		sb.WriteString("📊 <b>درآمد ۷ روز اخیر:</b>\n")
		maxRevenue := 0.0
		for _, d := range daily {
			if d.Revenue > maxRevenue {
				maxRevenue = d.Revenue
			}
		}
		for _, d := range daily {
			bar := buildBar(d.Revenue, maxRevenue, 10)
			sb.WriteString(fmt.Sprintf("   %s | %s %s (%d سفارش)\n",
				d.Date[5:], bar, rate.Display(d.Revenue, cur), d.Orders))
		}

		sb.WriteString("\n📅 <b>مقایسه ماهانه:</b>\n")
		sb.WriteString(fmt.Sprintf(
			"   این ماه: <b>%s</b> (%d سفارش)\n"+
				"   ماه قبل: %s (%d سفارش)\n",
			rate.Display(monthly.CurrentMonthRevenue, cur), monthly.CurrentMonthOrders,
			rate.Display(monthly.PreviousMonthRevenue, cur), monthly.PreviousMonthOrders,
		))

		growthIcon := "📈"
		if monthly.GrowthPercent < 0 {
			growthIcon = "📉"
		}
		sb.WriteString(fmt.Sprintf("   %s رشد: <b>%.1f%%</b>\n", growthIcon, monthly.GrowthPercent))

		if len(byGateway) > 0 {
			sb.WriteString("\n🚪 <b>درآمد ۳۰ روز اخیر به تفکیک درگاه:</b>\n")
			for gw, amount := range byGateway {
				sb.WriteString(fmt.Sprintf("   %s: %s\n", gatewayLabel(gw), rate.Display(amount, cur)))
			}
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("📄 گزارش متنی کامل", "adm_analytics_report"),
			),
			inline.Row(
				inline.Data("⬅️ برگشت", "adm_settings"),
			),
		)

		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func buildBar(value, max float64, width int) string {
	if max <= 0 {
		return strings.Repeat("░", width)
	}
	filled := int((value / max) * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func gatewayLabel(g string) string {
	switch g {
	case "tetraminator":
		return "💎 Tetraminator"
	case "cubepay":
		return "🧊 CubePay"
	case "abangateway":
		return "🌐 AbanGateway"
	case "manual_card":
		return "💳 کارت دستی"
	case "manual_charge":
		return "🎁 شارژ دستی"
	}
	return "—"
}

func AnalyticsReportHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		_ = c.Respond()

		cur := db.GetCurrency()
		daily, _ := db.GetDailyRevenue(30)
		monthly, _ := db.GetMonthlyComparison()
		stats := db.GetStats()
		byGateway, _ := db.GetRevenueByGateway(30)
		topProducts := db.GetTopProducts(time.Now().AddDate(0, -1, 0), 10)

		var sb strings.Builder
		sb.WriteString("═══════════════════════════════\n")
		sb.WriteString("        گزارش کامل ربات\n")
		sb.WriteString("═══════════════════════════════\n\n")

		sb.WriteString(fmt.Sprintf("🕐 تاریخ گزارش: %s\n\n", time.Now().Format("2006/01/02 15:04")))

		sb.WriteString("📊 آمار کلی\n")
		sb.WriteString("───────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("👥 کاربران: %d\n", stats.TotalUsers))
		sb.WriteString(fmt.Sprintf("🟢 فعال امروز: %d\n", stats.ActiveToday))
		sb.WriteString(fmt.Sprintf("📅 فعال هفته: %d\n", stats.ActiveWeek))
		sb.WriteString(fmt.Sprintf("📅 فعال ماه: %d\n", stats.ActiveMonth))
		sb.WriteString(fmt.Sprintf("💵 موجودی کل: %s\n", rate.Display(stats.TotalBalance, cur)))
		sb.WriteString(fmt.Sprintf("💵 میانگین: %s\n\n", rate.Display(stats.AvgBalance, cur)))

		sb.WriteString("💰 درآمد\n")
		sb.WriteString("───────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("این ماه: %s\n", rate.Display(monthly.CurrentMonthRevenue, cur)))
		sb.WriteString(fmt.Sprintf("ماه قبل: %s\n", rate.Display(monthly.PreviousMonthRevenue, cur)))
		sb.WriteString(fmt.Sprintf("رشد: %.1f%%\n", monthly.GrowthPercent))
		sb.WriteString(fmt.Sprintf("کل شارژ موفق: %s\n\n", rate.Display(stats.TotalCharged, cur)))

		sb.WriteString("📦 تراکنش‌ها\n")
		sb.WriteString("───────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("✅ موفق: %d\n", stats.PaidTx))
		sb.WriteString(fmt.Sprintf("⏳ در انتظار: %d\n", stats.PendingTx))
		sb.WriteString(fmt.Sprintf("❌ لغو: %d\n", stats.CanceledTx))
		sb.WriteString(fmt.Sprintf("⌛ منقضی: %d\n\n", stats.ExpiredTx))

		if len(byGateway) > 0 {
			sb.WriteString("🚪 درآمد به تفکیک درگاه\n")
			sb.WriteString("───────────────────────────────\n")
			for gw, amount := range byGateway {
				sb.WriteString(fmt.Sprintf("%s: %s\n", gatewayLabel(gw), rate.Display(amount, cur)))
			}
			sb.WriteString("\n")
		}

		if len(topProducts) > 0 {
			sb.WriteString("🏆 پرفروش‌ترین محصولات\n")
			sb.WriteString("───────────────────────────────\n")
			for i, p := range topProducts {
				sb.WriteString(fmt.Sprintf("%d. %s — %s (%d سفارش)\n",
					i+1, p.ProductName, rate.Display(p.TotalRevenue, cur), p.OrderCount))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("📅 درآمد ۳۰ روز اخیر\n")
		sb.WriteString("───────────────────────────────\n")
		for _, d := range daily {
			if d.Revenue > 0 {
				sb.WriteString(fmt.Sprintf("%s: %s (%d سفارش)\n",
					d.Date, rate.Display(d.Revenue, cur), d.Orders))
			}
		}

		sb.WriteString("\n═══════════════════════════════\n")
		sb.WriteString("    پایان گزارش\n")
		sb.WriteString("═══════════════════════════════")

		content := []byte(sb.String())
		doc := &AsliBot.Document{
			File:     AsliBot.FromReader(bytes.NewReader(content)),
			FileName: "report_" + time.Now().Format("20060102_1504") + ".txt",
			Caption:  "📄 گزارش کامل ربات",
		}

		return c.Send(doc, &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}
