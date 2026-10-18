// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"errors"
	"time"
)

var ErrPanelNotFound = errors.New("پنل پیدا نشد")

type XUIPanel struct {
	ID         int64
	PanelType  string
	Name       string
	BaseURL    string
	ServerIP   string
	AuthMethod string
	APIToken   string
	Username   string
	Password   string
	Active     bool
	CreatedAt  string
}

func (d *DB) AddXUIPanel(p *XUIPanel) (*XUIPanel, error) {
	now := time.Now().Format(time.RFC3339)
	if p.PanelType == "" {
		p.PanelType = "3x-ui"
	}
	res, err := d.Exec(
		`INSERT INTO xui_panels (panel_type, name, base_url, server_ip, auth_method, api_token, username, password, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.PanelType, p.Name, p.BaseURL, p.ServerIP, p.AuthMethod, p.APIToken, p.Username, p.Password, boolToInt(p.Active), now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	p.ID = id
	p.CreatedAt = now
	return p, nil
}

func (d *DB) GetXUIPanel(id int64) (*XUIPanel, error) {
	var p XUIPanel
	var active int
	err := d.QueryRow(
		`SELECT id, panel_type, name, base_url, server_ip, auth_method, api_token, username, password, active, created_at
		 FROM xui_panels WHERE id = ?`, id,
	).Scan(&p.ID, &p.PanelType, &p.Name, &p.BaseURL, &p.ServerIP, &p.AuthMethod, &p.APIToken, &p.Username, &p.Password, &active, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPanelNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Active = active == 1
	return &p, nil
}

func (d *DB) GetAllXUIPanels() ([]*XUIPanel, error) {
	rows, err := d.Query(
		`SELECT id, panel_type, name, base_url, server_ip, auth_method, api_token, username, password, active, created_at FROM xui_panels ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*XUIPanel
	for rows.Next() {
		var p XUIPanel
		var active int
		if err := rows.Scan(&p.ID, &p.PanelType, &p.Name, &p.BaseURL, &p.ServerIP, &p.AuthMethod, &p.APIToken, &p.Username, &p.Password, &active, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.Active = active == 1
		out = append(out, &p)
	}
	return out, nil
}

func (d *DB) UpdateXUIPanel(p *XUIPanel) error {
	_, err := d.Exec(
		`UPDATE xui_panels SET panel_type = ?, name = ?, base_url = ?, server_ip = ?, auth_method = ?, api_token = ?, username = ?, password = ?, active = ? WHERE id = ?`,
		p.PanelType, p.Name, p.BaseURL, p.ServerIP, p.AuthMethod, p.APIToken, p.Username, p.Password, boolToInt(p.Active), p.ID,
	)
	return err
}

func (d *DB) DeleteXUIPanel(id int64) error {
	_, err := d.Exec("DELETE FROM xui_panels WHERE id = ?", id)
	return err
}

func (d *DB) ToggleXUIPanelActive(id int64) error {
	_, err := d.Exec("UPDATE xui_panels SET active = CASE WHEN active = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}
