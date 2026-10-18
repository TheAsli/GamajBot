// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package handlers

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/notify"

	AsliBot "gopkg.in/telebot.v4"
)

var (
	cardRefAwaiting = make(map[int64]string)
	cardRefMu       sync.Mutex
)

func setCardRefAwaiting(uid int64, payID string) {
	cardRefMu.Lock()
	defer cardRefMu.Unlock()
	if payID == "" {
		delete(cardRefAwaiting, uid)
	} else {
		cardRefAwaiting[uid] = payID
	}
}

func getCardRefAwaiting(uid int64) string {
	cardRefMu.Lock()
	defer cardRefMu.Unlock()
	return cardRefAwaiting[uid]
}

func CardChargeRefHandler(c AsliBot.Context) error {
	uid := c.Sender().ID
	payMu.Lock()
	payID, ok := pendingPayments[uid]
	payMu.Unlock()
	if !ok {
		return c.Respond(&AsliBot.CallbackResponse{Text: "❌ درخواستی در انتظار نیست."})
	}
	setCardRefAwaiting(uid, payID)
	return c.Respond(&AsliBot.CallbackResponse{
		Text: "🖼 رسید (عکس یا فایل) + شماره پیگیری را بفرست.",
	})
}

func CardChargeRefInputHandler(c AsliBot.Context, cfg *config.Config, db *database.DB) error {
	uid := c.Sender().ID
	payID := getCardRefAwaiting(uid)
	if payID == "" {
		return nil
	}

	msg := c.Message()
	if msg == nil {
		return nil
	}

	var fileID, receiptType, caption string
	switch {
	case msg.Photo != nil:
		fileID = msg.Photo.File.FileID
		receiptType = "photo"
		caption = strings.TrimSpace(msg.Caption)
	case msg.Document != nil:
		fileID = msg.Document.File.FileID
		receiptType = "document"
		caption = strings.TrimSpace(msg.Caption)
	default:
		return c.Send("❌ لطفاً رسید را به‌صورت عکس یا فایل بفرست (نه متن).")
	}

	setCardRefAwaiting(uid, "")

	req := db.FindManualChargeByPayID(payID)
	if req == nil {
		return c.Send("❌ درخواست پیدا نشد.")
	}

	req.TxRef = caption
	_ = db.UpdateManualCharge(req)
	_ = db.SetReceipt(req.ID, fileID, receiptType)

	sendReceiptToAdmins(c, cfg, db, req)

	return c.Send("✅ رسید دریافت شد و برای بررسی به ادمین‌ها ارسال شد.")
}

func sendReceiptToAdmins(c AsliBot.Context, cfg *config.Config, db *database.DB, req *database.ManualChargeRequest) {
	inlinePrivate := &AsliBot.ReplyMarkup{}
	btnApprove := inlinePrivate.Data("✅ تایید", "adm_mc_approve|"+req.ID)
	btnReject := inlinePrivate.Data("❌ رد", "adm_mc_reject|"+req.ID)
	btnReview := inlinePrivate.Data("🔍 نیاز به بررسی", "adm_mc_review|"+req.ID)
	btnBlock := inlinePrivate.Data("🚫 رد و بلاک", "adm_mc_block|"+req.ID)
	inlinePrivate.Inline(
		inlinePrivate.Row(btnApprove, btnReject),
		inlinePrivate.Row(btnReview),
		inlinePrivate.Row(btnBlock),
	)

	inlineLog := &AsliBot.ReplyMarkup{}
	btnApproveLog := inlineLog.Data("✅ تایید", "adm_mc_approve|"+req.ID)
	btnRejectLog := inlineLog.Data("❌ رد", "adm_mc_reject|"+req.ID)
	btnReviewLog := inlineLog.Data("🔍 نیاز به بررسی", "adm_mc_review|"+req.ID)
	btnBlockLog := inlineLog.Data("🚫 رد و بلاک", "adm_mc_block|"+req.ID)
	inlineLog.Inline(
		inlineLog.Row(btnApproveLog, btnRejectLog),
		inlineLog.Row(btnReviewLog),
		inlineLog.Row(btnBlockLog),
	)

	msgText := fmt.Sprintf(
		"💳 <b>رسید کارت‌به‌کارت جدید</b>\n\n"+
			"👤 کاربر: %s (ID: <code>%d</code>)\n"+
			"💵 مبلغ: <b>%.2f دلار</b> (%s تومان)\n"+
			"🔢 شماره پیگیری: <code>%s</code>\n"+
			"🆔 شناسه: <code>%s</code>",
		req.TargetName, req.TargetUserID, req.AmountUSD,
		formatMoney(req.AmountToman), req.TxRef, req.ID,
	)

	photo := &AsliBot.Photo{
		File:    AsliBot.File{FileID: req.ReceiptID},
		Caption: msgText,
	}

	if botInstance != nil {
		notify.SendWithLogPhoto(botInstance, cfg, db, notify.LogTopicFinance, photo, inlinePrivate, inlineLog)
	}

	for _, a := range db.GetAllAdmins() {
		if a.UserID == cfg.ADMIN_ID || a.Role == database.RoleViewer {
			continue
		}
		_, _ = c.Bot().Send(AsliBot.ChatID(a.UserID), photo,
			&AsliBot.SendOptions{ParseMode: AsliBot.ModeHTML}, inlinePrivate)
	}
}

func bytesReaderWrap(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}
