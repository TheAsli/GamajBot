// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"fmt"

	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/i18n"

	AsliBot "gopkg.in/telebot.v4"
)

func LanguageHandler(c AsliBot.Context) error {
	_ = c.Respond()

	inline := &AsliBot.ReplyMarkup{}
	inline.Inline(
		inline.Row(
			inline.Data("🇮🇷 فارسی", "lang_set_fa"),
			inline.Data("🇬🇧 English", "lang_set_en"),
		),
	)

	return c.Send("🌍 زبان را انتخاب کن / Select language:", inline)
}

func LanguageSetHandler(c AsliBot.Context, db *database.DB, lang string) error {
	uid := c.Sender().ID
	if err := db.SetUserLanguage(uid, lang); err != nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
	}

	_ = c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(i18n.Lang(lang), "language_changed")})
	_ = c.Delete()

	msg := fmt.Sprintf("✅ %s", i18n.LangLabel(i18n.Lang(lang)))
	return c.Send(msg)
}
