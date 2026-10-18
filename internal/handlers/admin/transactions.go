// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package admin

import (
	"fmt"
	"strings"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"

	AsliBot "gopkg.in/telebot.v4"
)

const txPerPage = 5

func TxHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showTxPage(c, db, 0)
	}
}

func TxPageHandler(cfg *config.Config, db *database.DB, page int) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showTxPage(c, db, page)
	}
}

func showTxPage(c AsliBot.Context, db *database.DB, page int) error {
	txs := db.GetAllTransactions()
	if len(txs) == 0 {
		return c.Send("💰 هیچ تراکنشی ثبت نشده.")
	}

	total := len(txs)
	start := page * txPerPage
	if start >= total {
		page, start = 0, 0
	}
	end := start + txPerPage
	if end > total {
		end = total
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("💰 تراکنش‌ها (صفحه %d از %d)\n\n", page+1, (total+txPerPage-1)/txPerPage))
	for i := start; i < end; i++ {
		sb.WriteString(formatTxInfo(txs[i]))
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("مجموع: %d تراکنش", total))

	inline := &AsliBot.ReplyMarkup{}
	var row []AsliBot.Btn
	if page > 0 {
		row = append(row, inline.Data("⬅️ قبلی", fmt.Sprintf("adm_tx_%d", page-1)))
	}
	if end < total {
		row = append(row, inline.Data("➡️ بعدی", fmt.Sprintf("adm_tx_%d", page+1)))
	}
	if len(row) > 0 {
		inline.Inline(inline.Row(row...))
		return c.Send(sb.String(), inline)
	}
	return c.Send(sb.String())
}

func formatTxInfo(t database.TxInfo) string {
	createdAt := t.CreatedAt
	if len(createdAt) > 16 {
		createdAt = createdAt[:16]
	}

	var icon, status string
	switch t.Status {
	case "paid":
		icon, status = "✅", "موفق"
	case "pending":
		icon, status = "⏳", "در انتظار"
	case "canceled":
		icon, status = "❌", "لغو"
	case "expired":
		icon, status = "⌛", "منقضی"
	default:
		icon, status = "❓", t.Status
	}

	var typeLabel string
	switch t.Type {
	case "charge":
		typeLabel = "شارژ"
	case "manual_charge":
		typeLabel = "شارژ دستی"
	case "spend":
		typeLabel = "خرج"
	case "deduct":
		typeLabel = "کسر"
	default:
		typeLabel = t.Type
	}

	var gwLabel string
	switch t.Gateway {
	case "tetraminator":
		gwLabel = " | 💎 Tetraminator"
	case "cubepay":
		gwLabel = " | 🧊 CubePay"
	case "manual_card":
		gwLabel = " | 💳 کارت دستی"
	case "manual_charge":
		gwLabel = " | 🎁 شارژ دستی"
	}

	uname := t.Username
	if uname == "" {
		uname = "—"
	}

	return fmt.Sprintf(
		"%s %s | %s | %.2f دلار%s\n   👤 %s (@%s)\n   🕐 %s\n",
		icon, status, typeLabel, t.AmountUSD, gwLabel, t.FirstName, uname, createdAt,
	)
}
