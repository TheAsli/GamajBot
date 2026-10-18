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
	"github.com/theasli/GamajBot/internal/notify"

	AsliBot "gopkg.in/telebot.v4"
)

type adminChargeState struct {
	Step         string
	TargetUserID int64
	TargetName   string
	AmountUSD    float64
}

var (
	adminChargeStates = make(map[int64]*adminChargeState)
	adminChargeMu     sync.Mutex
)

func setAdminChargeState(uid int64, s *adminChargeState) {
	adminChargeMu.Lock()
	defer adminChargeMu.Unlock()
	if s == nil {
		delete(adminChargeStates, uid)
	} else {
		adminChargeStates[uid] = s
	}
}

func getAdminChargeState(uid int64) *adminChargeState {
	adminChargeMu.Lock()
	defer adminChargeMu.Unlock()
	return adminChargeStates[uid]
}

var (
	awaitingRejectReason = make(map[int64]string)
	rejectMu             sync.Mutex
)

func setAwaitingReject(adminID int64, action string) {
	rejectMu.Lock()
	defer rejectMu.Unlock()
	if action == "" {
		delete(awaitingRejectReason, adminID)
	} else {
		awaitingRejectReason[adminID] = action
	}
}

func getAwaitingReject(adminID int64) string {
	rejectMu.Lock()
	defer rejectMu.Unlock()
	return awaitingRejectReason[adminID]
}

func ManualChargeStartHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		if role == "" || role == database.RoleViewer {
			return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
		}
		setAdminChargeState(c.Sender().ID, &adminChargeState{Step: "awaiting_user"})
		_ = c.Respond(&AsliBot.CallbackResponse{})
		return c.Send("👤 آیدی عددی کاربری که می‌خواهی شارژ کنی را بفرست:")
	}
}

func ManualChargeTextHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		uid := c.Sender().ID

		if action := getAwaitingReject(uid); action != "" {
			parts := strings.SplitN(action, ":", 2)
			if len(parts) != 2 {
				setAwaitingReject(uid, "")
				return nil
			}
			kind, reqID := parts[0], parts[1]
			text := strings.TrimSpace(c.Text())
			if text == "" || strings.HasPrefix(text, "/") {
				return nil
			}
			setAwaitingReject(uid, "")

			switch kind {
			case "reject":
				req, err := db.RejectManualCharge(reqID, text)
				if err != nil {
					return c.Send("❌ " + err.Error())
				}
				notifyAdminResult(c, cfg, db, req, "rejected")
				notifyUserStatus(c, req, "rejected")

				admin := c.Sender()
				logActionTarget(botInstance, cfg, db, admin, "رد شارژ دستی", req.TargetName, req.TargetUserID,
					fmt.Sprintf("دلیل: %s", text))

				return c.Send("✅ رد شد و به کاربر اطلاع داده شد.")
			case "review":
				req, err := db.MarkManualChargeNeedsReview(reqID, text)
				if err != nil {
					return c.Send("❌ " + err.Error())
				}
				notifyAdminResult(c, cfg, db, req, "needs_review")
				sendToSuperAdminReview(c, cfg, db, req)

				admin := c.Sender()
				logActionTarget(botInstance, cfg, db, admin, "ارجاع به مدیر", req.TargetName, req.TargetUserID,
					fmt.Sprintf("یادداشت: %s", text))

				return c.Send("✅ برای مدیر اصلی ارسال شد.")
			case "block":
				req, err := db.BlockRejectManualCharge(reqID, text)
				if err != nil {
					return c.Send("❌ " + err.Error())
				}
				_ = db.BanUser(req.TargetUserID, "رسید جعلی: "+text)
				notifyAdminResult(c, cfg, db, req, "blocked")
				notifyUserStatus(c, req, "blocked")

				admin := c.Sender()
				logActionTarget(botInstance, cfg, db, admin, "بلاک کاربر بابت رسید جعلی", req.TargetName, req.TargetUserID,
					fmt.Sprintf("دلیل: %s", text))

				return c.Send("✅ رد و کاربر بلاک شد.")
			}
			return nil
		}

		st := getAdminChargeState(uid)
		if st == nil {
			return nil
		}
		text := strings.TrimSpace(c.Text())
		if strings.HasPrefix(text, "/") {
			return nil
		}

		switch st.Step {
		case "awaiting_user":
			targetID, err := strconv.ParseInt(text, 10, 64)
			if err != nil {
				return c.Send("❌ آیدی نامعتبر. یک عدد بفرست:")
			}
			target, err := db.GetUser(targetID)
			if err != nil {
				return c.Send("❌ کاربر پیدا نشد. دوباره بفرست:")
			}
			st.TargetUserID = targetID
			st.TargetName = target.FirstName
			st.Step = "awaiting_amount"
			return c.Send(fmt.Sprintf("👤 کاربر: %s (ID: %d)\n\n💵 مبلغ دلاری را بفرست:", target.FirstName, targetID))

		case "awaiting_amount":
			text = strings.Replace(text, ",", ".", 1)
			usd, err := strconv.ParseFloat(text, 64)
			if err != nil || usd <= 0 {
				return c.Send("❌ مبلغ نامعتبر. یک عدد مثبت بفرست:")
			}
			st.AmountUSD = usd
			setAdminChargeState(uid, nil)

			requester := c.Sender()
			req := &database.ManualChargeRequest{
				RequesterID:   requester.ID,
				RequesterName: requester.FirstName,
				TargetUserID:  st.TargetUserID,
				TargetName:    st.TargetName,
				AmountUSD:     usd,
				Gateway:       "manual_charge",
			}
			if err := db.CreateManualChargeRequest(req); err != nil {
				return c.Send("❌ " + err.Error())
			}
			if err := sendToSuperAdmin(c, cfg, db, req); err != nil {
				return c.Send("❌ " + err.Error())
			}
			return c.Send("✅ درخواست ثبت و به مدیر اصلی ارسال شد.")
		}
		return nil
	}
}

func sendToSuperAdmin(c AsliBot.Context, cfg *config.Config, db *database.DB, req *database.ManualChargeRequest) error {
	inlinePrivate := &AsliBot.ReplyMarkup{}
	btnApprove := inlinePrivate.Data("✅ تایید", "adm_mc_approve|"+req.ID)
	btnReject := inlinePrivate.Data("❌ رد", "adm_mc_reject|"+req.ID)
	inlinePrivate.Inline(inlinePrivate.Row(btnApprove, btnReject))

	inlineLog := &AsliBot.ReplyMarkup{}
	btnApproveLog := inlineLog.Data("✅ تایید", "adm_mc_approve|"+req.ID)
	btnRejectLog := inlineLog.Data("❌ رد", "adm_mc_reject|"+req.ID)
	inlineLog.Inline(inlineLog.Row(btnApproveLog, btnRejectLog))

	msg := fmt.Sprintf(
		"🔔 <b>درخواست شارژ دستی</b>\n\n"+
			"👮 ادمین: %s (ID: <code>%d</code>)\n"+
			"👤 کاربر: %s (ID: <code>%d</code>)\n"+
			"💵 مبلغ: <b>%.2f دلار</b>\n"+
			"🆔 شناسه: <code>%s</code>",
		req.RequesterName, req.RequesterID,
		req.TargetName, req.TargetUserID,
		req.AmountUSD, req.ID,
	)

	if botInstance != nil {
		notify.SendWithLog(botInstance, cfg, db, notify.LogTopicFinance, msg, inlinePrivate, inlineLog)
	}
	return nil
}

func sendToSuperAdminReview(c AsliBot.Context, cfg *config.Config, db *database.DB, req *database.ManualChargeRequest) {
	inlinePrivate := &AsliBot.ReplyMarkup{}
	btnApprove := inlinePrivate.Data("✅ تایید نهایی", "adm_mc_approve|"+req.ID)
	btnReject := inlinePrivate.Data("❌ رد نهایی", "adm_mc_reject|"+req.ID)
	btnBlock := inlinePrivate.Data("🚫 رد و بلاک", "adm_mc_block|"+req.ID)
	inlinePrivate.Inline(inlinePrivate.Row(btnApprove, btnReject), inlinePrivate.Row(btnBlock))

	inlineLog := &AsliBot.ReplyMarkup{}
	btnApproveLog := inlineLog.Data("✅ تایید نهایی", "adm_mc_approve|"+req.ID)
	btnRejectLog := inlineLog.Data("❌ رد نهایی", "adm_mc_reject|"+req.ID)
	btnBlockLog := inlineLog.Data("🚫 رد و بلاک", "adm_mc_block|"+req.ID)
	inlineLog.Inline(inlineLog.Row(btnApproveLog, btnRejectLog), inlineLog.Row(btnBlockLog))

	msg := fmt.Sprintf(
		"🔍 <b>درخواست بررسی نهایی</b>\n\n"+
			"👤 کاربر: %s (ID: <code>%d</code>)\n"+
			"💵 مبلغ: <b>%.2f دلار</b>\n"+
			"📝 یادداشت ادمین: %s\n"+
			"🆔 شناسه: <code>%s</code>",
		req.TargetName, req.TargetUserID, req.AmountUSD, req.ReviewNote, req.ID,
	)

	if botInstance != nil {
		notify.SendWithLog(botInstance, cfg, db, notify.LogTopicFinance, msg, inlinePrivate, inlineLog)
	}
}

func ManualChargeCallbackHandler(cfg *config.Config, db *database.DB) func(AsliBot.Context) error {
	return func(c AsliBot.Context) error {
		isSuper := c.Sender().ID == cfg.ADMIN_ID
		role := db.GetAdminRole(c.Sender().ID, cfg.ADMIN_ID)
		data := c.Callback().Data
		data = strings.TrimPrefix(data, "\f")

		switch {
		case strings.HasPrefix(data, "adm_mc_approve|"):
			reqID := strings.TrimPrefix(data, "adm_mc_approve|")
			req, _ := db.GetManualChargeRequest(reqID)
			if req == nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ درخواست پیدا نشد"})
			}
			if !isSuper && req.Status != "pending" {
				return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ فقط مدیر می‌تواند این را تایید کند"})
			}
			req2, newBalance, err := db.ApproveManualCharge(reqID)
			if err != nil {
				return c.Respond(&AsliBot.CallbackResponse{Text: "❌ " + err.Error()})
			}
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "✅ تایید شد"})
			_ = c.Edit(fmt.Sprintf(
				"✅ تایید شد.\n👤 %s\n💵 %.2f دلار\n💰 موجودی جدید: %.2f",
				req2.TargetName, req2.AmountUSD, newBalance,
			))
			notifyAdminResult(c, cfg, db, req2, "approved")
			notifyUserCharged(c, req2, newBalance)

			admin := c.Sender()
			logActionTarget(botInstance, cfg, db, admin, "تایید شارژ دستی", req2.TargetName, req2.TargetUserID,
				fmt.Sprintf("مبلغ: %.2f دلار — موجودی جدید: %.2f", req2.AmountUSD, newBalance))

			return nil

		case strings.HasPrefix(data, "adm_mc_reject|"):
			if role == "" || role == database.RoleViewer {
				return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
			}
			reqID := strings.TrimPrefix(data, "adm_mc_reject|")
			setAwaitingReject(c.Sender().ID, "reject:"+reqID)
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "دلیل رد را بنویس"})
			return c.Send("✍️ دلیل رد را بفرست:")

		case strings.HasPrefix(data, "adm_mc_review|"):
			if isSuper {
				return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ این گزینه برای ادمین‌هاست"})
			}
			if role == "" || role == database.RoleViewer {
				return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
			}
			reqID := strings.TrimPrefix(data, "adm_mc_review|")
			setAwaitingReject(c.Sender().ID, "review:"+reqID)
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "یادداشت بررسی را بنویس"})
			return c.Send("📝 یادداشت بررسی را بنویس:")

		case strings.HasPrefix(data, "adm_mc_block|"):
			if role == "" || role == database.RoleViewer {
				return c.Respond(&AsliBot.CallbackResponse{Text: "⛔ دسترسی نداری"})
			}
			reqID := strings.TrimPrefix(data, "adm_mc_block|")
			setAwaitingReject(c.Sender().ID, "block:"+reqID)
			_ = c.Respond(&AsliBot.CallbackResponse{Text: "دلیل بلاک را بنویس"})
			return c.Send("🚫 دلیل رد و بلاک را بنویس:")
		}
		return c.Respond(&AsliBot.CallbackResponse{})
	}
}

func notifyAdminResult(c AsliBot.Context, cfg *config.Config, db *database.DB, req *database.ManualChargeRequest, status string) {
	var title string
	switch status {
	case "approved":
		title = "✅ تایید شد"
	case "rejected":
		title = "❌ رد شد"
	case "needs_review":
		title = "🔍 نیاز به بررسی بیشتر"
	case "blocked":
		title = "🚫 رد و بلاک"
	}
	msg := fmt.Sprintf("%s\n\n👤 %s\n💵 %.2f دلار\n🆔 %s",
		title, req.TargetName, req.AmountUSD, req.ID)

	if botInstance != nil {
		_, _ = botInstance.Send(AsliBot.ChatID(req.RequesterID), msg,
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
		notify.AdminLog(botInstance, cfg, db, notify.LogTopicFinance, msg)
	}
}

func notifyUserCharged(c AsliBot.Context, req *database.ManualChargeRequest, newBalance float64) {
	msg := fmt.Sprintf(
		"🎁 <b>حساب شما شارژ شد!</b>\n\n"+
			"💵 مبلغ: <b>%.2f دلار</b>\n"+
			"💰 موجودی جدید: <b>%.2f دلار</b>\n\n"+
			"با تشکر 🌹",
		req.AmountUSD, newBalance,
	)
	_, _ = c.Bot().Send(AsliBot.ChatID(req.TargetUserID), msg,
		&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML})
}

func notifyUserStatus(c AsliBot.Context, req *database.ManualChargeRequest, status string) {
	var msg string
	switch status {
	case "rejected":
		msg = fmt.Sprintf("❌ رسید شما رد شد.\n📝 دلیل: %s", req.RejectReason)
	case "blocked":
		msg = fmt.Sprintf("🚫 رسید شما جعلی تشخیص داده شد و حساب شما بلاک شد.\n📝 دلیل: %s", req.RejectReason)
	}
	_, _ = c.Bot().Send(AsliBot.ChatID(req.TargetUserID), msg)
}
