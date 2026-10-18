// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package notify

import (
	"fmt"
	"log"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

type LogTopicKey string

const (
	LogTopicOrders    LogTopicKey = "orders"
	LogTopicFinance   LogTopicKey = "finance"
	LogTopicUsers     LogTopicKey = "users"
	LogTopicSupport   LogTopicKey = "support"
	LogTopicXUI       LogTopicKey = "xui"
	LogTopicErrors    LogTopicKey = "errors"
	LogTopicDaily     LogTopicKey = "daily"
	LogTopicAdminActs LogTopicKey = "admin_acts"
)

var LogTopicTitles = map[LogTopicKey]string{
	LogTopicOrders:    "📢 گزارش خرید",
	LogTopicFinance:   "💰 گزارش مالی",
	LogTopicUsers:     "👤 کاربران",
	LogTopicSupport:   "🎫 پشتیبانی",
	LogTopicXUI:       "🖥 3x-ui",
	LogTopicErrors:    "⚠️ خطاها",
	LogTopicDaily:     "📊 گزارش روزانه",
	LogTopicAdminActs: "⚙️ عملیات ادمین",
}

func AdminLog(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, topicKey LogTopicKey, text string, opts ...interface{}) {
	chatID := getLogChatID(db)
	if chatID == 0 {
		return
	}

	topic, err := db.GetLogTopic(string(topicKey))
	if err != nil {
		log.Printf("⚠️ تاپیک لاگ %s پیدا نشد: %v", topicKey, err)
		return
	}

	sendOpts := &AsliBot.SendOptions{
		ParseMode: AsliBot.ModeHTML,
		ThreadID:  int(topic.ThreadID),
	}

	var extraMarkup *AsliBot.ReplyMarkup
	for _, opt := range opts {
		if so, ok := opt.(*AsliBot.SendOptions); ok {
			if so.ParseMode != "" {
				sendOpts.ParseMode = so.ParseMode
			}
			if so.ReplyMarkup != nil {
				extraMarkup = so.ReplyMarkup
			}
		}
		if rm, ok := opt.(*AsliBot.ReplyMarkup); ok {
			extraMarkup = rm
		}
	}
	if extraMarkup != nil {
		sendOpts.ReplyMarkup = extraMarkup
	}

	_, err = bot.Send(AsliBot.ChatID(chatID), text, sendOpts)
	if err != nil {
		log.Printf("⚠️ خطا در ارسال به تاپیک %s: %v", topicKey, err)
	}
}

func AdminLogWithPhoto(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, topicKey LogTopicKey, photo *AsliBot.Photo, opts ...interface{}) {
	chatID := getLogChatID(db)
	if chatID == 0 {
		return
	}

	topic, err := db.GetLogTopic(string(topicKey))
	if err != nil {
		return
	}

	sendOpts := &AsliBot.SendOptions{
		ParseMode: AsliBot.ModeHTML,
		ThreadID:  int(topic.ThreadID),
	}
	var extraMarkup *AsliBot.ReplyMarkup
	for _, opt := range opts {
		if so, ok := opt.(*AsliBot.SendOptions); ok && so.ReplyMarkup != nil {
			extraMarkup = so.ReplyMarkup
		}
		if rm, ok := opt.(*AsliBot.ReplyMarkup); ok {
			extraMarkup = rm
		}
	}
	if extraMarkup != nil {
		sendOpts.ReplyMarkup = extraMarkup
	}

	_, _ = bot.Send(AsliBot.ChatID(chatID), photo, sendOpts)
}

// SendWithLog — هم به چت خصوصی مدیر می‌فرستد، هم به تاپیک لاگ
// اگر privateMarkup داده شود، در چت خصوصی با دکمه می‌رود
// اگر logMarkup داده شود، در تاپیک با دکمه می‌رود
// اگر logMarkup nil باشد، بدون دکمه در تاپیک می‌رود
func SendWithLog(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, topicKey LogTopicKey, text string, privateMarkup *AsliBot.ReplyMarkup, logMarkup *AsliBot.ReplyMarkup) {
	// ۱. به چت خصوصی مدیر
	if cfg.ADMIN_ID != 0 {
		opts := &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}
		if privateMarkup != nil {
			opts.ReplyMarkup = privateMarkup
		}
		_, _ = bot.Send(AsliBot.ChatID(cfg.ADMIN_ID), text, opts)
	}

	// ۲. به تاپیک لاگ گروه
	AdminLog(bot, cfg, db, topicKey, text, logMarkup)
}

// SendWithLogPhoto — نسخه عکس‌دار
func SendWithLogPhoto(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, topicKey LogTopicKey, photo *AsliBot.Photo, privateMarkup *AsliBot.ReplyMarkup, logMarkup *AsliBot.ReplyMarkup) {
	if cfg.ADMIN_ID != 0 {
		opts := &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}
		if privateMarkup != nil {
			opts.ReplyMarkup = privateMarkup
		}
		_, _ = bot.Send(AsliBot.ChatID(cfg.ADMIN_ID), photo, opts)
	}

	AdminLogWithPhoto(bot, cfg, db, topicKey, photo, logMarkup)
}

func getLogChatID(db *database.DB) int64 {
	s := db.GetSettings()
	if s == nil {
		return 0
	}
	if s.LogGroupChatID == 0 {
		return 0
	}
	return s.LogGroupChatID
}

func SetupLogTopics(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) error {
	chatID := getLogChatID(db)
	if chatID == 0 {
		return fmt.Errorf("گروه لاگ تنظیم نشده")
	}

	chat := &AsliBot.Chat{ID: chatID}

	for _, key := range []LogTopicKey{
		LogTopicOrders, LogTopicFinance, LogTopicUsers, LogTopicSupport,
		LogTopicXUI, LogTopicErrors, LogTopicDaily, LogTopicAdminActs,
	} {
		title := LogTopicTitles[key]

		existing, err := db.GetLogTopic(string(key))
		if err == nil && existing != nil {
			continue
		}

		topic, err := bot.CreateTopic(chat, &AsliBot.Topic{Name: title})
		if err != nil {
			log.Printf("⚠️ خطا در ساخت تاپیک %s: %v", title, err)
			continue
		}
		if topic == nil {
			continue
		}

		_ = db.SetLogTopic(string(key), int64(topic.ThreadID), title)

		welcome := fmt.Sprintf("✅ تاپیک <b>%s</b> آماده است.\n\nاز این به بعد همه رویدادهای مربوطه اینجا ارسال می‌شوند.", title)
		_, _ = bot.Send(chat, welcome, &AsliBot.SendOptions{
			ParseMode: AsliBot.ModeHTML,
			ThreadID:  topic.ThreadID,
		})
	}

	return nil
}
