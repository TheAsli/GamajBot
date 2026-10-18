// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("کاربر پیدا نشد")

type User struct {
	ID         int64
	FirstName  string
	Username   string
	BalanceUSD float64
	Banned     bool
	BanReason  string
	Language   string
}

func (d *DB) GetOrCreateUser(id int64, firstName, username string) (*User, error) {
	var u User
	var banned int
	err := d.QueryRow(
		"SELECT id, first_name, username, balance_usd, banned, ban_reason, language FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.FirstName, &u.Username, &u.BalanceUSD, &banned, &u.BanReason, &u.Language)

	if errors.Is(err, sql.ErrNoRows) {
		_, err = d.Exec(
			"INSERT INTO users (id, first_name, username, balance_usd, banned, ban_reason, language) VALUES (?, ?, ?, 0, 0, '', 'fa')",
			id, firstName, username,
		)
		if err != nil {
			return nil, fmt.Errorf("ساخت کاربر: %w", err)
		}
		u = User{ID: id, FirstName: firstName, Username: username, Language: "fa"}
		return &u, nil
	}
	if err != nil {
		return nil, fmt.Errorf("دریافت کاربر: %w", err)
	}

	u.Banned = banned == 1
	_, _ = d.Exec("UPDATE users SET first_name = ?, username = ? WHERE id = ?", firstName, username, id)
	return &u, nil
}

func (d *DB) GetUser(id int64) (*User, error) {
	var u User
	var banned int
	err := d.QueryRow(
		"SELECT id, first_name, username, balance_usd, banned, ban_reason, language FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.FirstName, &u.Username, &u.BalanceUSD, &banned, &u.BanReason, &u.Language)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u.Banned = banned == 1
	return &u, nil
}

func (d *DB) UpdateUserBalance(id int64, delta float64) (float64, error) {
	_, err := d.Exec("UPDATE users SET balance_usd = balance_usd + ? WHERE id = ?", delta, id)
	if err != nil {
		return 0, err
	}
	var newBal float64
	err = d.QueryRow("SELECT balance_usd FROM users WHERE id = ?", id).Scan(&newBal)
	if err != nil {
		return 0, err
	}
	return newBal, nil
}

func (d *DB) SetUserBalance(id int64, bal float64) error {
	_, err := d.Exec("UPDATE users SET balance_usd = ? WHERE id = ?", bal, id)
	return err
}

func (d *DB) SetUserLanguage(id int64, lang string) error {
	_, err := d.Exec("UPDATE users SET language = ? WHERE id = ?", lang, id)
	return err
}

func (d *DB) BanUser(id int64, reason string) error {
	_, err := d.Exec("UPDATE users SET banned = 1, ban_reason = ? WHERE id = ?", reason, id)
	return err
}

func (d *DB) UnbanUser(id int64) error {
	_, err := d.Exec("UPDATE users SET banned = 0, ban_reason = '' WHERE id = ?", id)
	return err
}

func (d *DB) IsUserBanned(id int64) bool {
	var banned int
	err := d.QueryRow("SELECT banned FROM users WHERE id = ?", id).Scan(&banned)
	if err != nil {
		return false
	}
	return banned == 1
}

type UserInfo struct {
	ID         int64
	FirstName  string
	Username   string
	BalanceUSD float64
	TxCount    int
	Banned     bool
}

func (d *DB) GetAllUsers() ([]UserInfo, error) {
	rows, err := d.Query(
		`SELECT u.id, u.first_name, u.username, u.balance_usd,
		       (SELECT COUNT(*) FROM transactions WHERE user_id = u.id) AS tx_count,
		       u.banned
		FROM users u
		ORDER BY u.balance_usd DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UserInfo
	for rows.Next() {
		var ui UserInfo
		var banned int
		if err := rows.Scan(&ui.ID, &ui.FirstName, &ui.Username, &ui.BalanceUSD, &ui.TxCount, &banned); err != nil {
			return nil, err
		}
		ui.Banned = banned == 1
		out = append(out, ui)
	}
	return out, nil
}
