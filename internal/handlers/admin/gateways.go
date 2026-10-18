// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"sync"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/payment"

	AsliBot "gopkg.in/telebot.v4"
)

var (
	gwAwaiting = make(map[int64]string)
	gwMu       sync.Mutex
)

func setGWAwaiting(uid int64, action string) {
	gwMu.Lock()
	defer gwMu.Unlock()
	if action == "" {
		delete(gwAwaiting, uid)
	} else {
		gwAwaiting[uid] = action
	}
}

func getGWAwaiting(uid int64) string {
	gwMu.Lock()
	defer gwMu.Unlock()
	return gwAwaiting[uid]
}

func GatewaysHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		g := db.GetPaymentGateways()

		msg := fmt.Sprintf(
			"🚪 مدیریت درگاه‌ها\n\n"+
				"📊 وضعیت:\n"+
				"   💳 کارت‌به‌کارت دستی: %s\n"+
				"   🧊 CubePay: %s\n"+
				"   💎 Tetraminator: %s\n"+
				"   🌐 AbanGateway: %s\n\n"+
				"💳 کارت دستی: <code>%s</code>\n"+
				"👤 صاحب کارت: %s",
			boolLabel(g.ManualCardEnabled),
			boolLabel(g.CubePayEnabled),
			boolLabel(g.TetraminatorEnabled),
			boolLabel(g.AbanEnabled),
			g.ManualCardNumber,
			g.ManualCardHolder,
		)

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(inline.Data("💳 کارت‌به‌کارت دستی", "adm_gw_manual")),
			inline.Row(inline.Data("🧊 CubePay", "adm_gw_cube")),
			inline.Row(inline.Data("💎 Tetraminator", "adm_gw_tetra")),
			inline.Row(inline.Data("🌐 AbanGateway", "adm_gw_aban")),
		)

		return c.Send(msg, &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func GatewayManualHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		g := db.GetPaymentGateways()

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(inline.Data("🔄 فعال/غیرفعال", "adm_gw_manual_toggle")),
			inline.Row(inline.Data("📝 شماره کارت", "adm_gw_manual_number")),
			inline.Row(inline.Data("👤 صاحب کارت", "adm_gw_manual_holder")),
		)

		return c.Send(fmt.Sprintf(
			"💳 کارت‌به‌کارت دستی\n\n"+
				"وضعیت: %s\n"+
				"شماره کارت: <code>%s</code>\n"+
				"صاحب کارت: %s",
			boolLabel(g.ManualCardEnabled),
			g.ManualCardNumber,
			g.ManualCardHolder,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func GatewayManualToggleHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		g := db.GetPaymentGateways()
		g.ManualCardEnabled = !g.ManualCardEnabled
		_ = db.UpdatePaymentGateways(g)

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر وضعیت کارت دستی", boolLabel(g.ManualCardEnabled))

		return c.Respond(&AsliBot.CallbackResponse{Text: "💳 کارت دستی: " + boolLabel(g.ManualCardEnabled)})
	}
}

func GatewayManualNumberHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setGWAwaiting(c.Sender().ID, "manual_number")
		return c.Respond(&AsliBot.CallbackResponse{Text: "📝 شماره کارت رو بفرست:"})
	}
}

func GatewayManualHolderHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setGWAwaiting(c.Sender().ID, "manual_holder")
		return c.Respond(&AsliBot.CallbackResponse{Text: "👤 نام صاحب کارت رو بفرست:"})
	}
}

func GatewayCubeHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		g := db.GetPaymentGateways()
		tokenMask := "—"
		if len(g.CubePayToken) > 8 {
			tokenMask = g.CubePayToken[:8] + "..."
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(inline.Data("🔄 فعال/غیرفعال", "adm_gw_cube_toggle")),
			inline.Row(inline.Data("🔑 توکن API", "adm_gw_cube_token")),
		)

		return c.Send(fmt.Sprintf(
			"🧊 CubePay\n\nوضعیت: %s\nتوکن: <code>%s</code>",
			boolLabel(g.CubePayEnabled), tokenMask,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func GatewayCubeToggleHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		g := db.GetPaymentGateways()
		g.CubePayEnabled = !g.CubePayEnabled
		_ = db.UpdatePaymentGateways(g)

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر وضعیت CubePay", boolLabel(g.CubePayEnabled))

		return c.Respond(&AsliBot.CallbackResponse{Text: "🧊 CubePay: " + boolLabel(g.CubePayEnabled)})
	}
}

func GatewayCubeTokenHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setGWAwaiting(c.Sender().ID, "cube_token")
		return c.Respond(&AsliBot.CallbackResponse{Text: "🔑 توکن API CubePay رو بفرست:"})
	}
}

func GatewayTetraHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		g := db.GetPaymentGateways()
		keyMask := "—"
		if len(g.TetraminatorAPIKey) > 8 {
			keyMask = g.TetraminatorAPIKey[:8] + "..."
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(inline.Data("🔄 فعال/غیرفعال", "adm_gw_tetra_toggle")),
			inline.Row(inline.Data("🔑 API Key", "adm_gw_tetra_key")),
			inline.Row(inline.Data("🌐 Base URL", "adm_gw_tetra_url")),
		)

		return c.Send(fmt.Sprintf(
			"💎 Tetraminator\n\n"+
				"وضعیت: %s\n"+
				"API Key: <code>%s</code>\n"+
				"Base URL: <code>%s</code>",
			boolLabel(g.TetraminatorEnabled), keyMask, g.TetraminatorBaseURL,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func GatewayTetraToggleHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		g := db.GetPaymentGateways()
		g.TetraminatorEnabled = !g.TetraminatorEnabled
		_ = db.UpdatePaymentGateways(g)

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر وضعیت Tetraminator", boolLabel(g.TetraminatorEnabled))

		return c.Respond(&AsliBot.CallbackResponse{Text: "💎 Tetraminator: " + boolLabel(g.TetraminatorEnabled)})
	}
}

func GatewayTetraKeyHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setGWAwaiting(c.Sender().ID, "tetra_key")
		return c.Respond(&AsliBot.CallbackResponse{Text: "🔑 API Key Tetraminator رو بفرست:"})
	}
}

func GatewayTetraURLHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setGWAwaiting(c.Sender().ID, "tetra_url")
		return c.Respond(&AsliBot.CallbackResponse{Text: "🌐 Base URL Tetraminator رو بفرست:"})
	}
}

func GatewayAbanHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		g := db.GetPaymentGateways()
		keyMask := "—"
		if len(g.AbanAPIKey) > 12 {
			keyMask = g.AbanAPIKey[:12] + "..."
		}
		secretMask := "—"
		if len(g.AbanWebhookSecret) > 8 {
			secretMask = g.AbanWebhookSecret[:8] + "..."
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(inline.Data("🔄 فعال/غیرفعال", "adm_gw_aban_toggle")),
			inline.Row(inline.Data("🔑 API Key", "adm_gw_aban_key")),
			inline.Row(inline.Data("🔒 Webhook Secret", "adm_gw_aban_secret")),
		)

		return c.Send(fmt.Sprintf(
			"🌐 AbanGateway\n\n"+
				"وضعیت: %s\n"+
				"API Key: <code>%s</code>\n"+
				"Webhook Secret: <code>%s</code>\n\n"+
				"📡 Webhook URL:\n<code>http://YOUR_SERVER:%s/webhook/abangateway</code>",
			boolLabel(g.AbanEnabled), keyMask, secretMask, cfg.WEBHOOK_PORT,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func GatewayAbanToggleHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		g := db.GetPaymentGateways()
		g.AbanEnabled = !g.AbanEnabled
		_ = db.UpdatePaymentGateways(g)

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر وضعیت AbanGateway", boolLabel(g.AbanEnabled))

		return c.Respond(&AsliBot.CallbackResponse{Text: "🌐 AbanGateway: " + boolLabel(g.AbanEnabled)})
	}
}

func GatewayAbanKeyHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setGWAwaiting(c.Sender().ID, "aban_key")
		return c.Respond(&AsliBot.CallbackResponse{Text: "🔑 API Key آبان گیت‌وی رو بفرست (ag_live_... یا ag_test_...):"})
	}
}

func GatewayAbanSecretHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setGWAwaiting(c.Sender().ID, "aban_secret")
		return c.Respond(&AsliBot.CallbackResponse{Text: "🔒 Webhook Secret آبان گیت‌وی رو بفرست:"})
	}
}

func GWInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	if db.GetAdminRole(uid, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		return nil
	}
	action := getGWAwaiting(uid)
	if action == "" {
		return nil
	}
	text := c.Text()
	setGWAwaiting(uid, "")

	switch action {
	case "manual_number":
		_ = db.SetPaymentGatewayField("manual_card_number", text)
		return c.Send("✅ شماره کارت ذخیره شد.")
	case "manual_holder":
		_ = db.SetPaymentGatewayField("manual_card_holder", text)
		return c.Send("✅ صاحب کارت ذخیره شد.")
	case "cube_token":
		_ = db.SetPaymentGatewayField("cubepay_token", text)
		payment.InvalidateCubeClient()
		return c.Send("✅ توکن CubePay ذخیره و کلاینت بازسازی شد.")
	case "tetra_key":
		_ = db.SetPaymentGatewayField("tetraminator_api_key", text)
		return c.Send("✅ API Key Tetraminator ذخیره شد.")
	case "tetra_url":
		_ = db.SetPaymentGatewayField("tetraminator_base_url", text)
		return c.Send("✅ Base URL Tetraminator ذخیره شد.")
	case "aban_key":
		_ = db.SetPaymentGatewayField("aban_api_key", text)
		return c.Send("✅ API Key AbanGateway ذخیره شد.")
	case "aban_secret":
		_ = db.SetPaymentGatewayField("aban_webhook_secret", text)
		return c.Send("✅ Webhook Secret AbanGateway ذخیره شد.")
	}
	return nil
}
