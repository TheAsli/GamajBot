// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/vpn"
	"github.com/theasli/GamajBot/internal/xui"

	AsliBot "gopkg.in/telebot.v4"
)

type xuiState struct {
	Step     string
	PanelID  int64
	Name     string
	BaseURL  string
	ServerIP string
	AuthMeth string
	Token    string
	Username string
	Password string
}

var (
	xuiStates = make(map[int64]*xuiState)
	xuiMu     sync.Mutex
)

func setXUIState(uid int64, s *xuiState) {
	xuiMu.Lock()
	defer xuiMu.Unlock()
	if s == nil {
		delete(xuiStates, uid)
	} else {
		xuiStates[uid] = s
	}
}

func getXUIState(uid int64) *xuiState {
	xuiMu.Lock()
	defer xuiMu.Unlock()
	return xuiStates[uid]
}

type xuiManageState struct {
	Step      string
	PanelID   int64
	InboundID int64
	Email     string
}

var (
	xuiManageStates = make(map[int64]*xuiManageState)
	xuiManageMu     sync.Mutex
)

func setXUIManageState(uid int64, s *xuiManageState) {
	xuiManageMu.Lock()
	defer xuiManageMu.Unlock()
	if s == nil {
		delete(xuiManageStates, uid)
	} else {
		xuiManageStates[uid] = s
	}
}

func getXUIManageState(uid int64) *xuiManageState {
	xuiManageMu.Lock()
	defer xuiManageMu.Unlock()
	return xuiManageStates[uid]
}

func getXUIClient(panel *database.XUIPanel) *xui.PanelClient {
	return xui.New(xui.Config{
		BaseURL:    panel.BaseURL,
		AuthMethod: xui.AuthMethod(panel.AuthMethod),
		APIToken:   panel.APIToken,
		Username:   panel.Username,
		Password:   panel.Password,
	})
}

func XUIPanelHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		return showXUIPanels(c, db)
	}
}

func showXUIPanels(c AsliBot.Context, db *database.DB) error {
	panels := db.GetAllXUIPanels()

	var sb strings.Builder
	sb.WriteString("🖥 پنل‌های 3x-ui\n\n")
	if len(panels) == 0 {
		sb.WriteString("هیچ پنلی ثبت نشده.")
	} else {
		sb.WriteString(fmt.Sprintf("تعداد: %d پنل", len(panels)))
	}

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	for _, p := range panels {
		status := "✅"
		if !p.Active {
			status = "❌"
		}
		authLabel := "🔑"
		if p.AuthMethod == "login" {
			authLabel = "👤"
		}
		panelTypeLabel := p.PanelType
		if panelTypeLabel == "" {
			panelTypeLabel = "3x-ui"
		}
		label := fmt.Sprintf("%s %s %s (%s)", status, authLabel, p.Name, panelTypeLabel)
		rows = append(rows, inline.Row(
			inline.Data(label, fmt.Sprintf("adm_xui_view_%d", p.ID)),
		))
	}

	rows = append(rows, inline.Row(
		inline.Data("➕ افزودن پنل", "adm_xui_add"),
	))

	inline.Inline(rows...)
	return c.Send(sb.String(), inline)
}

func XUIPanelViewHandler(cfg *config.Config, db *database.DB, panelID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		p := db.GetXUIPanel(panelID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ پنل پیدا نشد"})
		}
		return showXUIPanelDetail(c, db, p)
	}
}

func showXUIPanelDetail(c AsliBot.Context, db *database.DB, p *database.XUIPanel) error {
	status := "✅ فعال"
	if !p.Active {
		status = "❌ غیرفعال"
	}

	authInfo := "🔑 API Token"
	if p.AuthMethod == "login" {
		authInfo = "👤 یوزرنیم/پسورد"
	}

	panelTypeLabel := p.PanelType
	if panelTypeLabel == "" {
		panelTypeLabel = "3x-ui"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"🖥 <b>%s</b>\n\n"+
			"📦 نوع: %s\n"+
			"🌐 آدرس: <code>%s</code>\n"+
			"🆔 IP سرور: <code>%s</code>\n"+
			"🔐 احراز هویت: %s\n"+
			"📊 وضعیت: %s",
		p.Name, panelTypeLabel, p.BaseURL, p.ServerIP, authInfo, status,
	))

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	rows = append(rows, inline.Row(
		inline.Data("📊 آمار کامل", fmt.Sprintf("adm_xui_stats_%d", p.ID)),
	))

	rows = append(rows, inline.Row(
		inline.Data("🔌 تست اتصال", fmt.Sprintf("adm_xui_test_%d", p.ID)),
		inline.Data("📡 Inboundها", fmt.Sprintf("adm_xui_inbounds_%d", p.ID)),
	))

	if p.Active {
		rows = append(rows, inline.Row(
			inline.Data("❌ غیرفعال کردن", fmt.Sprintf("adm_xui_toggle_%d", p.ID)),
		))
	} else {
		rows = append(rows, inline.Row(
			inline.Data("✅ فعال کردن", fmt.Sprintf("adm_xui_toggle_%d", p.ID)),
		))
	}

	rows = append(rows, inline.Row(
		inline.Data("🗑 حذف پنل", fmt.Sprintf("adm_xui_del_%d", p.ID)),
	))
	rows = append(rows, inline.Row(
		inline.Data("⬅️ برگشت", "adm_xui_panels"),
	))

	inline.Inline(rows...)
	return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func XUIPanelAddStartHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setXUIState(c.Sender().ID, &xuiState{Step: "name"})
		_ = c.Respond()
		return c.Send("🏷 نام پنل را بفرست (مثلاً: پنل اصلی):")
	}
}

func XUIPanelToggleHandler(cfg *config.Config, db *database.DB, panelID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		if err := db.ToggleXUIPanelActive(panelID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "تغییر وضعیت پنل 3x-ui", fmt.Sprintf("پنل <code>%d</code>", panelID))

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ تغییر کرد"})
		p := db.GetXUIPanel(panelID)
		if p == nil {
			return showXUIPanels(c, db)
		}
		return showXUIPanelDetail(c, db, p)
	}
}

func XUIPanelDeleteHandler(cfg *config.Config, db *database.DB, panelID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		if err := db.DeleteXUIPanel(panelID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		admin := c.Sender()
		logAction(botInstance, cfg, db, admin, "حذف پنل 3x-ui", fmt.Sprintf("پنل <code>%d</code>", panelID))

		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ حذف شد"})
		return showXUIPanels(c, db)
	}
}

func XUIPanelTestHandler(cfg *config.Config, db *database.DB, panelID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		p := db.GetXUIPanel(panelID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ پنل پیدا نشد"})
		}

		cli := getXUIClient(p)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		type result struct {
			count int
			err   error
		}
		done := make(chan result, 1)
		go func() {
			inbounds, err := cli.ListInbounds()
			done <- result{count: len(inbounds), err: err}
		}()

		select {
		case r := <-done:
			if r.err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + r.err.Error()})
			}
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ اتصال موفق"})
			return c.Send(fmt.Sprintf(
				"✅ اتصال به پنل <b>%s</b> موفق بود.\n\n📡 تعداد inbound: %d",
				p.Name, r.count,
			), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		case <-ctx.Done():
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ timeout"})
		}
	}
}

func XUIPanelInboundsHandler(cfg *config.Config, db *database.DB, panelID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		p := db.GetXUIPanel(panelID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ پنل پیدا نشد"})
		}

		cli := getXUIClient(p)

		inbounds, err := cli.ListInbounds()
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		if len(inbounds) == 0 {
			return c.Send("📡 هیچ inboundی وجود ندارد.")
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📡 Inboundهای پنل <b>%s</b>\n\n", p.Name))

		for _, inb := range inbounds {
			status := "✅"
			if !inb.Enable {
				status = "❌"
			}
			sb.WriteString(fmt.Sprintf(
				"%s #%d — %s\n"+
					"   🔌 %s | 🚪 %d\n\n",
				status, inb.ID, inb.Remark, inb.Protocol, inb.Port,
			))
		}

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(inline.Row(
			inline.Data("⬅️ برگشت", fmt.Sprintf("adm_xui_view_%d", p.ID)),
		))

		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func XUIPanelStatsHandler(cfg *config.Config, db *database.DB, panelID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		_ = c.Respond()

		panel := db.GetXUIPanel(panelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := getXUIClient(panel)
		stats, err := cli.GetInboundStats()
		if err != nil {
			return c.Send("❌ " + err.Error())
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📊 <b>آمار پنل %s</b>\n\n", panel.Name))

		totalClients := 0
		totalTrafficGB := 0.0
		for _, inb := range stats {
			totalClients += inb.ClientCount
			totalTrafficGB += float64(inb.Up+inb.Down) / (1024 * 1024 * 1024)
		}

		sb.WriteString(fmt.Sprintf(
			"📡 تعداد inbound: <b>%d</b>\n"+
				"👥 تعداد کلاینت: <b>%d</b>\n"+
				"📥 مصرف کل: <b>%.2f GB</b>\n\n",
			len(stats), totalClients, totalTrafficGB,
		))

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		for _, inb := range stats {
			status := "✅"
			if !inb.Enable {
				status = "❌"
			}
			used := float64(inb.Up+inb.Down) / (1024 * 1024 * 1024)
			label := fmt.Sprintf("%s #%d %s (%d کلاینت، %.2f GB)",
				status, inb.ID, inb.Remark, inb.ClientCount, used)
			rows = append(rows, inline.Row(
				inline.Data(label, fmt.Sprintf("adm_xui_inbound_%d_%d", panelID, inb.ID)),
			))
		}

		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", fmt.Sprintf("adm_xui_view_%d", panelID)),
		))

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func XUIInboundClientsHandler(cfg *config.Config, db *database.DB, panelID, inboundID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		_ = c.Respond()

		panel := db.GetXUIPanel(panelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := getXUIClient(panel)

		inbounds, err := cli.ListInbounds()
		if err != nil {
			return c.Send("❌ " + err.Error())
		}

		var targetInbound *vpn.Inbound
		for i := range inbounds {
			if inbounds[i].ID == inboundID {
				targetInbound = &inbounds[i]
				break
			}
		}
		if targetInbound == nil {
			return c.Send("❌ inbound پیدا نشد")
		}

		stats, _ := cli.GetClientStats(inboundID)

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📡 <b>Inbound #%d — %s</b>\n\n", targetInbound.ID, targetInbound.Remark))
		sb.WriteString(fmt.Sprintf(
			"🔌 پروتکل: %s\n"+
				"🚪 پورت: %d\n"+
				"👥 تعداد کلاینت: <b>%d</b>\n\n",
			targetInbound.Protocol, targetInbound.Port, len(stats),
		))

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		if len(stats) == 0 {
			sb.WriteString("هیچ کلاینتی ثبت نشده.")
		} else {
			for _, st := range stats {
				status := "✅"
				if !st.Enable {
					status = "❌"
				}
				usedGB := float64(st.Up+st.Down) / (1024 * 1024 * 1024)
				totalGB := float64(st.Total) / (1024 * 1024 * 1024)
				expiryStr := "—"
				if st.ExpiryTime > 0 {
					expiryStr = time.UnixMilli(st.ExpiryTime).Format("06/01/02")
				}
				label := fmt.Sprintf("%s %s | %.1f/%.1fGB | %s",
					status, st.Email, usedGB, totalGB, expiryStr)
				rows = append(rows, inline.Row(
					inline.Data(label, fmt.Sprintf("adm_xui_client_%d_%d_%s", panelID, inboundID, st.Email)),
				))
			}
		}

		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", fmt.Sprintf("adm_xui_stats_%d", panelID)),
		))

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func XUIClientDetailHandler(cfg *config.Config, db *database.DB, panelID, inboundID int64, email string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		_ = c.Respond()

		panel := db.GetXUIPanel(panelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := getXUIClient(panel)

		client, err := cli.GetClient(inboundID, email)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}

		stats, _ := cli.GetClientTraffic(inboundID, email)

		usedGB := 0.0
		totalGB := float64(client.TotalGB) / (1024 * 1024 * 1024)
		upGB := 0.0
		downGB := 0.0
		if stats != nil {
			usedGB = float64(stats.Up+stats.Down) / (1024 * 1024 * 1024)
			upGB = float64(stats.Up) / (1024 * 1024 * 1024)
			downGB = float64(stats.Down) / (1024 * 1024 * 1024)
		}
		remaining := totalGB - usedGB
		if remaining < 0 {
			remaining = 0
		}

		expiry := "نامحدود"
		if client.ExpiryTime > 0 {
			expiry = time.UnixMilli(client.ExpiryTime).Format("2006/01/02 15:04")
		}

		statusStr := "❌ غیرفعال"
		if client.Enable {
			statusStr = "✅ فعال"
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(
			"👤 <b>%s</b>\n\n"+
				"🆔 <code>%s</code>\n"+
				"📊 وضعیت: %s\n"+
				"⏰ انقضا: %s\n\n"+
				"📥 آپلود: <b>%.2f GB</b>\n"+
				"📤 دانلود: <b>%.2f GB</b>\n"+
				"📦 کل: <b>%.2f GB</b>\n"+
				"📊 مصرف: <b>%.2f GB</b>\n"+
				"📤 باقی‌مانده: <b>%.2f GB</b>",
			client.Email, client.ID, statusStr, expiry,
			upGB, downGB, totalGB, usedGB, remaining,
		))

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row

		if client.Enable {
			rows = append(rows, inline.Row(
				inline.Data("❌ غیرفعال", fmt.Sprintf("adm_xui_cli_disable_%d_%d_%s", panelID, inboundID, email)),
			))
		} else {
			rows = append(rows, inline.Row(
				inline.Data("✅ فعال", fmt.Sprintf("adm_xui_cli_enable_%d_%d_%s", panelID, inboundID, email)),
			))
		}

		rows = append(rows, inline.Row(
			inline.Data("🔄 ریست ترافیک", fmt.Sprintf("adm_xui_cli_reset_%d_%d_%s", panelID, inboundID, email)),
		))

		rows = append(rows, inline.Row(
			inline.Data("📅 تمدید", fmt.Sprintf("adm_xui_cli_renew_%d_%d_%s", panelID, inboundID, email)),
		))

		rows = append(rows, inline.Row(
			inline.Data("🗑 حذف", fmt.Sprintf("adm_xui_cli_del_%d_%d_%s", panelID, inboundID, email)),
		))

		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", fmt.Sprintf("adm_xui_inbound_%d_%d", panelID, inboundID)),
		))

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func XUIClientActionHandler(cfg *config.Config, db *database.DB, action string, panelID, inboundID int64, email string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}

		panel := db.GetXUIPanel(panelID)
		if panel == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ پنل پیدا نشد"})
		}

		cli := getXUIClient(panel)

		switch action {
		case "enable":
			if err := cli.EnableClient(inboundID, email, true); err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ فعال شد"})
			logAction(botInstance, cfg, db, c.Sender(), "فعال کردن کلاینت 3x-ui",
				fmt.Sprintf("Email: <code>%s</code> — پنل: <code>%d</code>", email, panelID))

		case "disable":
			if err := cli.EnableClient(inboundID, email, false); err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ غیرفعال شد"})
			logAction(botInstance, cfg, db, c.Sender(), "غیرفعال کردن کلاینت 3x-ui",
				fmt.Sprintf("Email: <code>%s</code> — پنل: <code>%d</code>", email, panelID))

		case "reset":
			if err := cli.ResetClientTraffic(inboundID, email); err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ ترافیک ریست شد"})
			logAction(botInstance, cfg, db, c.Sender(), "ریست ترافیک کلاینت",
				fmt.Sprintf("Email: <code>%s</code> — پنل: <code>%d</code>", email, panelID))

		case "del":
			client, err := cli.GetClient(inboundID, email)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			if err := cli.DeleteClient(inboundID, client.ID); err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ حذف شد"})
			logAction(botInstance, cfg, db, c.Sender(), "حذف کلاینت 3x-ui",
				fmt.Sprintf("Email: <code>%s</code> — پنل: <code>%d</code>", email, panelID))
			return XUIInboundClientsHandler(cfg, db, panelID, inboundID)(c)

		case "renew":
			setXUIManageState(c.Sender().ID, &xuiManageState{
				Step:      "renew_days",
				PanelID:   panelID,
				InboundID: inboundID,
				Email:     email,
			})
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "تعداد روز را بفرست"})
			return c.Send("📅 چند روز تمدید شود؟ (مثال: 30):")
		}

		_ = c.Delete()
		return XUIClientDetailHandler(cfg, db, panelID, inboundID, email)(c)
	}
}

func XUIManageInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getXUIManageState(uid)
	if st == nil {
		return nil
	}
	if db.GetAdminRole(uid, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		setXUIManageState(uid, nil)
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	switch st.Step {
	case "renew_days":
		days, err := strconv.Atoi(text)
		if err != nil || days <= 0 {
			return c.Send("❌ عدد نامعتبر. یک عدد مثبت بفرست:")
		}

		panel := db.GetXUIPanel(st.PanelID)
		if panel == nil {
			setXUIManageState(uid, nil)
			return c.Send("❌ پنل پیدا نشد")
		}

		cli := getXUIClient(panel)

		client, err := cli.GetClient(st.InboundID, st.Email)
		if err != nil {
			setXUIManageState(uid, nil)
			return c.Send("❌ " + err.Error())
		}

		nowMs := time.Now().UnixMilli()
		base := client.ExpiryTime
		if base < nowMs {
			base = nowMs
		}
		newExpiry := base + int64(days)*24*60*60*1000
		client.ExpiryTime = newExpiry
		client.Enable = true

		if err := cli.UpdateClient(st.InboundID, *client); err != nil {
			setXUIManageState(uid, nil)
			return c.Send("❌ " + err.Error())
		}

		logAction(botInstance, cfg, db, c.Sender(), "تمدید کلاینت 3x-ui",
			fmt.Sprintf("Email: <code>%s</code> — %d روز — پنل: <code>%d</code>", st.Email, days, st.PanelID))

		setXUIManageState(uid, nil)
		return c.Send(fmt.Sprintf("✅ تمدید شد تا %s",
			time.UnixMilli(newExpiry).Format("2006/01/02 15:04")))
	}
	return nil
}

func XUIInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getXUIState(uid)
	if st == nil {
		return nil
	}
	if db.GetAdminRole(uid, cfg.ADMIN_ID) != database.RoleSuperAdmin {
		setXUIState(uid, nil)
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	switch st.Step {
	case "name":
		st.Name = text
		st.Step = "base_url"
		setXUIState(uid, st)
		return c.Send("🌐 آدرس پنل را بفرست (مثل: https://panel.example.com:2053):")

	case "base_url":
		st.BaseURL = text
		st.Step = "server_ip"
		setXUIState(uid, st)
		return c.Send("🆔 IP سرور (برای ساخت لینک) را بفرست:")

	case "server_ip":
		st.ServerIP = text
		st.Step = "auth_method"
		setXUIState(uid, st)
		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("🔑 API Token", "adm_xui_auth_token"),
				inline.Data("👤 یوزر/پسورد", "adm_xui_auth_login"),
			),
		)
		return c.Send("🔐 روش احراز هویت را انتخاب کن:", inline)

	case "token":
		st.Token = text
		st.AuthMeth = "token"
		return finalizeXUIPanel(c, db, st)

	case "username":
		st.Username = text
		st.Step = "password"
		setXUIState(uid, st)
		return c.Send("🔑 پسورد پنل را بفرست:")

	case "password":
		st.Password = text
		st.AuthMeth = "login"
		return finalizeXUIPanel(c, db, st)
	}
	return nil
}

func XUIAuthSelectHandler(cfg *config.Config, db *database.DB, method string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getXUIState(uid)
		if st == nil || st.Step != "auth_method" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		_ = c.Respond()
		switch method {
		case "token":
			st.Step = "token"
			setXUIState(uid, st)
			return c.Send("🔑 API Token را بفرست:")
		case "login":
			st.Step = "username"
			setXUIState(uid, st)
			return c.Send("👤 یوزرنیم پنل را بفرست:")
		}
		return c.Respond(&AsliBot.CallbackResponse{Text: "❌ روش نامعتبر"})
	}
}

func finalizeXUIPanel(c AsliBot.Context, db *database.DB, st *xuiState) error {
	p := &database.XUIPanel{
		PanelType:  "3x-ui",
		Name:       st.Name,
		BaseURL:    st.BaseURL,
		ServerIP:   st.ServerIP,
		AuthMethod: st.AuthMeth,
		APIToken:   st.Token,
		Username:   st.Username,
		Password:   st.Password,
		Active:     true,
	}
	if _, err := db.AddXUIPanel(p); err != nil {
		return c.Send("❌ " + err.Error())
	}
	setXUIState(c.Sender().ID, nil)
	return c.Send(fmt.Sprintf(
		"✅ پنل <b>%s</b> ساخته شد.\n\n🆔 شناسه: %d",
		p.Name, p.ID,
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
}
