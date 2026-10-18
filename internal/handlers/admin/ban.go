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
	banAwaiting = make(map[int64]string) // adminID → "ban" | "unban"
	banMu       sync.Mutex
)

func setBanAwaiting(uid int64, action string) {
	banMu.Lock()
	defer banMu.Unlock()
	if action == "" {
		delete(banAwaiting, uid)
	} else {
		banAwaiting[uid] = action
	}
}

func getBanAwaiting(uid int64) string {
	banMu.Lock()
	defer banMu.Unlock()
	return banAwaiting[uid]
}

func BanStartHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		setBanAwaiting(c.Sender().ID, "ban")
		return c.Respond(&AsliBot.CallbackResponse{})
	}
}

func UnbanStartHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		setBanAwaiting(c.Sender().ID, "unban")
		return c.Respond(&AsliBot.CallbackResponse{})
	}
}

// BanInputHandler — ورودی آیدی کاربر برای بلاک/آنبلاک
func BanInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	action := getBanAwaiting(uid)
	if action == "" {
		return nil
	}
	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}
	setBanAwaiting(uid, "")

	id, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return c.Send("❌ آیدی نامعتبر.")
	}
	u, err := db.GetUser(id)
	if err != nil {
		return c.Send("❌ کاربر پیدا نشد.")
	}

	if action == "ban" {
		if u.Banned {
			return c.Send("⚠️ این کاربر از قبل بلاک است.")
		}
		if err := db.BanUser(id, "بلاک توسط ادمین "+c.Sender().FirstName); err != nil {
			return c.Send("❌ " + err.Error())
		}
		_, _ = c.Bot().Send(AsliBot.ChatID(id), "🚫 حساب شما توسط ادمین بلاک شد.")
		return c.Send(fmt.Sprintf("✅ کاربر <code>%d</code> بلاک شد.", id),
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}

	// unban
	if !u.Banned {
		return c.Send("⚠️ این کاربر بلاک نیست.")
	}
	if err := db.UnbanUser(id); err != nil {
		return c.Send("❌ " + err.Error())
	}
	_, _ = c.Bot().Send(AsliBot.ChatID(id), "✅ حساب شما آنبلاک شد.")
	return c.Send(fmt.Sprintf("✅ کاربر <code>%d</code> آنبلاک شد.", id),
		&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
}
