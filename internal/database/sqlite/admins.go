// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrAlreadyAdmin = errors.New("این کاربر از قبل ادمین است")
	ErrNotAdmin     = errors.New("این کاربر ادمین نیست")
)

type AdminRole string

const (
	RoleSuperAdmin AdminRole = "super_admin"
	RoleFinance    AdminRole = "finance"
	RoleSupport    AdminRole = "support"
	RoleViewer     AdminRole = "viewer"
)

type Admin struct {
	UserID  int64
	Role    AdminRole
	DeptIDs []int64
	AddedBy int64
	AddedAt string
}

func (d *DB) GetAdminRole(userID, mainAdminID int64) AdminRole {
	if userID == mainAdminID {
		return RoleSuperAdmin
	}
	var role string
	err := d.QueryRow("SELECT role FROM admins WHERE user_id = ?", userID).Scan(&role)
	if err != nil {
		return ""
	}
	return AdminRole(role)
}

func (d *DB) IsAdmin(userID, mainAdminID int64) bool {
	return d.GetAdminRole(userID, mainAdminID) != ""
}

func (d *DB) AddAdmin(userID int64, role AdminRole, deptIDs []int64, addedBy int64) error {
	var exists int
	_ = d.QueryRow("SELECT COUNT(*) FROM admins WHERE user_id = ?", userID).Scan(&exists)
	if exists > 0 {
		return ErrAlreadyAdmin
	}
	deptJSON, _ := json.Marshal(deptIDs)
	_, err := d.Exec(
		"INSERT INTO admins (user_id, role, dept_ids, added_by, added_at) VALUES (?, ?, ?, ?, ?)",
		userID, string(role), string(deptJSON), addedBy, time.Now().Format(time.RFC3339),
	)
	return err
}

func (d *DB) RemoveAdmin(userID int64) error {
	res, err := d.Exec("DELETE FROM admins WHERE user_id = ?", userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotAdmin
	}
	return nil
}

func (d *DB) SetAdminRole(userID int64, role AdminRole) error {
	res, err := d.Exec("UPDATE admins SET role = ? WHERE user_id = ?", string(role), userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotAdmin
	}
	return nil
}

func (d *DB) SetAdminDepts(userID int64, deptIDs []int64) error {
	deptJSON, _ := json.Marshal(deptIDs)
	res, err := d.Exec("UPDATE admins SET dept_ids = ? WHERE user_id = ?", string(deptJSON), userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotAdmin
	}
	return nil
}

func (d *DB) GetAllAdmins() ([]*Admin, error) {
	rows, err := d.Query("SELECT user_id, role, dept_ids, added_by, added_at FROM admins ORDER BY user_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Admin
	for rows.Next() {
		var a Admin
		var deptJSON string
		if err := rows.Scan(&a.UserID, &a.Role, &deptJSON, &a.AddedBy, &a.AddedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(deptJSON), &a.DeptIDs)
		out = append(out, &a)
	}
	return out, nil
}

func (d *DB) GetAdminByID(userID int64) (*Admin, error) {
	var a Admin
	var deptJSON string
	err := d.QueryRow("SELECT user_id, role, dept_ids, added_by, added_at FROM admins WHERE user_id = ?", userID).
		Scan(&a.UserID, &a.Role, &deptJSON, &a.AddedBy, &a.AddedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotAdmin
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(deptJSON), &a.DeptIDs)
	return &a, nil
}
