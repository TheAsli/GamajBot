// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"encoding/json"
	"errors"
	"time"
)

type ActionRequestStatus string

const (
	ActionPending    ActionRequestStatus = "pending"
	ActionInProgress ActionRequestStatus = "in_progress"
	ActionDone       ActionRequestStatus = "done"
	ActionRejected   ActionRequestStatus = "rejected"
)

type ActionRequest struct {
	ID          string              `json:"id"`
	ActionKey   string              `json:"action_key"`
	ActionLabel string              `json:"action_label"`
	UserInput   string              `json:"user_input,omitempty"`
	Status      ActionRequestStatus `json:"status"`
	AdminNote   string              `json:"admin_note,omitempty"`
	AdminID     int64               `json:"admin_id,omitempty"`
	RequestedAt string              `json:"requested_at"`
	DoneAt      string              `json:"done_at,omitempty"`
}

func (d *DB) GetOrderDeliveryData(orderID string) (map[string]string, error) {
	var raw string
	err := d.QueryRow("SELECT delivery_data FROM orders WHERE id = ?", orderID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &out)
	}
	return out, nil
}

func (d *DB) SetOrderDeliveryData(orderID string, data map[string]string) error {
	raw, _ := json.Marshal(data)
	_, err := d.Exec("UPDATE orders SET delivery_data = ? WHERE id = ?", string(raw), orderID)
	return err
}

func (d *DB) GetOrderActionRequests(orderID string) ([]ActionRequest, error) {
	var raw string
	err := d.QueryRow("SELECT action_requests FROM orders WHERE id = ?", orderID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var out []ActionRequest
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &out)
	}
	return out, nil
}

func (d *DB) AddOrderActionRequest(orderID string, req ActionRequest) error {
	reqs, err := d.GetOrderActionRequests(orderID)
	if err != nil {
		return err
	}
	if req.ID == "" {
		req.ID = "req_" + time.Now().Format("20060102150405") + "_" + orderID
	}
	if req.Status == "" {
		req.Status = ActionPending
	}
	if req.RequestedAt == "" {
		req.RequestedAt = time.Now().Format(time.RFC3339)
	}
	reqs = append(reqs, req)
	raw, _ := json.Marshal(reqs)
	_, err = d.Exec("UPDATE orders SET action_requests = ? WHERE id = ?", string(raw), orderID)
	return err
}

func (d *DB) UpdateOrderActionStatus(orderID, reqID string, status ActionRequestStatus, adminID int64, note string) error {
	reqs, err := d.GetOrderActionRequests(orderID)
	if err != nil {
		return err
	}
	found := false
	for i := range reqs {
		if reqs[i].ID == reqID {
			reqs[i].Status = status
			reqs[i].AdminID = adminID
			if note != "" {
				reqs[i].AdminNote = note
			}
			if status == ActionDone || status == ActionRejected {
				reqs[i].DoneAt = time.Now().Format(time.RFC3339)
			}
			found = true
			break
		}
	}
	if !found {
		return errors.New("درخواست پیدا نشد")
	}
	raw, _ := json.Marshal(reqs)
	_, err = d.Exec("UPDATE orders SET action_requests = ? WHERE id = ?", string(raw), orderID)
	return err
}

func (d *DB) GetOrderActionRequest(orderID, reqID string) (*ActionRequest, error) {
	reqs, err := d.GetOrderActionRequests(orderID)
	if err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].ID == reqID {
			return &reqs[i], nil
		}
	}
	return nil, errors.New("درخواست پیدا نشد")
}
