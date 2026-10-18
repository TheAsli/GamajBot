// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package notify

import (
	"fmt"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

func LogAdminAction(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, adminName string, adminID int64, action string, details string) {
	msg := fmt.Sprintf(
		"⚙️ <b>عملیات ادمین</b>\n\n"+
			"👮 ادمین: %s (<code>%d</code>)\n"+
			"📝 عمل: <b>%s</b>\n"+
			"📋 جزئیات: %s",
		adminName, adminID, action, details,
	)
	AdminLog(bot, cfg, db, LogTopicAdminActs, msg)
}

func LogAdminActionWithTarget(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, adminName string, adminID int64, action string, targetName string, targetID int64, details string) {
	msg := fmt.Sprintf(
		"⚙️ <b>عملیات ادمین</b>\n\n"+
			"👮 ادمین: %s (<code>%d</code>)\n"+
			"📝 عمل: <b>%s</b>\n"+
			"👤 هدف: %s (<code>%d</code>)\n"+
			"📋 جزئیات: %s",
		adminName, adminID, action, targetName, targetID, details,
	)
	AdminLog(bot, cfg, db, LogTopicAdminActs, msg)
}
