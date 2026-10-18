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
	"github.com/theasli/GamajBot/internal/notify"

	AsliBot "gopkg.in/telebot.v4"
)

var (
	logSetupAwaiting = make(map[int64]string)
	logSetupMu       sync.Mutex
	botInstance      *AsliBot.Bot
)

func setLogSetupAwaiting(uid int64, action string) {
	logSetupMu.Lock()
	defer logSetupMu.Unlock()
	if action == "" {
		delete(logSetupAwaiting, uid)
	} else {
		logSetupAwaiting[uid] = action
	}
}

func getLogSetupAwaiting(uid int64) string {
	logSetupMu.Lock()
	defer logSetupMu.Unlock()
	return logSetupAwaiting[uid]
}

func SetBotInstance(b *AsliBot.Bot) {
	botInstance = b
}

func LogSetupHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		s := db.GetSettings()
		topics, _ := db.GetAllLogTopics()

		var sb strings.Builder
		sb.WriteString("📝 <b>تنظیمات گروه لاگ</b>\n\n")
		if s.LogGroupChatID == 0 {
			sb.WriteString("⚠️ گروه لاگ تنظیم نشده.\n\n")
			sb.WriteString("1. یک سوپرگروه بساز\n")
			sb.WriteString("2. حالت Topics (Forum) را فعال کن\n")
			sb.WriteString("3. ربات را ادمین کامل کن (با حق manage topics)\n")
			sb.WriteString("4. chat_id گروه را اینجا ثبت کن\n")
		} else {
			sb.WriteString(fmt.Sprintf("✅ گروه لاگ: <code>%d</code>\n", s.LogGroupChatID))
			sb.WriteString(fmt.Sprintf("📊 تعداد تاپیک‌ها: %d\n\n", len(topics)))
			for _, t := range topics {
				sb.WriteString(fmt.Sprintf("• %s (thread: %d)\n", t.Title, t.ThreadID))
			}
		}

		inline := &AsliBot.ReplyMarkup{}
		rows := []AsliBot.Row{
			inline.Row(
				inline.Data("📝 ثبت chat_id گروه", "adm_log_set_chat"),
			),
			inline.Row(
				inline.Data("🏗 ساخت تاپیک‌ها", "adm_log_build"),
			),
			inline.Row(
				inline.Data("🗑 پاک کردن تنظیمات", "adm_log_clear"),
			),
			inline.Row(
				inline.Data("⬅️ برگشت", "adm_settings"),
			),
		}
		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func LogSetupSetChatHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setLogSetupAwaiting(c.Sender().ID, "chat_id")
		_ = c.Respond()
		return c.Send("📝 chat_id گروه را بفرست (مثلاً <code>-1001234567890</code>):\n\n"+
			"برای گرفتن chat_id، ربات را در گروه ادمین کن و یک پیام بفرست، سپس از @userinfobot یا @RawDataBot استفاده کن.",
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}

func LogSetupBuildHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "⏳ در حال ساخت..."})

		if botInstance == nil {
			return c.Send("❌ نمونه ربات در دسترس نیست")
		}

		if err := notify.SetupLogTopics(botInstance, cfg, db); err != nil {
			return c.Send("❌ " + err.Error())
		}
		return c.Send("✅ تاپیک‌ها ساخته شدند.")
	}
}

func LogSetupClearHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		_ = db.DeleteAllLogTopics()
		s := db.GetSettings()
		s.LogGroupChatID = 0
		_ = db.UpdateSettings(s)
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ پاک شد"})
		return LogSetupHandler(cfg, db)(c)
	}
}

func LogSetupInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	action := getLogSetupAwaiting(uid)
	if action == "" {
		return nil
	}
	if db.GetAdminRole(uid, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		setLogSetupAwaiting(uid, "")
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}
	setLogSetupAwaiting(uid, "")

	if action == "chat_id" {
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil || id == 0 {
			return c.Send("❌ chat_id نامعتبر. یک عدد منفی بفرست (مثلاً -1001234567890):")
		}
		s := db.GetSettings()
		s.LogGroupChatID = id
		if err := db.UpdateSettings(s); err != nil {
			return c.Send("❌ " + err.Error())
		}
		return c.Send("✅ chat_id ثبت شد. حالا «🏗 ساخت تاپیک‌ها» را بزن.")
	}
	return nil
}
