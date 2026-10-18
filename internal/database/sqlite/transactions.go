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
	ErrInvalidAmount = errors.New("مبلغ نامعتبر")
	ErrPendingExists = errors.New("یه پرداخت در انتظار داری")
)

type Transaction struct {
	ID          int64
	UserID      int64
	PayID       string
	AmountUSD   float64
	AmountToman int64
	Type        string
	Status      string
	Description string
	Gateway     string
	CreatedAt   string
	ExpiresAt   string
}

func (d *DB) CreateCharge(userID int64, payID string, amountUSD float64, amountToman int64, expiresAt time.Time) error {
	var exists int
	err := d.QueryRow(
		"SELECT COUNT(*) FROM transactions WHERE user_id = ? AND status = 'pending'",
		userID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrPendingExists
	}

	_, err = d.Exec(
		`INSERT INTO transactions (user_id, pay_id, amount_usd, amount_toman, type, status, description, gateway, created_at, expires_at)
		 VALUES (?, ?, ?, ?, 'charge', 'pending', '', '', ?, ?)`,
		userID, payID, amountUSD, amountToman,
		time.Now().Format(time.RFC3339), expiresAt.Format(time.RFC3339),
	)
	return err
}

func (d *DB) SetTransactionGateway(userID int64, payID, gateway string) error {
	_, err := d.Exec(
		"UPDATE transactions SET gateway = ? WHERE user_id = ? AND pay_id = ? AND status = 'pending'",
		gateway, userID, payID,
	)
	return err
}

func (d *DB) ConfirmCharge(payID string) (int64, float64, error) {
	tx, err := d.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	var userID int64
	var amountUSD float64
	var expiresAt string
	err = tx.QueryRow(
		"SELECT user_id, amount_usd, expires_at FROM transactions WHERE pay_id = ? AND status = 'pending'",
		payID,
	).Scan(&userID, &amountUSD, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, errors.New("تراکنش پیدا نشد یا قبلاً تأیید شده")
	}
	if err != nil {
		return 0, 0, err
	}

	if expiresAt != "" {
		exp, _ := time.Parse(time.RFC3339, expiresAt)
		if !exp.IsZero() && time.Now().After(exp) {
			_, _ = tx.Exec("UPDATE transactions SET status = 'expired' WHERE pay_id = ?", payID)
			_ = tx.Commit()
			return 0, 0, errors.New("این فاکتور منقضی شده")
		}
	}

	if _, err := tx.Exec("UPDATE transactions SET status = 'paid' WHERE pay_id = ?", payID); err != nil {
		return 0, 0, err
	}
	if _, err := tx.Exec("UPDATE users SET balance_usd = balance_usd + ? WHERE id = ?", amountUSD, userID); err != nil {
		return 0, 0, err
	}

	var newBal float64
	if err := tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&newBal); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return userID, newBal, nil
}

func (d *DB) Spend(userID int64, amountUSD float64, description string) (float64, error) {
	return d.DeductBalanceWithType(userID, amountUSD, description, "spend")
}

func (d *DB) DeductBalance(userID int64, amountUSD float64, description string) (float64, error) {
	return d.DeductBalanceWithType(userID, amountUSD, description, "deduct")
}

func (d *DB) DeductBalanceWithType(userID int64, amountUSD float64, description, txType string) (float64, error) {
	if amountUSD <= 0 {
		return 0, ErrInvalidAmount
	}
	if txType == "" {
		txType = "deduct"
	}

	tx, err := d.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var bal float64
	err = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&bal)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}
	if bal < amountUSD {
		return 0, errors.New("موجودی کافی نیست")
	}

	if _, err := tx.Exec("UPDATE users SET balance_usd = balance_usd - ? WHERE id = ?", amountUSD, userID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO transactions (user_id, amount_usd, type, status, description, created_at)
		 VALUES (?, ?, ?, 'paid', ?, ?)`,
		userID, -amountUSD, txType, description, time.Now().Format(time.RFC3339),
	); err != nil {
		return 0, err
	}

	var newBal float64
	_ = tx.QueryRow("SELECT balance_usd FROM users WHERE id = ?", userID).Scan(&newBal)

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newBal, nil
}

func (d *DB) GetHistory(userID int64, limit int) ([]Transaction, error) {
	q := `SELECT id, user_id, pay_id, amount_usd, amount_toman, type, status, description, gateway, created_at, expires_at
	      FROM transactions WHERE user_id = ? ORDER BY id DESC`
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.Query(q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.ID, &t.UserID, &t.PayID, &t.AmountUSD, &t.AmountToman, &t.Type, &t.Status, &t.Description, &t.Gateway, &t.CreatedAt, &t.ExpiresAt)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (d *DB) GetPendingCharge(userID int64) (*Transaction, error) {
	var t Transaction
	err := d.QueryRow(
		`SELECT id, user_id, pay_id, amount_usd, amount_toman, type, status, description, gateway, created_at, expires_at
		 FROM transactions WHERE user_id = ? AND status = 'pending' ORDER BY id DESC LIMIT 1`,
		userID,
	).Scan(&t.ID, &t.UserID, &t.PayID, &t.AmountUSD, &t.AmountToman, &t.Type, &t.Status, &t.Description, &t.Gateway, &t.CreatedAt, &t.ExpiresAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("پرداخت در انتظاری نداری")
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *DB) CancelPending(userID int64) error {
	_, err := d.Exec(
		"UPDATE transactions SET status = 'canceled' WHERE user_id = ? AND status = 'pending'",
		userID,
	)
	return err
}

type TxInfo struct {
	UserID      int64
	FirstName   string
	Username    string
	PayID       string
	AmountUSD   float64
	AmountToman int64
	Type        string
	Status      string
	Gateway     string
	CreatedAt   string
}

func (d *DB) GetAllTransactions() ([]TxInfo, error) {
	rows, err := d.Query(`
		SELECT t.user_id, u.first_name, u.username, t.pay_id, t.amount_usd, t.amount_toman,
		       t.type, t.status, t.gateway, t.created_at
		FROM transactions t
		JOIN users u ON u.id = t.user_id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TxInfo
	for rows.Next() {
		var t TxInfo
		err := rows.Scan(&t.UserID, &t.FirstName, &t.Username, &t.PayID, &t.AmountUSD, &t.AmountToman, &t.Type, &t.Status, &t.Gateway, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}
