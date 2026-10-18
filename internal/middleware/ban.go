// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package middleware

import (
	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

func Ban(cfg *config.Config, db *database.DB) AsliBot.MiddlewareFunc {
	return func(next AsliBot.HandlerFunc) AsliBot.HandlerFunc {
		return func(c AsliBot.Context) error {
			user := c.Sender()
			if user == nil {
				return next(c)
			}
			// ادمین‌ها معاف
			if db.GetAdminRole(user.ID, cfg.ADMIN_ID) != "" {
				return next(c)
			}
			if db.IsUserBanned(user.ID) {
				msg := "🚫 حساب شما بلاک شده است. برای رفع بلاک با پشتیبانی تماس بگیرید."
				if c.Callback() != nil {
					return c.Respond(&AsliBot.CallbackResponse{Text: msg})
				}
				return c.Send(msg)
			}
			return next(c)
		}
	}
}
