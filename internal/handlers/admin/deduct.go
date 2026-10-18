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

type deductState struct {
	Step         string
	TargetUserID int64
	TargetName   string
	AmountUSD    float64
}

var (
	deductStates = make(map[int64]*deductState)
	deductMu     sync.Mutex
)

func setDeductState(uid int64, s *deductState) {
	deductMu.Lock()
	defer deductMu.Unlock()
	if s == nil {
		delete(deductStates, uid)
	} else {
		deductStates[uid] = s
	}
}

func getDeductState(uid int64) *deductState {
	deductMu.Lock()
	defer deductMu.Unlock()
	return deductStates[uid]
}

func DeductStartHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role != database.RoleSuperAdmin && role != database.RoleFinance {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		setDeductState(c.Sender().ID, &deductState{Step: "awaiting_user"})
		_ = c.Respond()
		return c.Send("👤 آیدی عددی کاربری که می‌خواهی از او کسر کنی را بفرست:")
	}
}

func DeductInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	st := getDeductState(uid)
	if st == nil {
		return nil
	}
	text := strings.TrimSpace(c.Text())
	if strings.HasPrefix(text, "/") {
		return nil
	}

	switch st.Step {
	case "awaiting_user":
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return c.Send("❌ آیدی نامعتبر. یک عدد بفرست:")
		}
		target, err := db.GetUser(id)
		if err != nil {
			return c.Send("❌ کاربر پیدا نشد. دوباره بفرست:")
		}
		st.TargetUserID = id
		st.TargetName = target.FirstName
		st.Step = "awaiting_amount"
		return c.Send(fmt.Sprintf(
			"👤 کاربر: %s (ID: %d)\n💰 موجودی: %.2f دلار\n\n💵 مبلغ دلاری برای کسر:",
			target.FirstName, id, target.BalanceUSD,
		))

	case "awaiting_amount":
		text = strings.Replace(text, ",", ".", 1)
		usd, err := strconv.ParseFloat(text, 64)
		if err != nil || usd <= 0 {
			return c.Send("❌ مبلغ نامعتبر. یک عدد مثبت بفرست:")
		}
		st.AmountUSD = usd
		st.Step = "awaiting_reason"

		req := &database.ManualChargeRequest{
			RequesterID:   uid,
			RequesterName: c.Sender().FirstName,
			TargetUserID:  st.TargetUserID,
			TargetName:    st.TargetName,
			AmountUSD:     usd,
			Gateway:       "deduct",
		}
		if err := db.CreateManualChargeRequest(req); err != nil {
			return c.Send("❌ " + err.Error())
		}
		setDeductState(uid, nil)

		inline := &AsliBot.ReplyMarkup{}
		btnApprove := inline.Data("✅ تایید کسر", "adm_deduct_approve|"+req.ID)
		btnReject := inline.Data("❌ رد", "adm_deduct_reject|"+req.ID)
		inline.Inline(inline.Row(btnApprove, btnReject))

		msg := fmt.Sprintf(
			"➖ <b>درخواست کسر موجودی</b>\n\n"+
				"👮 ادمین: %s (ID: <code>%d</code>)\n"+
				"👤 کاربر: %s (ID: <code>%d</code>)\n"+
				"💵 مبلغ: <b>%.2f دلار</b>\n"+
				"🆔 شناسه: <code>%s</code>",
			c.Sender().FirstName, uid,
			st.TargetName, st.TargetUserID, usd, req.ID,
		)
		_, _ = c.Bot().Send(AsliBot.ChatID(cfg.ADMIN_ID), msg,
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inline)
		return c.Send("✅ درخواست کسر برای مدیر ارسال شد.")
	}
	return nil
}

func DeductCallbackHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		if c.Sender().ID != cfg.ADMIN_ID {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط مدیر اصلی"})
		}
		data := c.Callback().Data
		data = strings.TrimPrefix(data, "\f")

		switch {
		case strings.HasPrefix(data, "adm_deduct_approve|"):
			reqID := strings.TrimPrefix(data, "adm_deduct_approve|")
			req, err := db.GetManualChargeRequest(reqID)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			if req.Status != "pending" {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ قبلاً بررسی شده"})
			}
			newBalance, err := db.DeductBalance(req.TargetUserID, req.AmountUSD, "کسر توسط ادمین "+req.RequesterName)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			req.Status = "approved"
			_ = db.UpdateManualCharge(req)
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ کسر شد"})
			_ = c.Edit(fmt.Sprintf("✅ کسر انجام شد.\n💰 موجودی جدید کاربر: %.2f دلار", newBalance))

			admin := c.Sender()
			logActionTarget(botInstance, cfg, db, admin, "کسر موجودی", req.TargetName, req.TargetUserID,
				fmt.Sprintf("مبلغ: %.2f دلار — موجودی جدید: %.2f", req.AmountUSD, newBalance))

			_, _ = c.Bot().Send(AsliBot.ChatID(req.TargetUserID), fmt.Sprintf(
				"➖ مبلغ %.2f دلار از حساب شما کسر شد.\n💰 موجودی: %.2f دلار",
				req.AmountUSD, newBalance))
			return nil

		case strings.HasPrefix(data, "adm_deduct_reject|"):
			reqID := strings.TrimPrefix(data, "adm_deduct_reject|")
			req, err := db.GetManualChargeRequest(reqID)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			if req.Status != "pending" {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ قبلاً بررسی شده"})
			}
			req.Status = "rejected"
			_ = db.UpdateManualCharge(req)
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "❌ رد شد"})
			_ = c.Edit("❌ درخواست کسر رد شد.")

			admin := c.Sender()
			logActionTarget(botInstance, cfg, db, admin, "رد درخواست کسر", req.TargetName, req.TargetUserID,
				fmt.Sprintf("مبلغ: %.2f دلار", req.AmountUSD))

			_, _ = c.Bot().Send(AsliBot.ChatID(req.RequesterID), "❌ درخواست کسر شما رد شد.")
			return nil
		}
		return c.Respond(&AsliBot.CallbackResponse{})
	}
}
