// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"encoding/json"
	"errors"
)

type DeliveryField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required"`
}

type ServiceAction struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	Emoji        string `json:"emoji,omitempty"`
	RequireInput bool   `json:"require_input"`
	InputLabel   string `json:"input_label,omitempty"`
}

func (d *DB) GetCategoryDeliveryFields(catID int64) ([]DeliveryField, error) {
	var raw string
	err := d.QueryRow("SELECT delivery_fields FROM categories WHERE id = ?", catID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var out []DeliveryField
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &out)
	}
	return out, nil
}

func (d *DB) SetCategoryDeliveryFields(catID int64, fields []DeliveryField) error {
	raw, _ := json.Marshal(fields)
	_, err := d.Exec("UPDATE categories SET delivery_fields = ? WHERE id = ?", string(raw), catID)
	return err
}

func (d *DB) GetCategoryServiceActions(catID int64) ([]ServiceAction, error) {
	var raw string
	err := d.QueryRow("SELECT service_actions FROM categories WHERE id = ?", catID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var out []ServiceAction
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &out)
	}
	return out, nil
}

func (d *DB) SetCategoryServiceActions(catID int64, actions []ServiceAction) error {
	raw, _ := json.Marshal(actions)
	_, err := d.Exec("UPDATE categories SET service_actions = ? WHERE id = ?", string(raw), catID)
	return err
}

func (d *DB) AddCategoryDeliveryField(catID int64, f DeliveryField) error {
	fields, err := d.GetCategoryDeliveryFields(catID)
	if err != nil {
		return err
	}
	for _, existing := range fields {
		if existing.Key == f.Key {
			return errors.New("این کلید قبلاً وجود دارد")
		}
	}
	fields = append(fields, f)
	return d.SetCategoryDeliveryFields(catID, fields)
}

func (d *DB) RemoveCategoryDeliveryField(catID int64, key string) error {
	fields, err := d.GetCategoryDeliveryFields(catID)
	if err != nil {
		return err
	}
	var out []DeliveryField
	for _, f := range fields {
		if f.Key != key {
			out = append(out, f)
		}
	}
	return d.SetCategoryDeliveryFields(catID, out)
}

func (d *DB) AddCategoryServiceAction(catID int64, a ServiceAction) error {
	actions, err := d.GetCategoryServiceActions(catID)
	if err != nil {
		return err
	}
	for _, existing := range actions {
		if existing.Key == a.Key {
			return errors.New("این اکشن قبلاً وجود دارد")
		}
	}
	actions = append(actions, a)
	return d.SetCategoryServiceActions(catID, actions)
}

func (d *DB) RemoveCategoryServiceAction(catID int64, key string) error {
	actions, err := d.GetCategoryServiceActions(catID)
	if err != nil {
		return err
	}
	var out []ServiceAction
	for _, a := range actions {
		if a.Key != key {
			out = append(out, a)
		}
	}
	return d.SetCategoryServiceActions(catID, out)
}
