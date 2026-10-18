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
	"github.com/theasli/GamajBot/internal/rate"

	AsliBot "gopkg.in/telebot.v4"
)

const usersPerPage = 8

var (
	userSearchAwaiting = make(map[int64]bool)
	userSearchMu       sync.Mutex
)

func setUserSearchAwaiting(uid int64, on bool) {
	userSearchMu.Lock()
	defer userSearchMu.Unlock()
	if on {
		userSearchAwaiting[uid] = true
	} else {
		delete(userSearchAwaiting, uid)
	}
}

func isUserSearchAwaiting(uid int64) bool {
	userSearchMu.Lock()
	defer userSearchMu.Unlock()
	return userSearchAwaiting[uid]
}

var mainAdminIDValue int64

func SetMainAdminID(id int64) {
	mainAdminIDValue = id
}

func appMainAdminID() int64 {
	return mainAdminIDValue
}

func UsersHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showUsersPage(c, db, 0)
	}
}

func UsersPageHandler(cfg *config.Config, db *database.DB, page int) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showUsersPage(c, db, page)
	}
}

func showUsersPage(c AsliBot.Context, db *database.DB, page int) error {
	users := db.GetAllUsers()
	if len(users) == 0 {
		return c.Send("👥 هیچ کاربری ثبت نشده.")
	}

	total := len(users)
	start := page * usersPerPage
	if start >= total {
		page, start = 0, 0
	}
	end := start + usersPerPage
	if end > total {
		end = total
	}

	cur := db.GetCurrency()

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	var pageRow []AsliBot.Btn
	if page > 0 {
		pageRow = append(pageRow, inline.Data("⬅️", fmt.Sprintf("adm_users_%d", page-1)))
	}
	pageRow = append(pageRow, inline.Data("🔍 جستجو", "adm_users_search"))
	if end < total {
		pageRow = append(pageRow, inline.Data("➡️", fmt.Sprintf("adm_users_%d", page+1)))
	}
	rows = append(rows, inline.Row(pageRow...))

	for _, u := range users[start:end] {
		banMark := ""
		if u.Banned {
			banMark = "🚫 "
		}
		bal := rate.Display(u.BalanceUSD, cur)
		btnLabel := fmt.Sprintf("%s👤 %s | %s", banMark, u.FirstName, bal)
		if u.Username != "" {
			btnLabel = fmt.Sprintf("%s👤 %s (@%s) | %s", banMark, u.FirstName, u.Username, bal)
		}
		rows = append(rows, inline.Row(
			inline.Data(btnLabel, fmt.Sprintf("adm_user_view_%d", u.ID)),
		))
	}

	inline.Inline(rows...)
	return c.Send(fmt.Sprintf(
		"👥 کاربران (صفحه %d از %d)\nمجموع: %d کاربر\n\nیک کاربر را انتخاب کن:",
		page+1, (total+usersPerPage-1)/usersPerPage, total,
	), inline)
}

func UsersSearchHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		setUserSearchAwaiting(c.Sender().ID, true)
		return c.Respond(&AsliBot.CallbackResponse{Text: "🔍 آیدی یا یوزرنیم را بفرست"})
	}
}

func UserSearchInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	if !isUserSearchAwaiting(uid) {
		return nil
	}
	if db.GetAdminRole(uid, cfg.ADMIN_ID) == "" {
		setUserSearchAwaiting(uid, false)
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}
	setUserSearchAwaiting(uid, false)

	text = strings.TrimPrefix(text, "@")

	var foundID int64
	if id, err := strconv.ParseInt(text, 10, 64); err == nil {
		if _, err := db.GetUser(id); err == nil {
			foundID = id
		}
	}
	if foundID == 0 {
		for _, u := range db.GetAllUsers() {
			if strings.EqualFold(u.Username, text) {
				foundID = u.ID
				break
			}
		}
	}
	if foundID == 0 {
		return c.Send("❌ کاربری پیدا نشد.")
	}

	return showUserDetail(c, db, foundID)
}

func UserViewHandler(cfg *config.Config, db *database.DB, userID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		_ = c.Respond()
		return showUserDetail(c, db, userID)
	}
}

func showUserDetail(c AsliBot.Context, db *database.DB, userID int64) error {
	u, err := db.GetUser(userID)
	if err != nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: "❌ کاربر پیدا نشد"})
	}

	uname := u.Username
	if uname == "" {
		uname = "—"
	}

	cur := db.GetCurrency()
	bal := rate.Display(u.BalanceUSD, cur)
	txCount := db.GetUserTxCount(userID)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"👤 <b>%s</b> (@%s)\n\n"+
			"🆔 <code>%d</code>\n"+
			"💵 موجودی: <b>%s</b>\n"+
			"📊 تراکنش‌ها: %d\n",
		u.FirstName, uname, u.ID, bal, txCount,
	))
	if u.Banned {
		sb.WriteString(fmt.Sprintf("🚫 <b>بلاک شده</b> | دلیل: %s\n", u.BanReason))
	} else {
		sb.WriteString("✅ فعال\n")
	}

	role := db.GetAdminRole(c.Sender().ID, appMainAdminID())
	isFinance := role == database.RoleSuperAdmin || role == database.RoleFinance
	isSupportOrFinance := role == database.RoleSuperAdmin || role == database.RoleFinance || role == database.RoleSupport

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	if isSupportOrFinance {
		rows = append(rows, inline.Row(
			inline.Data("💳 شارژ دستی", fmt.Sprintf("adm_user_charge_%d", u.ID)),
		))
	}
	if isFinance {
		rows = append(rows, inline.Row(
			inline.Data("➖ کسر موجودی", fmt.Sprintf("adm_user_deduct_%d", u.ID)),
		))
	}
	rows = append(rows, inline.Row(
		inline.Data("📜 تراکنش‌ها", fmt.Sprintf("adm_user_tx_%d_0", u.ID)),
	))

	if isSupportOrFinance {
		if u.Banned {
			rows = append(rows, inline.Row(
				inline.Data("✅ آنبلاک", fmt.Sprintf("adm_user_unban_%d", u.ID)),
			))
		} else {
			rows = append(rows, inline.Row(
				inline.Data("🚫 بلاک", fmt.Sprintf("adm_user_ban_%d", u.ID)),
			))
		}
	}

	rows = append(rows, inline.Row(
		inline.Data("⬅️ برگشت به لیست", "adm_users"),
	))

	inline.Inline(rows...)
	return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func UserTxPageHandler(cfg *config.Config, db *database.DB, userID int64, page int) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		_ = c.Respond()

		allTxs, err := db.GetHistory(userID, 0)
		if err != nil || len(allTxs) == 0 {
			return c.Send("📜 تراکنشی ندارد.")
		}

		total := len(allTxs)
		start := page * 5
		if start >= total {
			page, start = 0, 0
		}
		end := start + 5
		if end > total {
			end = total
		}

		cur := db.GetCurrency()
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📜 تراکنش‌های کاربر <code>%d</code> (صفحه %d از %d)\n\n",
			userID, page+1, (total+4)/5))

		for _, t := range allTxs[start:end] {
			createdAt := t.CreatedAt
			if len(createdAt) > 16 {
				createdAt = createdAt[:16]
			}
			icon, title := txIconTitle(t)
			sign := "+"
			if t.AmountUSD < 0 {
				sign = ""
			}
			gw := ""
			if t.Gateway != "" {
				gw = " | " + gatewayLabelShort(t.Gateway)
			}
			sb.WriteString(fmt.Sprintf(
				"%s %s | %s%s%s\n   🕐 %s\n\n",
				icon, title, sign, rate.Display(t.AmountUSD, cur), gw, createdAt,
			))
		}
		sb.WriteString(fmt.Sprintf("مجموع: %d تراکنش", total))

		inline := &AsliBot.ReplyMarkup{}
		var row []AsliBot.Btn
		if page > 0 {
			row = append(row, inline.Data("⬅️", fmt.Sprintf("adm_user_tx_%d_%d", userID, page-1)))
		}
		if end < total {
			row = append(row, inline.Data("➡️", fmt.Sprintf("adm_user_tx_%d_%d", userID, page+1)))
		}

		rows := []AsliBot.Row{}
		if len(row) > 0 {
			rows = append(rows, inline.Row(row...))
		}
		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت به کاربر", fmt.Sprintf("adm_user_view_%d", userID)),
		))
		inline.Inline(rows...)

		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func txIconTitle(t database.Transaction) (string, string) {
	switch {
	case t.Type == "charge" && t.Status == "paid":
		return "✅", "شارژ موفق"
	case t.Type == "charge" && t.Status == "pending":
		return "⏳", "در انتظار پرداخت"
	case t.Type == "charge" && t.Status == "expired":
		return "⌛", "منقضی"
	case t.Type == "charge" && t.Status == "canceled":
		return "❌", "لغو شده"
	case t.Type == "manual_charge":
		return "🎁", "شارژ دستی"
	case t.Type == "spend":
		return "💸", "خرج"
	case t.Type == "deduct":
		return "➖", "کسر"
	case t.Type == "purchase":
		return "🛒", "خرید"
	case t.Type == "refund":
		return "🔄", "بازگشت وجه"
	case t.Type == "renewal":
		return "🔁", "تمدید"
	}
	return "❓", t.Status
}

func gatewayLabelShort(g string) string {
	switch g {
	case "tetraminator":
		return "💎"
	case "cubepay":
		return "🧊"
	case "manual_card":
		return "💳"
	case "manual_charge":
		return "🎁"
	}
	return ""
}

func UserBanHandler(cfg *config.Config, db *database.DB, userID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		u, err := db.GetUser(userID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ کاربر پیدا نشد"})
		}
		if u.Banned {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⚠️ از قبل بلاک است"})
		}
		if err := db.BanUser(userID, "بلاک توسط ادمین "+c.Sender().FirstName); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		admin := c.Sender()
		logActionTarget(botInstance, cfg, db, admin, "بلاک کاربر", u.FirstName, userID,
			"دلیل: بلاک توسط ادمین")

		_, _ = c.Bot().Send(AsliBot.ChatID(userID), "🚫 حساب شما توسط ادمین بلاک شد.")
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ بلاک شد"})
		_ = c.Delete()
		return showUserDetail(c, db, userID)
	}
}

func UserUnbanHandler(cfg *config.Config, db *database.DB, userID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		u, err := db.GetUser(userID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ کاربر پیدا نشد"})
		}
		if !u.Banned {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⚠️ بلاک نیست"})
		}
		if err := db.UnbanUser(userID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}

		admin := c.Sender()
		logActionTarget(botInstance, cfg, db, admin, "آنبلاک کاربر", u.FirstName, userID, "")

		_, _ = c.Bot().Send(AsliBot.ChatID(userID), "✅ حساب شما آنبلاک شد.")
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ آنبلاک شد"})
		_ = c.Delete()
		return showUserDetail(c, db, userID)
	}
}

func UserChargeHandler(cfg *config.Config, db *database.DB, userID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role == "" || role == database.RoleViewer {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		u, err := db.GetUser(userID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ کاربر پیدا نشد"})
		}
		setAdminChargeState(c.Sender().ID, &adminChargeState{
			Step:         "awaiting_amount",
			TargetUserID: userID,
			TargetName:   u.FirstName,
		})
		_ = c.Respond(&AsliBot.CallbackResponse{})
		return c.Send(fmt.Sprintf(
			"💳 شارژ دستی برای <b>%s</b> (<code>%d</code>)\n\n💵 مبلغ دلاری را بفرست:",
			u.FirstName, userID,
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}

func UserDeductHandler(cfg *config.Config, db *database.DB, userID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		u, err := db.GetUser(userID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ کاربر پیدا نشد"})
		}
		cur := db.GetCurrency()
		setDeductState(c.Sender().ID, &deductState{
			Step:         "awaiting_amount",
			TargetUserID: userID,
			TargetName:   u.FirstName,
		})
		_ = c.Respond(&AsliBot.CallbackResponse{})
		return c.Send(fmt.Sprintf(
			"➖ کسر از <b>%s</b> (<code>%d</code>)\n💰 موجودی: %s\n\n💵 مبلغ دلاری برای کسر:",
			u.FirstName, userID, rate.Display(u.BalanceUSD, cur),
		), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
}
