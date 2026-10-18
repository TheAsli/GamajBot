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

type orderState struct {
	Step    string
	OrderID string
}

var (
	orderStates = make(map[int64]*orderState)
	orderMu     sync.Mutex
)

func setOrderState(uid int64, s *orderState) {
	orderMu.Lock()
	defer orderMu.Unlock()
	if s == nil {
		delete(orderStates, uid)
	} else {
		orderStates[uid] = s
	}
}

func getOrderState(uid int64) *orderState {
	orderMu.Lock()
	defer orderMu.Unlock()
	return orderStates[uid]
}

func OrdersHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showOrdersPage(c, db, "pending", 0)
	}
}

func OrdersPageHandler(cfg *config.Config, db *database.DB, filter string, page int) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showOrdersPage(c, db, filter, page)
	}
}

func showOrdersPage(c AsliBot.Context, db *database.DB, filter string, page int) error {
	all := db.GetAllOrders()
	var filtered []*database.Order
	for _, o := range all {
		switch filter {
		case "pending":
			if o.Status == database.OrderPending || o.Status == database.OrderCodeRequested {
				filtered = append(filtered, o)
			}
		case "delivered":
			if o.Status == database.OrderDelivered || o.Status == database.OrderCodeSent {
				filtered = append(filtered, o)
			}
		case "rejected":
			if o.Status == database.OrderRejected {
				filtered = append(filtered, o)
			}
		default:
			filtered = append(filtered, o)
		}
	}

	filterLabel := map[string]string{
		"pending":   "در انتظار",
		"delivered": "تحویل شده",
		"rejected":  "رد شده",
		"all":       "همه",
	}[filter]
	if filterLabel == "" {
		filterLabel = filter
	}

	if len(filtered) == 0 {
		return c.Send(fmt.Sprintf("🧾 سفارش‌ها — %s\n\nهیچ سفارشی نیست.", filterLabel))
	}

	perPage := 6
	total := len(filtered)
	start := page * perPage
	if start >= total {
		page, start = 0, 0
	}
	end := start + perPage
	if end > total {
		end = total
	}

	cur := db.GetCurrency()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🧾 سفارش‌ها — %s (صفحه %d از %d)\n\n",
		filterLabel, page+1, (total+perPage-1)/perPage))

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	for _, o := range filtered[start:end] {
		vpnMark := ""
		product := db.GetProduct(o.ProductID)
		if product != nil && product.ProductType == "vpn" {
			vpnMark = "🌐 "
		}
		sb.WriteString(fmt.Sprintf("🆔 %s\n📦 %s%s\n👤 %s\n💰 %s\n\n",
			o.ID, vpnMark, o.ProductName, o.UserName, rate.Display(o.PriceUSD, cur)))
		rows = append(rows, inline.Row(
			inline.Data(
				fmt.Sprintf("%s %s%s | %s", shortStatus(o.Status), vpnMark, o.ProductName, o.UserName),
				fmt.Sprintf("adm_order_view_%s", o.ID),
			),
		))
	}

	filterRow := []AsliBot.Btn{
		inline.Data("⏳ در انتظار", "adm_orders_pending_0"),
		inline.Data("✅ تحویل", "adm_orders_delivered_0"),
		inline.Data("❌ رد", "adm_orders_rejected_0"),
		inline.Data("📋 همه", "adm_orders_all_0"),
	}
	rows = append([]AsliBot.Row{inline.Row(filterRow...)}, rows...)

	var pageRow []AsliBot.Btn
	if page > 0 {
		pageRow = append(pageRow, inline.Data("⬅️", fmt.Sprintf("adm_orders_%s_%d", filter, page-1)))
	}
	if end < total {
		pageRow = append(pageRow, inline.Data("➡️", fmt.Sprintf("adm_orders_%s_%d", filter, page+1)))
	}
	if len(pageRow) > 0 {
		rows = append(rows, inline.Row(pageRow...))
	}

	inline.Inline(rows...)
	return c.Send(sb.String(), inline)
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
		return "📨"
	case database.OrderCodeSent:
		return "📨"
	}
	return "❓"
}

func OrderViewHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		_ = c.Respond()
		return showOrderDetail(c, db, orderID)
	}
}

func showOrderDetail(c AsliBot.Context, db *database.DB, orderID string) error {
	o := db.GetOrder(orderID)
	if o == nil {
		return c.Send("❌ سفارش پیدا نشد")
	}

	var statusLabel string
	switch o.Status {
	case database.OrderPending:
		statusLabel = "⏳ در انتظار تحویل"
	case database.OrderDelivered:
		statusLabel = "✅ تحویل شده"
	case database.OrderRejected:
		statusLabel = "❌ رد شده"
	case database.OrderCodeRequested:
		statusLabel = "📨 درخواست کد"
	case database.OrderCodeSent:
		statusLabel = "📨 کد ارسال شده"
	}

	product := db.GetProduct(o.ProductID)
	isVPN := product != nil && product.ProductType == "vpn"

	cur := db.GetCurrency()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"🧾 <b>سفارش %s</b>\n\n"+
			"📦 %s\n"+
			"👤 %s (<code>%d</code>)\n"+
			"💰 %s\n"+
			"📊 %s\n"+
			"🕐 %s\n",
		o.ID, o.ProductName, o.UserName, o.UserID,
		rate.Display(o.PriceUSD, cur), statusLabel, o.CreatedAt,
	))

	if isVPN {
		sb.WriteString(fmt.Sprintf(
			"\n🌐 <b>VPN</b>\n"+
				"   📊 ترافیک: %d GB\n"+
				"   ⏰ مدت: %d روز\n",
			product.TrafficGB, product.DurationDays,
		))
		if o.ClientEmail != "" {
			sb.WriteString(fmt.Sprintf("   📧 Email: <code>%s</code>\n", o.ClientEmail))
		}
	}

	if o.ServiceInfo != "" {
		sb.WriteString(fmt.Sprintf("\n📋 اطلاعات سرویس:\n<code>%s</code>", o.ServiceInfo))
	}
	if len(o.ClientLinks) > 0 {
		sb.WriteString(fmt.Sprintf("\n\n🔑 لینک‌های اتصال:\n<code>%s</code>",
			strings.Join(o.ClientLinks, "\n")))
	}
	if o.RejectNote != "" {
		sb.WriteString(fmt.Sprintf("\n📝 دلیل رد: %s", o.RejectNote))
	}
	if o.CodeValue != "" {
		sb.WriteString(fmt.Sprintf("\n🔑 کد: <code>%s</code>", o.CodeValue))
	}

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	if o.Status == database.OrderPending {
		rows = append(rows, inline.Row(
			inline.Data("📤 تحویل", "adm_order_deliver_"+o.ID),
			inline.Data("❌ رد", "adm_order_reject_"+o.ID),
		))
	}
	if o.Status == database.OrderCodeRequested {
		rows = append(rows, inline.Row(
			inline.Data("📨 ارسال کد", "adm_order_code_"+o.ID),
		))
	}
	rows = append(rows, inline.Row(
		inline.Data("⬅️ برگشت", "adm_orders_pending_0"),
	))

	inline.Inline(rows...)
	return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func OrderDeliverHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return DeliveryStartHandler(cfg, db, orderID)(c)
	}
}

func OrderRejectHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
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
		setOrderState(c.Sender().ID, &orderState{Step: "reject_note", OrderID: orderID})
		_ = c.Respond()
		return c.Send("📝 دلیل رد سفارش را بفرست:")
	}
}

func OrderCodeHandler(cfg *config.Config, db *database.DB, orderID string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		o := db.GetOrder(orderID)
		if o == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ سفارش پیدا نشد"})
		}
		if o.Status != database.OrderCodeRequested {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ درخواست کدی نیست"})
		}
		setOrderState(c.Sender().ID, &orderState{Step: "send_code", OrderID: orderID})
		_ = c.Respond()
		return c.Send(fmt.Sprintf(
			"📨 <b>ارسال کد برای سفارش %s</b>\n\n📦 %s\n\nکد را بفرست:",
			o.ID, o.ProductName,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}

func OrderInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getOrderState(uid)
	if st == nil {
		return nil
	}
	role := db.GetAdminRole(uid, cfg.ADMIN_ID)
	if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
		setOrderState(uid, nil)
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	switch st.Step {
	case "deliver_info":
		o := db.GetOrder(st.OrderID)
		if o == nil {
			return c.Send("❌ سفارش پیدا نشد.")
		}
		product := db.GetProduct(o.ProductID)
		if product != nil && product.ProductType == "vpn" {
			links := strings.Split(text, "\n")
			o.ClientLinks = links
			o.Status = database.OrderDelivered
			o.DeliveredAt = time.Now().Format(time.RFC3339)
			o.ServiceInfo = text
			_ = db.UpdateOrder(o)
			setOrderState(uid, nil)

			admin := c.Sender()
			logActionTarget(botInstance, cfg, db, admin, "تحویل سفارش VPN", o.UserName, o.UserID,
				fmt.Sprintf("سفارش <code>%s</code> — %s", o.ID, o.ProductName))

			_, _ = c.Bot().Send(AsliBot.ChatID(o.UserID), fmt.Sprintf(
				"📦 <b>سفارش شما تحویل شد!</b>\n\n"+
					"🆔 <code>%s</code>\n"+
					"📦 %s\n\n"+
					"🔑 لینک اتصال:\n<code>%s</code>",
				o.ID, o.ProductName, text,
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
			return c.Send("✅ لینک‌ها به کاربر ارسال شد.")
		}

		o2, err := db.DeliverOrder(st.OrderID, text)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}
		setOrderState(uid, nil)

		admin := c.Sender()
		logActionTarget(botInstance, cfg, db, admin, "تحویل سفارش", o2.UserName, o2.UserID,
			fmt.Sprintf("سفارش <code>%s</code> — %s", o2.ID, o2.ProductName))

		_, _ = c.Bot().Send(AsliBot.ChatID(o2.UserID), fmt.Sprintf(
			"📦 <b>سفارش شما تحویل شد!</b>\n\n"+
				"🆔 <code>%s</code>\n"+
				"📦 %s\n\n"+
				"📋 اطلاعات سرویس:\n<code>%s</code>",
			o2.ID, o2.ProductName, text,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		return c.Send("✅ سفارش تحویل شد و به کاربر اطلاع داده شد.")

	case "reject_note":
		o, newBal, err := db.RejectOrder(st.OrderID, text)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}
		setOrderState(uid, nil)
		cur := db.GetCurrency()

		admin := c.Sender()
		logActionTarget(botInstance, cfg, db, admin, "رد سفارش", o.UserName, o.UserID,
			fmt.Sprintf("سفارش <code>%s</code> — %s — دلیل: %s", o.ID, o.ProductName, text))

		_, _ = c.Bot().Send(AsliBot.ChatID(o.UserID), fmt.Sprintf(
			"❌ <b>سفارش شما رد شد</b>\n\n"+
				"🆔 <code>%s</code>\n"+
				"📦 %s\n"+
				"📝 دلیل: %s\n\n"+
				"💰 مبلغ %s به کیف پول شما بازگشت.\n"+
				"💵 موجودی جدید: <b>%s</b>",
			o.ID, o.ProductName, text,
			rate.Display(o.PriceUSD, cur),
			rate.Display(newBal, cur),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		return c.Send("✅ سفارش رد شد و وجه به کاربر برگشت.")

	case "send_code":
		o, err := db.SendOrderCode(st.OrderID, text)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}
		setOrderState(uid, nil)

		admin := c.Sender()
		logActionTarget(botInstance, cfg, db, admin, "ارسال کد", o.UserName, o.UserID,
			fmt.Sprintf("سفارش <code>%s</code> — %s", o.ID, o.ProductName))

		_, _ = c.Bot().Send(AsliBot.ChatID(o.UserID), fmt.Sprintf(
			"📨 <b>کد دریافت شد</b>\n\n"+
				"🆔 <code>%s</code>\n"+
				"📦 %s\n\n"+
				"🔑 کد: <code>%s</code>",
			o.ID, o.ProductName, text,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		return c.Send("✅ کد ارسال شد.")
	}
	return nil
}
