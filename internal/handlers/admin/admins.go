// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

var (
	adminAwaiting = make(map[int64]string)
	adminStateMu  sync.Mutex
)

func SetAwaiting(userID int64, action string) {
	adminStateMu.Lock()
	defer adminStateMu.Unlock()
	if action == "" {
		delete(adminAwaiting, userID)
	} else {
		adminAwaiting[userID] = action
	}
}

func getAwaiting(userID int64) string {
	adminStateMu.Lock()
	defer adminStateMu.Unlock()
	return adminAwaiting[userID]
}

// ═══════════════════════════════════════════
// ─── لیست ادمین‌ها ───
// ═══════════════════════════════════════════

func AdminsHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		admins := db.GetAllAdmins()

		var sb strings.Builder
		sb.WriteString("👮 ادمین‌ها\n\n")
		sb.WriteString(fmt.Sprintf("👑 ادمین اصلی: <code>%d</code>\n\n", cfg.ADMIN_ID))

		if len(admins) == 0 {
			sb.WriteString("هیچ ادمین اضافه‌ای ثبت نشده.")
		} else {
			sb.WriteString("ادمین‌های اضافه:\n")
			for i, a := range admins {
				deptNames := ""
				for _, dID := range a.DeptIDs {
					d := db.GetDepartmentByID(dID)
					if d != nil {
						deptNames += d.Emoji + " "
					}
				}
				sb.WriteString(fmt.Sprintf(
					"%d. <code>%d</code> — %s %s\n",
					i+1, a.UserID, roleLabel(a.Role), deptNames,
				))
			}
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		for _, a := range admins {
			if a.Role != database.RoleSupport && a.Role != database.RoleSuperAdmin {
				continue
			}
			btn := inline.Data(
				fmt.Sprintf("🏢 %d", a.UserID),
				fmt.Sprintf("adm_admin_depts_%d", a.UserID),
			)
			rows = append(rows, inline.Row(btn))
		}

		btnAdd := inline.Data("➕ افزودن", "adm_add")
		btnRm := inline.Data("➖ حذف", "adm_rm")
		rows = append(rows, inline.Row(btnAdd, btnRm))

		inline.Inline(rows...)

		return c.Send(sb.String(), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		}, inline)
	}
}

// ═══════════════════════════════════════════
// ─── اختصاص دپارتمان‌ها ───
// ═══════════════════════════════════════════

func AdminDeptsHandler(cfg *config.Config, db *database.DB, adminID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		depts := db.GetAllDepartments()
		if len(depts) == 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ هیچ دپارتمانی نیست"})
		}

		admin := db.GetAdminByID(adminID)
		if admin == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ ادمین پیدا نشد"})
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		for _, d := range depts {
			selected := "⬜"
			for _, did := range admin.DeptIDs {
				if did == d.ID {
					selected = "✅"
					break
				}
			}
			btn := inline.Data(
				fmt.Sprintf("%s %s %s", selected, d.Emoji, d.Name),
				fmt.Sprintf("adm_dept_toggle_%d_%d", adminID, d.ID),
			)
			rows = append(rows, inline.Row(btn))
		}

		inline.Inline(rows...)

		return c.Send(fmt.Sprintf(
			"🏢 دپارتمان‌های ادمین <code>%d</code>:\n\n"+
				"هر دپارتمانی که می‌خوای رو کلیک کن:",
			adminID,
		), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		}, inline)
	}
}

// ═══════════════════════════════════════════
// ─── تغییر وضعیت یه دپارتمان ───
// ═══════════════════════════════════════════

func AdminDeptToggleHandler(cfg *config.Config, db *database.DB, adminID, deptID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		admin := db.GetAdminByID(adminID)
		if admin == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ ادمین پیدا نشد"})
		}

		var newDepts []int64
		found := false
		for _, d := range admin.DeptIDs {
			if d == deptID {
				found = true
				continue
			}
			newDepts = append(newDepts, d)
		}
		if !found {
			newDepts = append(newDepts, deptID)
		}

		if err := db.SetAdminDepts(adminID, newDepts); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		// نوتیف به ادمین
		if notifier != nil {
			var depts []*database.Department
			for _, dID := range newDepts {
				d := db.GetDepartmentByID(dID)
				if d != nil {
					depts = append(depts, d)
				}
			}
			notifier.AdminDeptsChanged(adminID, depts)
		}

		return c.Respond(&AsliBot.CallbackResponse{
			Text: "✅ تغییر کرد.",
		})
	}
}

// ═══════════════════════════════════════════
// ─── افزودن ادمین ───
// ═══════════════════════════════════════════

func AddAdminHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		SetAwaiting(c.Sender().ID, "add_admin_id")
		return c.Send(
			"👤 آیدی عددی ادمین جدید رو بفرست:\n\n" +
				"برای لغو /cancel رو بزن.",
		)
	}
}

// ═══════════════════════════════════════════
// ─── حذف ادمین ───
// ═══════════════════════════════════════════

func RemoveAdminHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		SetAwaiting(c.Sender().ID, "remove_admin_id")
		return c.Send(
			"👤 آیدی عددی ادمین رو بفرست:\n\n" +
				"برای لغو /cancel رو بزن.",
		)
	}
}

// ═══════════════════════════════════════════
// ─── هندلر ورودی ───
// ═══════════════════════════════════════════

func InputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	user := c.Sender()
	if db.GetAdminRole(user.ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		return nil
	}

	action := getAwaiting(user.ID)
	if action == "" {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	switch action {
	case "add_admin_id":
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return c.Send("❌ آیدی نامعتبره.")
		}
		adminStateMu.Lock()
		adminAwaiting[user.ID] = fmt.Sprintf("add_admin_role:%d", id)
		adminStateMu.Unlock()

		inline := &AsliBot.ReplyMarkup{}
		btnSuper := inline.Data("👑 ادمین اصلی", "role_super")
		btnFinance := inline.Data("💰 مالی", "role_finance")
		btnSupport := inline.Data("👥 پشتیبانی", "role_support")
		btnViewer := inline.Data("👤 فقط دیدن", "role_viewer")
		inline.Inline(
			inline.Row(btnSuper),
			inline.Row(btnFinance, btnSupport),
			inline.Row(btnViewer),
		)
		return c.Send("🎭 نقش رو انتخاب کن:", inline)

	case "remove_admin_id":
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return c.Send("❌ آیدی نامعتبره.")
		}
		SetAwaiting(user.ID, "")
		if err := db.RemoveAdmin(id); err != nil {
			return c.Send("❌ " + err.Error())
		}
		return c.Send(fmt.Sprintf("✅ ادمین <code>%d</code> حذف شد.", id), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		})
	}

	return nil
}

// ═══════════════════════════════════════════
// ─── انتخاب نقش ───
// ═══════════════════════════════════════════

func AddAdminRoleHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		adminStateMu.Lock()
		pending := adminAwaiting[c.Sender().ID]
		adminStateMu.Unlock()

		if !strings.HasPrefix(pending, "add_admin_role:") {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}

		idStr := strings.TrimPrefix(pending, "add_admin_role:")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		data := c.Callback().Data
		data = strings.TrimPrefix(data, "\f")

		var role database.AdminRole
		switch data {
		case "role_super":
			role = database.RoleSuperAdmin
		case "role_finance":
			role = database.RoleFinance
		case "role_support":
			role = database.RoleSupport
		case "role_viewer":
			role = database.RoleViewer
		default:
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}

		SetAwaiting(c.Sender().ID, "")

		if err := db.AddAdmin(id, role, nil, c.Sender().ID); err != nil {
			return c.Send("❌ " + err.Error())
		}

		_ = c.Delete()
		return c.Send(fmt.Sprintf(
			"✅ ادمین <code>%d</code> با نقش %s اضافه شد.",
			id, roleLabel(role),
		), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		})
	}
}

// ═══════════════════════════════════════════
// ─── لیبل نقش ───
// ═══════════════════════════════════════════

func roleLabel(role database.AdminRole) string {
	switch role {
	case database.RoleSuperAdmin:
		return "👑 ادمین اصلی"
	case database.RoleFinance:
		return "💰 مالی"
	case database.RoleSupport:
		return "👥 پشتیبانی"
	case database.RoleViewer:
		return "👤 فقط دیدن"
	}
	return "❓"
}
