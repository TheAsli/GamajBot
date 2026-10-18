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

	AsliBot "gopkg.in/telebot.v4"
)

var (
	deptAwaiting = make(map[int64]string)
	deptMu       sync.Mutex
)

func setDeptAwaiting(userID int64, action string) {
	deptMu.Lock()
	defer deptMu.Unlock()
	if action == "" {
		delete(deptAwaiting, userID)
	} else {
		deptAwaiting[userID] = action
	}
}

func getDeptAwaiting(userID int64) string {
	deptMu.Lock()
	defer deptMu.Unlock()
	return deptAwaiting[userID]
}

// DepartmentsHandler — لیست دپارتمان‌ها
func DepartmentsHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		depts := db.GetAllDepartments()

		var sb strings.Builder
		sb.WriteString("🏢 دپارتمان‌ها\n\n")

		if len(depts) == 0 {
			sb.WriteString("هیچ دپارتمانی ثبت نشده.")
		} else {
			for i, d := range depts {
				status := "✅"
				if !d.Active {
					status = "❌"
				}
				sb.WriteString(fmt.Sprintf(
					"%d. %s %s %s\n   %s\n\n",
					i+1, d.Emoji, d.Name, status, d.Description,
				))
			}
		}

		inline := &AsliBot.ReplyMarkup{}
		btnAdd := inline.Data("➕ افزودن", "dept_add")
		btnRm := inline.Data("➖ حذف", "dept_rm")
		btnToggle := inline.Data("🔄 فعال/غیرفعال", "dept_toggle")
		inline.Inline(
			inline.Row(btnAdd, btnRm),
			inline.Row(btnToggle),
		)

		return c.Send(sb.String(), inline)
	}
}

// AddDeptHandler — افزودن دپارتمان
func AddDeptHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setDeptAwaiting(c.Sender().ID, "add_dept")
		return c.Send(
			"🏢 اسم دپارتمان رو بفرست:\n\n" +
				"مثال: 🛠 پشتیبانی فنی | سوالات فنی",
		)
	}
}

// RemoveDeptHandler — حذف دپارتمان
func RemoveDeptHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setDeptAwaiting(c.Sender().ID, "remove_dept")
		return c.Send("🏢 آیدی عددی دپارتمان رو بفرست:")
	}
}

// ToggleDeptHandler — فعال/غیرفعال
func ToggleDeptHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setDeptAwaiting(c.Sender().ID, "toggle_dept")
		return c.Send("🏢 آیدی عددی دپارتمان رو بفرست:")
	}
}

// DeptInputHandler — هندلر ورودی
func DeptInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	user := c.Sender()
	if db.GetAdminRole(user.ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		return nil
	}

	action := getDeptAwaiting(user.ID)
	if action == "" {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	setDeptAwaiting(user.ID, "")

	switch action {
	case "add_dept":
		parts := strings.SplitN(text, "|", 2)
		nameEmoji := strings.TrimSpace(parts[0])
		desc := ""
		if len(parts) > 1 {
			desc = strings.TrimSpace(parts[1])
		}

		// جدا کردن ایموجی از اسم
		fields := strings.Fields(nameEmoji)
		emoji := "🏢"
		name := nameEmoji
		if len(fields) > 0 {
			emoji = fields[0]
			name = strings.Join(fields[1:], " ")
		}
		if name == "" {
			name = nameEmoji
		}

		d, err := db.AddDepartment(name, emoji, desc)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}
		return c.Send(fmt.Sprintf("✅ دپارتمان <code>%d</code> (%s %s) اضافه شد.", d.ID, d.Emoji, d.Name), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		})

	case "remove_dept":
		var id int64
		_, err := fmt.Sscanf(text, "%d", &id)
		if err != nil {
			return c.Send("❌ آیدی نامعتبره.")
		}
		if err := db.RemoveDepartment(id); err != nil {
			return c.Send("❌ " + err.Error())
		}
		return c.Send(fmt.Sprintf("✅ دپارتمان <code>%d</code> حذف شد.", id), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		})

	case "toggle_dept":
		var id int64
		_, err := fmt.Sscanf(text, "%d", &id)
		if err != nil {
			return c.Send("❌ آیدی نامعتبره.")
		}
		if err := db.ToggleDepartment(id); err != nil {
			return c.Send("❌ " + err.Error())
		}
		return c.Send(fmt.Sprintf("✅ وضعیت دپارتمان <code>%d</code> تغییر کرد.", id), &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
		})
	}

	return nil
}
