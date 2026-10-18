// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package menus

import AsliBot "gopkg.in/telebot.v4"

var MainMenu = &AsliBot.ReplyMarkup{ResizeKeyboard: true}

var (
	BtnBuySubscription   = MainMenu.Text("🏠 خرید اشتراک")
	BtnBuyOthersProducts = MainMenu.Text("🛒 سایر محصولات")
	BtnMyServices        = MainMenu.Text("📦 سرویس های من")
	BtnMyAccount         = MainMenu.Text("👤 حساب من")
	BtnCharge            = MainMenu.Text("💰 افزایش موجودی")
	BtnSupport           = MainMenu.Text("🎫 پشتیبانی")
	BtnHelp              = MainMenu.Text("❓ راهنما")
	BtnDev               = MainMenu.Text("🛠 توسعه دهنده")
	BtnCom               = MainMenu.Text("🚧 به زودی")
)

func Setup(placeholder string) {
	MainMenu.Placeholder = placeholder
	MainMenu.Reply(
		MainMenu.Row(BtnBuyOthersProducts, BtnBuySubscription),
		MainMenu.Row(BtnMyServices, BtnMyAccount),
		MainMenu.Row(BtnSupport, BtnHelp, BtnCharge),
		MainMenu.Row(BtnDev, BtnCom),
	)
}
