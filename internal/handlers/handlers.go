// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/handlers/admin"
	"github.com/theasli/GamajBot/internal/handlers/support"
	"github.com/theasli/GamajBot/internal/i18n"
	"github.com/theasli/GamajBot/internal/menus"
	"github.com/theasli/GamajBot/internal/middleware"
	"github.com/theasli/GamajBot/internal/notify"

	AsliBot "gopkg.in/telebot.v4"
)

const (
	MsgWelcome = "سلام %s 👋\nبه ربات خوش اومدی!"
	MsgHelp    = "📖 راهنما:\n" +
		"/start - شروع\n" +
		"/myaccount - حساب من\n" +
		"/charge - افزایش موجودی\n" +
		"/support - پشتیبانی\n" +
		"/cancel - لغو\n" +
		"/help - راهنما"
)

var botInstance *AsliBot.Bot

func Register(b *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	botInstance = b

	b.Use(middleware.Ban(cfg, db))
	b.Use(middleware.Maintenance(cfg, db))
	b.Use(middleware.ForceJoin(b, cfg))

	b.Handle("/start", StartHandler)
	b.Handle("/BuySubscription", SubscriptionHandler(cfg, db))
	b.Handle("/BuyOthersProducts", ShopHandler(cfg, db))
	b.Handle("/MyServices", MyServicesHandler(cfg, db))
	b.Handle("/myaccount", MyAccountHandler)
	b.Handle("/charge", ChargeStartHandler)
	b.Handle("/support", support.SupportHandler(cfg, db))
	b.Handle("/cancel", CancelCommandHandler)
	b.Handle("/help", HelpHandler)
	b.Handle("/history", HistoryHandler)
	b.Handle("/admin", admin.PanelHandler(cfg, db))

	b.Handle(&menus.BtnBuySubscription, SubscriptionHandler(cfg, db))
	b.Handle(&menus.BtnBuyOthersProducts, ShopHandler(cfg, db))
	b.Handle(&menus.BtnMyServices, MyServicesHandler(cfg, db))
	b.Handle(&menus.BtnMyAccount, MyAccountHandler)
	b.Handle(&menus.BtnCharge, ChargeStartHandler)
	b.Handle(&menus.BtnSupport, support.SupportHandler(cfg, db))
	b.Handle(&menus.BtnHelp, HelpHandler)
	b.Handle(&menus.BtnDev, DevHandler(cfg))
	b.Handle(&menus.BtnCom, ComingSoon)

	b.Handle("\fwallet_check", CheckPaymentHandler)
	b.Handle("\fwallet_cancel", CancelPaymentHandler)
	b.Handle("\fwallet_manual_ref", CardChargeRefHandler)
	b.Handle("\fforce_join_check", ForceJoinCheckHandler(b, cfg))

	b.Handle("\faccount_history", HistoryHandler)
	for i := 0; i < 100; i++ {
		page := i
		b.Handle(
			fmt.Sprintf("\faccount_history_%d", page),
			func(c AsliBot.Context) error { return showHistoryPage(c, db, page) },
		)
	}

	b.Handle("\fcharge_confirm", ChargeConfirmHandler)
	b.Handle("\fcharge_cancel_confirm", ChargeCancelConfirmHandler)

	b.Handle("\fadm_stats", admin.StatsHandler(cfg, db))
	b.Handle("\fadm_users", admin.UsersHandler(cfg, db))
	b.Handle("\fadm_users_search", admin.UsersSearchHandler(cfg, db))
	b.Handle("\fadm_tx", admin.TxHandler(cfg, db))
	for i := 0; i < 100; i++ {
		page := i
		b.Handle(fmt.Sprintf("\fadm_users_%d", page), admin.UsersPageHandler(cfg, db, page))
		b.Handle(fmt.Sprintf("\fadm_tx_%d", page), admin.TxPageHandler(cfg, db, page))
	}

	b.Handle("\fadm_report_today", admin.ReportsHandler(cfg, db, "today"))
	b.Handle("\fadm_report_week", admin.ReportsHandler(cfg, db, "week"))
	b.Handle("\fadm_report_month", admin.ReportsHandler(cfg, db, "month"))
	b.Handle("\fadm_report_all", admin.ReportsHandler(cfg, db, "all"))
	b.Handle("\fadm_report_top_today", admin.TopProductsHandler(cfg, db, "today"))
	b.Handle("\fadm_report_top_week", admin.TopProductsHandler(cfg, db, "week"))
	b.Handle("\fadm_report_top_month", admin.TopProductsHandler(cfg, db, "month"))
	b.Handle("\fadm_report_top_all", admin.TopProductsHandler(cfg, db, "all"))

	b.Handle("\fadm_analytics", admin.AnalyticsHandler(cfg, db))
	b.Handle("\fadm_analytics_report", admin.AnalyticsReportHandler(cfg, db))

	b.Handle("\fadm_finance", admin.FinancePanelHandler(cfg, db))
	b.Handle("\fadm_gateways", admin.GatewaysHandler(cfg, db))
	b.Handle("\fadm_gw_manual", admin.GatewayManualHandler(cfg, db))
	b.Handle("\fadm_gw_manual_toggle", admin.GatewayManualToggleHandler(cfg, db))
	b.Handle("\fadm_gw_manual_number", admin.GatewayManualNumberHandler(cfg, db))
	b.Handle("\fadm_gw_manual_holder", admin.GatewayManualHolderHandler(cfg, db))
	b.Handle("\fadm_gw_cube", admin.GatewayCubeHandler(cfg, db))
	b.Handle("\fadm_gw_cube_toggle", admin.GatewayCubeToggleHandler(cfg, db))
	b.Handle("\fadm_gw_cube_token", admin.GatewayCubeTokenHandler(cfg, db))
	b.Handle("\fadm_gw_tetra", admin.GatewayTetraHandler(cfg, db))
	b.Handle("\fadm_gw_tetra_toggle", admin.GatewayTetraToggleHandler(cfg, db))
	b.Handle("\fadm_gw_tetra_key", admin.GatewayTetraKeyHandler(cfg, db))
	b.Handle("\fadm_gw_tetra_url", admin.GatewayTetraURLHandler(cfg, db))
	b.Handle("\fadm_gw_aban", admin.GatewayAbanHandler(cfg, db))
	b.Handle("\fadm_gw_aban_toggle", admin.GatewayAbanToggleHandler(cfg, db))
	b.Handle("\fadm_gw_aban_key", admin.GatewayAbanKeyHandler(cfg, db))
	b.Handle("\fadm_gw_aban_secret", admin.GatewayAbanSecretHandler(cfg, db))

	b.Handle("\fadm_deduct", admin.DeductStartHandler(cfg, db))

	b.Handle("\fadm_admins", admin.AdminsHandler(cfg, db))
	b.Handle("\fadm_add", admin.AddAdminHandler(cfg, db))
	b.Handle("\fadm_rm", admin.RemoveAdminHandler(cfg, db))
	b.Handle("\frole_super", admin.AddAdminRoleHandler(cfg, db))
	b.Handle("\frole_finance", admin.AddAdminRoleHandler(cfg, db))
	b.Handle("\frole_support", admin.AddAdminRoleHandler(cfg, db))
	b.Handle("\frole_viewer", admin.AddAdminRoleHandler(cfg, db))

	for i := 1; i < 1000; i++ {
		adminID := int64(i)
		b.Handle(fmt.Sprintf("\fadm_admin_depts_%d", adminID),
			admin.AdminDeptsHandler(cfg, db, adminID))
		for j := 1; j < 100; j++ {
			deptID := int64(j)
			b.Handle(fmt.Sprintf("\fadm_dept_toggle_%d_%d", adminID, deptID),
				admin.AdminDeptToggleHandler(cfg, db, adminID, deptID))
		}
	}

	b.Handle("\fadm_settings", admin.SettingsHandler(cfg, db))
	b.Handle("\fset_currency", admin.SetCurrencyHandler(cfg, db))
	b.Handle("\fset_cur_USD", admin.SetCurrencySelectHandler(cfg, db, database.CurrencyUSD))
	b.Handle("\fset_cur_USDT", admin.SetCurrencySelectHandler(cfg, db, database.CurrencyUSDT))
	b.Handle("\fset_cur_TOMAN", admin.SetCurrencySelectHandler(cfg, db, database.CurrencyToman))
	b.Handle("\fset_cur_RIAL", admin.SetCurrencySelectHandler(cfg, db, database.CurrencyRial))
	b.Handle("\fset_margin", admin.SetMarginHandler(cfg, db))
	b.Handle("\fset_min", admin.SetMinHandler(cfg, db))
	b.Handle("\fset_max", admin.SetMaxHandler(cfg, db))
	b.Handle("\fset_quick", admin.SetQuickHandler(cfg, db))
	b.Handle("\fset_maintenance", admin.ToggleMaintenanceHandler(cfg, db))
	b.Handle("\fset_support", admin.ToggleSupportHandler(cfg, db))

	b.Handle("\fadm_log_setup", admin.LogSetupHandler(cfg, db))
	b.Handle("\fadm_log_set_chat", admin.LogSetupSetChatHandler(cfg, db))
	b.Handle("\fadm_log_build", admin.LogSetupBuildHandler(cfg, db))
	b.Handle("\fadm_log_clear", admin.LogSetupClearHandler(cfg, db))

	b.Handle("\fadm_depts", admin.DepartmentsHandler(cfg, db))
	b.Handle("\fdept_add", admin.AddDeptHandler(cfg, db))
	b.Handle("\fdept_rm", admin.RemoveDeptHandler(cfg, db))
	b.Handle("\fdept_toggle", admin.ToggleDeptHandler(cfg, db))

	b.Handle("\fadm_tickets", admin.TicketsHandler(cfg, db))
	for i := 1; i < 1000; i++ {
		id := int64(i)
		b.Handle(fmt.Sprintf("\fticket_view_%d", id), admin.ViewTicketHandler(cfg, db, id))
		b.Handle(fmt.Sprintf("\fticket_reply_%d", id), admin.ReplyTicketHandler(cfg, db, id))
		b.Handle(fmt.Sprintf("\fticket_close_%d", id), admin.CloseTicketHandler(cfg, db, id))
		b.Handle(fmt.Sprintf("\fticket_pending_%d", id), admin.PendingTicketHandler(cfg, db, id))
		b.Handle(fmt.Sprintf("\fticket_prio_low_%d", id), admin.SetTicketPriorityHandler(cfg, db, id, database.PriorityLow))
		b.Handle(fmt.Sprintf("\fticket_prio_normal_%d", id), admin.SetTicketPriorityHandler(cfg, db, id, database.PriorityNormal))
		b.Handle(fmt.Sprintf("\fticket_prio_high_%d", id), admin.SetTicketPriorityHandler(cfg, db, id, database.PriorityHigh))
		b.Handle(fmt.Sprintf("\fticket_prio_urgent_%d", id), admin.SetTicketPriorityHandler(cfg, db, id, database.PriorityUrgent))
		b.Handle(fmt.Sprintf("\fticket_transfer_%d", id), admin.TransferTicketHandler(cfg, db, id))
		for j := 1; j < 100; j++ {
			deptID := int64(j)
			b.Handle(fmt.Sprintf("\fticket_transfer_to_%d_%d", id, deptID),
				admin.TransferTicketToHandler(cfg, db, id, deptID))
		}
	}

	b.Handle("\fadm_broadcast", admin.BroadcastHandler(cfg, db))
	b.Handle("\fadm_manual_charge", admin.ManualChargeStartHandler(cfg, db))

	b.Handle("\fadm_products", admin.ProductsAdminHandler(cfg, db))
	b.Handle("\fadm_orders", admin.OrdersHandler(cfg, db))

	b.Handle("\fadm_xui_panels", admin.XUIPanelHandler(cfg, db))
	b.Handle("\fadm_xui_add", admin.XUIPanelAddStartHandler(cfg, db))
	b.Handle("\fadm_xui_auth_token", admin.XUIAuthSelectHandler(cfg, db, "token"))
	b.Handle("\fadm_xui_auth_login", admin.XUIAuthSelectHandler(cfg, db, "login"))

	b.Handle("\fsup_new", support.NewTicketHandler(cfg, db))
	b.Handle("\fsup_my", support.MyTicketsHandler(cfg, db))
	b.Handle("\fsup_prio_low", support.SelectPriorityHandler(cfg, db, database.PriorityLow))
	b.Handle("\fsup_prio_normal", support.SelectPriorityHandler(cfg, db, database.PriorityNormal))
	b.Handle("\fsup_prio_high", support.SelectPriorityHandler(cfg, db, database.PriorityHigh))
	b.Handle("\fsup_prio_urgent", support.SelectPriorityHandler(cfg, db, database.PriorityUrgent))
	for i := 1; i < 1000; i++ {
		id := int64(i)
		b.Handle(fmt.Sprintf("\fsup_dept_%d", id), support.SelectDeptHandler(cfg, db, id))
		b.Handle(fmt.Sprintf("\fsup_view_%d", id), support.ViewMyTicketHandler(cfg, db, id))
		b.Handle(fmt.Sprintf("\fsup_reply_%d", id), support.ReplyMyTicketHandler(cfg, db, id))
	}

	b.Handle(AsliBot.OnCallback, func(c AsliBot.Context) error {
		data := c.Callback().Data
		data = strings.TrimPrefix(data, "\f")

		switch {
		case strings.HasPrefix(data, "charge_gw_"):
			return ChargeGatewaySelectHandler(c)
		case strings.HasPrefix(data, "charge_amt_"):
			return ChargeAmountSelectedHandler(c)
		case strings.HasPrefix(data, "charge_custom_"):
			return ChargeCustomHandler(c)

		case data == "account_language":
			return LanguageHandler(c)

		case data == "lang_set_fa":
			return LanguageSetHandler(c, db, "fa")

		case data == "lang_set_en":
			return LanguageSetHandler(c, db, "en")

		case strings.HasPrefix(data, "adm_mc_approve|") ||
			strings.HasPrefix(data, "adm_mc_reject|") ||
			strings.HasPrefix(data, "adm_mc_review|") ||
			strings.HasPrefix(data, "adm_mc_block|"):
			return admin.ManualChargeCallbackHandler(cfg, db)(c)

		case strings.HasPrefix(data, "adm_deduct_approve|") ||
			strings.HasPrefix(data, "adm_deduct_reject|"):
			return admin.DeductCallbackHandler(cfg, db)(c)

		case strings.HasPrefix(data, "adm_user_view_"):
			idStr := strings.TrimPrefix(data, "adm_user_view_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.UserViewHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_user_charge_"):
			idStr := strings.TrimPrefix(data, "adm_user_charge_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.UserChargeHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_user_deduct_"):
			idStr := strings.TrimPrefix(data, "adm_user_deduct_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.UserDeductHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_user_ban_"):
			idStr := strings.TrimPrefix(data, "adm_user_ban_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.UserBanHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_user_unban_"):
			idStr := strings.TrimPrefix(data, "adm_user_unban_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.UserUnbanHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_user_tx_"):
			rest := strings.TrimPrefix(data, "adm_user_tx_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			id, err1 := strconv.ParseInt(parts[0], 10, 64)
			page, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.UserTxPageHandler(cfg, db, id, page)(c)

		case strings.HasPrefix(data, "adm_cat_view_"):
			idStr := strings.TrimPrefix(data, "adm_cat_view_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryViewHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cat_add_"):
			idStr := strings.TrimPrefix(data, "adm_cat_add_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryAddStartHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cat_subs_"):
			idStr := strings.TrimPrefix(data, "adm_cat_subs_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategorySubsHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cat_products_"):
			idStr := strings.TrimPrefix(data, "adm_cat_products_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryProductsHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cat_toggle_"):
			idStr := strings.TrimPrefix(data, "adm_cat_toggle_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryToggleHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cat_del_"):
			idStr := strings.TrimPrefix(data, "adm_cat_del_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryDeleteHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cf_view_"):
			idStr := strings.TrimPrefix(data, "adm_cf_view_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryFieldsHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cf_add_"):
			idStr := strings.TrimPrefix(data, "adm_cf_add_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryFieldAddHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cf_addact_"):
			idStr := strings.TrimPrefix(data, "adm_cf_addact_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryActionAddHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_cf_ft_"):
			fieldType := strings.TrimPrefix(data, "adm_cf_ft_")
			return admin.CategoryFieldTypeSelectHandler(cfg, db, fieldType)(c)

		case data == "adm_cf_fr_yes":
			return admin.CategoryFieldRequiredHandler(cfg, db, true)(c)
		case data == "adm_cf_fr_no":
			return admin.CategoryFieldRequiredHandler(cfg, db, false)(c)
		case data == "adm_cf_ar_yes":
			return admin.CategoryActionRequireInputHandler(cfg, db, true)(c)
		case data == "adm_cf_ar_no":
			return admin.CategoryActionRequireInputHandler(cfg, db, false)(c)

		case strings.HasPrefix(data, "adm_cf_delf_"):
			rest := strings.TrimPrefix(data, "adm_cf_delf_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryFieldDeleteHandler(cfg, db, id, parts[1])(c)

		case strings.HasPrefix(data, "adm_cf_dela_"):
			rest := strings.TrimPrefix(data, "adm_cf_dela_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.CategoryActionDeleteHandler(cfg, db, id, parts[1])(c)

		case strings.HasPrefix(data, "adm_prod_view_"):
			idStr := strings.TrimPrefix(data, "adm_prod_view_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductViewHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_prod_add_"):
			idStr := strings.TrimPrefix(data, "adm_prod_add_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductAddStartHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_prod_toggle_avail_"):
			idStr := strings.TrimPrefix(data, "adm_prod_toggle_avail_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductToggleAvailableHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_prod_toggle_active_"):
			idStr := strings.TrimPrefix(data, "adm_prod_toggle_active_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductToggleActiveHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_prod_del_"):
			idStr := strings.TrimPrefix(data, "adm_prod_del_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductDeleteHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_prod_price_"):
			idStr := strings.TrimPrefix(data, "adm_prod_price_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductPriceStartHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_prod_delivery_"):
			delivery := strings.TrimPrefix(data, "adm_prod_delivery_")
			return admin.ProductDeliverySelectHandler(cfg, db, delivery)(c)

		case data == "adm_prod_type_vpn":
			return admin.ProductTypeSelectHandler(cfg, db, "vpn")(c)
		case data == "adm_prod_type_normal":
			return admin.ProductTypeSelectHandler(cfg, db, "normal")(c)

		case strings.HasPrefix(data, "adm_prod_panel_"):
			idStr := strings.TrimPrefix(data, "adm_prod_panel_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductPanelSelectHandler(cfg, db, id)(c)

		case data == "adm_prod_inbound_done":
			return admin.ProductInboundDoneHandler(cfg, db)(c)

		case data == "adm_prod_inbound_back":
			return admin.ProductDeliverySelectHandler(cfg, db, "manual")(c)

		case strings.HasPrefix(data, "adm_prod_inbound_"):
			idStr := strings.TrimPrefix(data, "adm_prod_inbound_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ProductInboundSelectHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_order_view_"):
			orderID := strings.TrimPrefix(data, "adm_order_view_")
			return admin.OrderViewHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "adm_order_deliver_"):
			orderID := strings.TrimPrefix(data, "adm_order_deliver_")
			return admin.OrderDeliverHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "adm_order_reject_"):
			orderID := strings.TrimPrefix(data, "adm_order_reject_")
			return admin.OrderRejectHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "adm_order_code_"):
			orderID := strings.TrimPrefix(data, "adm_order_code_")
			return admin.OrderCodeHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "adm_orders_"):
			rest := strings.TrimPrefix(data, "adm_orders_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			filter := parts[0]
			page, err := strconv.Atoi(parts[1])
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.OrdersPageHandler(cfg, db, filter, page)(c)

		case strings.HasPrefix(data, "adm_ar_view_"):
			rest := strings.TrimPrefix(data, "adm_ar_view_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ActionRequestViewHandler(cfg, db, parts[0], parts[1])(c)

		case strings.HasPrefix(data, "adm_ar_done_"):
			rest := strings.TrimPrefix(data, "adm_ar_done_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ActionRequestDoneHandler(cfg, db, parts[0], parts[1])(c)

		case strings.HasPrefix(data, "adm_ar_rej_"):
			rest := strings.TrimPrefix(data, "adm_ar_rej_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.ActionRequestRejectHandler(cfg, db, parts[0], parts[1])(c)

		case strings.HasPrefix(data, "adm_xui_stats_"):
			idStr := strings.TrimPrefix(data, "adm_xui_stats_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIPanelStatsHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_xui_inbound_"):
			rest := strings.TrimPrefix(data, "adm_xui_inbound_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			panelID, err1 := strconv.ParseInt(parts[0], 10, 64)
			inboundID, err2 := strconv.ParseInt(parts[1], 10, 64)
			if err1 != nil || err2 != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIInboundClientsHandler(cfg, db, panelID, inboundID)(c)

		case strings.HasPrefix(data, "adm_xui_client_"):
			rest := strings.TrimPrefix(data, "adm_xui_client_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			panelID, err1 := strconv.ParseInt(parts[0], 10, 64)
			inboundID, err2 := strconv.ParseInt(parts[1], 10, 64)
			if err1 != nil || err2 != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIClientDetailHandler(cfg, db, panelID, inboundID, parts[2])(c)

		case strings.HasPrefix(data, "adm_xui_cli_enable_"):
			rest := strings.TrimPrefix(data, "adm_xui_cli_enable_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			panelID, _ := strconv.ParseInt(parts[0], 10, 64)
			inboundID, _ := strconv.ParseInt(parts[1], 10, 64)
			return admin.XUIClientActionHandler(cfg, db, "enable", panelID, inboundID, parts[2])(c)

		case strings.HasPrefix(data, "adm_xui_cli_disable_"):
			rest := strings.TrimPrefix(data, "adm_xui_cli_disable_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			panelID, _ := strconv.ParseInt(parts[0], 10, 64)
			inboundID, _ := strconv.ParseInt(parts[1], 10, 64)
			return admin.XUIClientActionHandler(cfg, db, "disable", panelID, inboundID, parts[2])(c)

		case strings.HasPrefix(data, "adm_xui_cli_reset_"):
			rest := strings.TrimPrefix(data, "adm_xui_cli_reset_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			panelID, _ := strconv.ParseInt(parts[0], 10, 64)
			inboundID, _ := strconv.ParseInt(parts[1], 10, 64)
			return admin.XUIClientActionHandler(cfg, db, "reset", panelID, inboundID, parts[2])(c)

		case strings.HasPrefix(data, "adm_xui_cli_renew_"):
			rest := strings.TrimPrefix(data, "adm_xui_cli_renew_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			panelID, _ := strconv.ParseInt(parts[0], 10, 64)
			inboundID, _ := strconv.ParseInt(parts[1], 10, 64)
			return admin.XUIClientActionHandler(cfg, db, "renew", panelID, inboundID, parts[2])(c)

		case strings.HasPrefix(data, "adm_xui_cli_del_"):
			rest := strings.TrimPrefix(data, "adm_xui_cli_del_")
			parts := strings.SplitN(rest, "_", 3)
			if len(parts) != 3 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			panelID, _ := strconv.ParseInt(parts[0], 10, 64)
			inboundID, _ := strconv.ParseInt(parts[1], 10, 64)
			return admin.XUIClientActionHandler(cfg, db, "del", panelID, inboundID, parts[2])(c)

		case strings.HasPrefix(data, "adm_xui_view_"):
			idStr := strings.TrimPrefix(data, "adm_xui_view_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIPanelViewHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_xui_test_"):
			idStr := strings.TrimPrefix(data, "adm_xui_test_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIPanelTestHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_xui_inbounds_"):
			idStr := strings.TrimPrefix(data, "adm_xui_inbounds_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIPanelInboundsHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_xui_toggle_"):
			idStr := strings.TrimPrefix(data, "adm_xui_toggle_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIPanelToggleHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "adm_xui_del_"):
			idStr := strings.TrimPrefix(data, "adm_xui_del_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return admin.XUIPanelDeleteHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "shop_cat_"):
			idStr := strings.TrimPrefix(data, "shop_cat_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return ShopCategoryHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "shop_prod_"):
			idStr := strings.TrimPrefix(data, "shop_prod_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return ShopProductHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "shop_buy_confirm_"):
			idStr := strings.TrimPrefix(data, "shop_buy_confirm_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return ShopBuyConfirmHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "shop_buy_"):
			idStr := strings.TrimPrefix(data, "shop_buy_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return ShopBuyHandler(cfg, db, id)(c)

		case data == "shop_buy_cancel":
			return ShopBuyCancelHandler(c)

		case data == "shop_noop":
			return ShopNoopHandler(c)

		case data == "shop_root":
			return ShopRootHandler(c)

		case strings.HasPrefix(data, "sub_fixed_"):
			idStr := strings.TrimPrefix(data, "sub_fixed_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return SubscriptionFixedHandler(cfg, db, id)(c)

		case data == "sub_dynamic":
			return SubscriptionDynamicHandler(cfg, db)(c)

		case strings.HasPrefix(data, "sub_buy_fixed_"):
			idStr := strings.TrimPrefix(data, "sub_buy_fixed_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return SubscriptionBuyFixedHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "sub_buy_dynamic_"):
			idStr := strings.TrimPrefix(data, "sub_buy_dynamic_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return SubscriptionBuyDynamicHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "sub_cat_"):
			idStr := strings.TrimPrefix(data, "sub_cat_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return SubscriptionCategoryHandler(cfg, db, id)(c)

		case strings.HasPrefix(data, "sub_prod_"):
			idStr := strings.TrimPrefix(data, "sub_prod_")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return SubscriptionProductHandler(cfg, db, id)(c)

		case data == "sub_root":
			return SubscriptionRootHandler(c)

		case strings.HasPrefix(data, "svc_view_"):
			orderID := strings.TrimPrefix(data, "svc_view_")
			return ServiceViewHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "svc_code_"):
			orderID := strings.TrimPrefix(data, "svc_code_")
			return ServiceRequestCodeHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "svc_stats_"):
			orderID := strings.TrimPrefix(data, "svc_stats_")
			return ServiceStatsHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "svc_relink_"):
			orderID := strings.TrimPrefix(data, "svc_relink_")
			return ServiceRelinkHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "svc_renew_"):
			orderID := strings.TrimPrefix(data, "svc_renew_")
			return ServiceRenewHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "svc_delete_"):
			orderID := strings.TrimPrefix(data, "svc_delete_")
			return ServiceDeleteHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "svc_qr_"):
			orderID := strings.TrimPrefix(data, "svc_qr_")
			return ServiceQRHandler(cfg, db, orderID)(c)

		case strings.HasPrefix(data, "svc_action_"):
			rest := strings.TrimPrefix(data, "svc_action_")
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ خطا"})
			}
			return ServiceActionRequestHandler(cfg, db, parts[0], parts[1])(c)

		case data == "svc_list":
			return ServicesListHandler(c)
		}
		return nil
	})

	b.Handle(AsliBot.OnText, func(c AsliBot.Context) error {
		if err := admin.GWInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.XUIInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.XUIManageInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.CategoryFieldInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.CategoryActionInputLabelHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.CategoryDeliveryInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.LogSetupInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.UserSearchInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.ProductInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.OrderInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := support.InputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.TicketReplyInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.SettingsInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.DeptInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.BroadcastInputHandler(b, c, cfg, db); err != nil {
			return err
		}
		if err := admin.InputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := admin.ManualChargeTextHandler(cfg, db)(c); err != nil {
			return err
		}
		if err := admin.DeductInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := ServiceActionInputHandler(c, cfg, db); err != nil {
			return err
		}
		if err := CardChargeRefInputHandler(c, cfg, db); err != nil {
			return err
		}
		return ChargeAmountHandler(c)
	})

	b.Handle(AsliBot.OnPhoto, func(c AsliBot.Context) error {
		return CardChargeRefInputHandler(c, cfg, db)
	})
	b.Handle(AsliBot.OnDocument, func(c AsliBot.Context) error {
		return CardChargeRefInputHandler(c, cfg, db)
	})
}

func StartHandler(c AsliBot.Context) error {
	user := c.Sender()
	walletDB.GetOrCreateUser(user.ID, user.FirstName, user.Username)
	lang := userLang(user.ID)

	go func() {
		if botInstance == nil {
			return
		}
		msg := fmt.Sprintf(
			"👋 <b>کاربر استارت زد</b>\n\n"+
				"👤 %s (@%s)\n"+
				"🆔 <code>%d</code>",
			user.FirstName, user.Username, user.ID,
		)
		notify.AdminLog(botInstance, appCfg, walletDB, notify.LogTopicUsers, msg)
	}()

	return c.Send(fmt.Sprintf(i18n.T(lang, "welcome"), user.FirstName), menus.MainMenu)
}

func HelpHandler(c AsliBot.Context) error {
	return c.Send(i18n.T(userLang(c.Sender().ID), "help"), menus.MainMenu)
}

func DevHandler(cfg *config.Config) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		return c.Send(cfg.ASLICODE_TEXTS)
	}
}

func ComingSoon(c AsliBot.Context) error {
	return c.Send("🚧 این بخش هنوز آماده نیست. به زودی...", menus.MainMenu)
}
