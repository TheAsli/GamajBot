// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"strings"
	"sync"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/handlers/notify"

	AsliBot "gopkg.in/telebot.v4"
)

var (
	ticketReply = make(map[int64]int64)
	ticketMu    sync.Mutex
	notifier    *notify.Notifier
)

func InitNotifier(n *notify.Notifier) {
	notifier = n
}

func setTicketReply(adminID, ticketID int64) {
	ticketMu.Lock()
	defer ticketMu.Unlock()
	if ticketID == 0 {
		delete(ticketReply, adminID)
	} else {
		ticketReply[adminID] = ticketID
	}
}

func getTicketReply(adminID int64) int64 {
	ticketMu.Lock()
	defer ticketMu.Unlock()
	return ticketReply[adminID]
}

// ═══════════════════════════════════════════
// ─── لیست تیکت‌ها ───
// ═══════════════════════════════════════════

func TicketsHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		var tickets []*database.Ticket
		if role == database.RoleSuperAdmin {
			tickets = db.GetAllTickets(true)
		} else {
			admin := db.GetAdminByID(c.Sender().ID)
			if admin == nil || len(admin.DeptIDs) == 0 {
				return c.Send("🎫 هیچ تیکتی نداری.")
			}
			for _, deptID := range admin.DeptIDs {
				tickets = append(tickets, db.GetTicketsByDept(deptID, true)...)
			}
		}

		if len(tickets) == 0 {
			return c.Send("🎫 هیچ تیکت بازی وجود نداره.")
		}

		stats := db.GetTicketStats()

		var sb strings.Builder
		sb.WriteString("🎫 مدیریت تیکت‌ها\n\n")
		sb.WriteString(fmt.Sprintf(
			"📊 آمار:\n"+
				"   🟢 باز: %d\n"+
				"   🟡 در حال بررسی: %d\n"+
				"   🔵 پاسخ داده: %d\n"+
				"   🟠 منتظر کاربر: %d\n"+
				"   ⚫ بسته: %d\n"+
				"   📈 کل: %d\n\n",
			stats.Open, stats.Pending, stats.Answered,
			stats.WaitingUser, stats.Closed, stats.Total,
		))

		sb.WriteString(fmt.Sprintf("🎫 تیکت‌های باز (%d):\n\n", len(tickets)))

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		for i, t := range tickets {
			if i >= 10 {
				break
			}
			dept := db.GetDepartmentByID(t.DeptID)
			deptName := "—"
			if dept != nil {
				deptName = dept.Emoji + " " + dept.Name
			}

			sb.WriteString(fmt.Sprintf(
				"#%d — %s\n"+
					"👤 %s | 🏢 %s\n"+
					"📊 %s | ⚡ %s\n\n",
				t.ID, t.Subject,
				t.UserFirstName, deptName,
				notify.StatusLabel(t.Status),
				notify.PriorityLabel(t.Priority),
			))

			btn := inline.Data(
				fmt.Sprintf("#%d — %s", t.ID, t.Subject),
				fmt.Sprintf("ticket_view_%d", t.ID),
			)
			rows = append(rows, inline.Row(btn))
		}

		inline.Inline(rows...)
		return c.Send(sb.String(), inline)
	}
}

// ═══════════════════════════════════════════
// ─── دیدن تیکت ───
// ═══════════════════════════════════════════

func ViewTicketHandler(cfg *config.Config, db *database.DB, ticketID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		t, err := db.GetTicket(ticketID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ تیکت پیدا نشد"})
		}

		dept := db.GetDepartmentByID(t.DeptID)
		deptName := "—"
		if dept != nil {
			deptName = dept.Emoji + " " + dept.Name
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"🎫 تیکت #%d\n\n"+
				"👤 کاربر: %s (@%s)\n"+
				"🆔 آیدی: <code>%d</code>\n"+
				"🏢 دپارتمان: %s\n"+
				"📝 موضوع: %s\n"+
				"📊 وضعیت: %s\n"+
				"⚡ اولویت: %s\n"+
				"🕐 ایجاد: %s\n\n"+
				"💬 پیام‌ها:\n",
			t.ID,
			t.UserFirstName, t.UserUsername,
			t.UserID,
			deptName,
			t.Subject,
			notify.StatusLabel(t.Status),
			notify.PriorityLabel(t.Priority),
			t.CreatedAt[:16],
		))

		for _, m := range t.Messages {
			who := "👤 کاربر"
			if m.IsAdmin {
				who = "👮 ادمین"
			}
			ts := m.CreatedAt
			if len(ts) > 16 {
				ts = ts[:16]
			}
			sb.WriteString(fmt.Sprintf("%s (%s):\n%s\n\n", who, ts, m.Text))
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		if t.Status != database.TicketClosed {
			btnReply := inline.Data("💬 پاسخ", fmt.Sprintf("ticket_reply_%d", t.ID))
			btnClose := inline.Data("🔒 بستن", fmt.Sprintf("ticket_close_%d", t.ID))
			btnPending := inline.Data("🟡 در حال بررسی", fmt.Sprintf("ticket_pending_%d", t.ID))
			btnTransfer := inline.Data("🔄 انتقال", fmt.Sprintf("ticket_transfer_%d", t.ID))

			rows = append(rows, inline.Row(btnReply))
			rows = append(rows, inline.Row(btnPending, btnClose))
			rows = append(rows, inline.Row(btnTransfer))
		}

		btnPrioLow := inline.Data("🟢", fmt.Sprintf("ticket_prio_low_%d", t.ID))
		btnPrioNormal := inline.Data("🟡", fmt.Sprintf("ticket_prio_normal_%d", t.ID))
		btnPrioHigh := inline.Data("🟠", fmt.Sprintf("ticket_prio_high_%d", t.ID))
		btnPrioUrgent := inline.Data("🔴", fmt.Sprintf("ticket_prio_urgent_%d", t.ID))
		rows = append(rows, inline.Row(btnPrioLow, btnPrioNormal, btnPrioHigh, btnPrioUrgent))

		inline.Inline(rows...)

		return c.Send(sb.String(), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		}, inline)
	}
}

// ═══════════════════════════════════════════
// ─── پاسخ به تیکت ───
// ═══════════════════════════════════════════

func ReplyTicketHandler(cfg *config.Config, db *database.DB, ticketID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		setTicketReply(c.Sender().ID, ticketID)
		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("💬 پاسخ برای تیکت #%d رو بفرست.", ticketID),
		})
	}
}

// ═══════════════════════════════════════════
// ─── بستن تیکت ───
// ═══════════════════════════════════════════

func CloseTicketHandler(cfg *config.Config, db *database.DB, ticketID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		t, err := db.GetTicket(ticketID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ تیکت پیدا نشد"})
		}

		oldStatus := t.Status

		if err := db.CloseTicket(ticketID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		if notifier != nil && oldStatus != database.TicketClosed {
			notifier.StatusChanged(t, oldStatus, database.TicketClosed)
		}

		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("✅ تیکت #%d بسته شد.", ticketID),
		})
	}
}

// ═══════════════════════════════════════════
// ─── تغییر وضعیت به pending ───
// ═══════════════════════════════════════════

func PendingTicketHandler(cfg *config.Config, db *database.DB, ticketID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		t, err := db.GetTicket(ticketID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ تیکت پیدا نشد"})
		}

		oldStatus := t.Status

		if err := db.SetTicketStatus(ticketID, database.TicketPending); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		if notifier != nil && oldStatus != database.TicketPending {
			notifier.StatusChanged(t, oldStatus, database.TicketPending)
		}

		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("✅ وضعیت تیکت #%d به «در حال بررسی» تغییر کرد.", ticketID),
		})
	}
}

// ═══════════════════════════════════════════
// ─── تغییر اولویت ───
// ═══════════════════════════════════════════

func SetTicketPriorityHandler(cfg *config.Config, db *database.DB, ticketID int64, priority database.TicketPriority) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		t, err := db.GetTicket(ticketID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ تیکت پیدا نشد"})
		}

		oldPriority := t.Priority

		if err := db.SetTicketPriority(ticketID, priority); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		if notifier != nil && oldPriority != priority {
			notifier.PriorityChanged(t, oldPriority, priority)
		}

		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("✅ اولویت تیکت #%d به %s تغییر کرد.", ticketID, notify.PriorityLabel(priority)),
		})
	}
}

// ═══════════════════════════════════════════
// ─── انتقال تیکت ───
// ═══════════════════════════════════════════

func TransferTicketHandler(cfg *config.Config, db *database.DB, ticketID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		depts := db.GetActiveDepartments()
		if len(depts) == 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دپارتمانی وجود نداره"})
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row
		for _, d := range depts {
			btn := inline.Data(
				fmt.Sprintf("%s %s", d.Emoji, d.Name),
				fmt.Sprintf("ticket_transfer_to_%d_%d", ticketID, d.ID),
			)
			rows = append(rows, inline.Row(btn))
		}
		inline.Inline(rows...)

		return c.Send(fmt.Sprintf("🔄 تیکت #%d رو به کدوم دپارتمان منتقل کنم؟", ticketID), inline)
	}
}

// ═══════════════════════════════════════════
// ─── انجام انتقال ───
// ═══════════════════════════════════════════

func TransferTicketToHandler(cfg *config.Config, db *database.DB, ticketID, newDeptID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		t, err := db.GetTicket(ticketID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ تیکت پیدا نشد"})
		}

		newDept := db.GetDepartmentByID(newDeptID)
		if newDept == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دپارتمان پیدا نشد"})
		}

		oldDeptID := t.DeptID

		if err := db.TransferTicket(ticketID, newDeptID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		// نوتیف به کاربر
		if notifier != nil {
			notifier.TicketTransferred(t, oldDeptID, newDeptID)
		}

		// نوتیف به ادمین‌های دپارتمان جدید
		t.DeptID = newDeptID
		if notifier != nil {
			notifier.NotifyDeptAdmins(t, "🔄 تیکت جدید (منتقل‌شده)")
		}

		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("✅ تیکت #%d به %s %s منتقل شد.", ticketID, newDept.Emoji, newDept.Name),
		})
	}
}

// ═══════════════════════════════════════════
// ─── هندلر ورودی پاسخ ───
// ═══════════════════════════════════════════

func TicketReplyInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	user := c.Sender()
	role := db.GetAdminRole(user.ID, cfg.ADMIN_ID)
	if role != database.RoleSuperAdmin && role != database.RoleSupport {
		return nil
	}

	ticketID := getTicketReply(user.ID)
	if ticketID == 0 {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	setTicketReply(user.ID, 0)

	t, err := db.GetTicket(ticketID)
	if err != nil {
		return c.Send("❌ تیکت پیدا نشد.")
	}

	if err := db.AddTicketMessage(ticketID, user.ID, true, text); err != nil {
		return c.Send("❌ " + err.Error())
	}

	if notifier != nil {
		notifier.AdminReply(t, text)
	}

	return c.Send(fmt.Sprintf("✅ پاسخ برای تیکت #%d ارسال شد.", ticketID))
}
