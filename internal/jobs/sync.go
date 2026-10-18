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

func StartSyncJob(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	go func() {
		time.Sleep(5 * time.Minute)
		runSync(bot, cfg, db)

		ticker := time.NewTicker(6 * time.Hour)
		for range ticker.C {
			runSync(bot, cfg, db)
		}
	}()
}

func runSync(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	orders := db.GetAllOrders()
	checked := 0
	deleted := 0

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

		_, err := cli.GetClient(product.XUIInboundIDs[0], o.ClientEmail)
		if err == nil {
			checked++
			continue
		}

		deleted++
		o.Status = database.OrderRejected
		o.RejectNote = "کلاینت در 3x-ui حذف شد (sync خودکار)"
		_ = db.UpdateOrder(o)

		notify.AdminLog(bot, cfg, db, notify.LogTopicXUI,
			fmt.Sprintf(
				"⚠️ <b>کلاینت حذف‌شده در 3x-ui</b>\n\n"+
					"🆔 سفارش: <code>%s</code>\n"+
					"👤 کاربر: %s (<code>%d</code>)\n"+
					"📦 محصول: %s\n"+
					"📧 Email: <code>%s</code>\n\n"+
					"وضعیت سفارش به «رد شده» تغییر کرد.",
				o.ID, o.UserName, o.UserID, o.ProductName, o.ClientEmail,
			))
	}

	log.Printf("✅ sync کامل شد: %d چک شد، %d حذف‌شده", checked, deleted)
}
