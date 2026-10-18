// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package support

import (
	"fmt"
	"strings"
	"sync"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/handlers/notify"
	"github.com/theasli/GamajBot/internal/menus"

	AsliBot "gopkg.in/telebot.v4"
)

var (
	supportAwaiting = make(map[int64]*SupportState)
	supportMu       sync.Mutex
	notifier        *notify.Notifier
)

type SupportState struct {
	Step     string
	DeptID   int64
	Priority database.TicketPriority
	Subject  string
	TicketID int64
}

func Init(n *notify.Notifier) {
	notifier = n
}

func setSupportState(userID int64, s *SupportState) {
	supportMu.Lock()
	defer supportMu.Unlock()
	if s == nil {
		delete(supportAwaiting, userID)
	} else {
		supportAwaiting[userID] = s
	}
}

func getSupportState(userID int64) *SupportState {
	supportMu.Lock()
	defer supportMu.Unlock()
	return supportAwaiting[userID]
}

// ─── منوی پشتیبانی ───

func SupportHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		s := db.GetSettings()
		if !s.SupportEnabled {
			return c.Send("🎫 پشتیبانی غیرفعاله.")
		}

		inline := &AsliBot.ReplyMarkup{}
		btnNew := inline.Data("🆕 تیکت جدید", "sup_new")
		btnMy := inline.Data("📋 تیکت‌های من", "sup_my")
		inline.Inline(
			inline.Row(btnNew),
			inline.Row(btnMy),
		)

		return c.Send(
			"🎫 پشتیبانی\n\n"+
				"چیکار می‌خوای بکنی؟",
			inline,
		)
	}
}

// ─── تیکت جدید ───

func NewTicketHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		depts := db.GetActiveDepartments()
		if len(depts) == 0 {
			return c.Respond(&AsliBot.CallbackResponse{
				Text: "❌ هیچ دپارتمانی فعال نیست.",
			})
		}

		setSupportState(c.Sender().ID, &SupportState{Step: "dept"})

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row
		for _, d := range depts {
			btn := inline.Data(
				fmt.Sprintf("%s %s", d.Emoji, d.Name),
				fmt.Sprintf("sup_dept_%d", d.ID),
			)
			rows = append(rows, inline.Row(btn))
		}
		inline.Inline(rows...)

		return c.Send("🏢 دپارتمان رو انتخاب کن:", inline)
	}
}

// ─── انتخاب دپارتمان ───

func SelectDeptHandler(cfg *config.Config, db *database.DB, deptID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		state := getSupportState(c.Sender().ID)
		if state == nil || state.Step != "dept" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}

		dept := db.GetDepartmentByID(deptID)
		if dept == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دپارتمان پیدا نشد"})
		}

		state.DeptID = deptID
		state.Step = "priority"
		setSupportState(c.Sender().ID, state)

		inline := &AsliBot.ReplyMarkup{}
		btnLow := inline.Data("🟢 کم", "sup_prio_low")
		btnNormal := inline.Data("🟡 معمولی", "sup_prio_normal")
		btnHigh := inline.Data("🟠 زیاد", "sup_prio_high")
		btnUrgent := inline.Data("🔴 فوری", "sup_prio_urgent")
		inline.Inline(
			inline.Row(btnLow, btnNormal),
			inline.Row(btnHigh, btnUrgent),
		)

		_ = c.Delete()
		return c.Send(fmt.Sprintf(
			"🏢 دپارتمان: %s %s\n\n⚡ اولویت تیکت رو انتخاب کن:",
			dept.Emoji, dept.Name,
		), inline)
	}
}

// ─── انتخاب اولویت ───

func SelectPriorityHandler(cfg *config.Config, db *database.DB, priority database.TicketPriority) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		state := getSupportState(c.Sender().ID)
		if state == nil || state.Step != "priority" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}

		state.Priority = priority
		state.Step = "subject"
		setSupportState(c.Sender().ID, state)

		_ = c.Delete()
		return c.Send(fmt.Sprintf(
			"⚡ اولویت: %s\n\n📝 موضوع تیکت رو بفرست:",
			notify.PriorityLabel(priority),
		))
	}
}

// ─── تیکت‌های من ───

func MyTicketsHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		tickets := db.GetUserTickets(c.Sender().ID)
		if len(tickets) == 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "📋 تیکتی نداری."})
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📋 تیکت‌های من (%d)\n\n", len(tickets)))

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		for i, t := range tickets {
			if i >= 10 {
				break
			}
			sb.WriteString(fmt.Sprintf(
				"#%d — %s\n%s | %s\n\n",
				t.ID, t.Subject,
				notify.StatusLabel(t.Status),
				notify.PriorityLabel(t.Priority),
			))

			btn := inline.Data(
				fmt.Sprintf("#%d — %s", t.ID, t.Subject),
				fmt.Sprintf("sup_view_%d", t.ID),
			)
			rows = append(rows, inline.Row(btn))
		}
		inline.Inline(rows...)

		return c.Send(sb.String(), inline)
	}
}

// ─── دیدن تیکت ───

func ViewMyTicketHandler(cfg *config.Config, db *database.DB, ticketID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		t, err := db.GetTicket(ticketID)
		if err != nil || t.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ تیکت پیدا نشد"})
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"🎫 تیکت #%d\n"+
				"📝 %s\n"+
				"📊 %s | ⚡ %s\n\n"+
				"💬 پیام‌ها:\n",
			t.ID, t.Subject,
			notify.StatusLabel(t.Status),
			notify.PriorityLabel(t.Priority),
		))

		for _, m := range t.Messages {
			who := "👤 شما"
			if m.IsAdmin {
				who = "👮 پشتیبانی"
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", who, m.Text))
		}

		var inline *AsliBot.ReplyMarkup
		if t.Status != database.TicketClosed {
			inline = &AsliBot.ReplyMarkup{}
			btnReply := inline.Data("💬 پاسخ", fmt.Sprintf("sup_reply_%d", t.ID))
			inline.Inline(inline.Row(btnReply))
		}

		if inline != nil {
			return c.Send(sb.String(), inline)
		}
		return c.Send(sb.String())
	}
}

// ─── پاسخ به تیکت ───

func ReplyMyTicketHandler(cfg *config.Config, db *database.DB, ticketID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		t, err := db.GetTicket(ticketID)
		if err != nil || t.UserID != c.Sender().ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ تیکت پیدا نشد"})
		}

		setSupportState(c.Sender().ID, &SupportState{
			Step:     "reply",
			TicketID: ticketID,
		})

		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("💬 پاسخ برای تیکت #%d رو بفرست.", ticketID),
		})
	}
}

// ─── هندلر ورودی ───

func InputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	state := getSupportState(c.Sender().ID)
	if state == nil {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	user := c.Sender()

	switch state.Step {
	case "subject":
		state.Subject = text
		state.Step = "message"
		setSupportState(user.ID, state)
		return c.Send("📝 متن تیکت رو بفرست:")

	case "message":
		t, err := db.CreateTicket(
			user.ID, user.FirstName, user.Username,
			state.DeptID, state.Subject, text, state.Priority,
		)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}
		setSupportState(user.ID, nil)

		if notifier != nil {
			notifier.NewTicket(t)
		}

		return c.Send(fmt.Sprintf(
			"✅ تیکت <code>#%d</code> ثبت شد.\n\n"+
				"⚡ اولویت: %s\n"+
				"📊 وضعیت: %s\n\n"+
				"به‌زودی پشتیبانی پاسخ می‌ده.",
			t.ID,
			notify.PriorityLabel(t.Priority),
			notify.StatusLabel(t.Status),
		), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		}, menus.MainMenu)

	case "reply":
		t, err := db.GetTicket(state.TicketID)
		if err != nil {
			setSupportState(user.ID, nil)
			return c.Send("❌ تیکت پیدا نشد.")
		}

		if err := db.AddTicketMessage(state.TicketID, user.ID, false, text); err != nil {
			return c.Send("❌ " + err.Error())
		}
		setSupportState(user.ID, nil)

		if notifier != nil {
			notifier.UserReply(t, text)
		}

		return c.Send(fmt.Sprintf("✅ پاسخ برای تیکت #%d ثبت شد.", state.TicketID), menus.MainMenu)
	}

	return nil
}
