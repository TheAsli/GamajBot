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

type catFieldState struct {
	Step        string
	CatID       int64
	FieldKey    string
	FieldLabel  string
	FieldType   string
	FieldPlace  string
	FieldReq    bool
	ActionKey   string
	ActionLabel string
	ActionEmoji string
	ActionReq   bool
	ActionInLbl string
	IsAction    bool
}

var (
	catFieldStates = make(map[int64]*catFieldState)
	catFieldMu     sync.Mutex
)

func setCatFieldState(uid int64, s *catFieldState) {
	catFieldMu.Lock()
	defer catFieldMu.Unlock()
	if s == nil {
		delete(catFieldStates, uid)
	} else {
		catFieldStates[uid] = s
	}
}

func getCatFieldState(uid int64) *catFieldState {
	catFieldMu.Lock()
	defer catFieldMu.Unlock()
	return catFieldStates[uid]
}

func CategoryFieldsHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}

		cat := db.GetCategory(catID)
		if cat == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دسته پیدا نشد"})
		}

		fields, _ := db.GetCategoryDeliveryFields(catID)
		actions, _ := db.GetCategoryServiceActions(catID)

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📁 <b>%s %s</b>\n\n", cat.Emoji, cat.Name))

		sb.WriteString(fmt.Sprintf("📝 <b>فیلدهای تحویل (%d):</b>\n", len(fields)))
		if len(fields) == 0 {
			sb.WriteString("   —\n")
		} else {
			for _, f := range fields {
				req := ""
				if f.Required {
					req = " *"
				}
				sb.WriteString(fmt.Sprintf("   • <code>%s</code> — %s (%s)%s\n", f.Key, f.Label, f.Type, req))
			}
		}

		sb.WriteString(fmt.Sprintf("\n⚙️ <b>اکشن‌های سرویس (%d):</b>\n", len(actions)))
		if len(actions) == 0 {
			sb.WriteString("   —\n")
		} else {
			for _, a := range actions {
				req := ""
				if a.RequireInput {
					req = " ✏️"
				}
				sb.WriteString(fmt.Sprintf("   • <b>%s %s</b> (<code>%s</code>)%s\n", a.Emoji, a.Label, a.Key, req))
			}
		}

		inline := &AsliBot.ReplyMarkup{}
		rows := []AsliBot.Row{
			inline.Row(
				inline.Data("➕ فیلد تحویل", fmt.Sprintf("adm_cf_add_%d", catID)),
			),
			inline.Row(
				inline.Data("➕ اکشن سرویس", fmt.Sprintf("adm_cf_addact_%d", catID)),
			),
		}

		for _, f := range fields {
			rows = append(rows, inline.Row(
				inline.Data(fmt.Sprintf("🗑 حذف فیلد %s", f.Label), fmt.Sprintf("adm_cf_delf_%d_%s", catID, f.Key)),
			))
		}
		for _, a := range actions {
			rows = append(rows, inline.Row(
				inline.Data(fmt.Sprintf("🗑 حذف اکشن %s", a.Label), fmt.Sprintf("adm_cf_dela_%d_%s", catID, a.Key)),
			))
		}

		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", fmt.Sprintf("adm_cat_view_%d", catID)),
		))

		inline.Inline(rows...)
		return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
	}
}

func CategoryFieldAddHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setCatFieldState(c.Sender().ID, &catFieldState{Step: "field_key", CatID: catID})
		_ = c.Respond()
		return c.Send("🔑 کلید فیلد را بفرست (انگلیسی، بدون فاصله — مثلاً: ip):")
	}
}

func CategoryActionAddHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		setCatFieldState(c.Sender().ID, &catFieldState{Step: "act_key", CatID: catID, IsAction: true})
		_ = c.Respond()
		return c.Send("🔑 کلید اکشن را بفرست (انگلیسی، بدون فاصله — مثلاً: rebuild):")
	}
}

func CategoryFieldInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getCatFieldState(uid)
	if st == nil {
		return nil
	}
	role := db.GetAdminRole(uid, cfg.ADMIN_ID)
	if role != database.RoleSuperAdmin && role != database.RoleFinance {
		setCatFieldState(uid, nil)
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	if !st.IsAction {
		switch st.Step {
		case "field_key":
			st.FieldKey = strings.ToLower(text)
			st.Step = "field_label"
			setCatFieldState(uid, st)
			return c.Send("📝 برچسب فیلد را بفرست (مثلاً: آدرس IP):")

		case "field_label":
			st.FieldLabel = text
			st.Step = "field_type"
			setCatFieldState(uid, st)
			inline := &AsliBot.ReplyMarkup{}
			inline.Inline(
				inline.Row(
					inline.Data("📝 متن", "adm_cf_ft_text"),
					inline.Data("🔒 پسورد", "adm_cf_ft_password"),
				),
				inline.Row(
					inline.Data("📄 متن بلند", "adm_cf_ft_textarea"),
					inline.Data("🔢 عدد", "adm_cf_ft_number"),
				),
			)
			return c.Send("نوع فیلد را انتخاب کن:", inline)

		case "field_placeholder":
			if text != "-" {
				st.FieldPlace = text
			}
			st.Step = "field_required"
			setCatFieldState(uid, st)
			inline := &AsliBot.ReplyMarkup{}
			inline.Inline(
				inline.Row(
					inline.Data("✅ اجباری", "adm_cf_fr_yes"),
					inline.Data("⚪ اختیاری", "adm_cf_fr_no"),
				),
			)
			return c.Send("این فیلد اجباری است؟", inline)
		}
	} else {
		switch st.Step {
		case "act_key":
			st.ActionKey = strings.ToLower(text)
			st.Step = "act_label"
			setCatFieldState(uid, st)
			return c.Send("📝 برچسب اکشن را بفرست (مثلاً: ریبیلد):")

		case "act_label":
			st.ActionLabel = text
			st.Step = "act_emoji"
			setCatFieldState(uid, st)
			return c.Send("ایموجی اکشن را بفرست (مثلاً: 🔄 — یا «-» برای خالی):")

		case "act_emoji":
			if text != "-" {
				st.ActionEmoji = text
			}
			st.Step = "act_req_input"
			setCatFieldState(uid, st)
			inline := &AsliBot.ReplyMarkup{}
			inline.Inline(
				inline.Row(
					inline.Data("📝 نیاز به ورودی", "adm_cf_ar_yes"),
					inline.Data("⚡ بدون ورودی", "adm_cf_ar_no"),
				),
			)
			return c.Send("این اکشن نیاز به ورودی از کاربر دارد؟", inline)
		}
	}
	return nil
}

func CategoryFieldTypeSelectHandler(cfg *config.Config, db *database.DB, fieldType string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getCatFieldState(uid)
		if st == nil || st.Step != "field_type" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		st.FieldType = fieldType
		st.Step = "field_placeholder"
		setCatFieldState(uid, st)
		_ = c.Respond()
		return c.Send("متن راهنما (placeholder) را بفرست (یا «-» برای خالی):")
	}
}

func CategoryFieldRequiredHandler(cfg *config.Config, db *database.DB, required bool) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getCatFieldState(uid)
		if st == nil || st.Step != "field_required" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		st.FieldReq = required

		field := database.DeliveryField{
			Key:         st.FieldKey,
			Label:       st.FieldLabel,
			Type:        st.FieldType,
			Placeholder: st.FieldPlace,
			Required:    st.FieldReq,
		}
		if err := db.AddCategoryDeliveryField(st.CatID, field); err != nil {
			setCatFieldState(uid, nil)
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		setCatFieldState(uid, nil)
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ اضافه شد"})
		return CategoryFieldsHandler(cfg, db, st.CatID)(c)
	}
}

func CategoryActionRequireInputHandler(cfg *config.Config, db *database.DB, require bool) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getCatFieldState(uid)
		if st == nil || st.Step != "act_req_input" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		st.ActionReq = require
		if require {
			st.Step = "act_input_label"
			setCatFieldState(uid, st)
			_ = c.Respond()
			return c.Send("📝 برچسب ورودی را بفرست (مثلاً: پسورد جدید):")
		}

		action := database.ServiceAction{
			Key:          st.ActionKey,
			Label:        st.ActionLabel,
			Emoji:        st.ActionEmoji,
			RequireInput: false,
		}
		if err := db.AddCategoryServiceAction(st.CatID, action); err != nil {
			setCatFieldState(uid, nil)
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		setCatFieldState(uid, nil)
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ اضافه شد"})
		return CategoryFieldsHandler(cfg, db, st.CatID)(c)
	}
}

func CategoryActionInputLabelHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getCatFieldState(uid)
	if st == nil || st.Step != "act_input_label" {
		return nil
	}
	text := strings.TrimSpace(c.Text())
	if text == "" || strings.HasPrefix(text, "/") {
		return nil
	}
	st.ActionInLbl = text

	action := database.ServiceAction{
		Key:          st.ActionKey,
		Label:        st.ActionLabel,
		Emoji:        st.ActionEmoji,
		RequireInput: true,
		InputLabel:   st.ActionInLbl,
	}
	if err := db.AddCategoryServiceAction(st.CatID, action); err != nil {
		setCatFieldState(uid, nil)
		return c.Send("❌ " + err.Error())
	}
	setCatFieldState(uid, nil)
	return CategoryFieldsHandler(cfg, db, st.CatID)(c)
}

func CategoryFieldDeleteHandler(cfg *config.Config, db *database.DB, catID int64, key string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		if err := db.RemoveCategoryDeliveryField(catID, key); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ حذف شد"})
		return CategoryFieldsHandler(cfg, db, catID)(c)
	}
}

func CategoryActionDeleteHandler(cfg *config.Config, db *database.DB, catID int64, key string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔"})
		}
		if err := db.RemoveCategoryServiceAction(catID, key); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ حذف شد"})
		return CategoryFieldsHandler(cfg, db, catID)(c)
	}
}

func unusedStrconv(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
