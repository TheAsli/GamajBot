// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package middleware

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/theasli/GamajBot/internal/config"

	AsliBot "gopkg.in/telebot.v4"
)

func ForceJoin(b *AsliBot.Bot, cfg *config.Config) AsliBot.MiddlewareFunc {
	return func(next AsliBot.HandlerFunc) AsliBot.HandlerFunc {
		return func(c AsliBot.Context) error {
			if !cfg.ForceJoinEnabled || cfg.ForceJoinChannel == "" {
				return next(c)
			}

			user := c.Sender()
			if user == nil {
				return next(c)
			}

			if cfg.ADMIN_ID != 0 && user.ID == cfg.ADMIN_ID {
				return next(c)
			}

			if IsMember(b, cfg.ForceJoinChannel, user) {
				return next(c)
			}

			return SendJoinPrompt(c, cfg)
		}
	}
}

// IsMember با استفاده مستقیم از Telegram API
func IsMember(b *AsliBot.Bot, channel string, user *AsliBot.User) bool {
	log.Printf("🔍 چک: channel=%s, user=%d", channel, user.ID)

	// درخواست مستقیم به Telegram API
	resp, err := b.Raw("getChatMember", map[string]string{
		"chat_id": channel,
		"user_id": strconv.FormatInt(user.ID, 10),
	})
	if err != nil {
		log.Printf("❌ Raw: %v", err)
		return false
	}

	// پارس کردن پاسخ
	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			Status string `json:"status"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		log.Printf("❌ Unmarshal: %v", err)
		return false
	}

	log.Printf("📊 status=%s, ok=%v", result.Result.Status, result.Ok)

	if !result.Ok {
		return false
	}

	switch result.Result.Status {
	case "creator", "administrator", "member", "restricted":
		return true
	default:
		return false
	}
}

func SendJoinPrompt(c AsliBot.Context, cfg *config.Config) error {
	inline := &AsliBot.ReplyMarkup{}
	btnJoin := inline.URL("📢 عضویت در کانال", cfg.ForceJoinChannelURL)
	btnCheck := inline.Data("✅ عضو شدم", "force_join_check")
	inline.Inline(
		inline.Row(btnJoin),
		inline.Row(btnCheck),
	)

	return c.Send(
		"⚠️ برای استفاده از ربات، باید اول توی کانال ما عضو بشی:\n\n"+
			"بعد از عضویت، روی «✅ عضو شدم» بزن.",
		inline,
	)
}
