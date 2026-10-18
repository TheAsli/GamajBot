// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/menus"
	"github.com/theasli/GamajBot/internal/rate"
	"github.com/theasli/GamajBot/internal/vpn"
	"github.com/theasli/GamajBot/internal/xui"
	AsliBot "gopkg.in/telebot.v4"
)

func SubscriptionHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		return showSubscriptionRoot(c, db)
	}
}

func SubscriptionRootHandler(c AsliBot.Context) error {
	_ = c.Respond()
	return showSubscriptionRoot(c, walletDB)
}

func showSubscriptionRoot(c AsliBot.Context, db *database.DB) error {
	plans, _ := db.GetActiveSubscriptionPlans()
	prices, _ := db.GetDynamicPrices()
	cur := db.GetCurrency()

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	for _, p := range plans {
		if p.Type != database.SubTypeFixed {
			continue
		}
		label := fmt.Sprintf("%s | %dGB | %dروز | %s",
			p.Name, p.TrafficGB, p.DurationDays, rate.Display(p.PriceUSD, cur))
		rows = append(rows, inline.Row(
			inline.Data(label, fmt.Sprintf("sub_fixed_%d", p.ID)),
		))
	}

	if len(prices) > 0 {
		rows = append(rows, inline.Row(
			inline.Data("🎛 پلن دلخواه (داینامیک)", "sub_dynamic"),
		))
	}

	if len(rows) == 0 {
		return c.Send("🏠 در حال حاضر پلنی برای فروش موجود نیست.", menus.MainMenu)
	}

	inline.Inline(rows...)
	return c.Send("🏠 <b>خرید اشتراک</b>\n\nیک پلن انتخاب کن:", &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func SubscriptionFixedHandler(cfg *config.Config, db *database.DB, planID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		p, err := db.GetSubscriptionPlan(planID)
		if err != nil || !p.Active || p.Type != database.SubTypeFixed {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ پلن پیدا نشد"})
		}
		_ = c.Respond()

		u, _ := db.GetUser(c.Sender().ID)
		balance := 0.0
		if u != nil {
			balance = u.BalanceUSD
		}
		cur := db.GetCurrency()

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("🛒 خرید", fmt.Sprintf("sub_buy_fixed_%d", p.ID)),
				inline.Data("⬅️ برگشت", "sub_root"),
			),
		)

		return c.Send(fmt.Sprintf(
			"🏠 <b>%s</b>\n\n"+
				"📝 %s\n\n"+
				"📊 ترافیک: %d GB\n"+
				"⏰ مدت: %d روز\n"+
				"👥 محدودیت IP: %d\n"+
				"💰 قیمت: <b>%s</b>\n"+
				"💵 موجودی شما: %s",
			p.Name, p.Description, p.TrafficGB, p.DurationDays, p.LimitIP,
			rate.Display(p.PriceUSD, cur), rate.Display(balance, cur),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func SubscriptionDynamicHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		prices, _ := db.GetDynamicPrices()
		if len(prices) == 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ هیچ قیمت داینامیکی تعریف نشده"})
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row
		cur := db.GetCurrency()

		for _, dp := range prices {
			label := fmt.Sprintf("%dGB | %dروز | %dIP | %s",
				dp.TrafficGB, dp.DurationDays, dp.LimitIP, rate.Display(dp.PriceUSD, cur))
			rows = append(rows, inline.Row(
				inline.Data(label, fmt.Sprintf("sub_buy_dynamic_%d", dp.ID)),
			))
		}
		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", "sub_root"),
		))

		inline.Inline(rows...)
		return c.Send("🎛 <b>پلن دلخواه</b>\n\nیکی از ترکیب‌ها را انتخاب کن:", &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func SubscriptionBuyFixedHandler(cfg *config.Config, db *database.DB, planID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID

		p, err := db.GetSubscriptionPlan(planID)
		if err != nil || !p.Active {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ پلن پیدا نشد"})
		}

		u, err := db.GetUser(uid)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ حساب پیدا نشد"})
		}
		if u.BalanceUSD < p.PriceUSD {
			cur := db.GetCurrency()
			return c.Respond(&AsliBot.CallbackResponse{
				Text: fmt.Sprintf("❌ موجودی کافی نیست. نیاز: %s", rate.Display(p.PriceUSD, cur)),
			})
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳ در حال خرید..."})

		order, newBal, err := db.CreatePlanOrder(uid, c.Sender().FirstName, p.Name, p.PriceUSD, p.TrafficGB, p.DurationDays, p.LimitIP, p.XUIPanelID, p.XUIInboundIDs)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}

		if err := provisionPlanVPN(db, order, p.TrafficGB, p.DurationDays, p.XUIPanelID, p.XUIInboundIDs); err != nil {
			_, _, _ = db.RejectOrder(order.ID, "خطا در ساخت: "+err.Error())
			cur := db.GetCurrency()
			return c.Send(fmt.Sprintf(
				"⚠️ <b>ساخت ناموفق بود</b>\n\n💰 %s به کیف پول بازگشت.\n📝 خطا: %s",
				rate.Display(p.PriceUSD, cur), err.Error(),
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
		}

		if botInstance != nil {
			notifyAdminsNewOrder(c, cfg, db, order)
		}

		cur := db.GetCurrency()
		linksText := strings.Join(order.ClientLinks, "\n")
		return c.Send(fmt.Sprintf(
			"✅ <b>خرید موفق</b>\n\n"+
				"🆔 <code>%s</code>\n"+
				"📦 %s\n"+
				"💰 %s\n"+
				"💵 موجودی جدید: <b>%s</b>\n\n"+
				"🔑 لینک:\n<code>%s</code>",
			order.ID, order.ProductName,
			rate.Display(p.PriceUSD, cur),
			rate.Display(newBal, cur),
			linksText,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
	}
}

func SubscriptionBuyDynamicHandler(cfg *config.Config, db *database.DB, priceID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID

		dp, err := db.GetDynamicPrice(priceID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ قیمت پیدا نشد"})
		}

		u, err := db.GetUser(uid)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ حساب پیدا نشد"})
		}
		if u.BalanceUSD < dp.PriceUSD {
			cur := db.GetCurrency()
			return c.Respond(&AsliBot.CallbackResponse{
				Text: fmt.Sprintf("❌ موجودی کافی نیست. نیاز: %s", rate.Display(dp.PriceUSD, cur)),
			})
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳ در حال خرید..."})

		planName := fmt.Sprintf("پلن %dGB / %d روز / %d IP", dp.TrafficGB, dp.DurationDays, dp.LimitIP)
		order, newBal, err := db.CreatePlanOrder(uid, c.Sender().FirstName, planName, dp.PriceUSD, dp.TrafficGB, dp.DurationDays, dp.LimitIP, dp.XUIPanelID, dp.XUIInboundIDs)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}

		if err := provisionPlanVPN(db, order, dp.TrafficGB, dp.DurationDays, dp.XUIPanelID, dp.XUIInboundIDs); err != nil {
			_, _, _ = db.RejectOrder(order.ID, "خطا در ساخت: "+err.Error())
			cur := db.GetCurrency()
			return c.Send(fmt.Sprintf(
				"⚠️ <b>ساخت ناموفق بود</b>\n\n💰 %s به کیف پول بازگشت.\n📝 خطا: %s",
				rate.Display(dp.PriceUSD, cur), err.Error(),
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
		}

		if botInstance != nil {
			notifyAdminsNewOrder(c, cfg, db, order)
		}

		cur := db.GetCurrency()
		linksText := strings.Join(order.ClientLinks, "\n")
		return c.Send(fmt.Sprintf(
			"✅ <b>خرید موفق</b>\n\n"+
				"🆔 <code>%s</code>\n"+
				"📦 %s\n"+
				"💰 %s\n"+
				"💵 موجودی جدید: <b>%s</b>\n\n"+
				"🔑 لینک:\n<code>%s</code>",
			order.ID, order.ProductName,
			rate.Display(dp.PriceUSD, cur),
			rate.Display(newBal, cur),
			linksText,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
	}
}

func SubscriptionCategoryHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		cat := db.GetCategory(catID)
		if cat == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دسته پیدا نشد"})
		}
		_ = c.Respond()

		subs := db.GetCategoriesByParent(catID)
		if len(subs) > 0 {
			inline := &AsliBot.ReplyMarkup{}
			var rows []AsliBot.Row
			for _, sub := range subs {
				rows = append(rows, inline.Row(
					inline.Data(fmt.Sprintf("%s %s", sub.Emoji, sub.Name), fmt.Sprintf("sub_cat_%d", sub.ID)),
				))
			}
			rows = append(rows, inline.Row(
				inline.Data("⬅️ برگشت", "sub_root"),
			))
			inline.Inline(rows...)
			return c.Send(fmt.Sprintf("🏠 <b>%s %s</b>\n\nیک زیردسته را انتخاب کن:", cat.Emoji, cat.Name), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
		}

		products := db.GetProductsByCategory(catID)
		var vpnProducts []*database.Product
		for _, p := range products {
			if p.ProductType == "vpn" && p.Active {
				vpnProducts = append(vpnProducts, p)
			}
		}
		if len(vpnProducts) == 0 {
			return c.Send("📭 در این دسته پلن VPN موجود نیست.", menus.MainMenu)
		}

		cur := db.GetCurrency()
		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row
		for _, p := range vpnProducts {
			status := "✅"
			if !p.Available {
				status = "❌"
			}
			label := fmt.Sprintf("%s %s | %s | %dGB / %dروز",
				status, p.Name, rate.Display(p.PriceUSD, cur), p.TrafficGB, p.DurationDays)
			rows = append(rows, inline.Row(
				inline.Data(label, fmt.Sprintf("sub_prod_%d", p.ID)),
			))
		}
		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", "sub_root"),
		))
		inline.Inline(rows...)
		return c.Send(fmt.Sprintf("🏠 <b>%s %s</b>\n\nیک پلن انتخاب کن:", cat.Emoji, cat.Name), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func SubscriptionProductHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		p := db.GetProduct(productID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ محصول پیدا نشد"})
		}
		if p.ProductType != "vpn" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ این محصول VPN نیست"})
		}
		_ = c.Respond()

		if !p.Active {
			return c.Send("❌ این پلن غیرفعال است.", menus.MainMenu)
		}

		u, _ := db.GetUser(c.Sender().ID)
		balance := 0.0
		if u != nil {
			balance = u.BalanceUSD
		}

		cur := db.GetCurrency()
		price := rate.Display(p.PriceUSD, cur)
		bal := rate.Display(balance, cur)

		status := "✅ موجود"
		if !p.Available {
			status = "❌ ناموجود"
		}

		panel := db.GetXUIPanel(p.XUIPanelID)
		panelName := "—"
		if panel != nil {
			panelName = panel.Name
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"🏠 <b>%s</b>\n\n"+
				"📝 %s\n\n"+
				"💰 قیمت: <b>%s</b>\n"+
				"🌐 ترافیک: <b>%d GB</b>\n"+
				"⏰ مدت: <b>%d روز</b>\n"+
				"🖥 سرور: %s\n"+
				"📊 وضعیت: %s\n"+
				"💵 موجودی شما: <b>%s</b>",
			p.Name, p.Description, price, p.TrafficGB, p.DurationDays, panelName, status, bal,
		))

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		if p.Available {
			rows = append(rows, inline.Row(
				inline.Data("🛒 خرید اشتراک", fmt.Sprintf("shop_buy_%d", p.ID)),
			))
		} else {
			rows = append(rows, inline.Row(
				inline.Data("❌ ناموجود", "shop_noop"),
			))
		}

		cat := db.GetCategory(p.CategoryID)
		if cat != nil && cat.ParentID != 0 {
			rows = append(rows, inline.Row(
				inline.Data("⬅️ برگشت", fmt.Sprintf("sub_cat_%d", cat.ParentID)),
			))
		} else {
			rows = append(rows, inline.Row(
				inline.Data("⬅️ برگشت", "sub_root"),
			))
		}

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func provisionPlanVPN(db *database.DB, order *database.Order, trafficGB int64, durationDays int, panelID int64, inboundIDs []int64) error {
	panel := db.GetXUIPanel(panelID)
	if panel == nil {
		return fmt.Errorf("پنل پیدا نشد")
	}
	if !panel.Active {
		return fmt.Errorf("پنل غیرفعال است")
	}
	if len(inboundIDs) == 0 {
		return fmt.Errorf("هیچ inbound انتخاب نشده")
	}

	cli := xui.New(xui.Config{
		BaseURL:    panel.BaseURL,
		AuthMethod: xui.AuthMethod(panel.AuthMethod),
		APIToken:   panel.APIToken,
		Username:   panel.Username,
		Password:   panel.Password,
	})

	email := fmt.Sprintf("u%d_%d", order.UserID, time.Now().UnixNano())
	totalBytes := trafficGB * 1024 * 1024 * 1024
	expiryTime := time.Now().Add(time.Duration(durationDays) * 24 * time.Hour)

	client := vpn.ClientAddRequest{
		Email:      email,
		TotalGB:    totalBytes,
		ExpiryTime: expiryTime,
		Enable:     true,
		InboundIDs: inboundIDs,
	}

	if err := cli.AddClient(client); err != nil {
		return fmt.Errorf("خطا در ساخت کلاینت: %w", err)
	}

	links, err := cli.GetClientLinks(email)
	if err != nil {
		return fmt.Errorf("خطا در دریافت لینک: %w", err)
	}

	order.ClientEmail = email
	order.ClientLinks = links
	order.Status = database.OrderDelivered
	order.DeliveredAt = time.Now().Format(time.RFC3339)
	order.ServiceInfo = fmt.Sprintf("Email: %s | Traffic: %d GB | %d days",
		email, trafficGB, durationDays)

	_ = db.UpdateOrder(order)
	return nil
}
