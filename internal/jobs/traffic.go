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
	"github.com/theasli/GamajBot/internal/xui"

	AsliBot "gopkg.in/telebot.v4"
)

func StartTrafficJob(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	go func() {
		time.Sleep(10 * time.Minute)
		runTrafficReport(bot, cfg, db)

		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			runTrafficReport(bot, cfg, db)
		}
	}()
}

func runTrafficReport(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	orders := db.GetAllOrders()

	totalUsers := 0
	totalUsedGB := 0.0
	highUsageUsers := 0

	for _, o := range orders {
		if o.Status != database.OrderDelivered {
			continue
		}
		if o.ClientEmail == "" {
			continue
		}

		product := db.GetProduct(o.ProductID)
		if product == nil || product.ProductType != "vpn" {
			continue
		}
		if len(product.XUIInboundIDs) == 0 {
			continue
		}

		panel := db.GetXUIPanel(product.XUIPanelID)
		if panel == nil || !panel.Active {
			continue
		}

		cli := xui.New(xui.Config{
			BaseURL:    panel.BaseURL,
			AuthMethod: xui.AuthMethod(panel.AuthMethod),
			APIToken:   panel.APIToken,
			Username:   panel.Username,
			Password:   panel.Password,
		})

		stats, err := cli.GetClientTraffic(product.XUIInboundIDs[0], o.ClientEmail)
		if err != nil {
			continue
		}

		usedGB := float64(stats.Up+stats.Down) / (1024 * 1024 * 1024)
		totalGB := float64(stats.TotalGB) / (1024 * 1024 * 1024)
		totalUsedGB += usedGB
		totalUsers++

		if totalGB > 0 && usedGB > totalGB*0.8 {
			highUsageUsers++
		}

		cur := db.GetCurrency()
		if totalGB > 0 && usedGB > totalGB*0.8 {
			msg := fmt.Sprintf(
				"⚠️ <b>هشدار مصرف</b>\n\n"+
					"📦 %s\n"+
					"📥 مصرف: <b>%.2f GB</b> از <b>%.2f GB</b>\n"+
					"📊 درصد: <b>%.0f%%</b>\n\n"+
					"برای تمدید از «📦 سرویس‌های من» استفاده کن.",
				o.ProductName, usedGB, totalGB, (usedGB/totalGB)*100,
			)
			inline := &AsliBot.ReplyMarkup{}
			inline.Inline(
				inline.Row(
					inline.Data("🔁 تمدید", fmt.Sprintf("svc_renew_%s", o.ID)),
				),
			)
			_, _ = bot.Send(AsliBot.ChatID(o.UserID), msg,
				&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
		}

		_ = cur
	}

	if totalUsers > 0 {
		summary := fmt.Sprintf(
			"📊 <b>گزارش ترافیک روزانه</b>\n\n"+
				"👥 کاربران فعال: <b>%d</b>\n"+
				"📥 مصرف کل: <b>%.2f GB</b>\n"+
				"⚠️ کاربران با مصرف > ۸۰٪: <b>%d</b>\n"+
				"🕐 %s",
			totalUsers, totalUsedGB, highUsageUsers, time.Now().Format("2006/01/02 15:04"),
		)
		notify.AdminLog(bot, cfg, db, notify.LogTopicDaily, summary)
	}

	log.Printf("✅ گزارش ترافیک ارسال شد: %d کاربر، %.2f GB", totalUsers, totalUsedGB)
}
