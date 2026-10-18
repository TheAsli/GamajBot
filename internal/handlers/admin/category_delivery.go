// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/rate"

	AsliBot "gopkg.in/telebot.v4"
)

type fieldCollectionState struct {
	OrderID string
	Fields  []database.DeliveryField
	Index   int
	Values  map[string]string
}

var (
	fieldStates = make(map[int64]*fieldCollectionState)
	fieldMu     sync.Mutex
)

func setFieldState(uid int64, fs *fieldCollectionState) {
	fieldMu.Lock()
	defer fieldMu.Unlock()
	if fs == nil {
		delete(fieldStates, uid)
	} else {
		fieldStates[uid] = fs
	}
}

func getFieldState(uid int64) *fieldCollectionState {
	fieldMu.Lock()
	defer fieldMu.Unlock()
	return fieldStates[uid]
}

func DeliveryStartHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		o := db.GetOrder(orderID)
		if o == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سفارش پیدا نشد"})
		}
		if o.Status != database.OrderPending {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ این سفارش قبلاً بررسی شده"})
		}

		product := db.GetProduct(o.ProductID)
		if product != nil && product.ProductType == "vpn" {
			_ = c.Respond()
			setOrderState(c.Sender().ID, &orderState{Step: "deliver_info", OrderID: orderID})
			return c.Send(fmt.Sprintf(
				"📤 <b>تحویل سفارش %s</b>\n\n📦 %s\n\n🔑 لینک‌های اتصال را بفرست (هر لینک در یک خط):",
				o.ID, o.ProductName,
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		}

		fields, err := db.GetCategoryDeliveryFields(o.CategoryID)
		if err != nil || len(fields) == 0 {
			_ = c.Respond()
			setOrderState(c.Sender().ID, &orderState{Step: "deliver_info", OrderID: orderID})
			return c.Send(fmt.Sprintf(
				"📤 <b>تحویل سفارش %s</b>\n\n📦 %s\n\nاطلاعات سرویس را بفرست:",
				o.ID, o.ProductName,
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		}

		_ = c.Respond()
		setOrderState(c.Sender().ID, &orderState{
			Step:    "deliver_fields",
			OrderID: orderID,
		})
		setFieldState(c.Sender().ID, &fieldCollectionState{
			OrderID: orderID,
			Fields:  fields,
			Index:   0,
			Values:  map[string]string{},
		})
		return startFieldCollection(c, db, orderID, fields, 0, map[string]string{})
	}
}

func startFieldCollection(c AsliBot.Context, db *database.DB, orderID string, fields []database.DeliveryField, idx int, values map[string]string) error {
	if idx >= len(fields) {
		if err := db.SetOrderDeliveryData(orderID, values); err != nil {
			return c.Send("❌ خطا در ذخیره: " + err.Error())
		}
		o := db.GetOrder(orderID)
		if o != nil {
			o.Status = database.OrderDelivered
			o.DeliveredAt = time.Now().Format(time.RFC3339)
			var parts []string
			for _, f := range fields {
				if v, ok := values[f.Key]; ok && v != "" {
					parts = append(parts, fmt.Sprintf("%s: %s", f.Label, v))
				}
			}
			o.ServiceInfo = strings.Join(parts, "\n")
			_ = db.UpdateOrder(o)
		}
		setOrderState(c.Sender().ID, nil)
		setFieldState(c.Sender().ID, nil)
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ تحویل شد"})

		cur := db.GetCurrency()
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"📦 <b>سفارش شما تحویل شد!</b>\n\n🆔 <code>%s</code>\n📦 %s\n💰 %s\n\n📋 اطلاعات سرویس:\n",
			o.ID, o.ProductName, rate.Display(o.PriceUSD, cur),
		))
		for _, f := range fields {
			if v, ok := values[f.Key]; ok && v != "" {
				sb.WriteString(fmt.Sprintf("• %s: <code>%s</code>\n", f.Label, v))
			}
		}
		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("📦 مشاهده سرویس", fmt.Sprintf("svc_view_%s", o.ID)),
			),
		)
		_, _ = c.Bot().Send(AsliBot.ChatID(o.UserID), sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
		return c.Send("✅ فیلدها ذخیره و سفارش تحویل شد.")
	}

	f := fields[idx]
	placeholder := ""
	if f.Placeholder != "" {
		placeholder = fmt.Sprintf(" (%s)", f.Placeholder)
	}
	req := ""
	if f.Required {
		req = " *"
	}
	return c.Send(fmt.Sprintf(
		"📝 فیلد %d از %d\n\n<b>%s</b>%s%s:\n\n(برای لغو /cancel بزن)",
		idx+1, len(fields), f.Label, req, placeholder,
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
}

func CategoryDeliveryInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getOrderState(uid)
	if st == nil || st.Step != "deliver_fields" {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	o := db.GetOrder(st.OrderID)
	if o == nil {
		setOrderState(uid, nil)
		setFieldState(uid, nil)
		return c.Send("❌ سفارش پیدا نشد")
	}

	fs := getFieldState(uid)
	if fs == nil {
		fields, _ := db.GetCategoryDeliveryFields(o.CategoryID)
		fs = &fieldCollectionState{
			OrderID: st.OrderID,
			Fields:  fields,
			Index:   0,
			Values:  map[string]string{},
		}
		setFieldState(uid, fs)
	}

	if fs.Index >= len(fs.Fields) {
		setOrderState(uid, nil)
		setFieldState(uid, nil)
		return c.Send("❌ خطا در ترتیب فیلدها")
	}

	f := fs.Fields[fs.Index]
	if text == "" && f.Required {
		return c.Send("❌ این فیلد اجباری است. دوباره بفرست:")
	}
	fs.Values[f.Key] = text
	fs.Index++
	setFieldState(uid, fs)
	return startFieldCollection(c, db, st.OrderID, fs.Fields, fs.Index, fs.Values)
}

func ActionRequestViewHandler(cfg *config.Config, db *database.DB, orderID, reqID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		_ = c.Respond()

		o := db.GetOrder(orderID)
		if o == nil {
			return c.Send("❌ سفارش پیدا نشد")
		}
		req, err := db.GetOrderActionRequest(orderID, reqID)
		if err != nil {
			return c.Send("❌ درخواست پیدا نشد")
		}

		var statusStr string
		switch req.Status {
		case database.ActionPending:
			statusStr = "⏳ در انتظار"
		case database.ActionInProgress:
			statusStr = "🔄 در حال انجام"
		case database.ActionDone:
			statusStr = "✅ انجام شد"
		case database.ActionRejected:
			statusStr = "❌ رد شد"
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"⚙️ <b>درخواست %s</b>\n\n"+
				"📦 سفارش: <code>%s</code>\n"+
				"👤 کاربر: %s (<code>%d</code>)\n"+
				"📊 وضعیت: %s\n"+
				"🕐 زمان: %s\n",
			req.ActionLabel, o.ID, o.UserName, o.UserID, statusStr, req.RequestedAt,
		))
		if req.UserInput != "" {
			sb.WriteString(fmt.Sprintf("\n📝 ورودی کاربر:\n<code>%s</code>", req.UserInput))
		}
		if req.AdminNote != "" {
			sb.WriteString(fmt.Sprintf("\n📝 یادداشت ادمین: %s", req.AdminNote))
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row
		if req.Status == database.ActionPending || req.Status == database.ActionInProgress {
			rows = append(rows, inline.Row(
				inline.Data("✅ انجام شد", fmt.Sprintf("adm_ar_done_%s_%s", orderID, reqID)),
				inline.Data("❌ رد", fmt.Sprintf("adm_ar_rej_%s_%s", orderID, reqID)),
			))
		}
		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", "adm_orders_pending_0"),
		))

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func ActionRequestDoneHandler(cfg *config.Config, db *database.DB, orderID, reqID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}

		o := db.GetOrder(orderID)
		if o == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سفارش پیدا نشد"})
		}
		req, err := db.GetOrderActionRequest(orderID, reqID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ درخواست پیدا نشد"})
		}

		if err := db.UpdateOrderActionStatus(orderID, reqID, database.ActionDone, c.Sender().ID, ""); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ انجام شد"})

		_, _ = c.Bot().Send(AsliBot.ChatID(o.UserID), fmt.Sprintf(
			"✅ <b>درخواست %s شما انجام شد</b>\n\n📦 %s\n\nبا تشکر",
			req.ActionLabel, o.ProductName,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})

		return c.Send("✅ به کاربر اطلاع داده شد.")
	}
}

func ActionRequestRejectHandler(cfg *config.Config, db *database.DB, orderID, reqID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}

		o := db.GetOrder(orderID)
		if o == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سفارش پیدا نشد"})
		}
		req, err := db.GetOrderActionRequest(orderID, reqID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ درخواست پیدا نشد"})
		}

		if err := db.UpdateOrderActionStatus(orderID, reqID, database.ActionRejected, c.Sender().ID, "رد شد توسط ادمین"); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ رد شد"})

		_, _ = c.Bot().Send(AsliBot.ChatID(o.UserID), fmt.Sprintf(
			"❌ <b>درخواست %s شما رد شد</b>\n\n📦 %s\n\nبرای اطلاعات بیشتر با پشتیبانی تماس بگیر.",
			req.ActionLabel, o.ProductName,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})

		return c.Send("✅ به کاربر اطلاع داده شد.")
	}
}
