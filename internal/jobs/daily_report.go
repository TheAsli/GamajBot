// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/notify"
	"github.com/theasli/GamajBot/internal/rate"

	AsliBot "gopkg.in/telebot.v4"
)

func StartDailyReportJob(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	go func() {
		time.Sleep(15 * time.Minute)
		runDailyReport(bot, cfg, db)

		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			runDailyReport(bot, cfg, db)
		}
	}()
}

func runDailyReport(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	since := time.Now().AddDate(0, 0, -1)
	stats := db.GetSalesStats(since)
	topProducts := db.GetTopProducts(since, 5)
	cur := db.GetCurrency()

	var sb string
	sb += "📊 <b>گزارش روزانه</b>\n\n"
	sb += fmt.Sprintf("🕐 %s\n\n", time.Now().Format("2006/01/02 15:04"))
	sb += "📦 <b>سفارش‌ها:</b>\n"
	sb += fmt.Sprintf("   کل: <b>%d</b>\n", stats.TotalOrders)
	sb += fmt.Sprintf("   ✅ تحویل شده: %d\n", stats.DeliveredOrders)
	sb += fmt.Sprintf("   ⏳ در انتظار: %d\n", stats.PendingOrders)
	sb += fmt.Sprintf("   ❌ رد شده: %d\n\n", stats.RejectedOrders)
	sb += "💰 <b>درآمد:</b>\n"
	sb += fmt.Sprintf("   درآمد کل: <b>%s</b>\n", rate.Display(stats.TotalRevenue, cur))
	sb += fmt.Sprintf("   بازگشتی: %s\n", rate.Display(stats.RefundedAmount, cur))
	sb += fmt.Sprintf("   خالص: <b>%s</b>\n\n", rate.Display(stats.NetRevenue, cur))

	if len(topProducts) > 0 {
		sb += "🏆 <b>پرفروش‌ترین‌ها:</b>\n"
		for i, p := range topProducts {
			sb += fmt.Sprintf("%d. %s — %s\n", i+1, p.ProductName, rate.Display(p.TotalRevenue, cur))
		}
		sb += "\n"
	}

	allStats := db.GetStats()
	sb += "📈 <b>آمار کلی:</b>\n"
	sb += fmt.Sprintf("   👥 کاربران: %d\n", allStats.TotalUsers)
	sb += fmt.Sprintf("   💵 موجودی کل: %s\n", rate.Display(allStats.TotalBalance, cur))
	sb += fmt.Sprintf("   🎫 تیکت‌های باز: %d\n", allStats.OpenTickets)

	notify.AdminLog(bot, cfg, db, notify.LogTopicDaily, sb)
	log.Printf("✅ گزارش روزانه ارسال شد")
}
