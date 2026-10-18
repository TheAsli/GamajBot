// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"errors"
)

var ErrDeptNotFound = errors.New("دپارتمان پیدا نشد")

type Department struct {
	ID          int64
	Name        string
	Emoji       string
	Description string
	Active      bool
}

func (d *DB) AddDepartment(name, emoji, description string) (*Department, error) {
	res, err := d.Exec(
		"INSERT INTO departments (name, emoji, description, active) VALUES (?, ?, ?, 1)",
		name, emoji, description,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Department{ID: id, Name: name, Emoji: emoji, Description: description, Active: true}, nil
}

func (d *DB) RemoveDepartment(id int64) error {
	_, err := d.Exec("DELETE FROM departments WHERE id = ?", id)
	return err
}

func (d *DB) GetAllDepartments() ([]*Department, error) {
	rows, err := d.Query("SELECT id, name, emoji, description, active FROM departments ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Department
	for rows.Next() {
		var dept Department
		var active int
		if err := rows.Scan(&dept.ID, &dept.Name, &dept.Emoji, &dept.Description, &active); err != nil {
			return nil, err
		}
		dept.Active = active == 1
		out = append(out, &dept)
	}
	return out, nil
}

func (d *DB) GetActiveDepartments() ([]*Department, error) {
	rows, err := d.Query("SELECT id, name, emoji, description, active FROM departments WHERE active = 1 ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Department
	for rows.Next() {
		var dept Department
		var active int
		if err := rows.Scan(&dept.ID, &dept.Name, &dept.Emoji, &dept.Description, &active); err != nil {
			return nil, err
		}
		dept.Active = active == 1
		out = append(out, &dept)
	}
	return out, nil
}

func (d *DB) GetDepartmentByID(id int64) (*Department, error) {
	var dept Department
	var active int
	err := d.QueryRow(
		"SELECT id, name, emoji, description, active FROM departments WHERE id = ?", id,
	).Scan(&dept.ID, &dept.Name, &dept.Emoji, &dept.Description, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDeptNotFound
	}
	if err != nil {
		return nil, err
	}
	dept.Active = active == 1
	return &dept, nil
}

func (d *DB) ToggleDepartment(id int64) error {
	_, err := d.Exec("UPDATE departments SET active = CASE WHEN active = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}
