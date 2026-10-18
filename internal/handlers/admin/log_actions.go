// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/notify"

	AsliBot "gopkg.in/telebot.v4"
)

func logAction(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, admin *AsliBot.User, action, details string) {
	if bot == nil {
		return
	}
	notify.LogAdminAction(bot, cfg, db, admin.FirstName, admin.ID, action, details)
}

func logActionTarget(bot *AsliBot.Bot, cfg *config.Config, db *database.DB, admin *AsliBot.User, action, targetName string, targetID int64, details string) {
	if bot == nil {
		return
	}
	notify.LogAdminActionWithTarget(bot, cfg, db, admin.FirstName, admin.ID, action, targetName, targetID, details)
}

func logActionSimple(cfg *config.Config, db *database.DB, adminID int64, adminName, action, details string) {
	if botInstance == nil {
		return
	}
	notify.LogAdminAction(botInstance, cfg, db, adminName, adminID, action, details)
}

func logActionSimpleTarget(cfg *config.Config, db *database.DB, adminID int64, adminName, action, targetName string, targetID int64, details string) {
	if botInstance == nil {
		return
	}
	notify.LogAdminActionWithTarget(botInstance, cfg, db, adminName, adminID, action, targetName, targetID, details)
}

var _ = fmt.Sprintf
