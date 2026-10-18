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

	AsliBot "gopkg.in/telebot.v4"
)

var (
	settingsAwaiting = make(map[int64]string)
	settingsMu       sync.Mutex
)

func setSettingsAwaiting(userID int64, key string) {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	if key == "" {
		delete(settingsAwaiting, userID)
	} else {
		settingsAwaiting[userID] = key
	}
}

func getSettingsAwaiting(userID int64) string {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	return settingsAwaiting[userID]
}

func SettingsHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		s := db.GetSettings()
		currency := db.GetCurrency()

		msg := fmt.Sprintf(
			"⚙️ تنظیمات ربات\n\n"+
				"💱 واحد پول: <b>%s</b>\n"+
				"💰 درصد سود دلار: <b>%.1f%%</b>\n"+
				"💵 حداقل شارژ: <b>%d تومان</b>\n"+
				"💵 حداکثر شارژ: <b>%d تومان</b>\n"+
				"📊 مبالغ آماده: <b>%v دلار</b>\n"+
				"🔧 حالت تعمیر: <b>%s</b>\n"+
				"🎫 پشتیبانی: <b>%s</b>",
			currencyLabel(currency),
			s.DollarRateMargin,
			s.ChargeMin,
			s.ChargeMax,
			s.QuickAmounts,
			boolLabel(s.MaintenanceMode),
			boolLabel(s.SupportEnabled),
		)

		inline := &AsliBot.ReplyMarkup{}
		btnCurrency := inline.Data("💱 واحد پول", "set_currency")
		btnMargin := inline.Data("💰 درصد سود", "set_margin")
		btnMin := inline.Data("📉 حداقل شارژ", "set_min")
		btnMax := inline.Data("📈 حداکثر شارژ", "set_max")
		btnQuick := inline.Data("📊 مبالغ آماده", "set_quick")
		btnMaintenance := inline.Data("🔧 حالت تعمیر", "set_maintenance")
		btnSupport := inline.Data("🎫 پشتیبانی", "set_support")
		btnAnalytics := inline.Data("📈 آمار پیشرفته", "adm_analytics")

		inline.Inline(
			inline.Row(btnCurrency),
			inline.Row(btnMargin),
			inline.Row(btnMin, btnMax),
			inline.Row(btnQuick),
			inline.Row(btnMaintenance, btnSupport),
			inline.Row(btnAnalytics),
		)

		return c.Send(msg, &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func SetCurrencyHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(inline.Data("💵 دلار (USD)", "set_cur_USD")),
			inline.Row(inline.Data("🪙 تتر (USDT)", "set_cur_USDT")),
			inline.Row(inline.Data("💰 تومان (TOMAN)", "set_cur_TOMAN")),
			inline.Row(inline.Data("💴 ریال (RIAL)", "set_cur_RIAL")),
		)

		return c.Send("💱 واحد پول ربات را انتخاب کن:", inline)
	}
}

func SetCurrencySelectHandler(cfg *config.Config, db *database.DB, currency database.CurrencyType) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		if err := db.SetCurrency(currency); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر واحد پول", currencyLabel(currency))

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ واحد پول تغییر کرد"})
		_ = c.Delete()
		return SettingsHandler(cfg, db)(c)
	}
}

func SetMarginHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setSettingsAwaiting(c.Sender().ID, "margin")
		return c.Respond(&AsliBot.CallbackResponse{Text: "✏️ درصد سود رو بفرست (مثال: 5)"})
	}
}

func SetMinHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setSettingsAwaiting(c.Sender().ID, "min")
		return c.Respond(&AsliBot.CallbackResponse{Text: "✏️ حداقل شارژ (تومان) رو بفرست"})
	}
}

func SetMaxHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setSettingsAwaiting(c.Sender().ID, "max")
		return c.Respond(&AsliBot.CallbackResponse{Text: "✏️ حداکثر شارژ (تومان) رو بفرست"})
	}
}

func SetQuickHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setSettingsAwaiting(c.Sender().ID, "quick")
		return c.Respond(&AsliBot.CallbackResponse{Text: "✏️ مبالغ دلاری با کاما (مثال: 1,5,10,20)"})
	}
}

func ToggleMaintenanceHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		s := db.GetSettings()
		s.MaintenanceMode = !s.MaintenanceMode
		_ = db.UpdateSettings(s)

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر حالت تعمیر", boolLabel(s.MaintenanceMode))

		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("🔧 حالت تعمیر: %s", boolLabel(s.MaintenanceMode)),
		})
	}
}

func ToggleSupportHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		s := db.GetSettings()
		s.SupportEnabled = !s.SupportEnabled
		_ = db.UpdateSettings(s)

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر وضعیت پشتیبانی", boolLabel(s.SupportEnabled))

		return c.Respond(&AsliBot.CallbackResponse{
			Text: fmt.Sprintf("🎫 پشتیبانی: %s", boolLabel(s.SupportEnabled)),
		})
	}
}

func SettingsInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	user := c.Sender()
	if db.GetAdminRole(user.ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		return nil
	}

	key := getSettingsAwaiting(user.ID)
	if key == "" {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	setSettingsAwaiting(user.ID, "")

	s := db.GetSettings()

	switch key {
	case "margin":
		v, err := strconv.ParseFloat(text, 64)
		if err != nil || v < 0 {
			return c.Send("❌ عدد نامعتبره.")
		}
		s.DollarRateMargin = v
		_ = db.UpdateSettings(s)
		logAction(botInstance, cfg, db, user, "تغییر درصد سود", fmt.Sprintf("%.1f%%", v))
		return c.Send(fmt.Sprintf("✅ درصد سود: %.1f%%", v))

	case "min":
		v, err := strconv.ParseInt(text, 10, 64)
		if err != nil || v < 0 {
			return c.Send("❌ عدد نامعتبره.")
		}
		s.ChargeMin = v
		_ = db.UpdateSettings(s)
		logAction(botInstance, cfg, db, user, "تغییر حداقل شارژ", fmt.Sprintf("%d تومان", v))
		return c.Send(fmt.Sprintf("✅ حداقل شارژ: %d تومان", v))

	case "max":
		v, err := strconv.ParseInt(text, 10, 64)
		if err != nil || v < 0 {
			return c.Send("❌ عدد نامعتبره.")
		}
		s.ChargeMax = v
		_ = db.UpdateSettings(s)
		logAction(botInstance, cfg, db, user, "تغییر حداکثر شارژ", fmt.Sprintf("%d تومان", v))
		return c.Send(fmt.Sprintf("✅ حداکثر شارژ: %d تومان", v))

	case "quick":
		var amounts []int64
		for _, p := range strings.Split(text, ",") {
			n, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
			if err == nil && n > 0 {
				amounts = append(amounts, n)
			}
		}
		if len(amounts) == 0 {
			return c.Send("❌ هیچ مبلغ معتبری پیدا نشد.")
		}
		s.QuickAmounts = amounts
		_ = db.UpdateSettings(s)
		logAction(botInstance, cfg, db, user, "تغییر مبالغ آماده", fmt.Sprintf("%v دلار", amounts))
		return c.Send(fmt.Sprintf("✅ مبالغ آماده: %v دلار", amounts))
	}

	return nil
}

func boolLabel(b bool) string {
	if b {
		return "✅ فعال"
	}
	return "❌ غیرفعال"
}

func currencyLabel(c database.CurrencyType) string {
	switch c {
	case database.CurrencyUSD:
		return "💵 دلار (USD)"
	case database.CurrencyUSDT:
		return "🪙 تتر (USDT)"
	case database.CurrencyToman:
		return "💰 تومان (TOMAN)"
	case database.CurrencyRial:
		return "💴 ریال (RIAL)"
	}
	return string(c)
}
