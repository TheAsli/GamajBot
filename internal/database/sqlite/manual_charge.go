// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrManualChargeNotFound = errors.New("درخواست شارژ پیدا نشد")
	ErrManualChargeDone     = errors.New("این درخواست قبلاً بررسی شده")
)

type ManualChargeRequest struct {
	ID            string
	RequesterID   int64
	RequesterName string
	TargetUserID  int64
	TargetName    string
	AmountUSD     float64
	AmountToman   int64
	Status        string
	RejectReason  string
	ReviewNote    string
	Gateway       string
	PayID         string
	TxRef         string
	ReceiptID     string
	ReceiptType   string
	CreatedAt     string
	ReviewedAt    string
}

func (d *DB) CreateManualChargeRequest(req *ManualChargeRequest) error {
	if req.ID == "" {
		req.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	req.Status = "pending"
	req.CreatedAt = time.Now().Format(time.RFC3339)

	_, err := d.Exec(
		`INSERT INTO manual_charge_requests (id, requester_id, requester_name, target_user_id, target_name, amount_usd, amount_toman, status, reject_reason, review_note, gateway, pay_id, tx_ref, receipt_id, receipt_type, created_at, reviewed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', '', '', ?, ?, ?, ?, ?, ?, '')`,
		req.ID, req.RequesterID, req.RequesterName, req.TargetUserID, req.TargetName,
		req.AmountUSD, req.AmountToman, req.Gateway, req.PayID, req.TxRef, req.ReceiptID, req.ReceiptType, req.CreatedAt,
	)
	return err
}

func (d *DB) GetManualChargeRequest(id string) (*ManualChargeRequest, error) {
	var r ManualChargeRequest
	err := d.QueryRow(
		`SELECT id, requester_id, requester_name, target_user_id, target_name, amount_usd, amount_toman, status, reject_reason, review_note, gateway, pay_id, tx_ref, receipt_id, receipt_type, created_at, reviewed_at
		 FROM manual_charge_requests WHERE id = ?`, id,
	).Scan(&r.ID, &r.RequesterID, &r.RequesterName, &r.TargetUserID, &r.TargetName, &r.AmountUSD, &r.AmountToman, &r.Status, &r.RejectReason, &r.ReviewNote, &r.Gateway, &r.PayID, &r.TxRef, &r.ReceiptID, &r.ReceiptType, &r.CreatedAt, &r.ReviewedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrManualChargeNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) ApproveManualCharge(id string) (*ManualChargeRequest, float64, error) {
	tx, err := d.Begin()
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var r ManualChargeRequest
	err = tx.QueryRow(
		`SELECT id, requester_id, requester_name, target_user_id, target_name, amount_usd, amount_toman, status, reject_reason, review_note, gateway, pay_id, tx_ref, receipt_id, receipt_type, created_at, reviewed_at
		 FROM manual_charge_requests WHERE id = ?`, id,
	).Scan(&r.ID, &r.RequesterID, &r.RequesterName, &r.TargetUserID, &r.TargetName, &r.AmountUSD, &r.AmountToman, &r.Status, &r.RejectReason, &r.ReviewNote, &r.Gateway, &r.PayID, &r.TxRef, &r.ReceiptID, &r.ReceiptType, &r.CreatedAt, &r.ReviewedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, ErrManualChargeNotFound
	}
	if err != nil {
		return nil, 0, err
	}
	if r.Status != "pending" && r.Status != "needs_review" {
		return nil, 0, ErrManualChargeDone
	}

	if _, err := tx.Exec("UPDATE users SET balance_usd = balance_usd + ? WHERE id = ?", r.AmountUSD, r.TargetUserID); err != nil {
		return nil, 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO transactions (user_id, pay_id, amount_usd, amount_toman, type, status, description, gateway, created_at)
		 VALUES (?, ?, ?, ?, 'manual_charge', 'paid', ?, ?, ?)`,
		r.TargetUserID, r.PayID, r.AmountUSD, r.AmountToman,
		"شارژ دستی توسط ادمین "+r.RequesterName, r.Gateway, time.Now().Format(time.RFC3339),
	); err != nil {
		return nil, 0, err
	}
	reviewedAt := time.Now().Format(time.RFC3339)
	if _, err := tx.Exec("UPDATE manual_charge_requests SET status = 'approved', reviewed_at = ? WHERE id = ?", reviewedAt, id); err != nil {
		return nil, 0, err
	}

	var newBal float64
	_ = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", r.TargetUserID).Scan(&newBal)

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	r.Status = "approved"
	r.ReviewedAt = reviewedAt
	return &r, newBal, nil
}

func (d *DB) RejectManualCharge(id, reason string) (*ManualChargeRequest, error) {
	return d.rejectWithStatus(id, reason, "rejected")
}

func (d *DB) BlockRejectManualCharge(id, reason string) (*ManualChargeRequest, error) {
	return d.rejectWithStatus(id, reason, "blocked")
}

func (d *DB) rejectWithStatus(id, reason, status string) (*ManualChargeRequest, error) {
	reviewedAt := time.Now().Format(time.RFC3339)
	res, err := d.Exec(
		"UPDATE manual_charge_requests SET status = ?, reject_reason = ?, reviewed_at = ? WHERE id = ? AND status IN ('pending', 'needs_review')",
		status, reason, reviewedAt, id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrManualChargeDone
	}
	return d.GetManualChargeRequest(id)
}

func (d *DB) MarkManualChargeNeedsReview(id, note string) (*ManualChargeRequest, error) {
	res, err := d.Exec(
		"UPDATE manual_charge_requests SET status = 'needs_review', review_note = ? WHERE id = ? AND status = 'pending'",
		note, id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, ErrManualChargeDone
	}
	return d.GetManualChargeRequest(id)
}

func (d *DB) SetReceipt(id, fileID, receiptType string) error {
	_, err := d.Exec(
		"UPDATE manual_charge_requests SET receipt_id = ?, receipt_type = ? WHERE id = ?",
		fileID, receiptType, id,
	)
	return err
}

func (d *DB) FindManualChargeByPayID(payID string) (*ManualChargeRequest, error) {
	var r ManualChargeRequest
	err := d.QueryRow(
		`SELECT id, requester_id, requester_name, target_user_id, target_name, amount_usd, amount_toman, status, reject_reason, review_note, gateway, pay_id, tx_ref, receipt_id, receipt_type, created_at, reviewed_at
		 FROM manual_charge_requests WHERE pay_id = ?`, payID,
	).Scan(&r.ID, &r.RequesterID, &r.RequesterName, &r.TargetUserID, &r.TargetName, &r.AmountUSD, &r.AmountToman, &r.Status, &r.RejectReason, &r.ReviewNote, &r.Gateway, &r.PayID, &r.TxRef, &r.ReceiptID, &r.ReceiptType, &r.CreatedAt, &r.ReviewedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrManualChargeNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) UpdateManualCharge(req *ManualChargeRequest) error {
	_, err := d.Exec(
		`UPDATE manual_charge_requests SET status = ?, reject_reason = ?, review_note = ?, gateway = ?, pay_id = ?, tx_ref = ?, receipt_id = ?, receipt_type = ?, reviewed_at = ?
		 WHERE id = ?`,
		req.Status, req.RejectReason, req.ReviewNote, req.Gateway, req.PayID, req.TxRef,
		req.ReceiptID, req.ReceiptType, req.ReviewedAt, req.ID,
	)
	return err
}
