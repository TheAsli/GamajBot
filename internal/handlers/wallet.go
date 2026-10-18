// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/i18n"
	"github.com/theasli/GamajBot/internal/menus"
	"github.com/theasli/GamajBot/internal/payment"
	"github.com/theasli/GamajBot/internal/rate"

	AsliBot "gopkg.in/telebot.v4"
)

const (
	chargeTimeout  = 15 * time.Minute
	historyPerPage = 5
)

var (
	walletDB    *database.DB
	tetraClient *payment.Client
	appCfg      *config.Config

	awaitingAmount = make(map[int64]bool)
	stateMu        sync.Mutex

	awaitingConfirm = make(map[int64]*pendingInvoice)
	confirmMu       sync.Mutex

	pendingPayments = make(map[int64]string)
	payMu           sync.Mutex

	chargeGatewayMap = make(map[int64]string)
	chargeGwMu       sync.Mutex
)

type pendingInvoice struct {
	AmountUSD   float64
	AmountToman int64
	Gateway     string
}

func InitWallet(db *database.DB, tc *payment.Client, cfg *config.Config) {
	walletDB = db
	tetraClient = tc
	appCfg = cfg
}

func getCubeClient() *payment.CubePayClient {
	g := walletDB.GetPaymentGateways()
	return payment.GetCubeClient(g.CubePayToken)
}

func userLang(uid int64) i18n.Lang {
	u, err := walletDB.GetUser(uid)
	if err != nil || u == nil || u.Language == "" {
		return i18n.LangFA
	}
	return i18n.Lang(u.Language)
}

func setAwaiting(userID int64, on bool) {
	stateMu.Lock()
	defer stateMu.Unlock()
	if on {
		awaitingAmount[userID] = true
	} else {
		delete(awaitingAmount, userID)
	}
}

func isAwaiting(userID int64) bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	return awaitingAmount[userID]
}

func setAwaitingConfirm(userID int64, inv *pendingInvoice) {
	confirmMu.Lock()
	defer confirmMu.Unlock()
	if inv == nil {
		delete(awaitingConfirm, userID)
	} else {
		awaitingConfirm[userID] = inv
	}
}

func getAwaitingConfirm(userID int64) *pendingInvoice {
	confirmMu.Lock()
	defer confirmMu.Unlock()
	return awaitingConfirm[userID]
}

func setChargeGateway(uid int64, gw string) {
	chargeGwMu.Lock()
	defer chargeGwMu.Unlock()
	if gw == "" {
		delete(chargeGatewayMap, uid)
	} else {
		chargeGatewayMap[uid] = gw
	}
}

func getChargeGateway(uid int64) string {
	chargeGwMu.Lock()
	defer chargeGwMu.Unlock()
	return chargeGatewayMap[uid]
}

func MyAccountHandler(c AsliBot.Context) error {
	user := c.Sender()
	u := walletDB.GetOrCreateUser(user.ID, user.FirstName, user.Username)
	lang := userLang(user.ID)

	finalRate, err := rate.GetFinalRate(appCfg.DollarRateAPI, appCfg.DollarRateCache, appCfg.DollarRateMargin)
	tomanEq := int64(0)
	if err == nil {
		tomanEq = int64(u.BalanceUSD * float64(finalRate))
	}

	msg := i18n.T(lang, "my_account", u.ID, u.FirstName, u.Username, u.BalanceUSD)
	if tomanEq > 0 {
		msg += fmt.Sprintf("\n💰 %s تومان", formatMoney(tomanEq))
	}

	inline := &AsliBot.ReplyMarkup{}
	btnHistory := inline.Data(i18n.T(lang, "history"), "account_history_0")
	btnLang := inline.Data(i18n.T(lang, "language_btn"), "account_language")
	inline.Inline(
		inline.Row(btnHistory),
		inline.Row(btnLang),
	)

	return c.Send(msg, &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func HistoryHandler(c AsliBot.Context) error {
	return showHistoryPage(c, walletDB, 0)
}

func HistoryPageHandler(c AsliBot.Context) error {
	data := c.Callback().Data
	data = strings.TrimPrefix(data, "\f")
	data = strings.TrimPrefix(data, "account_history_")
	page, _ := strconv.Atoi(data)
	return showHistoryPage(c, walletDB, page)
}

func showHistoryPage(c AsliBot.Context, db *database.DB, page int) error {
	user := c.Sender()
	lang := userLang(user.ID)
	db.GetOrCreateUser(user.ID, user.FirstName, user.Username)

	allTxs, err := db.GetHistory(user.ID, 0)
	if err != nil || len(allTxs) == 0 {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "no_history")})
	}

	total := len(allTxs)
	start := page * historyPerPage
	if start >= total {
		page, start = 0, 0
	}
	end := start + historyPerPage
	if end > total {
		end = total
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(i18n.T(lang, "history_page"), page+1, (total+historyPerPage-1)/historyPerPage))
	for _, t := range allTxs[start:end] {
		sb.WriteString(formatTransaction(t, lang))
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf(i18n.T(lang, "history_total"), total))

	inline := &AsliBot.ReplyMarkup{}
	var row []AsliBot.Btn
	if page > 0 {
		row = append(row, inline.Data(i18n.T(lang, "prev"), fmt.Sprintf("account_history_%d", page-1)))
	}
	if end < total {
		row = append(row, inline.Data(i18n.T(lang, "next"), fmt.Sprintf("account_history_%d", page+1)))
	}

	if len(row) > 0 {
		inline.Inline(inline.Row(row...))
		return c.Send(sb.String(), inline)
	}
	return c.Send(sb.String())
}

func ChargeStartHandler(c AsliBot.Context) error {
	user := c.Sender()
	lang := userLang(user.ID)
	walletDB.GetOrCreateUser(user.ID, user.FirstName, user.Username)

	if _, err := walletDB.GetPendingCharge(user.ID); err == nil {
		return c.Send(i18n.T(lang, "charge_pending"))
	}

	g := walletDB.GetPaymentGateways()
	active := []string{}
	if g.TetraminatorEnabled {
		active = append(active, "tetraminator")
	}
	if g.CubePayEnabled {
		active = append(active, "cubepay")
	}
	if g.AbanEnabled {
		active = append(active, "abangateway")
	}
	if g.ManualCardEnabled {
		active = append(active, "manual_card")
	}

	if len(active) == 0 {
		return c.Send(i18n.T(lang, "no_gateway"))
	}
	if len(active) == 1 {
		return startChargeWithGateway(c, active[0])
	}

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row
	if g.TetraminatorEnabled {
		rows = append(rows, inline.Row(inline.Data("💎 Tetraminator (تایید آنی)", "charge_gw_tetraminator")))
	}
	if g.CubePayEnabled {
		rows = append(rows, inline.Row(inline.Data("🧊 CubePay (تایید آنی)", "charge_gw_cubepay")))
	}
	if g.AbanEnabled {
		rows = append(rows, inline.Row(inline.Data("🌐 AbanGateway (تایید آنی)", "charge_gw_abangateway")))
	}
	if g.ManualCardEnabled {
		rows = append(rows, inline.Row(inline.Data("💳 کارت‌به‌کارت (تایید دستی)", "charge_gw_manual_card")))
	}
	inline.Inline(rows...)

	return c.Send(i18n.T(lang, "choose_gateway"), inline)
}

func ChargeGatewaySelectHandler(c AsliBot.Context) error {
	data := c.Callback().Data
	data = strings.TrimPrefix(data, "\f")
	data = strings.TrimPrefix(data, "charge_gw_")
	if data != "tetraminator" && data != "cubepay" && data != "abangateway" && data != "manual_card" {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(userLang(c.Sender().ID), "error")})
	}
	_ = c.Respond()
	return startChargeWithGateway(c, data)
}

func startChargeWithGateway(c AsliBot.Context, gateway string) error {
	lang := userLang(c.Sender().ID)

	finalRate, err := rate.GetFinalRate(appCfg.DollarRateAPI, appCfg.DollarRateCache, appCfg.DollarRateMargin)
	if err != nil {
		return c.Send(i18n.T(lang, "rate_error"))
	}

	inline := &AsliBot.ReplyMarkup{}
	var rows []AsliBot.Row
	for i := 0; i < len(appCfg.QuickAmounts); i += 2 {
		var row []AsliBot.Btn
		row = append(row, inline.Data(
			fmt.Sprintf("%d دلار", appCfg.QuickAmounts[i]),
			fmt.Sprintf("charge_amt_%s_%d", gateway, appCfg.QuickAmounts[i]),
		))
		if i+1 < len(appCfg.QuickAmounts) {
			row = append(row, inline.Data(
				fmt.Sprintf("%d دلار", appCfg.QuickAmounts[i+1]),
				fmt.Sprintf("charge_amt_%s_%d", gateway, appCfg.QuickAmounts[i+1]),
			))
		}
		rows = append(rows, inline.Row(row...))
	}
	rows = append(rows, inline.Row(inline.Data(i18n.T(lang, "custom_amount"), "charge_custom_"+gateway)))
	inline.Inline(rows...)

	return c.Send(fmt.Sprintf(
		i18n.T(lang, "choose_amount"),
		gatewayLabel(gateway), formatMoney(finalRate),
	), inline)
}

func ChargeAmountSelectedHandler(c AsliBot.Context) error {
	user := c.Sender()
	data := c.Callback().Data
	data = strings.TrimPrefix(data, "\f")
	data = strings.TrimPrefix(data, "charge_amt_")

	idx := strings.LastIndex(data, "_")
	if idx <= 0 || idx == len(data)-1 {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(userLang(user.ID), "invalid_input")})
	}
	gateway := data[:idx]
	usdStr := data[idx+1:]
	usd, err := strconv.ParseFloat(usdStr, 64)
	if err != nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(userLang(user.ID), "invalid_input")})
	}

	_ = c.Respond()
	return calculateAndConfirm(c, user, usd, gateway)
}

func ChargeCustomHandler(c AsliBot.Context) error {
	user := c.Sender()
	lang := userLang(user.ID)
	data := c.Callback().Data
	data = strings.TrimPrefix(data, "\f")
	data = strings.TrimPrefix(data, "charge_custom_")

	setAwaiting(user.ID, true)
	setChargeGateway(user.ID, data)

	return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "custom_amount_send")})
}

func calculateAndConfirm(c AsliBot.Context, user *AsliBot.User, usd float64, gateway string) error {
	lang := userLang(user.ID)

	if usd <= 0 {
		return c.Send(i18n.T(lang, "invalid_amount"))
	}

	finalRate, err := rate.GetFinalRate(appCfg.DollarRateAPI, appCfg.DollarRateCache, appCfg.DollarRateMargin)
	if err != nil {
		return c.Send(i18n.T(lang, "rate_error"))
	}

	toman := rate.CalculateToman(usd, finalRate)

	if toman < appCfg.CHARGE_MIN {
		return c.Send(fmt.Sprintf(i18n.T(lang, "min_charge"), formatMoney(appCfg.CHARGE_MIN)))
	}
	if toman > appCfg.CHARGE_MAX {
		return c.Send(fmt.Sprintf(i18n.T(lang, "max_charge"), formatMoney(appCfg.CHARGE_MAX)))
	}

	setAwaiting(user.ID, false)
	setAwaitingConfirm(user.ID, &pendingInvoice{
		AmountUSD:   usd,
		AmountToman: toman,
		Gateway:     gateway,
	})

	inline := &AsliBot.ReplyMarkup{}
	btnConfirm := inline.Data(i18n.T(lang, "confirm_btn"), "charge_confirm")
	btnCancel := inline.Data(i18n.T(lang, "cancel_btn"), "charge_cancel_confirm")
	inline.Inline(inline.Row(btnConfirm, btnCancel))

	return c.Send(fmt.Sprintf(
		i18n.T(lang, "confirm_charge"),
		usd, formatMoney(finalRate), formatMoney(toman), gatewayLabel(gateway),
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func ChargeConfirmHandler(c AsliBot.Context) error {
	user := c.Sender()
	lang := userLang(user.ID)
	inv := getAwaitingConfirm(user.ID)
	if inv == nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "invalid_input")})
	}
	setAwaitingConfirm(user.ID, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	orderID := strconv.FormatInt(user.ID, 10) + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	baseURL := appCfg.WEBHOOK_BASE_URL
	if baseURL == "" {
		baseURL = "http://localhost:" + appCfg.WEBHOOK_PORT
	}
	callbackURL := baseURL + "/webhook/abangateway?order_id=" + orderID

	switch inv.Gateway {
	case "tetraminator":
		return chargeViaTetra(c, user.ID, inv, ctx, callbackURL)
	case "cubepay":
		return chargeViaCube(c, user.ID, inv, ctx, orderID, callbackURL)
	case "abangateway":
		return chargeViaAban(c, user.ID, inv, ctx, orderID, callbackURL)
	case "manual_card":
		return chargeViaManualCard(c, user.ID, inv)
	}
	return c.Send(i18n.T(lang, "error"))
}

func chargeViaTetra(c AsliBot.Context, uid int64, inv *pendingInvoice, ctx context.Context, callbackURL string) error {
	lang := userLang(uid)
	if tetraClient == nil {
		return c.Send(i18n.T(lang, "error"))
	}

	invoice, err := tetraClient.CreateInvoice(ctx, inv.AmountToman, callbackURL)
	if err != nil {
		return c.Send(i18n.T(lang, "error") + ": " + err.Error())
	}

	expiresAt := time.Now().Add(chargeTimeout)
	if err := walletDB.CreateCharge(uid, invoice.PayID, inv.AmountUSD, inv.AmountToman, expiresAt); err != nil {
		return c.Send(i18n.T(lang, "error") + ": " + err.Error())
	}
	walletDB.SetTransactionGateway(uid, invoice.PayID, "tetraminator")

	payMu.Lock()
	pendingPayments[uid] = invoice.PayID
	payMu.Unlock()

	inline := &AsliBot.ReplyMarkup{}
	btnPay := inline.URL(i18n.T(lang, "pay_btn"), invoice.PaymentLink)
	btnCheck := inline.Data(i18n.T(lang, "paid_btn"), "wallet_check")
	btnCancel := inline.Data(i18n.T(lang, "cancel_btn"), "wallet_cancel")
	inline.Inline(inline.Row(btnPay), inline.Row(btnCheck, btnCancel))

	return c.Send(fmt.Sprintf(
		i18n.T(lang, "invoice_created"),
		gatewayLabel("tetraminator"),
		inv.AmountUSD, formatMoney(inv.AmountToman), invoice.PayID, expiresAt.Format("15:04"),
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func chargeViaCube(c AsliBot.Context, uid int64, inv *pendingInvoice, ctx context.Context, orderID, callbackURL string) error {
	lang := userLang(uid)
	g := walletDB.GetPaymentGateways()
	if g.CubePayToken == "" {
		return c.Send(i18n.T(lang, "error"))
	}

	resp, err := getCubeClient().CreatePayment(ctx, orderID, inv.AmountToman, callbackURL, "شارژ حساب")
	if err != nil {
		return c.Send(i18n.T(lang, "error") + ": " + err.Error())
	}

	expiresAt := time.Now().Add(chargeTimeout)
	if err := walletDB.CreateCharge(uid, resp.Authority, inv.AmountUSD, inv.AmountToman, expiresAt); err != nil {
		return c.Send(i18n.T(lang, "error") + ": " + err.Error())
	}
	walletDB.SetTransactionGateway(uid, resp.Authority, "cubepay")

	payMu.Lock()
	pendingPayments[uid] = resp.Authority
	payMu.Unlock()

	inline := &AsliBot.ReplyMarkup{}
	btnPay := inline.URL(i18n.T(lang, "pay_btn"), resp.PaymentLink)
	btnCheck := inline.Data(i18n.T(lang, "paid_btn"), "wallet_check")
	btnCancel := inline.Data(i18n.T(lang, "cancel_btn"), "wallet_cancel")
	inline.Inline(inline.Row(btnPay), inline.Row(btnCheck, btnCancel))

	return c.Send(fmt.Sprintf(
		i18n.T(lang, "invoice_created"),
		gatewayLabel("cubepay"),
		inv.AmountUSD, formatMoney(inv.AmountToman), resp.Authority, expiresAt.Format("15:04"),
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func chargeViaAban(c AsliBot.Context, uid int64, inv *pendingInvoice, ctx context.Context, orderID, callbackURL string) error {
	lang := userLang(uid)
	g := walletDB.GetPaymentGateways()
	if g.AbanAPIKey == "" {
		return c.Send(i18n.T(lang, "error"))
	}

	cli := payment.NewAbanClient(g.AbanAPIKey)
	amountRial := inv.AmountToman * 10

	invoice, err := cli.CreateInvoice(ctx, orderID, amountRial, callbackURL)
	if err != nil {
		return c.Send(i18n.T(lang, "error") + ": " + err.Error())
	}

	expiresAt := time.Now().Add(chargeTimeout)
	if err := walletDB.CreateCharge(uid, invoice.PayID, inv.AmountUSD, inv.AmountToman, expiresAt); err != nil {
		return c.Send(i18n.T(lang, "error") + ": " + err.Error())
	}
	walletDB.SetTransactionGateway(uid, invoice.PayID, "abangateway")

	payMu.Lock()
	pendingPayments[uid] = invoice.PayID
	payMu.Unlock()

	inline := &AsliBot.ReplyMarkup{}
	btnPay := inline.URL(i18n.T(lang, "pay_btn"), invoice.PaymentLink)
	btnCheck := inline.Data(i18n.T(lang, "paid_btn"), "wallet_check")
	btnCancel := inline.Data(i18n.T(lang, "cancel_btn"), "wallet_cancel")
	inline.Inline(inline.Row(btnPay), inline.Row(btnCheck, btnCancel))

	return c.Send(fmt.Sprintf(
		i18n.T(lang, "invoice_created"),
		gatewayLabel("abangateway"),
		inv.AmountUSD, formatMoney(inv.AmountToman), invoice.PayID, expiresAt.Format("15:04"),
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func chargeViaManualCard(c AsliBot.Context, uid int64, inv *pendingInvoice) error {
	lang := userLang(uid)
	g := walletDB.GetPaymentGateways()
	if g.ManualCardNumber == "" {
		return c.Send(i18n.T(lang, "manual_card_no_number"))
	}

	req := &database.ManualChargeRequest{
		RequesterID:   uid,
		RequesterName: c.Sender().FirstName,
		TargetUserID:  uid,
		TargetName:    c.Sender().FirstName,
		AmountUSD:     inv.AmountUSD,
		AmountToman:   inv.AmountToman,
		Gateway:       "manual_card",
		PayID:         strconv.FormatInt(time.Now().UnixNano(), 10),
	}
	if err := walletDB.CreateManualChargeRequest(req); err != nil {
		return c.Send(i18n.T(lang, "error") + ": " + err.Error())
	}

	payMu.Lock()
	pendingPayments[uid] = req.PayID
	payMu.Unlock()

	inline := &AsliBot.ReplyMarkup{}
	btnSent := inline.Data("✅ پرداخت کردم (شماره پیگیری می‌فرستم)", "wallet_manual_ref")
	btnCancel := inline.Data(i18n.T(lang, "cancel_btn"), "wallet_cancel")
	inline.Inline(inline.Row(btnSent), inline.Row(btnCancel))

	return c.Send(fmt.Sprintf(
		i18n.T(lang, "manual_card_amount"),
		inv.AmountUSD, formatMoney(inv.AmountToman),
		g.ManualCardNumber, g.ManualCardHolder,
	), &AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
}

func ChargeCancelConfirmHandler(c AsliBot.Context) error {
	setAwaitingConfirm(c.Sender().ID, nil)
	_ = c.Delete()
	return c.Send(i18n.T(userLang(c.Sender().ID), "canceled"), menus.MainMenu)
}

func ChargeAmountHandler(c AsliBot.Context) error {
	user := c.Sender()

	if !isAwaiting(user.ID) {
		return nil
	}

	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	text = strings.Replace(text, ",", ".", 1)
	usd, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return c.Send(i18n.T(userLang(user.ID), "invalid_amount"))
	}

	gateway := getChargeGateway(user.ID)
	setChargeGateway(user.ID, "")
	setAwaiting(user.ID, false)

	return calculateAndConfirm(c, user, usd, gateway)
}

func CheckPaymentHandler(c AsliBot.Context) error {
	user := c.Sender()
	lang := userLang(user.ID)
	log.Printf("🔘 پرداخت کردم: user=%d", user.ID)

	payMu.Lock()
	payID, ok := pendingPayments[user.ID]
	payMu.Unlock()

	if !ok {
		tx, err := walletDB.GetPendingCharge(user.ID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "no_pending")})
		}
		payID = tx.PayID
	}

	tx, err := walletDB.GetPendingCharge(user.ID)
	if err != nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "no_pending")})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	switch tx.Gateway {
	case "cubepay":
		status, err := getCubeClient().CheckStatus(ctx, payID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
		}
		if status.Status != "paid" {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "payment_pending")})
		}
		if _, err := getCubeClient().VerifyPayment(ctx, payID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
		}
		return finalizeCharge(c, user.ID, payID)

	case "abangateway":
		g := walletDB.GetPaymentGateways()
		if g.AbanAPIKey == "" {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error")})
		}
		cli := payment.NewAbanClient(g.AbanAPIKey)
		status, err := cli.Inquiry(ctx, payID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
		}
		if status != payment.StatusPaid {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "payment_pending")})
		}
		if err := cli.Verify(ctx, payID); err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
		}
		return finalizeCharge(c, user.ID, payID)

	case "tetraminator", "":
		if tetraClient == nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error")})
		}
		inquiry, err := tetraClient.Inquiry(ctx, payID)
		if err != nil {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
		}
		if !inquiry.Status || inquiry.PaymentStatus != "paid" {
			return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "payment_pending")})
		}
		return finalizeCharge(c, user.ID, payID)

	default:
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error")})
	}
}

func finalizeCharge(c AsliBot.Context, uid int64, payID string) error {
	lang := userLang(uid)
	_, newBalance, err := walletDB.ConfirmCharge(payID)
	if err != nil {
		return c.Respond(&AsliBot.CallbackResponse{Text: i18n.T(lang, "error") + ": " + err.Error()})
	}

	payMu.Lock()
	delete(pendingPayments, uid)
	payMu.Unlock()

	return c.Send(fmt.Sprintf(i18n.T(lang, "payment_confirmed"), newBalance),
		&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
}

func CancelPaymentHandler(c AsliBot.Context) error {
	user := c.Sender()
	lang := userLang(user.ID)
	walletDB.CancelPending(user.ID)

	payMu.Lock()
	delete(pendingPayments, user.ID)
	payMu.Unlock()

	setAwaiting(user.ID, false)
	setAwaitingConfirm(user.ID, nil)

	_ = c.Delete()
	return c.Send(i18n.T(lang, "payment_canceled"), menus.MainMenu)
}

func CancelCommandHandler(c AsliBot.Context) error {
	uid := c.Sender().ID
	lang := userLang(uid)
	setAwaiting(uid, false)
	setAwaitingConfirm(uid, nil)
	setChargeGateway(uid, "")
	return c.Send(i18n.T(lang, "canceled"), menus.MainMenu)
}

func formatTransaction(t database.Transaction, lang i18n.Lang) string {
	createdAt, _ := time.Parse(time.RFC3339, t.CreatedAt)
	dateStr := createdAt.Format("2006/01/02 15:04")

	var icon, titleKey string
	switch {
	case t.Type == "charge" && t.Status == "paid":
		icon, titleKey = "✅", "tx_charge_paid"
	case t.Type == "charge" && t.Status == "pending":
		icon, titleKey = "⏳", "tx_charge_pending"
	case t.Type == "charge" && t.Status == "expired":
		icon, titleKey = "⌛", "tx_charge_expired"
	case t.Type == "charge" && t.Status == "canceled":
		icon, titleKey = "❌", "tx_charge_canceled"
	case t.Type == "manual_charge":
		icon, titleKey = "🎁", "tx_manual_charge"
	case t.Type == "spend":
		icon, titleKey = "💸", "tx_spend"
	case t.Type == "deduct":
		icon, titleKey = "➖", "tx_deduct"
	case t.Type == "purchase":
		icon, titleKey = "🛒", "tx_purchase"
	case t.Type == "refund":
		icon, titleKey = "🔄", "tx_refund"
	case t.Type == "renewal":
		icon, titleKey = "🔁", "tx_renewal"
	default:
		icon, titleKey = "❓", t.Status
	}

	gw := ""
	if t.Gateway != "" {
		gw = " | " + gatewayLabel(t.Gateway)
	}

	// %+.2f renders the sign itself: "+1.50" for credits, "-1.50" for debits.
	// Passing the float directly keeps the argument type-safe for go vet.
	return fmt.Sprintf(
		"%s %s | %+.2f دلار%s\n   🕐 %s\n",
		icon, i18n.T(lang, titleKey),
		t.AmountUSD, gw, dateStr,
	)
}

func gatewayLabel(g string) string {
	switch g {
	case "tetraminator":
		return "💎 Tetraminator"
	case "cubepay":
		return "🧊 CubePay"
	case "abangateway":
		return "🌐 AbanGateway"
	case "manual_card":
		return "💳 کارت دستی"
	case "manual_charge":
		return "🎁 شارژ دستی"
	}
	return "—"
}

func formatMoney(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var out []byte
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(ch))
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

func formatMoneyFa(n int64) string {
	s := formatMoney(n)
	replacer := strings.NewReplacer(
		"0", "۰", "1", "۱", "2", "۲", "3", "۳", "4", "۴",
		"5", "۵", "6", "۶", "7", "۷", "8", "۸", "9", "۹",
	)
	return replacer.Replace(s)
}
