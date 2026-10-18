// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"log"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/xui"

	AsliBot "gopkg.in/telebot.v4"
)

func StartExpiryNotifier(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	go func() {
		time.Sleep(30 * time.Second)
		runExpiryCheck(bot, cfg, db)

		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			runExpiryCheck(bot, cfg, db)
		}
	}()
}

func runExpiryCheck(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	orders := db.GetAllOrders()
	now := time.Now()

	for _, o := range orders {
		if o.Status != database.OrderDelivered {
			continue
		}

		product := db.GetProduct(o.ProductID)
		if product == nil || product.ProductType != "vpn" {
			continue
		}
		if o.ClientEmail == "" {
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

		if len(product.XUIInboundIDs) == 0 {
			continue
		}

		stats, err := cli.GetClientTraffic(product.XUIInboundIDs[0], o.ClientEmail)
		if err != nil || stats == nil {
			log.Printf("expiry check %s: %v", o.ClientEmail, err)
			continue
		}

		if stats.ExpiryTime <= 0 {
			continue
		}

		expiry := time.UnixMilli(stats.ExpiryTime)
		daysLeft := int(expiry.Sub(now).Hours() / 24)

		if daysLeft < 0 {
			continue
		}

		if daysLeft > 3 {
			continue
		}

		var daysStr string
		if daysLeft == 0 {
			daysStr = "امروز"
		} else {
			daysStr = fmt.Sprintf("%d روز دیگر", daysLeft)
		}

		cur := db.GetCurrency()
		msg := fmt.Sprintf(
			"⚠️ <b>سرویس VPN شما به‌زودی منقضی می‌شود</b>\n\n"+
				"📦 %s\n"+
				"⏰ انقضا: <b>%s</b> (%s)\n\n"+
				"برای تمدید، از «📦 سرویس‌های من» استفاده کن.",
			o.ProductName, expiry.Format("2006/01/02 15:04"), daysStr,
		)

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("🔁 تمدید", fmt.Sprintf("svc_renew_%s", o.ID)),
				inline.Data("📦 سرویس‌ها", "svc_list"),
			),
		)

		_, _ = bot.Send(AsliBot.ChatID(o.UserID), msg,
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)

		_ = cur
	}
}
