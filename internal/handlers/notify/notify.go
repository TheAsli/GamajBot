// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package notify

import (
	"fmt"
	"log"
	"strings"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

type Notifier struct {
	bot *AsliBot.Bot
	cfg *config.Config
	db  *database.DB
}

func New(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) *Notifier {
	return &Notifier{bot: bot, cfg: cfg, db: db}
}

// ═══════════════════════════════════════════
// ─── نوتیف تیکت جدید ───
// ═══════════════════════════════════════════

func (n *Notifier) NewTicket(t *database.Ticket) {
	n.NotifyDeptAdmins(t, "🆕 تیکت جدید")
}

// ═══════════════════════════════════════════
// ─── نوتیف به ادمین‌های دپارتمان ───
// ═══════════════════════════════════════════

func (n *Notifier) NotifyDeptAdmins(t *database.Ticket, title string) {
	// ادمین اصلی
	n.sendToAdmin(n.cfg.ADMIN_ID, t, title)

	admins := n.db.GetAllAdmins()
	for _, a := range admins {
		if a.Role != database.RoleSupport && a.Role != database.RoleSuperAdmin {
			continue
		}
		if a.UserID == n.cfg.ADMIN_ID {
			continue
		}
		hasAccess := false
		for _, deptID := range a.DeptIDs {
			if deptID == t.DeptID {
				hasAccess = true
				break
			}
		}
		if hasAccess || a.Role == database.RoleSuperAdmin {
			n.sendToAdmin(a.UserID, t, title)
		}
	}
}

func (n *Notifier) sendToAdmin(adminID int64, t *database.Ticket, title string) {
	dept := n.db.GetDepartmentByID(t.DeptID)
	deptName := "—"
	if dept != nil {
		deptName = dept.Emoji + " " + dept.Name
	}

	firstMsg := ""
	if len(t.Messages) > 0 {
		firstMsg = t.Messages[0].Text
	}

	msg := fmt.Sprintf(
		"%s\n\n"+
			"🎫 #%d\n"+
			"👤 کاربر: %s (@%s)\n"+
			"🏢 دپارتمان: %s\n"+
			"⚡ اولویت: %s\n"+
			"📝 موضوع: %s\n\n"+
			"💬 %s",
		title,
		t.ID,
		t.UserFirstName, t.UserUsername,
		deptName,
		PriorityLabel(t.Priority),
		t.Subject,
		firstMsg,
	)

	inline := &AsliBot.ReplyMarkup{}
	btnView := inline.Data("👁 دیدن", fmt.Sprintf("ticket_view_%d", t.ID))
	inline.Inline(inline.Row(btnView))

	_, err := n.bot.Send(AsliBot.ChatID(adminID), msg, inline)
	if err != nil {
		log.Printf("❌ خطا در ارسال نوتیف به ادمین %d: %v", adminID, err)
	}
}

// ═══════════════════════════════════════════
// ─── نوتیف پاسخ ادمین به کاربر ───
// ═══════════════════════════════════════════

func (n *Notifier) AdminReply(t *database.Ticket, text string) {
	msg := fmt.Sprintf(
		"💬 پاسخ پشتیبانی\n\n"+
			"🎫 تیکت #%d\n"+
			"📝 موضوع: %s\n\n"+
			"%s",
		t.ID, t.Subject, text,
	)

	inline := &AsliBot.ReplyMarkup{}
	btnView := inline.Data("👁 دیدن تیکت", fmt.Sprintf("sup_view_%d", t.ID))
	inline.Inline(inline.Row(btnView))

	_, err := n.bot.Send(AsliBot.ChatID(t.UserID), msg, inline)
	if err != nil {
		log.Printf("❌ خطا در ارسال پاسخ به کاربر %d: %v", t.UserID, err)
	}
}

// ═══════════════════════════════════════════
// ─── نوتیف پاسخ کاربر به ادمین ───
// ═══════════════════════════════════════════

func (n *Notifier) UserReply(t *database.Ticket, text string) {
	if t.AssignedTo != 0 {
		n.sendReplyNotify(t.AssignedTo, t, text)
		return
	}

	admins := n.db.GetAllAdmins()
	for _, a := range admins {
		if a.Role != database.RoleSupport && a.Role != database.RoleSuperAdmin {
			continue
		}
		hasAccess := false
		for _, deptID := range a.DeptIDs {
			if deptID == t.DeptID {
				hasAccess = true
				break
			}
		}
		if hasAccess || a.Role == database.RoleSuperAdmin {
			n.sendReplyNotify(a.UserID, t, text)
		}
	}
}

func (n *Notifier) sendReplyNotify(adminID int64, t *database.Ticket, text string) {
	msg := fmt.Sprintf(
		"💬 پاسخ جدید از کاربر\n\n"+
			"🎫 تیکت #%d\n"+
			"👤 %s\n\n"+
			"%s",
		t.ID, t.UserFirstName, text,
	)

	inline := &AsliBot.ReplyMarkup{}
	btnView := inline.Data("👁 دیدن", fmt.Sprintf("ticket_view_%d", t.ID))
	inline.Inline(inline.Row(btnView))

	_, _ = n.bot.Send(AsliBot.ChatID(adminID), msg, inline)
}

// ═══════════════════════════════════════════
// ─── نوتیف بسته شدن تیکت ───
// ═══════════════════════════════════════════

func (n *Notifier) TicketClosed(t *database.Ticket) {
	msg := fmt.Sprintf(
		"🔒 تیکت #%d بسته شد.\n\n"+
			"📝 موضوع: %s",
		t.ID, t.Subject,
	)
	_, _ = n.bot.Send(AsliBot.ChatID(t.UserID), msg)
}

// ═══════════════════════════════════════════
// ─── نوتیف تغییر وضعیت ───
// ═══════════════════════════════════════════

func (n *Notifier) StatusChanged(t *database.Ticket, oldStatus, newStatus database.TicketStatus) {
	msg := fmt.Sprintf(
		"📊 وضعیت تیکت شما تغییر کرد\n\n"+
			"🎫 تیکت #%d\n"+
			"📝 موضوع: %s\n\n"+
			"از: %s\n"+
			"به: %s",
		t.ID, t.Subject,
		StatusLabel(oldStatus),
		StatusLabel(newStatus),
	)

	inline := &AsliBot.ReplyMarkup{}
	btnView := inline.Data("👁 دیدن تیکت", fmt.Sprintf("sup_view_%d", t.ID))
	inline.Inline(inline.Row(btnView))

	_, _ = n.bot.Send(AsliBot.ChatID(t.UserID), msg, inline)
}

// ═══════════════════════════════════════════
// ─── نوتیف تغییر اولویت ───
// ═══════════════════════════════════════════

func (n *Notifier) PriorityChanged(t *database.Ticket, oldPrio, newPrio database.TicketPriority) {
	msg := fmt.Sprintf(
		"⚡ اولویت تیکت شما تغییر کرد\n\n"+
			"🎫 تیکت #%d\n"+
			"📝 موضوع: %s\n\n"+
			"از: %s\n"+
			"به: %s",
		t.ID, t.Subject,
		PriorityLabel(oldPrio),
		PriorityLabel(newPrio),
	)

	inline := &AsliBot.ReplyMarkup{}
	btnView := inline.Data("👁 دیدن تیکت", fmt.Sprintf("sup_view_%d", t.ID))
	inline.Inline(inline.Row(btnView))

	_, _ = n.bot.Send(AsliBot.ChatID(t.UserID), msg, inline)
}

// ═══════════════════════════════════════════
// ─── نوتیف انتقال تیکت ───
// ═══════════════════════════════════════════

func (n *Notifier) TicketTransferred(t *database.Ticket, oldDeptID, newDeptID int64) {
	oldDept := n.db.GetDepartmentByID(oldDeptID)
	newDept := n.db.GetDepartmentByID(newDeptID)

	oldName := "—"
	if oldDept != nil {
		oldName = oldDept.Emoji + " " + oldDept.Name
	}
	newName := "—"
	if newDept != nil {
		newName = newDept.Emoji + " " + newDept.Name
	}

	msg := fmt.Sprintf(
		"🔄 تیکت شما به دپارتمان دیگه ارجاع شد\n\n"+
			"🎫 تیکت #%d\n"+
			"📝 موضوع: %s\n\n"+
			"از: %s\n"+
			"به: %s",
		t.ID, t.Subject,
		oldName, newName,
	)

	inline := &AsliBot.ReplyMarkup{}
	btnView := inline.Data("👁 دیدن تیکت", fmt.Sprintf("sup_view_%d", t.ID))
	inline.Inline(inline.Row(btnView))

	_, _ = n.bot.Send(AsliBot.ChatID(t.UserID), msg, inline)
}

// ═══════════════════════════════════════════
// ─── نوتیف تغییر دپارتمان ادمین ───
// ═══════════════════════════════════════════

func (n *Notifier) AdminDeptsChanged(adminID int64, depts []*database.Department) {
	var sb strings.Builder
	sb.WriteString("📋 دپارتمان‌های شما تغییر کرد\n\n")

	if len(depts) == 0 {
		sb.WriteString("❌ هیچ دپارتمانی به شما اختصاص داده نشده.")
	} else {
		sb.WriteString("✅ دپارتمان‌های جدید شما:\n")
		for _, d := range depts {
			sb.WriteString(fmt.Sprintf("• %s %s\n", d.Emoji, d.Name))
		}
	}

	_, _ = n.bot.Send(AsliBot.ChatID(adminID), sb.String())
}

// ═══════════════════════════════════════════
// ─── لیبل‌ها ───
// ═══════════════════════════════════════════

func PriorityLabel(p database.TicketPriority) string {
	switch p {
	case database.PriorityLow:
		return "🟢 کم"
	case database.PriorityNormal:
		return "🟡 معمولی"
	case database.PriorityHigh:
		return "🟠 زیاد"
	case database.PriorityUrgent:
		return "🔴 فوری"
	}
	return "🟡 معمولی"
}

func StatusLabel(s database.TicketStatus) string {
	switch s {
	case database.TicketOpen:
		return "🟢 باز"
	case database.TicketPending:
		return "🟡 در حال بررسی"
	case database.TicketAnswered:
		return "🔵 پاسخ داده شده"
	case database.TicketWaitingUser:
		return "🟠 منتظر کاربر"
	case database.TicketClosed:
		return "⚫ بسته"
	}
	return "❓"
}
