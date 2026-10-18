// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCategoryNotFound    = errors.New("دسته‌بندی پیدا نشد")
	ErrProductNotFound     = errors.New("محصول پیدا نشد")
	ErrProductUnavailable  = errors.New("این محصول ناموجود است")
	ErrInsufficientBalance = errors.New("موجودی کیف پول کافی نیست")
)

type Category struct {
	ID       int64
	ParentID int64
	Name     string
	Emoji    string
	Active   bool
	Order    int
}

type Product struct {
	ID            int64
	CategoryID    int64
	Name          string
	Description   string
	PriceUSD      float64
	Available     bool
	DeliveryType  string
	Active        bool
	CreatedAt     string
	ProductType   string
	XUIPanelID    int64
	XUIInboundIDs []int64
	TrafficGB     int64
	DurationDays  int
}

func (d *DB) CreateCategory(parentID int64, name, emoji string) (*Category, error) {
	if name == "" {
		return nil, errors.New("نام دسته‌بندی الزامی است")
	}
	res, err := d.Exec(
		"INSERT INTO categories (parent_id, name, emoji, active, sort_order) VALUES (?, ?, ?, 1, 0)",
		parentID, name, emoji,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Category{ID: id, ParentID: parentID, Name: name, Emoji: emoji, Active: true}, nil
}

func (d *DB) GetCategory(id int64) (*Category, error) {
	var c Category
	var active int
	err := d.QueryRow(
		"SELECT id, parent_id, name, emoji, active, sort_order FROM categories WHERE id = ?", id,
	).Scan(&c.ID, &c.ParentID, &c.Name, &c.Emoji, &active, &c.Order)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}
	c.Active = active == 1
	return &c, nil
}

func (d *DB) GetCategoriesByParent(parentID int64) ([]*Category, error) {
	rows, err := d.Query(
		"SELECT id, parent_id, name, emoji, active, sort_order FROM categories WHERE parent_id = ? AND active = 1 ORDER BY sort_order, id",
		parentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Category
	for rows.Next() {
		var c Category
		var active int
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Emoji, &active, &c.Order); err != nil {
			return nil, err
		}
		c.Active = active == 1
		out = append(out, &c)
	}
	return out, nil
}

func (d *DB) GetAllCategories() ([]*Category, error) {
	rows, err := d.Query("SELECT id, parent_id, name, emoji, active, sort_order FROM categories ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Category
	for rows.Next() {
		var c Category
		var active int
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Emoji, &active, &c.Order); err != nil {
			return nil, err
		}
		c.Active = active == 1
		out = append(out, &c)
	}
	return out, nil
}

func (d *DB) DeleteCategory(id int64) error {
	_, err := d.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}

func (d *DB) ToggleCategory(id int64) error {
	_, err := d.Exec("UPDATE categories SET active = CASE WHEN active = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}

func (d *DB) CreateProduct(categoryID int64, name, description string, priceUSD float64, deliveryType string) (*Product, error) {
	if name == "" {
		return nil, errors.New("نام محصول الزامی است")
	}
	if priceUSD <= 0 {
		return nil, ErrInvalidAmount
	}
	if deliveryType == "" {
		deliveryType = "manual"
	}
	now := time.Now().Format(time.RFC3339)
	res, err := d.Exec(
		`INSERT INTO products (category_id, name, description, price_usd, available, delivery_type, active, product_type, xui_panel_id, xui_inbound_ids, traffic_gb, duration_days, created_at)
		 VALUES (?, ?, ?, ?, 1, ?, 1, '', 0, '[]', 0, 0, ?)`,
		categoryID, name, description, priceUSD, deliveryType, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Product{
		ID: id, CategoryID: categoryID, Name: name, Description: description,
		PriceUSD: priceUSD, Available: true, DeliveryType: deliveryType, Active: true, CreatedAt: now,
	}, nil
}

func (d *DB) SetProductVPN(id, panelID int64, inboundIDs []int64, trafficGB int64, durationDays int) error {
	inboundJSON, _ := json.Marshal(inboundIDs)
	_, err := d.Exec(
		`UPDATE products SET product_type = 'vpn', xui_panel_id = ?, xui_inbound_ids = ?, traffic_gb = ?, duration_days = ? WHERE id = ?`,
		panelID, string(inboundJSON), trafficGB, durationDays, id,
	)
	return err
}

func (d *DB) GetProduct(id int64) (*Product, error) {
	var p Product
	var available, active int
	var inboundJSON string
	err := d.QueryRow(
		`SELECT id, category_id, name, description, price_usd, available, delivery_type, active, created_at, product_type, xui_panel_id, xui_inbound_ids, traffic_gb, duration_days
		 FROM products WHERE id = ?`, id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.PriceUSD, &available, &p.DeliveryType, &active, &p.CreatedAt, &p.ProductType, &p.XUIPanelID, &inboundJSON, &p.TrafficGB, &p.DurationDays)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Available = available == 1
	p.Active = active == 1
	_ = json.Unmarshal([]byte(inboundJSON), &p.XUIInboundIDs)
	return &p, nil
}

func (d *DB) GetProductsByCategory(categoryID int64) ([]*Product, error) {
	rows, err := d.Query(
		`SELECT id, category_id, name, description, price_usd, available, delivery_type, active, created_at, product_type, xui_panel_id, xui_inbound_ids, traffic_gb, duration_days
		 FROM products WHERE category_id = ? AND active = 1`, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Product
	for rows.Next() {
		var p Product
		var available, active int
		var inboundJSON string
		err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.PriceUSD, &available, &p.DeliveryType, &active, &p.CreatedAt, &p.ProductType, &p.XUIPanelID, &inboundJSON, &p.TrafficGB, &p.DurationDays)
		if err != nil {
			return nil, err
		}
		p.Available = available == 1
		p.Active = active == 1
		_ = json.Unmarshal([]byte(inboundJSON), &p.XUIInboundIDs)
		out = append(out, &p)
	}
	return out, nil
}

func (d *DB) GetAllProducts() ([]*Product, error) {
	rows, err := d.Query(
		`SELECT id, category_id, name, description, price_usd, available, delivery_type, active, created_at, product_type, xui_panel_id, xui_inbound_ids, traffic_gb, duration_days
		 FROM products ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Product
	for rows.Next() {
		var p Product
		var available, active int
		var inboundJSON string
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.PriceUSD, &available, &p.DeliveryType, &active, &p.CreatedAt, &p.ProductType, &p.XUIPanelID, &inboundJSON, &p.TrafficGB, &p.DurationDays); err != nil {
			return nil, err
		}
		p.Available = available == 1
		p.Active = active == 1
		_ = json.Unmarshal([]byte(inboundJSON), &p.XUIInboundIDs)
		out = append(out, &p)
	}
	return out, nil
}

func (d *DB) DeleteProduct(id int64) error {
	_, err := d.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}

func (d *DB) ToggleProductAvailable(id int64) error {
	_, err := d.Exec("UPDATE products SET available = CASE WHEN available = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}

func (d *DB) ToggleProductActive(id int64) error {
	_, err := d.Exec("UPDATE products SET active = CASE WHEN active = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}

func (d *DB) UpdateProductPrice(id int64, priceUSD float64) error {
	if priceUSD <= 0 {
		return ErrInvalidAmount
	}
	_, err := d.Exec("UPDATE products SET price_usd = ? WHERE id = ?", priceUSD, id)
	return err
}

func (d *DB) DecrementProductStock(id int64) error {
	_, err := d.Exec("UPDATE products SET available = 0 WHERE id = ? AND available = 1", id)
	return err
}

func (d *DB) CountProductsInCategory(catID int64) (int, error) {
	var c int
	err := d.QueryRow("SELECT COUNT(*) FROM products WHERE category_id = ?", catID).Scan(&c)
	return c, err
}

func (d *DB) CountSubCategories(catID int64) (int, error) {
	var c int
	err := d.QueryRow("SELECT COUNT(*) FROM categories WHERE parent_id = ?", catID).Scan(&c)
	return c, err
}

var _ = fmt.Sprintf
