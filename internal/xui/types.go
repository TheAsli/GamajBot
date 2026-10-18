// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package xui

import "encoding/json"

type AuthMethod string

const (
	AuthToken AuthMethod = "token"
	AuthLogin AuthMethod = "login"
)

type Config struct {
	BaseURL    string
	AuthMethod AuthMethod
	APIToken   string
	Username   string
	Password   string
}

type Inbound struct {
	ID             int64           `json:"id"`
	Up             int64           `json:"up"`
	Down           int64           `json:"down"`
	Total          int64           `json:"total"`
	Remark         string          `json:"remark"`
	Enable         bool            `json:"enable"`
	ExpiryTime     int64           `json:"expiryTime"`
	Listen         string          `json:"listen"`
	Port           int64           `json:"port"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings"`
	StreamSettings json.RawMessage `json:"streamSettings"`
	Tag            string          `json:"tag"`
	Sniffing       json.RawMessage `json:"sniffing"`
}

type InboundSettings struct {
	Clients []InboundClient `json:"clients"`
}

type InboundClient struct {
	ID         string `json:"id,omitempty"`
	Flow       string `json:"flow,omitempty"`
	Email      string `json:"email"`
	LimitIP    int    `json:"limitIp"`
	TotalGB    int64  `json:"totalGB"`
	ExpiryTime int64  `json:"expiryTime"`
	Enable     bool   `json:"enable"`
	TgID       int64  `json:"tgId"`
	SubID      string `json:"subId,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Reset      int    `json:"reset"`
}

type ClientAddRequest struct {
	Client     InboundClient `json:"client"`
	InboundIDs []int64       `json:"inboundIds"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj,omitempty"`
}
