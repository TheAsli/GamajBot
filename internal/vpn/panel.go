// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package vpn

import "time"

type PanelType string

const (
	PanelTypeXUI     PanelType = "3x-ui"
	PanelTypeMarzban PanelType = "marzban"
	PanelTypeHiddify PanelType = "hiddify"
)

type PanelConfig struct {
	BaseURL    string
	AuthMethod string
	APIToken   string
	Username   string
	Password   string
}

type Inbound struct {
	ID          int64
	Remark      string
	Protocol    string
	Port        int64
	Enable      bool
	ClientCount int
	Up          int64
	Down        int64
	Total       int64
}

type Client struct {
	ID         string
	Email      string
	Enable     bool
	ExpiryTime int64
	TotalGB    int64
	Up         int64
	Down       int64
}

type ClientAddRequest struct {
	Email      string
	TotalGB    int64
	ExpiryTime time.Time
	Enable     bool
	InboundIDs []int64
}

type Panel interface {
	Type() PanelType
	ListInbounds() ([]Inbound, error)
	GetInboundStats() ([]Inbound, error)
	GetClient(inboundID int64, email string) (*Client, error)
	AddClient(req ClientAddRequest) error
	UpdateClient(inboundID int64, client Client) error
	DeleteClient(inboundID int64, clientID string) error
	EnableClient(inboundID int64, clientID string, enable bool) error
	ResetClientTraffic(inboundID int64, email string) error
	GetClientLinks(email string) ([]string, error)
	GetClientTraffic(inboundID int64, email string) (*Client, error)
}
