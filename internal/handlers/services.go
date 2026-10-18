// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/menus"
	"github.com/theasli/GamajBot/internal/notify"
	"github.com/theasli/GamajBot/internal/qrcode"
	"github.com/theasli/GamajBot/internal/rate"
	"github.com/theasli/GamajBot/internal/xui"

	AsliBot "gopkg.in/telebot.v4"
)

func ServiceRenewHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سرویس پیدا نشد"})
		}
		if o.Status != database.OrderDelivered {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ فقط سرویس‌های تحویل‌شده قابل تمدید هستند"})
		}

		product := db.GetProduct(o.ProductID)
		if product == nil || product.ProductType != "vpn" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ این سرویس VPN نیست"})
		}

		u, err := db.GetUser(c.Sender().ID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ حساب پیدا نشد"})
		}
		if u.BalanceUSD < product.PriceUSD {
			cur := db.GetCurrency()
			return c.Respond(&AsliBot.CallbackResponse{
				Text: fmt.Sprintf("❌ موجودی کافی نیست. نیاز: %s", rate.Display(product.PriceUSD, cur)),
			})
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳ در حال تمدید..."})

		panel := db.GetXUIPanel(product.XUIPanelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := xui.New(xui.Config{
			BaseURL:    panel.BaseURL,
			AuthMethod: xui.AuthMethod(panel.AuthMethod),
			APIToken:   panel.APIToken,
			Username:   panel.Username,
			Password:   panel.Password,
		})

		if len(product.XUIInboundIDs) == 0 {
			return c.Send("❌ هیچ inbound برای این محصول ثبت نشده")
		}

		client, err := cli.GetClient(product.XUIInboundIDs[0], o.ClientEmail)
		if err != nil {
			return c.Send("❌ خطا در دریافت کلاینت: " + err.Error())
		}

		nowMs := time.Now().UnixMilli()
		baseExpiry := client.ExpiryTime
		if baseExpiry < nowMs {
			baseExpiry = nowMs
		}
		newExpiry := baseExpiry + int64(product.DurationDays)*24*60*60*1000

		client.ExpiryTime = newExpiry
		client.Enable = true

		if err := cli.UpdateClient(product.XUIInboundIDs[0], *client); err != nil {
			return c.Send("❌ خطا در تمدید: " + err.Error())
		}

		newBal, err := db.DeductBalanceWithType(u.ID, product.PriceUSD, "تمدید "+o.ProductName, "renewal")
		if err != nil {
			return c.Send("❌ خطا در کسر موجودی: " + err.Error())
		}

		cur := db.GetCurrency()
		return c.Send(fmt.Sprintf(
			"✅ <b>سرویس تمدید شد</b>\n\n"+
				"📦 %s\n"+
				"💰 %s کسر شد\n"+
				"⏰ انقضای جدید: %s\n"+
				"💵 موجودی جدید: <b>%s</b>",
			o.ProductName,
			rate.Display(product.PriceUSD, cur),
			time.UnixMilli(newExpiry).Format("2006/01/02 15:04"),
			rate.Display(newBal, cur),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}

func ServiceRelinkHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سرویس پیدا نشد"})
		}
		if o.Status != database.OrderDelivered {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ فقط سرویس‌های تحویل‌شده"})
		}

		product := db.GetProduct(o.ProductID)
		if product == nil || product.ProductType != "vpn" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ این سرویس VPN نیست"})
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳ در حال دریافت..."})

		panel := db.GetXUIPanel(product.XUIPanelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := xui.New(xui.Config{
			BaseURL:    panel.BaseURL,
			AuthMethod: xui.AuthMethod(panel.AuthMethod),
			APIToken:   panel.APIToken,
			Username:   panel.Username,
			Password:   panel.Password,
		})

		links, err := cli.GetClientLinks(o.ClientEmail)
		if err != nil {
			return c.Send("❌ خطا: " + err.Error())
		}
		if len(links) == 0 {
			return c.Send("❌ لینکی پیدا نشد")
		}

		o.ClientLinks = links
		_ = db.UpdateOrder(o)

		return c.Send(fmt.Sprintf(
			"🔑 <b>لینک‌های اتصال %s</b>\n\n<code>%s</code>",
			o.ProductName, strings.Join(links, "\n"),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}

func ServiceStatsHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سرویس پیدا نشد"})
		}
		if o.Status != database.OrderDelivered {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ فقط سرویس‌های تحویل‌شده"})
		}

		product := db.GetProduct(o.ProductID)
		if product == nil || product.ProductType != "vpn" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ این سرویس VPN نیست"})
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳..."})

		panel := db.GetXUIPanel(product.XUIPanelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := xui.New(xui.Config{
			BaseURL:    panel.BaseURL,
			AuthMethod: xui.AuthMethod(panel.AuthMethod),
			APIToken:   panel.APIToken,
			Username:   panel.Username,
			Password:   panel.Password,
		})

		if len(product.XUIInboundIDs) == 0 {
			return c.Send("❌ هیچ inbound برای این محصول ثبت نشده")
		}

		stats, err := cli.GetClientTraffic(product.XUIInboundIDs[0], o.ClientEmail)
		if err != nil {
			return c.Send("❌ خطا: " + err.Error())
		}

		usedGB := float64(stats.Up+stats.Down) / (1024 * 1024 * 1024)
		totalGB := float64(stats.TotalGB) / (1024 * 1024 * 1024)
		remaining := totalGB - usedGB
		if remaining < 0 {
			remaining = 0
		}

		expiry := "نامحدود"
		if stats.ExpiryTime > 0 {
			expiry = time.UnixMilli(stats.ExpiryTime).Format("2006/01/02 15:04")
		}

		statusStr := "❌ غیرفعال"
		if stats.Enable {
			statusStr = "✅ فعال"
		}

		return c.Send(fmt.Sprintf(
			"📊 <b>آمار %s</b>\n\n"+
				"📥 مصرف: <b>%.2f GB</b>\n"+
				"📦 کل: <b>%.2f GB</b>\n"+
				"📤 باقی‌مانده: <b>%.2f GB</b>\n"+
				"⏰ انقضا: %s\n"+
				"📊 وضعیت: %s",
			o.ProductName, usedGB, totalGB, remaining, expiry, statusStr,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}

func ServiceDeleteHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سرویس پیدا نشد"})
		}
		if o.Status != database.OrderDelivered {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ فقط سرویس‌های تحویل‌شده"})
		}

		product := db.GetProduct(o.ProductID)
		if product == nil || product.ProductType != "vpn" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ این سرویس VPN نیست"})
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳ در حال حذف..."})

		panel := db.GetXUIPanel(product.XUIPanelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := xui.New(xui.Config{
			BaseURL:    panel.BaseURL,
			AuthMethod: xui.AuthMethod(panel.AuthMethod),
			APIToken:   panel.APIToken,
			Username:   panel.Username,
			Password:   panel.Password,
		})

		if len(product.XUIInboundIDs) == 0 {
			return c.Send("❌ هیچ inbound برای این محصول ثبت نشده")
		}

		client, err := cli.GetClient(product.XUIInboundIDs[0], o.ClientEmail)
		if err != nil {
			return c.Send("❌ خطا در دریافت کلاینت: " + err.Error())
		}

		if err := cli.DeleteClient(product.XUIInboundIDs[0], client.ID); err != nil {
			return c.Send("❌ خطا در حذف: " + err.Error())
		}

		o.Status = database.OrderRejected
		o.RejectNote = "حذف شده توسط کاربر"
		_ = db.UpdateOrder(o)

		cur := db.GetCurrency()
		return c.Send(fmt.Sprintf(
			"✅ سرویس <b>%s</b> حذف شد.\n\n"+
				"💰 مبلغ %s به کیف پول شما بازگشت.",
			o.ProductName, rate.Display(o.PriceUSD, cur),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, menus.MainMenu)
	}
}

func ServiceQRHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سرویس پیدا نشد"})
		}
		if o.Status != database.OrderDelivered {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ فقط سرویس‌های تحویل‌شده"})
		}
		if len(o.ClientLinks) == 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ لینکی برای این سرویس نیست"})
		}

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳ در حال ساخت..."})

		link := o.ClientLinks[0]
		pngData, err := qrcode.GeneratePNG(link, 512)
		if err != nil {
			return c.Send("❌ خطا در ساخت QR: " + err.Error())
		}

		photo := &AsliBot.Photo{
			File:    AsliBot.FromReader(bytes.NewReader(pngData)),
			Caption: fmt.Sprintf("📱 QR Code سرویس <b>%s</b>\n\nبا اپ VPN اسکن کن.", o.ProductName),
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("📋 کپی لینک", fmt.Sprintf("svc_relink_%s", o.ID)),
				inline.Data("⬅️ برگشت", fmt.Sprintf("svc_view_%s", o.ID)),
			),
		)

		return c.Send(photo, &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func ServiceActionRequestHandler(cfg *config.Config, db *database.DB, orderID, actionKey string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		o := db.GetOrder(orderID)
		if o == nil || o.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سرویس پیدا نشد"})
		}
		if o.Status != database.OrderDelivered {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ فقط سرویس‌های تحویل‌شده"})
		}

		actions, _ := db.GetCategoryServiceActions(o.CategoryID)
		var target *database.ServiceAction
		for i := range actions {
			if actions[i].Key == actionKey {
				target = &actions[i]
				break
			}
		}
		if target == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ اکشن پیدا نشد"})
		}

		_ = c.Respond()

		if target.RequireInput {
			setServiceActionState(c.Sender().ID, &actionInputState{
				OrderID:   orderID,
				ActionKey: actionKey,
				ActionLbl: target.Label,
			})
			label := target.InputLabel
			if label == "" {
				label = "ورودی"
			}
			return c.Send(fmt.Sprintf(
				"⚙️ <b>%s %s</b>\n\n📦 %s\n\n📝 %s را بفرست:\n\n(برای لغو /cancel بزن)",
				target.Emoji, target.Label, o.ProductName, label,
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		}

		req := database.ActionRequest{
			ActionKey:   actionKey,
			ActionLabel: target.Label,
			Status:      database.ActionPending,
		}
		if err := db.AddOrderActionRequest(orderID, req); err != nil {
			return c.Send("❌ " + err.Error())
		}
		notifyAdminsActionRequest(c, cfg, db, o, req)
		return c.Send(fmt.Sprintf(
			"✅ <b>درخواست %s ثبت شد</b>\n\n📦 %s\n\nپشتیبانی به‌زودی بررسی می‌کند.",
			target.Label, o.ProductName,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}

type actionInputState struct {
	OrderID   string
	ActionKey string
	ActionLbl string
}

var (
	actionInputStates = make(map[int64]*actionInputState)
	actionInputMu     sync.Mutex
)

func setServiceActionState(uid int64, s *actionInputState) {
	actionInputMu.Lock()
	defer actionInputMu.Unlock()
	if s == nil {
		delete(actionInputStates, uid)
	} else {
		actionInputStates[uid] = s
	}
}

func getServiceActionState(uid int64) *actionInputState {
	actionInputMu.Lock()
	defer actionInputMu.Unlock()
	return actionInputStates[uid]
}

func ServiceActionInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getServiceActionState(uid)
	if st == nil {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}
	if text == "" {
		return c.Send("❌ ورودی خالی است. دوباره بفرست:")
	}

	o := db.GetOrder(st.OrderID)
	if o == nil {
		setServiceActionState(uid, nil)
		return c.Send("❌ سرویس پیدا نشد")
	}

	req := database.ActionRequest{
		ActionKey:   st.ActionKey,
		ActionLabel: st.ActionLbl,
		UserInput:   text,
		Status:      database.ActionPending,
	}
	if err := db.AddOrderActionRequest(st.OrderID, req); err != nil {
		setServiceActionState(uid, nil)
		return c.Send("❌ " + err.Error())
	}
	setServiceActionState(uid, nil)

	notifyAdminsActionRequest(c, cfg, db, o, req)
	return c.Send(fmt.Sprintf(
		"✅ <b>درخواست %s ثبت شد</b>\n\n📦 %s\n\nپشتیبانی به‌زودی بررسی می‌کند.",
		st.ActionLbl, o.ProductName,
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
}

func notifyAdminsActionRequest(c AsliBot.Context, cfg *config.Config, db *database.DB, o *database.Order, req database.ActionRequest) {
	inlinePrivate := &AsliBot.ReplyMarkup{}
	inlinePrivate.Inline(
		inlinePrivate.Row(
			inlinePrivate.Data("👁 مشاهده", fmt.Sprintf("adm_ar_view_%s_%s", o.ID, req.ID)),
		),
	)

	inlineLog := &AsliBot.ReplyMarkup{}
	inlineLog.Inline(
		inlineLog.Row(
			inlineLog.Data("👁 مشاهده", fmt.Sprintf("adm_ar_view_%s_%s", o.ID, req.ID)),
		),
	)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"⚙️ <b>درخواست %s</b>\n\n"+
			"📦 سفارش: <code>%s</code>\n"+
			"👤 کاربر: %s (<code>%d</code>)\n"+
			"🕐 زمان: %s\n",
		req.ActionLabel, o.ID, o.UserName, o.UserID, req.RequestedAt,
	))
	if req.UserInput != "" {
		sb.WriteString(fmt.Sprintf("\n📝 ورودی:\n<code>%s</code>", req.UserInput))
	}

	if botInstance != nil {
		notify.SendWithLog(botInstance, cfg, db, notify.LogTopicOrders, sb.String(), inlinePrivate, inlineLog)
	}
}
