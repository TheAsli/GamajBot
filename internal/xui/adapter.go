// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package xui

import (
	"github.com/theasli/GamajBot/internal/vpn"
)

func (c *PanelClient) Type() vpn.PanelType {
	return vpn.PanelTypeXUI
}

func (c *PanelClient) ListInbounds() ([]vpn.Inbound, error) {
	inbounds, err := c.listInboundsRaw()
	if err != nil {
		return nil, err
	}
	out := make([]vpn.Inbound, 0, len(inbounds))
	for _, inb := range inbounds {
		out = append(out, vpn.Inbound{
			ID:       inb.ID,
			Remark:   inb.Remark,
			Protocol: inb.Protocol,
			Port:     inb.Port,
			Enable:   inb.Enable,
			Up:       inb.Up,
			Down:     inb.Down,
			Total:    inb.Total,
		})
	}
	return out, nil
}

func (c *PanelClient) GetInboundStats() ([]vpn.Inbound, error) {
	inbounds, err := c.ListInbounds()
	if err != nil {
		return nil, err
	}
	for i := range inbounds {
		clients, err := c.getClientStatsRaw(inbounds[i].ID)
		if err == nil {
			inbounds[i].ClientCount = len(clients)
		}
	}
	return inbounds, nil
}

func (c *PanelClient) GetClient(inboundID int64, email string) (*vpn.Client, error) {
	cl, err := c.getClientRaw(inboundID, email)
	if err != nil {
		return nil, err
	}
	return &vpn.Client{
		ID:         cl.ID,
		Email:      cl.Email,
		Enable:     cl.Enable,
		ExpiryTime: cl.ExpiryTime,
		TotalGB:    cl.TotalGB,
	}, nil
}

func (c *PanelClient) AddClient(req vpn.ClientAddRequest) error {
	return c.addClientRaw(ClientAddRequest{
		Client: InboundClient{
			Email:      req.Email,
			TotalGB:    req.TotalGB,
			ExpiryTime: req.ExpiryTime.UnixMilli(),
			Enable:     req.Enable,
		},
		InboundIDs: req.InboundIDs,
	})
}

func (c *PanelClient) UpdateClient(inboundID int64, client vpn.Client) error {
	return c.updateClientRaw(inboundID, InboundClient{
		ID:         client.ID,
		Email:      client.Email,
		Enable:     client.Enable,
		ExpiryTime: client.ExpiryTime,
		TotalGB:    client.TotalGB,
	})
}

func (c *PanelClient) DeleteClient(inboundID int64, clientID string) error {
	return c.deleteClientRaw(inboundID, clientID)
}

func (c *PanelClient) EnableClient(inboundID int64, clientID string, enable bool) error {
	return c.enableClientRaw(inboundID, clientID, enable)
}

func (c *PanelClient) ResetClientTraffic(inboundID int64, email string) error {
	return c.resetClientTrafficRaw(inboundID, email)
}

func (c *PanelClient) GetClientLinks(email string) ([]string, error) {
	return c.getClientLinksRaw(email)
}

func (c *PanelClient) GetClientTraffic(inboundID int64, email string) (*vpn.Client, error) {
	st, err := c.getClientTrafficRaw(inboundID, email)
	if err != nil {
		return nil, err
	}
	return &vpn.Client{
		Email:      st.Email,
		Enable:     st.Enable,
		ExpiryTime: st.ExpiryTime,
		TotalGB:    st.Total,
		Up:         st.Up,
		Down:       st.Down,
	}, nil
}
