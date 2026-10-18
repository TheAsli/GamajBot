// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package Services

import (
	"fmt"

	"github.com/theasli/GamajBot/internal/menus"
	AsliBot "gopkg.in/telebot.v4"
)

const (
	MsgCoom = "سلام %s 👋\nاین بخش از ربات هنوز آماده نیست. 🚧 به زودی..."
)

// ═══════════════════════════════════════════
// ─── منوی اصلی ───
// ═══════════════════════════════════════════

func BuySubscriptionHandler(c AsliBot.Context) error {
	msg := fmt.Sprintf(MsgCoom, c.Sender().FirstName)
	return c.Send(msg, menus.MainMenu)
}

func BuyOthersProducts(c AsliBot.Context) error {
	msg := fmt.Sprintf(MsgCoom, c.Sender().FirstName)
	return c.Send(msg, menus.MainMenu)
}

func MyServicesHandler(c AsliBot.Context) error {
	msg := fmt.Sprintf(MsgCoom, c.Sender().FirstName)
	return c.Send(msg, menus.MainMenu)
}

//func BuySubscriptionHandler(c AsliBot.Context) error {
//	return c.Send(MsgCoom, menus.MainMenu)
//}

//func MyServicesHandler(c AsliBot.Context) error {
//	return c.Send(MsgCoom, menus.MainMenu)
//}

//func StartHandler(c AsliBot.Context) error {
//	msg := fmt.Sprintf(MsgWelcome, c.Sender().FirstName)
//	return c.Send(msg, menus.MainMenu)
//}
