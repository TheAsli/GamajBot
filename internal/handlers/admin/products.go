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
	"github.com/theasli/GamajBot/internal/xui"

	AsliBot "gopkg.in/telebot.v4"
)

type productState struct {
	Step         string
	ParentID     int64
	CategoryID   int64
	ProductID    int64
	Name         string
	Description  string
	PriceUSD     float64
	Delivery     string
	ProductType  string
	PanelID      int64
	InboundIDs   []int64
	TrafficGB    int64
	DurationDays int
}

var (
	productStates = make(map[int64]*productState)
	productMu     sync.Mutex
)

func setProductState(uid int64, s *productState) {
	productMu.Lock()
	defer productMu.Unlock()
	if s == nil {
		delete(productStates, uid)
	} else {
		productStates[uid] = s
	}
}

func getProductState(uid int64) *productState {
	productMu.Lock()
	defer productMu.Unlock()
	return productStates[uid]
}

func ProductsAdminHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showAdminCategories(c, db, 0)
	}
}

func showAdminCategories(c AsliBot.Context, db *database.DB, parentID int64) error {
	cats := db.GetCategoriesByParent(parentID)

	var title string
	if parentID == 0 {
		title = "📦 سایر محصولات — دسته‌های اصلی"
	} else {
		parent := db.GetCategory(parentID)
		if parent == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دسته پیدا نشد"})
		}
		title = fmt.Sprintf("%s %s — زیردسته‌ها", parent.Emoji, parent.Name)
	}

	var sb strings.Builder
	sb.WriteString(title + "\n\n")
	if len(cats) == 0 {
		sb.WriteString("هیچ دسته‌ای ثبت نشده.")
	} else {
		sb.WriteString(fmt.Sprintf("تعداد: %d دسته", len(cats)))
	}

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	for _, cat := range cats {
		status := "✅"
		if !cat.Active {
			status = "❌"
		}
		subCount := len(db.GetCategoriesByParent(cat.ID))
		prodCount := len(db.GetProductsByCategory(cat.ID))
		label := fmt.Sprintf("%s %s %s (📁%d 📦%d)", status, cat.Emoji, cat.Name, subCount, prodCount)
		rows = append(rows, inline.Row(
			inline.Data(label, fmt.Sprintf("adm_cat_view_%d", cat.ID)),
		))
	}

	rows = append(rows, inline.Row(
		inline.Data("➕ افزودن دسته", fmt.Sprintf("adm_cat_add_%d", parentID)),
		inline.Data("📦 محصولات", fmt.Sprintf("adm_cat_products_%d", parentID)),
	))

	if parentID != 0 {
		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", "adm_products"),
		))
	}

	inline.Inline(rows...)
	return c.Send(sb.String(), inline)
}

func CategoryViewHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		cat := db.GetCategory(catID)
		if cat == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دسته پیدا نشد"})
		}
		return showAdminCategoryView(c, db, cat)
	}
}

func showAdminCategoryView(c AsliBot.Context, db *database.DB, cat *database.Category) error {
	subs := db.GetCategoriesByParent(cat.ID)
	products := db.GetProductsByCategory(cat.ID)
	fields, _ := db.GetCategoryDeliveryFields(cat.ID)
	actions, _ := db.GetCategoryServiceActions(cat.ID)

	status := "✅ فعال"
	if !cat.Active {
		status = "❌ غیرفعال"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"%s <b>%s</b>\n\nوضعیت: %s\n📁 زیردسته: %d\n📦 محصول: %d\n⚙️ فیلد تحویل: %d\n🔧 اکشن سرویس: %d\n",
		cat.Emoji, cat.Name, status, len(subs), len(products), len(fields), len(actions),
	))

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	rows = append(rows, inline.Row(
		inline.Data("📁 زیردسته‌ها", fmt.Sprintf("adm_cat_subs_%d", cat.ID)),
		inline.Data("📦 محصولات", fmt.Sprintf("adm_cat_products_%d", cat.ID)),
	))
	rows = append(rows, inline.Row(
		inline.Data("➕ افزودن زیردسته", fmt.Sprintf("adm_cat_add_%d", cat.ID)),
		inline.Data("➕ افزودن محصول", fmt.Sprintf("adm_prod_add_%d", cat.ID)),
	))
	rows = append(rows, inline.Row(
		inline.Data("⚙️ فیلدها و اکشن‌ها", fmt.Sprintf("adm_cf_view_%d", cat.ID)),
	))

	if cat.Active {
		rows = append(rows, inline.Row(
			inline.Data("❌ غیرفعال کردن", fmt.Sprintf("adm_cat_toggle_%d", cat.ID)),
		))
	} else {
		rows = append(rows, inline.Row(
			inline.Data("✅ فعال کردن", fmt.Sprintf("adm_cat_toggle_%d", cat.ID)),
		))
	}

	rows = append(rows, inline.Row(
		inline.Data("🗑 حذف دسته", fmt.Sprintf("adm_cat_del_%d", cat.ID)),
	))

	if cat.ParentID == 0 {
		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", "adm_products"),
		))
	} else {
		rows = append(rows, inline.Row(
			inline.Data("⬅️ برگشت", fmt.Sprintf("adm_cat_view_%d", cat.ParentID)),
		))
	}

	inline.Inline(rows...)
	return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func CategoryToggleHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		if err := db.ToggleCategory(catID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ تغییر کرد"})
		cat := db.GetCategory(catID)
		if cat == nil {
			return c.Send("❌ دسته پیدا نشد")
		}
		return showAdminCategoryView(c, db, cat)
	}
}

func CategoryDeleteHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		cat := db.GetCategory(catID)
		if cat == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دسته پیدا نشد"})
		}
		if len(db.GetCategoriesByParent(catID)) > 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ اول زیردسته‌ها را حذف کن"})
		}
		if len(db.GetProductsByCategory(catID)) > 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ اول محصولات را حذف کن"})
		}
		parent := cat.ParentID
		if err := db.DeleteCategory(catID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ حذف شد"})
		return showAdminCategories(c, db, parent)
	}
}

func CategoryAddStartHandler(cfg *config.Config, db *database.DB, parentID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		setProductState(c.Sender().ID, &productState{
			Step:     "cat_name",
			ParentID: parentID,
		})
		_ = c.Respond(&AsliBot.CallbackResponse{})
		return c.Send("📁 نام دسته‌بندی را بفرست:\n\nمثال: 🖥 سرور مجازی خارج")
	}
}

func CategorySubsHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showAdminCategories(c, db, catID)
	}
}

func CategoryProductsHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		return showAdminProducts(c, db, catID)
	}
}

func showAdminProducts(c AsliBot.Context, db *database.DB, catID int64) error {
	cat := db.GetCategory(catID)
	if cat == nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: "❌ دسته پیدا نشد"})
	}
	products := db.GetProductsByCategory(catID)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📦 محصولات %s %s\n\n", cat.Emoji, cat.Name))
	if len(products) == 0 {
		sb.WriteString("هیچ محصولی ثبت نشده.")
	} else {
		sb.WriteString(fmt.Sprintf("تعداد: %d", len(products)))
	}

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	cur := db.GetCurrency()
	for _, p := range products {
		status := "✅"
		if !p.Available {
			status = "❌"
		}
		label := fmt.Sprintf("%s %s — %s", status, p.Name, rate.Display(p.PriceUSD, cur))
		rows = append(rows, inline.Row(
			inline.Data(label, fmt.Sprintf("adm_prod_view_%d", p.ID)),
		))
	}

	rows = append(rows, inline.Row(
		inline.Data("➕ افزودن محصول", fmt.Sprintf("adm_prod_add_%d", catID)),
	))
	rows = append(rows, inline.Row(
		inline.Data("⬅️ برگشت", fmt.Sprintf("adm_cat_view_%d", catID)),
	))

	inline.Inline(rows...)
	return c.Send(sb.String(), inline)
}

func ProductViewHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		p := db.GetProduct(productID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ محصول پیدا نشد"})
		}
		return showAdminProductView(c, db, p)
	}
}

func showAdminProductView(c AsliBot.Context, db *database.DB, p *database.Product) error {
	status := "✅ موجود"
	if !p.Available {
		status = "❌ ناموجود"
	}
	active := "✅ فعال"
	if !p.Active {
		active = "❌ غیرفعال"
	}
	delivery := "دستی"
	if p.DeliveryType == "code_request" {
		delivery = "دریافت کد"
	}

	cur := db.GetCurrency()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"📦 <b>%s</b>\n\n"+
			"📁 دسته: %d\n"+
			"💰 قیمت: <b>%s</b>\n"+
			"📊 وضعیت: %s\n"+
			"🔘 فعال: %s\n"+
			"🚚 نوع تحویل: %s\n",
		p.Name, p.CategoryID, rate.Display(p.PriceUSD, cur), status, active, delivery,
	))

	if p.ProductType == "vpn" {
		panel := db.GetXUIPanel(p.XUIPanelID)
		panelName := "—"
		if panel != nil {
			panelName = panel.Name
		}
		sb.WriteString(fmt.Sprintf(
			"\n🌐 <b>VPN</b>\n"+
				"   🖥 پنل: %s\n"+
				"   📡 inboundها: %v\n"+
				"   📊 ترافیک: %d GB\n"+
				"   ⏰ مدت: %d روز\n",
			panelName, p.XUIInboundIDs, p.TrafficGB, p.DurationDays,
		))
	}

	sb.WriteString(fmt.Sprintf("\n📝 توضیحات:\n%s", p.Description))

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row

	if p.Available {
		rows = append(rows, inline.Row(
			inline.Data("❌ ناموجود کن", fmt.Sprintf("adm_prod_toggle_avail_%d", p.ID)),
		))
	} else {
		rows = append(rows, inline.Row(
			inline.Data("✅ موجود کن", fmt.Sprintf("adm_prod_toggle_avail_%d", p.ID)),
		))
	}

	rows = append(rows, inline.Row(
		inline.Data("✏️ تغییر قیمت", fmt.Sprintf("adm_prod_price_%d", p.ID)),
	))

	if p.Active {
		rows = append(rows, inline.Row(
			inline.Data("❌ غیرفعال کردن", fmt.Sprintf("adm_prod_toggle_active_%d", p.ID)),
		))
	} else {
		rows = append(rows, inline.Row(
			inline.Data("✅ فعال کردن", fmt.Sprintf("adm_prod_toggle_active_%d", p.ID)),
		))
	}

	rows = append(rows, inline.Row(
		inline.Data("🗑 حذف محصول", fmt.Sprintf("adm_prod_del_%d", p.ID)),
	))
	rows = append(rows, inline.Row(
		inline.Data("⬅️ برگشت", fmt.Sprintf("adm_cat_products_%d", p.CategoryID)),
	))

	inline.Inline(rows...)
	return c.Send(sb.String(), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func ProductToggleAvailableHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance && role != database.RoleSupport {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		p, err := db.ToggleProductAvailable(productID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ تغییر کرد"})
		return showAdminProductView(c, db, p)
	}
}

func ProductToggleActiveHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		if err := db.ToggleProductActive(productID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ تغییر کرد"})
		p := db.GetProduct(productID)
		if p == nil {
			return c.Send("❌ محصول پیدا نشد")
		}
		return showAdminProductView(c, db, p)
	}
}

func ProductDeleteHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		p := db.GetProduct(productID)
		if p == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ محصول پیدا نشد"})
		}
		catID := p.CategoryID
		if err := db.DeleteProduct(productID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
		}
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ حذف شد"})
		return showAdminProducts(c, db, catID)
	}
}

func ProductPriceStartHandler(cfg *config.Config, db *database.DB, productID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID) != database.RoleSuperAdmin {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط ادمین اصلی"})
		}
		setProductState(c.Sender().ID, &productState{
			Step:      "prod_price_edit",
			ProductID: productID,
		})
		_ = c.Respond(&AsliBot.CallbackResponse{})
		return c.Send("💰 قیمت جدید را به دلار بفرست (مثال: 5.5):")
	}
}

func ProductAddStartHandler(cfg *config.Config, db *database.DB, catID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		setProductState(c.Sender().ID, &productState{
			Step:       "prod_name",
			CategoryID: catID,
		})
		_ = c.Respond(&AsliBot.CallbackResponse{})
		return c.Send("📦 نام محصول را بفرست:\n\nمثال: سرور آلمان")
	}
}

func ProductInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getProductState(uid)
	if st == nil {
		return nil
	}
	role := db.GetAdminRole(uid, cfg.ADMIN_ID)
	if role != database.RoleSuperAdmin && role != database.RoleFinance {
		setProductState(uid, nil)
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	switch st.Step {
	case "cat_name":
		if text == "" {
			return c.Send("❌ نام نمی‌تواند خالی باشد. دوباره بفرست:")
		}
		emoji, name := splitEmojiName(text)
		cat, err := db.CreateCategory(st.ParentID, name, emoji)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}
		setProductState(uid, nil)
		return c.Send(fmt.Sprintf("✅ دسته <code>%d</code> با نام «%s %s» ساخته شد.", cat.ID, cat.Emoji, cat.Name),
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})

	case "prod_name":
		if text == "" {
			return c.Send("❌ نام نمی‌تواند خالی باشد. دوباره بفرست:")
		}
		st.Name = text
		st.Step = "prod_desc"
		setProductState(uid, st)
		return c.Send("📝 توضیحات محصول را بفرست (یا بنویس «-» برای خالی):")

	case "prod_desc":
		if text == "-" {
			text = ""
		}
		st.Description = text
		st.Step = "prod_price"
		setProductState(uid, st)
		return c.Send("💰 قیمت را به دلار بفرست (مثال: 5.5):")

	case "prod_price":
		text = strings.Replace(text, ",", ".", 1)
		price, err := strconv.ParseFloat(text, 64)
		if err != nil || price <= 0 {
			return c.Send("❌ قیمت نامعتبر. یک عدد مثبت بفرست:")
		}
		st.PriceUSD = price
		st.Step = "prod_delivery"
		setProductState(uid, st)

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("🚚 تحویل دستی", "adm_prod_delivery_manual"),
				inline.Data("🔑 دریافت کد", "adm_prod_delivery_code"),
			),
		)
		return c.Send("🚚 نوع تحویل را انتخاب کن:", inline)

	case "prod_price_edit":
		text = strings.Replace(text, ",", ".", 1)
		price, err := strconv.ParseFloat(text, 64)
		if err != nil || price <= 0 {
			return c.Send("❌ قیمت نامعتبر. یک عدد مثبت بفرست:")
		}
		if err := db.UpdateProductPrice(st.ProductID, price); err != nil {
			return c.Send("❌ " + err.Error())
		}
		setProductState(uid, nil)
		return c.Send(fmt.Sprintf("✅ قیمت به %.2f دلار تغییر کرد.", price))

	case "prod_traffic":
		gb, err := strconv.ParseInt(text, 10, 64)
		if err != nil || gb <= 0 {
			return c.Send("❌ عدد نامعتبر. یک عدد مثبت بفرست:")
		}
		st.TrafficGB = gb
		st.Step = "prod_duration"
		setProductState(uid, st)
		return c.Send("⏰ مدت زمان محصول را به روز بفرست (مثال: 30):")

	case "prod_duration":
		days, err := strconv.Atoi(text)
		if err != nil || days <= 0 {
			return c.Send("❌ عدد نامعتبر. یک عدد مثبت بفرست:")
		}
		st.DurationDays = days

		p, err := db.CreateProduct(st.CategoryID, st.Name, st.Description, st.PriceUSD, st.Delivery)
		if err != nil {
			return c.Send("❌ " + err.Error())
		}
		if err := db.SetProductVPN(p.ID, st.PanelID, st.InboundIDs, st.TrafficGB, st.DurationDays); err != nil {
			return c.Send("❌ " + err.Error())
		}
		setProductState(uid, nil)
		return c.Send(fmt.Sprintf("✅ محصول VPN ساخته شد.\n\n🆔 %d", p.ID),
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
	}
	return nil
}

func ProductDeliverySelectHandler(cfg *config.Config, db *database.DB, delivery string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getProductState(uid)
		if st == nil || st.Step != "prod_delivery" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		if delivery != "manual" && delivery != "code_request" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ نوع نامعتبر"})
		}
		st.Delivery = delivery
		st.Step = "prod_type"
		setProductState(uid, st)
		_ = c.Respond()

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("🌐 VPN (ساخت خودکار)", "adm_prod_type_vpn"),
				inline.Data("📦 محصول عادی", "adm_prod_type_normal"),
			),
		)
		return c.Send("🌐 این محصول VPN است یا عادی؟", inline)
	}
}

func ProductTypeSelectHandler(cfg *config.Config, db *database.DB, productType string) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getProductState(uid)
		if st == nil || st.Step != "prod_type" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		_ = c.Respond()

		if productType == "normal" {
			p, err := db.CreateProduct(st.CategoryID, st.Name, st.Description, st.PriceUSD, st.Delivery)
			if err != nil {
				return c.Send("❌ " + err.Error())
			}
			setProductState(uid, nil)
			return showAdminProductView(c, db, p)
		}

		panels := db.GetAllXUIPanels()
		active := []*database.XUIPanel{}
		for _, p := range panels {
			if p.Active {
				active = append(active, p)
			}
		}
		if len(active) == 0 {
			return c.Send("⚠️ هیچ پنل 3x-ui فعالی ثبت نشده.\n\nاول از «🖥 پنل‌های 3x-ui» یک پنل اضافه کن.")
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row
		for _, p := range active {
			rows = append(rows, inline.Row(
				inline.Data(p.Name, fmt.Sprintf("adm_prod_panel_%d", p.ID)),
			))
		}
		inline.Inline(rows...)
		st.Step = "prod_panel"
		setProductState(uid, st)
		return c.Send("🖥 پنل 3x-ui را انتخاب کن:", inline)
	}
}

func ProductPanelSelectHandler(cfg *config.Config, db *database.DB, panelID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getProductState(uid)
		if st == nil || st.Step != "prod_panel" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		_ = c.Respond()

		panel := db.GetXUIPanel(panelID)
		if panel == nil {
			return c.Send("❌ پنل پیدا نشد")
		}
		st.PanelID = panelID
		st.Step = "prod_inbound"
		setProductState(uid, st)

		cli := xui.New(xui.Config{
			BaseURL:    panel.BaseURL,
			AuthMethod: xui.AuthMethod(panel.AuthMethod),
			APIToken:   panel.APIToken,
			Username:   panel.Username,
			Password:   panel.Password,
		})

		inbounds, err := cli.ListInbounds()
		if err != nil {
			return c.Send("❌ خطا در دریافت inboundها: " + err.Error())
		}
		if len(inbounds) == 0 {
			return c.Send("📡 هیچ inboundی در این پنل نیست.")
		}

		inline := &AsliBot.ReplyMarkup{}
		var rows []AsliBot.Row
		for _, inb := range inbounds {
			if !inb.Enable {
				continue
			}
			label := fmt.Sprintf("#%d %s (%s:%d)", inb.ID, inb.Remark, inb.Protocol, inb.Port)
			rows = append(rows, inline.Row(
				inline.Data(label, fmt.Sprintf("adm_prod_inbound_%d", inb.ID)),
			))
		}
		inline.Inline(rows...)
		return c.Send("📡 inbound(ها) را انتخاب کن:", inline)
	}
}

func ProductInboundSelectHandler(cfg *config.Config, db *database.DB, inboundID int64) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getProductState(uid)
		if st == nil || st.Step != "prod_inbound" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		found := false
		for _, id := range st.InboundIDs {
			if id == inboundID {
				found = true
				break
			}
		}
		if !found {
			st.InboundIDs = append(st.InboundIDs, inboundID)
		}
		setProductState(uid, st)

		inline := &AsliBot.ReplyMarkup{}
		inline.Inline(
			inline.Row(
				inline.Data("✅ ادامه", "adm_prod_inbound_done"),
				inline.Data("🔙 برگشت", "adm_prod_inbound_back"),
			),
		)
		_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ اضافه شد"})
		return c.Send(fmt.Sprintf("📡 inbound #%d اضافه شد. (تعداد: %d)\n\nادامه بده یا برگرد.",
			inboundID, len(st.InboundIDs)), inline)
	}
}

func ProductInboundDoneHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID
		st := getProductState(uid)
		if st == nil || st.Step != "prod_inbound" {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
		}
		if len(st.InboundIDs) == 0 {
			return c.Respond(&AsliBot.CallbackResponse{Text: "❌ حداقل یک inbound انتخاب کن"})
		}
		_ = c.Respond()
		st.Step = "prod_traffic"
		setProductState(uid, st)
		return c.Send("📊 ترافیک محصول را به GB بفرست (مثال: 30):")
	}
}

func splitEmojiName(text string) (string, string) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "📁", text
	}
	first := fields[0]
	runes := []rune(first)
	if len(runes) > 0 && (runes[0] >= 0x1F300) {
		name := strings.Join(fields[1:], " ")
		if name == "" {
			name = first
			first = "📁"
		}
		return first, name
	}
	return "📁", text
}
