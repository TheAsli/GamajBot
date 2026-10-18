// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type legacyDB struct {
	Users                map[string]*legacyUser         `json:"users"`
	Admins               []*legacyAdmin                 `json:"admins"`
	Settings             *legacySettings                `json:"settings"`
	Departments          []*legacyDepartment            `json:"departments"`
	Tickets              []*legacyTicket                `json:"tickets"`
	ManualChargeRequests map[string]*legacyManualCharge `json:"manual_charge_requests"`
	PaymentGateways      *legacyPaymentGateways         `json:"payment_gateways"`
	Categories           []*legacyCategory              `json:"categories"`
	Products             []*legacyProduct               `json:"products"`
	Orders               []*legacyOrder                 `json:"orders"`
	NextCategoryID       int64                          `json:"next_category_id"`
	NextProductID        int64                          `json:"next_product_id"`
	XUIPanels            []*legacyXUIPanel              `json:"xui_panels"`
	NextXUIPanelID       int64                          `json:"next_xui_panel_id"`
}

type legacyUser struct {
	ID           int64               `json:"id"`
	FirstName    string              `json:"first_name"`
	Username     string              `json:"username"`
	BalanceUSD   float64             `json:"balance_usd"`
	Banned       bool                `json:"banned"`
	BanReason    string              `json:"ban_reason"`
	Transactions []legacyTransaction `json:"transactions"`
}

type legacyTransaction struct {
	PayID       string  `json:"pay_id"`
	AmountUSD   float64 `json:"amount_usd"`
	AmountToman int64   `json:"amount_toman"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   string  `json:"expires_at"`
	Gateway     string  `json:"gateway"`
}

type legacyAdmin struct {
	UserID  int64   `json:"user_id"`
	Role    string  `json:"role"`
	DeptIDs []int64 `json:"dept_ids"`
	AddedBy int64   `json:"added_by"`
	AddedAt string  `json:"added_at"`
}

type legacySettings struct {
	DollarRateMargin float64 `json:"dollar_rate_margin"`
	ChargeMin        int64   `json:"charge_min"`
	ChargeMax        int64   `json:"charge_max"`
	QuickAmounts     []int64 `json:"quick_amounts"`
	MaintenanceMode  bool    `json:"maintenance_mode"`
	WelcomeText      string  `json:"welcome_text"`
	ForceJoinText    string  `json:"force_join_text"`
	ForceJoinChannel string  `json:"force_join_channel"`
	ForceJoinURL     string  `json:"force_join_url"`
	SupportEnabled   bool    `json:"support_enabled"`
	Currency         string  `json:"currency"`
}

type legacyDepartment struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Emoji       string `json:"emoji"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
}

type legacyTicket struct {
	ID            int64                 `json:"id"`
	UserID        int64                 `json:"user_id"`
	UserFirstName string                `json:"user_first_name"`
	UserUsername  string                `json:"user_username"`
	DeptID        int64                 `json:"dept_id"`
	Subject       string                `json:"subject"`
	Status        string                `json:"status"`
	Priority      string                `json:"priority"`
	Messages      []legacyTicketMessage `json:"messages"`
	AssignedTo    int64                 `json:"assigned_to"`
	CreatedAt     string                `json:"created_at"`
	UpdatedAt     string                `json:"updated_at"`
	ClosedAt      string                `json:"closed_at"`
}

type legacyTicketMessage struct {
	FromID    int64  `json:"from_id"`
	IsAdmin   bool   `json:"is_admin"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type legacyManualCharge struct {
	ID            string  `json:"id"`
	RequesterID   int64   `json:"requester_id"`
	RequesterName string  `json:"requester_name"`
	TargetUserID  int64   `json:"target_user_id"`
	TargetName    string  `json:"target_name"`
	AmountUSD     float64 `json:"amount_usd"`
	AmountToman   int64   `json:"amount_toman"`
	Status        string  `json:"status"`
	RejectReason  string  `json:"reject_reason"`
	ReviewNote    string  `json:"review_note"`
	Gateway       string  `json:"gateway"`
	PayID         string  `json:"pay_id"`
	TxRef         string  `json:"tx_ref"`
	ReceiptID     string  `json:"receipt_id"`
	ReceiptType   string  `json:"receipt_type"`
	CreatedAt     string  `json:"created_at"`
	ReviewedAt    string  `json:"reviewed_at"`
}

type legacyPaymentGateways struct {
	TetraminatorEnabled bool   `json:"tetraminator_enabled"`
	CubePayEnabled      bool   `json:"cubepay_enabled"`
	ManualCardEnabled   bool   `json:"manual_card_enabled"`
	ManualCardNumber    string `json:"manual_card_number"`
	ManualCardHolder    string `json:"manual_card_holder"`
	CubePayToken        string `json:"cubepay_token"`
	TetraminatorAPIKey  string `json:"tetraminator_api_key"`
	TetraminatorBaseURL string `json:"tetraminator_base_url"`
}

type legacyCategory struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parent_id"`
	Name     string `json:"name"`
	Emoji    string `json:"emoji"`
	Active   bool   `json:"active"`
	Order    int    `json:"order"`
}

type legacyProduct struct {
	ID            int64   `json:"id"`
	CategoryID    int64   `json:"category_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	PriceUSD      float64 `json:"price_usd"`
	Available     bool    `json:"available"`
	DeliveryType  string  `json:"delivery_type"`
	Active        bool    `json:"active"`
	CreatedAt     string  `json:"created_at"`
	ProductType   string  `json:"product_type"`
	XUIPanelID    int64   `json:"xui_panel_id"`
	XUIInboundIDs []int64 `json:"xui_inbound_ids"`
	TrafficGB     int64   `json:"traffic_gb"`
	DurationDays  int     `json:"duration_days"`
}

type legacyOrder struct {
	ID              string   `json:"id"`
	UserID          int64    `json:"user_id"`
	UserName        string   `json:"user_name"`
	ProductID       int64    `json:"product_id"`
	ProductName     string   `json:"product_name"`
	PriceUSD        float64  `json:"price_usd"`
	Status          string   `json:"status"`
	ServiceInfo     string   `json:"service_info"`
	RejectNote      string   `json:"reject_note"`
	CreatedAt       string   `json:"created_at"`
	DeliveredAt     string   `json:"delivered_at"`
	CodeRequestedAt string   `json:"code_requested_at"`
	CodeValue       string   `json:"code_value"`
	ClientEmail     string   `json:"client_email"`
	ClientLinks     []string `json:"client_links"`
}

type legacyXUIPanel struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	BaseURL    string `json:"base_url"`
	ServerIP   string `json:"server_ip"`
	AuthMethod string `json:"auth_method"`
	APIToken   string `json:"api_token"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Active     bool   `json:"active"`
	CreatedAt  string `json:"created_at"`
}

func (d *DB) MigrateFromJSON(jsonPath string) error {
	var count int
	if err := d.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return fmt.Errorf("چک کردن جدول users: %w", err)
	}
	if count > 0 {
		return nil
	}

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("خواندن JSON: %w", err)
	}

	var old legacyDB
	if err := json.Unmarshal(data, &old); err != nil {
		return fmt.Errorf("پارس JSON: %w", err)
	}

	backupDir := filepath.Join(filepath.Dir(jsonPath), "backups")
	_ = os.MkdirAll(backupDir, 0755)
	backupPath := filepath.Join(backupDir, fmt.Sprintf("wallets_%s.json", time.Now().Format("20060102_150405")))
	_ = os.WriteFile(backupPath, data, 0644)

	tx, err := d.Begin()
	if err != nil {
		return fmt.Errorf("شروع تراکنش: %w", err)
	}
	defer tx.Rollback()

	for _, u := range old.Users {
		_, err := tx.Exec(
			"INSERT INTO users (id, first_name, username, balance_usd, banned, ban_reason) VALUES (?, ?, ?, ?, ?, ?)",
			u.ID, u.FirstName, u.Username, u.BalanceUSD, boolToInt(u.Banned), u.BanReason,
		)
		if err != nil {
			return fmt.Errorf("درج کاربر %d: %w", u.ID, err)
		}

		for _, t := range u.Transactions {
			_, err := tx.Exec(
				`INSERT INTO transactions (user_id, pay_id, amount_usd, amount_toman, type, status, description, gateway, created_at, expires_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				u.ID, t.PayID, t.AmountUSD, t.AmountToman, t.Type, t.Status, t.Description, t.Gateway, t.CreatedAt, t.ExpiresAt,
			)
			if err != nil {
				return fmt.Errorf("درج تراکنش کاربر %d: %w", u.ID, err)
			}
		}
	}

	for _, a := range old.Admins {
		deptJSON, _ := json.Marshal(a.DeptIDs)
		_, err := tx.Exec(
			"INSERT INTO admins (user_id, role, dept_ids, added_by, added_at) VALUES (?, ?, ?, ?, ?)",
			a.UserID, a.Role, string(deptJSON), a.AddedBy, a.AddedAt,
		)
		if err != nil {
			return fmt.Errorf("درج ادمین %d: %w", a.UserID, err)
		}
	}

	for _, dept := range old.Departments {
		_, err := tx.Exec(
			"INSERT INTO departments (id, name, emoji, description, active) VALUES (?, ?, ?, ?, ?)",
			dept.ID, dept.Name, dept.Emoji, dept.Description, boolToInt(dept.Active),
		)
		if err != nil {
			return fmt.Errorf("درج دپارتمان %d: %w", dept.ID, err)
		}
	}

	for _, t := range old.Tickets {
		_, err := tx.Exec(
			`INSERT INTO tickets (id, user_id, user_first_name, user_username, dept_id, subject, status, priority, assigned_to, created_at, updated_at, closed_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.ID, t.UserID, t.UserFirstName, t.UserUsername, t.DeptID, t.Subject, t.Status, t.Priority, t.AssignedTo, t.CreatedAt, t.UpdatedAt, t.ClosedAt,
		)
		if err != nil {
			return fmt.Errorf("درج تیکت %d: %w", t.ID, err)
		}
		for _, m := range t.Messages {
			_, err := tx.Exec(
				"INSERT INTO ticket_messages (ticket_id, from_id, is_admin, text, created_at) VALUES (?, ?, ?, ?, ?)",
				t.ID, m.FromID, boolToInt(m.IsAdmin), m.Text, m.CreatedAt,
			)
			if err != nil {
				return fmt.Errorf("درج پیام تیکت %d: %w", t.ID, err)
			}
		}
	}

	for _, c := range old.Categories {
		_, err := tx.Exec(
			"INSERT INTO categories (id, parent_id, name, emoji, active, sort_order) VALUES (?, ?, ?, ?, ?, ?)",
			c.ID, c.ParentID, c.Name, c.Emoji, boolToInt(c.Active), c.Order,
		)
		if err != nil {
			return fmt.Errorf("درج دسته %d: %w", c.ID, err)
		}
	}

	for _, p := range old.Products {
		inboundJSON, _ := json.Marshal(p.XUIInboundIDs)
		_, err := tx.Exec(
			`INSERT INTO products (id, category_id, name, description, price_usd, available, delivery_type, active, product_type, xui_panel_id, xui_inbound_ids, traffic_gb, duration_days, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.ID, p.CategoryID, p.Name, p.Description, p.PriceUSD, boolToInt(p.Available), p.DeliveryType, boolToInt(p.Active), p.ProductType, p.XUIPanelID, string(inboundJSON), p.TrafficGB, p.DurationDays, p.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("درج محصول %d: %w", p.ID, err)
		}
	}

	for _, o := range old.Orders {
		linksJSON, _ := json.Marshal(o.ClientLinks)
		_, err := tx.Exec(
			`INSERT INTO orders (id, user_id, user_name, product_id, product_name, price_usd, status, service_info, reject_note, code_value, client_email, client_links, created_at, delivered_at, code_requested_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			o.ID, o.UserID, o.UserName, o.ProductID, o.ProductName, o.PriceUSD, o.Status, o.ServiceInfo, o.RejectNote, o.CodeValue, o.ClientEmail, string(linksJSON), o.CreatedAt, o.DeliveredAt, o.CodeRequestedAt,
		)
		if err != nil {
			return fmt.Errorf("درج سفارش %s: %w", o.ID, err)
		}
	}

	for _, mc := range old.ManualChargeRequests {
		_, err := tx.Exec(
			`INSERT INTO manual_charge_requests (id, requester_id, requester_name, target_user_id, target_name, amount_usd, amount_toman, status, reject_reason, review_note, gateway, pay_id, tx_ref, receipt_id, receipt_type, created_at, reviewed_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			mc.ID, mc.RequesterID, mc.RequesterName, mc.TargetUserID, mc.TargetName, mc.AmountUSD, mc.AmountToman, mc.Status, mc.RejectReason, mc.ReviewNote, mc.Gateway, mc.PayID, mc.TxRef, mc.ReceiptID, mc.ReceiptType, mc.CreatedAt, mc.ReviewedAt,
		)
		if err != nil {
			return fmt.Errorf("درج شارژ دستی %s: %w", mc.ID, err)
		}
	}

	if old.Settings != nil {
		quickJSON, _ := json.Marshal(old.Settings.QuickAmounts)
		settingsMap := map[string]string{
			"dollar_rate_margin": fmt.Sprintf("%f", old.Settings.DollarRateMargin),
			"charge_min":         fmt.Sprintf("%d", old.Settings.ChargeMin),
			"charge_max":         fmt.Sprintf("%d", old.Settings.ChargeMax),
			"quick_amounts":      string(quickJSON),
			"maintenance_mode":   boolToStr(old.Settings.MaintenanceMode),
			"welcome_text":       old.Settings.WelcomeText,
			"force_join_text":    old.Settings.ForceJoinText,
			"force_join_channel": old.Settings.ForceJoinChannel,
			"force_join_url":     old.Settings.ForceJoinURL,
			"support_enabled":    boolToStr(old.Settings.SupportEnabled),
			"currency":           old.Settings.Currency,
		}
		for k, v := range settingsMap {
			_, err := tx.Exec("INSERT INTO settings (key, value) VALUES (?, ?)", k, v)
			if err != nil {
				return fmt.Errorf("درج تنظیم %s: %w", k, err)
			}
		}
	}

	if old.PaymentGateways != nil {
		pgMap := map[string]string{
			"tetraminator_enabled":  boolToStr(old.PaymentGateways.TetraminatorEnabled),
			"cubepay_enabled":       boolToStr(old.PaymentGateways.CubePayEnabled),
			"manual_card_enabled":   boolToStr(old.PaymentGateways.ManualCardEnabled),
			"manual_card_number":    old.PaymentGateways.ManualCardNumber,
			"manual_card_holder":    old.PaymentGateways.ManualCardHolder,
			"cubepay_token":         old.PaymentGateways.CubePayToken,
			"tetraminator_api_key":  old.PaymentGateways.TetraminatorAPIKey,
			"tetraminator_base_url": old.PaymentGateways.TetraminatorBaseURL,
		}
		for k, v := range pgMap {
			_, err := tx.Exec("INSERT INTO payment_gateways (key, value) VALUES (?, ?)", k, v)
			if err != nil {
				return fmt.Errorf("درج درگاه %s: %w", k, err)
			}
		}
	}

	for _, xp := range old.XUIPanels {
		_, err := tx.Exec(
			`INSERT INTO xui_panels (id, name, base_url, server_ip, auth_method, api_token, username, password, active, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			xp.ID, xp.Name, xp.BaseURL, xp.ServerIP, xp.AuthMethod, xp.APIToken, xp.Username, xp.Password, boolToInt(xp.Active), xp.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("درج پنل %d: %w", xp.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("کامیت تراکنش: %w", err)
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
