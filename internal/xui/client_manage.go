// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package xui

type ClientStats struct {
	ID         int64  `json:"id"`
	InboundID  int64  `json:"inboundId"`
	Enable     bool   `json:"enable"`
	Email      string `json:"email"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
	ExpiryTime int64  `json:"expiryTime"`
	Total      int64  `json:"total"`
	Reset      int    `json:"reset"`
}

type InboundTraffic struct {
	ID          int64  `json:"id"`
	Up          int64  `json:"up"`
	Down        int64  `json:"down"`
	Total       int64  `json:"total"`
	Remark      string `json:"remark"`
	Enable      bool   `json:"enable"`
	ExpiryTime  int64  `json:"expiryTime"`
	ClientCount int    `json:"clientCount"`
}

func (c *PanelClient) GetClientStats(inboundID int64) ([]ClientStats, error) {
	return c.getClientStatsRaw(inboundID)
}

func (c *PanelClient) GetSubLink(email string) string {
	return c.base() + "/sub/" + email
}
