// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/i18n"
	"github.com/theasli/GamajBot/internal/menus"
	"github.com/theasli/GamajBot/internal/notify"
	"github.com/theasli/GamajBot/internal/rate"
	"github.com/theasli/GamajBot/internal/vpn"
	"github.com/theasli/GamajBot/internal/xui"

	AsliBot "gopkg.in/telebot.v4"
)

var (
	buyConfirm = make(map[int64]int64)
	buyMu      sync.Mutex
)

func setBuyConfirm(uid, productID int64) {
	buyMu.Lock()
	defer buyMu.Unlock()
	if productID == 0 {
		delete(buyConfirm, uid)
	} else {
		buyConfirm[uid] = productID
	}
}

func getBuyConfirm(uid int64) int64 {
	buyMu.Lock()
	defer buyMu.Unlock()
	return buyConfirm[uid]
}

func shopPrice(usd float64) string {
	return rate.Display(usd, walletDB.GetCurrency())
}

func ShopHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		return showShopCategories(c, db, 0)
	}
}

func showShopCategories(c AsliBot.Context, db *database.DB, parentID int64) error {
	lang := userLang(c.Sender().ID)
	cats := db.GetCategoriesByParent(parentID)

	var title string
	if parentID == 0 {
		title = i18n.T(lang, "shop_choose_cat")
	} else {
		parent := db.GetCategory(parentID)
		if parent == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "not_found")})
		}
		title = fmt.Sprintf(i18n.T(lang, "shop_choose_sub"), parent.Emoji, parent.Name)
	}

	if len(cats) == 0 {
		products := db.GetProductsByCategory(parentID)
		if len(products) == 0 {
			return c.Send(i18n.T(lang, "shop_empty"), menus.MainMenu)
		}
		return showShopProducts(c, db, parentID)
	}

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	for _, cat := range cats {
		label := fmt.Sprintf("%s %s", cat.Emoji, cat.Name)
		rows = append(rows, inline.Row(
			inline.Data(label, fmt.Sprintf("shop_cat_%d", cat.ID)),
		))
	}

	if parentID != 0 {
		parent := db.GetCategory(parentID)
		if parent != nil && parent.ParentID != 0 {
			rows = append(rows, inline.Row(
				inline.Data(i18n.T(lang, "back"), fmt.Sprintf("shop_cat_%d", parent.ParentID)),
			))
		} else {
			rows = append(rows, inline.Row(
				inline.Data(i18n.T(lang, "back"), "shop_root"),
			))
		}
	}

	inline.Inline(rows...)
	return c.Send(title, inline)
}

func ShopCategoryHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		cat := db.GetCategory(catID)
		if cat == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(userLang(c.Sender().ID), "not_found")})
		}
		_ = c.Respond()

		subs := db.GetCategoriesByParent(catID)
		if len(subs) > 0 {
			return showShopCategories(c, db, catID)
		}
		return showShopProducts(c, db, catID)
	}
}

func ShopRootHandler(c AsliBot.Context) error {
	_ = c.Respond()
	return showShopCategories(c, walletDB, 0)
}

func showShopProducts(c AsliBot.Context, db *database.DB, catID int64) error {
	lang := userLang(c.Sender().ID)
	cat := db.GetCategory(catID)
	if cat == nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "not_found")})
	}
	products := db.GetProductsByCategory(catID)

	if len(products) == 0 {
		return c.Send(i18n.T(lang, "shop_no_product"), menus.MainMenu)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(i18n.T(lang, "shop_products"), cat.Emoji, cat.Name))

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	for _, p := range products {
		status := "✅"
		if !p.Available {
			status = i18n.T(lang, "shop_status_no")
		}
		price := shopPrice(p.PriceUSD)
		sb.WriteString(fmt.Sprintf("%s <b>%s</b> — %s\n", status, p.Name, price))
		label := fmt.Sprintf("📦 %s | %s", p.Name, price)
		if !p.Available {
			label = "❌ " + label
		}
		rows = append(rows, inline.Row(
			inline.Data(label, fmt.Sprintf("shop_prod_%d", p.ID)),
		))
	}

	if cat.ParentID == 0 {
		rows = append(rows, inline.Row(inline.Data(i18n.T(lang, "back"), "shop_root")))
	} else {
		rows = append(rows, inline.Row(
			inline.Data(i18n.T(lang, "back"), fmt.Sprintf("shop_cat_%d", cat.ParentID)),
		))
	}

	inline.Inline(rows...)
	return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func ShopProductHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		lang := userLang(c.Sender().ID)
		p := db.GetProduct(productID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "not_found")})
		}
		_ = c.Respond()

		if !p.Active {
			return c.Send(i18n.T(lang, "shop_unavailable"), menus.MainMenu)
		}

		status := i18n.T(lang, "shop_status_ok")
		if !p.Available {
			status = i18n.T(lang, "shop_status_no")
		}

		u, _ := db.GetUser(c.Sender().ID)
		balance := 0.0
		if u != nil {
			balance = u.BalanceUSD
		}

		price := shopPrice(p.PriceUSD)
		balanceDisplay := rate.Display(balance, db.GetCurrency())

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"📦 <b>%s</b>\n\n📝 %s\n\n%s\n%s\n%s",
			p.Name, p.Description,
			fmt.Sprintf(i18n.T(lang, "shop_price"), price),
			fmt.Sprintf("📊 %s", status),
			fmt.Sprintf(i18n.T(lang, "shop_balance"), balanceDisplay),
		))

		if p.ProductType == "vpn" {
			sb.WriteString(fmt.Sprintf(
				"\n\n🌐 ترافیک: <b>%d GB</b>\n⏰ مدت: <b>%d روز</b>",
				p.TrafficGB, p.DurationDays,
			))
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		if p.Available {
			rows = append(rows, inline.Row(
				inline.Data(i18n.T(lang, "shop_buy"), fmt.Sprintf("shop_buy_%d", p.ID)),
			))
		} else {
			rows = append(rows, inline.Row(
				inline.Data(i18n.T(lang, "shop_unavailable"), "shop_noop"),
			))
		}

		cat := db.GetCategory(p.CategoryID)
		if cat != nil && cat.ParentID != 0 {
			rows = append(rows, inline.Row(
				inline.Data(i18n.T(lang, "back"), fmt.Sprintf("shop_cat_%d", cat.ParentID)),
			))
		} else {
			rows = append(rows, inline.Row(
				inline.Data(i18n.T(lang, "back"), "shop_root"),
			))
		}

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func ShopBuyHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		lang := userLang(c.Sender().ID)
		p := db.GetProduct(productID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "not_found")})
		}
		if !p.Active || !p.Available {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "shop_out_of_stock")})
		}

		u, err := db.GetUser(c.Sender().ID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "not_found")})
		}
		if u.BalanceUSD < p.PriceUSD {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "shop_insufficient")})
		}

		setBuyConfirm(c.Sender().ID, productID)
		_ = c.Respond()

		cur := db.GetCurrency()
		price := rate.Display(p.PriceUSD, cur)
		bal := rate.Display(u.BalanceUSD, cur)
		after := rate.Display(u.BalanceUSD-p.PriceUSD, cur)

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data(i18n.T(lang, "shop_confirm_btn"), fmt.Sprintf("shop_buy_confirm_%d", p.ID)),
				inline.Data(i18n.T(lang, "shop_buy_cancel"), "shop_buy_cancel"),
			),
		)

		return c.Send(fmt.Sprintf(
			i18n.T(lang, "shop_confirm"),
			p.Name, price, bal, after,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func ShopBuyConfirmHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		lang := userLang(uid)
		if getBuyConfirm(uid) != productID {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "invalid_input")})
		}
		setBuyConfirm(uid, 0)

		name := c.Sender().FirstName
		order, newBal, err := db.CreateOrder(uid, name, productID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
		}

		product := db.GetProduct(productID)

		if product != nil && product.ProductType == "vpn" {
			if err := provisionVPN(c, db, order, product); err != nil {
				_, _, _ = db.RejectOrder(order.ID, "خطا در ساخت خودکار: "+err.Error())
				_ = c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error")})
				_ = c.Delete()
				cur := db.GetCurrency()
				return c.Send(fmt.Sprintf(
					i18n.T(lang, "shop_purchase_error"),
					rate.Display(order.PriceUSD, cur), err.Error(),
				), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
			}

			_ = c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "success")})
			_ = c.Delete()

			notifyAdminsNewOrder(c, cfg, db, order)

			cur := db.GetCurrency()
			linksText := strings.Join(order.ClientLinks, "\n")
			return c.Send(fmt.Sprintf(
				i18n.T(lang, "shop_vpn_success"),
				order.ID, order.ProductName,
				rate.Display(order.PriceUSD, cur),
				rate.Display(newBal, cur),
				linksText,
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "success")})
		_ = c.Delete()

		notifyAdminsNewOrder(c, cfg, db, order)

		cur := db.GetCurrency()
		return c.Send(fmt.Sprintf(
			i18n.T(lang, "shop_success"),
			order.ID, order.ProductName,
			rate.Display(order.PriceUSD, cur),
			rate.Display(newBal, cur),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
	}
}

func provisionVPN(c AsliBot.Context, db *database.DB, order *database.Order, product *database.Product) error {
	panel := db.GetXUIPanel(product.XUIPanelID)
	if panel == nil {
		return fmt.Errorf("پنل پیدا نشد")
	}
	if !panel.Active {
		return fmt.Errorf("پنل غیرفعال است")
	}
	if len(product.XUIInboundIDs) == 0 {
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
	totalBytes := product.TrafficGB * 1024 * 1024 * 1024
	expiryTime := time.Now().Add(time.Duration(product.DurationDays) * 24 * time.Hour)

	client := vpn.ClientAddRequest{
		Email:      email,
		TotalGB:    totalBytes,
		ExpiryTime: expiryTime,
		Enable:     true,
		InboundIDs: product.XUIInboundIDs,
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
		email, product.TrafficGB, product.DurationDays)

	_ = db.UpdateOrder(order)
	return nil
}

func ShopBuyCancelHandler(c AsliBot.Context) error {
	setBuyConfirm(c.Sender().ID, 0)
	_ = c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(userLang(c.Sender().ID), "canceled")})
	_ = c.Delete()
	return c.Send(i18n.T(userLang(c.Sender().ID), "canceled"), menus.MainMenu)
}

func ShopNoopHandler(c AsliBot.Context) error {
	return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(userLang(c.Sender().ID), "shop_out_of_stock")})
}

func notifyAdminsNewOrder(c AsliBot.Context, cfg *config.Config, db *database.DB, order *database.Order) {
	inlinePrivate := &AsliBot.ReplyMarkup{}
	btnDeliver := inlinePrivate.Data("📤 تحویل", "adm_order_deliver_"+order.ID)
	btnReject := inlinePrivate.Data("❌ رد", "adm_order_reject_"+order.ID)
	inlinePrivate.Inline(inlinePrivate.Row(btnDeliver, btnReject))

	inlineLog := &AsliBot.ReplyMarkup{}
	btnDeliverLog := inlineLog.Data("📤 تحویل", "adm_order_deliver_"+order.ID)
	btnRejectLog := inlineLog.Data("❌ رد", "adm_order_reject_"+order.ID)
	inlineLog.Inline(inlineLog.Row(btnDeliverLog, btnRejectLog))

	cur := db.GetCurrency()
	msg := fmt.Sprintf(
		"🔔 <b>سفارش جدید</b>\n\n"+
			"👤 %s (<code>%d</code>)\n"+
			"📦 %s\n"+
			"💰 %s\n"+
			"🆔 <code>%s</code>",
		order.UserName, order.UserID, order.ProductName,
		rate.Display(order.PriceUSD, cur),
		order.ID,
	)

	if botInstance != nil {
		notify.SendWithLog(botInstance, cfg, db, notify.LogTopicOrders, msg, inlinePrivate, inlineLog)
	}
}

func MyServicesHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		lang := userLang(c.Sender().ID)
		orders := db.GetOrdersByUser(c.Sender().ID)
		if len(orders) == 0 {
			return c.Send(i18n.T(lang, "services_empty"), menus.MainMenu)
		}

		var sb strings.Builder
		sb.WriteString(i18n.T(lang, "services_header"))
		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		for _, o := range orders {
			var statusKey string
			switch o.Status {
			case database.OrderPending:
				statusKey = "services_pending"
			case database.OrderDelivered:
				statusKey = "services_delivered"
			case database.OrderRejected:
				statusKey = "services_rejected"
			case database.OrderCodeRequested:
				statusKey = "services_code_req"
			case database.OrderCodeSent:
				statusKey = "services_code_sent"
			}
			sb.WriteString(fmt.Sprintf("📦 %s — %s\n", o.ProductName, i18n.T(lang, statusKey)))
			rows = append(rows, inline.Row(
				inline.Data(
					fmt.Sprintf("%s | %s", o.ProductName, shortStatus(o.Status)),
					fmt.Sprintf("svc_view_%s", o.ID),
				),
			))
		}

		inline.Inline(rows...)
		return c.Send(sb.String(), inline)
	}
}

func shortStatus(s database.OrderStatus) string {
	switch s {
	case database.OrderPending:
		return "⏳"
	case database.OrderDelivered:
		return "✅"
	case database.OrderRejected:
		return "❌"
	case database.OrderCodeRequested:
		return "🔄"
	case database.OrderCodeSent:
		return "📨"
	}
	return "❓"
}

func ServiceViewHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		lang := userLang(c.Sender().ID)
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "not_found")})
		}
		_ = c.Respond()

		cur := db.GetCurrency()
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			i18n.T(lang, "service_detail"),
			o.ProductName, o.ID, rate.Display(o.PriceUSD, cur),
		))

		product := db.GetProduct(o.ProductID)
		isVPN := product != nil && product.ProductType == "vpn"

		if isVPN {
			sb.WriteString(fmt.Sprintf("🌐 ترافیک: %d GB\n⏰ مدت: %d روز\n", product.TrafficGB, product.DurationDays))
		}

		switch o.Status {
		case database.OrderPending:
			sb.WriteString("\n" + i18n.T(lang, "service_pending_msg"))
		case database.OrderDelivered:
			sb.WriteString("\n" + i18n.T(lang, "service_delivered") + "\n")
			if len(o.DeliveryData) > 0 {
				fields, _ := db.GetCategoryDeliveryFields(o.CategoryID)
				if len(fields) > 0 {
					sb.WriteString("\n" + i18n.T(lang, "service_delivery_data"))
					for _, f := range fields {
						if v, ok := o.DeliveryData[f.Key]; ok && v != "" {
							sb.WriteString(fmt.Sprintf("• %s: <code>%s</code>\n", f.Label, v))
						}
					}
				}
			}
			if len(o.ClientLinks) > 0 {
				sb.WriteString(fmt.Sprintf("\n"+i18n.T(lang, "service_links"),
					strings.Join(o.ClientLinks, "\n")))
			} else if o.ServiceInfo != "" && len(o.DeliveryData) == 0 {
				sb.WriteString(fmt.Sprintf("\n📋 اطلاعات سرویس:\n<code>%s</code>", o.ServiceInfo))
			}
		case database.OrderRejected:
			sb.WriteString(fmt.Sprintf("\n"+i18n.T(lang, "service_rejected"), o.RejectNote))
		case database.OrderCodeRequested:
			sb.WriteString("\n" + i18n.T(lang, "service_code_req"))
		case database.OrderCodeSent:
			sb.WriteString(fmt.Sprintf("\n"+i18n.T(lang, "service_code_sent"), o.CodeValue))
		}

		if len(o.ActionRequests) > 0 {
			sb.WriteString("\n\n" + i18n.T(lang, "service_actions"))
			for _, ar := range o.ActionRequests {
				var statusStr string
				switch ar.Status {
				case database.ActionPending:
					statusStr = "⏳"
				case database.ActionInProgress:
					statusStr = "🔄"
				case database.ActionDone:
					statusStr = "✅"
				case database.ActionRejected:
					statusStr = "❌"
				}
				sb.WriteString(fmt.Sprintf("%s %s\n", statusStr, ar.ActionLabel))
			}
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		if o.Status == database.OrderDelivered {
			if isVPN {
				rows = append(rows, inline.Row(
					inline.Data(i18n.T(lang, "service_stats_btn"), fmt.Sprintf("svc_stats_%s", o.ID)),
					inline.Data(i18n.T(lang, "service_relink_btn"), fmt.Sprintf("svc_relink_%s", o.ID)),
				))
				rows = append(rows, inline.Row(
					inline.Data(i18n.T(lang, "service_qr_btn"), fmt.Sprintf("svc_qr_%s", o.ID)),
					inline.Data(i18n.T(lang, "service_renew_btn"), fmt.Sprintf("svc_renew_%s", o.ID)),
				))
				rows = append(rows, inline.Row(
					inline.Data(i18n.T(lang, "service_delete_btn"), fmt.Sprintf("svc_delete_%s", o.ID)),
				))
			} else {
				if product != nil && product.DeliveryType == "code_request" {
					rows = append(rows, inline.Row(
						inline.Data(i18n.T(lang, "service_code_btn"), fmt.Sprintf("svc_code_%s", o.ID)),
					))
				}
			}

			actions, _ := db.GetCategoryServiceActions(o.CategoryID)
			for _, a := range actions {
				label := a.Label
				if a.Emoji != "" {
					label = a.Emoji + " " + a.Label
				}
				rows = append(rows, inline.Row(
					inline.Data(label, fmt.Sprintf("svc_action_%s_%s", o.ID, a.Key)),
				))
			}
		}

		rows = append(rows, inline.Row(
			inline.Data(i18n.T(lang, "service_back_btn"), "svc_list"),
		))

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func ServiceRequestCodeHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		lang := userLang(c.Sender().ID)
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "not_found")})
		}
		if o.Status != database.OrderDelivered {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error")})
		}
		p := db.GetProduct(o.ProductID)
		if p == nil || p.DeliveryType != "code_request" {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error")})
		}
		req, err := db.RequestOrderCode(orderID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "success")})
		notifyAdminsCodeRequest(c, cfg, db, req)
		return c.Send(i18n.T(lang, "service_code_req"))
	}
}

func notifyAdminsCodeRequest(c AsliBot.Context, cfg *config.Config, db *database.DB, o *database.Order) {
	inlinePrivate := &AsliBot.ReplyMarkup{}
	btnSend := inlinePrivate.Data("📨 ارسال کد", "adm_order_code_"+o.ID)
	inlinePrivate.Inline(inlinePrivate.Row(btnSend))

	inlineLog := &AsliBot.ReplyMarkup{}
	btnSendLog := inlineLog.Data("📨 ارسال کد", "adm_order_code_"+o.ID)
	inlineLog.Inline(inlineLog.Row(btnSendLog))

	msg := fmt.Sprintf(
		"📨 <b>درخواست کد</b>\n\n"+
			"👤 %s (<code>%d</code>)\n"+
			"📦 %s\n"+
			"🆔 <code>%s</code>\n\n"+
			"اطلاعات سرویس:\n<code>%s</code>",
		o.UserName, o.UserID, o.ProductName, o.ID, o.ServiceInfo,
	)

	if botInstance != nil {
		notify.SendWithLog(botInstance, cfg, db, notify.LogTopicOrders, msg, inlinePrivate, inlineLog)
	}
}

func ServicesListHandler(c AsliBot.Context) error {
	_ = c.Respond()
	return MyServicesHandler(appCfg, walletDB)(c)
}
