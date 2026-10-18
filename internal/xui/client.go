// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package xui

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

type PanelClient struct {
	cfg    Config
	http   *http.Client
	mu     sync.Mutex
	csrf   string
	logged bool
}

func New(cfg Config) *PanelClient {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &PanelClient{
		cfg: cfg,
		http: &http.Client{
			Timeout:   30 * time.Second,
			Jar:       jar,
			Transport: transport,
		},
	}
}

func (c *PanelClient) base() string {
	return strings.TrimRight(c.cfg.BaseURL, "/")
}

func (c *PanelClient) do(method, path string, body []byte) (*http.Response, error) {
	fullURL := c.base() + path
	req, err := http.NewRequest(method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if c.cfg.AuthMethod == AuthToken && c.cfg.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.mu.Lock()
	csrf := c.csrf
	c.mu.Unlock()
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}

	return c.http.Do(req)
}

func (c *PanelClient) ensureLogin() error {
	if c.cfg.AuthMethod != AuthLogin {
		return nil
	}
	c.mu.Lock()
	if c.logged {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()
	return c.login()
}

func (c *PanelClient) login() error {
	loginURL := c.base() + "/login"

	resp, err := c.http.Get(loginURL)
	if err != nil {
		return fmt.Errorf("خطا در دریافت صفحه لاگین: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	c.mu.Lock()
	c.csrf = resp.Header.Get("X-CSRF-Token")
	c.mu.Unlock()

	payload, _ := json.Marshal(map[string]string{
		"username": c.cfg.Username,
		"password": c.cfg.Password,
	})

	req, err := http.NewRequest("POST", loginURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", c.csrf)

	resp2, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("خطا در ارسال لاگین: %w", err)
	}
	defer resp2.Body.Close()

	data, _ := io.ReadAll(resp2.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("پاسخ لاگین نامعتبر: %s", string(data))
	}
	if !apiResp.Success {
		return fmt.Errorf("لاگین ناموفق: %s", apiResp.Msg)
	}

	c.mu.Lock()
	c.logged = true
	c.mu.Unlock()
	return nil
}

func (c *PanelClient) listInboundsRaw() ([]Inbound, error) {
	if err := c.ensureLogin(); err != nil {
		return nil, err
	}

	resp, err := c.do("GET", "/panel/api/inbounds/list", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var apiResp struct {
		Success bool      `json:"success"`
		Msg     string    `json:"msg"`
		Obj     []Inbound `json:"obj"`
	}
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("پاسخ inbound نامعتبر: %s", string(data))
	}
	if !apiResp.Success {
		return nil, fmt.Errorf("خطا در دریافت inboundها: %s", apiResp.Msg)
	}
	return apiResp.Obj, nil
}

func (c *PanelClient) addClientRaw(req ClientAddRequest) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	payload, _ := json.Marshal(req)
	resp, err := c.do("POST", "/panel/api/clients/add", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("پاسخ addClient نامعتبر: %s", string(data))
	}
	if !apiResp.Success {
		return fmt.Errorf("addClient ناموفق: %s", apiResp.Msg)
	}
	return nil
}

func (c *PanelClient) getClientLinksRaw(email string) ([]string, error) {
	inbounds, err := c.listInboundsRaw()
	if err != nil {
		return nil, err
	}

	var links []string
	for _, inb := range inbounds {
		var settings InboundSettings
		if err := json.Unmarshal(inb.Settings, &settings); err != nil {
			continue
		}
		for _, cl := range settings.Clients {
			if cl.Email != email {
				continue
			}
			link := buildLink(inb, cl)
			if link != "" {
				links = append(links, link)
			}
		}
	}
	return links, nil
}

func (c *PanelClient) getClientRaw(inboundID int64, email string) (*InboundClient, error) {
	inbounds, err := c.listInboundsRaw()
	if err != nil {
		return nil, err
	}

	for _, inb := range inbounds {
		if inb.ID != inboundID {
			continue
		}
		var settings InboundSettings
		if err := json.Unmarshal(inb.Settings, &settings); err != nil {
			continue
		}
		for _, cl := range settings.Clients {
			if cl.Email == email {
				return &cl, nil
			}
		}
	}
	return nil, fmt.Errorf("کلاینت پیدا نشد")
}

func (c *PanelClient) updateClientRaw(inboundID int64, client InboundClient) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	settings := map[string]interface{}{
		"clients": []InboundClient{client},
	}
	settingsJSON, _ := json.Marshal(settings)

	form := url.Values{}
	form.Set("id", fmt.Sprintf("%d", inboundID))
	form.Set("settings", string(settingsJSON))

	fullURL := c.base() + "/panel/api/inbounds/updateClient/" + client.ID
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.cfg.AuthMethod == AuthToken {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
	}
	c.mu.Lock()
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	c.mu.Unlock()

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("پاسخ updateClient نامعتبر: %s", string(data))
	}
	if !apiResp.Success {
		return fmt.Errorf("updateClient ناموفق: %s", apiResp.Msg)
	}
	return nil
}

func (c *PanelClient) deleteClientRaw(inboundID int64, clientID string) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	form := url.Values{}
	form.Set("id", fmt.Sprintf("%d", inboundID))

	fullURL := fmt.Sprintf("%s/panel/api/inbounds/%d/delClient/%s", c.base(), inboundID, clientID)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.cfg.AuthMethod == AuthToken {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
	}
	c.mu.Lock()
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	c.mu.Unlock()

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("پاسخ deleteClient نامعتبر: %s", string(data))
	}
	if !apiResp.Success {
		return fmt.Errorf("deleteClient ناموفق: %s", apiResp.Msg)
	}
	return nil
}

func (c *PanelClient) enableClientRaw(inboundID int64, clientID string, enable bool) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	action := "enable"
	if !enable {
		action = "disable"
	}

	form := url.Values{}
	form.Set("id", fmt.Sprintf("%d", inboundID))

	fullURL := fmt.Sprintf("%s/panel/api/inbounds/%sClient/%s", c.base(), action, clientID)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.cfg.AuthMethod == AuthToken {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
	}
	c.mu.Lock()
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	c.mu.Unlock()

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("پاسخ %s نامعتبر: %s", action, string(data))
	}
	if !apiResp.Success {
		return fmt.Errorf("%s ناموفق: %s", action, apiResp.Msg)
	}
	return nil
}

func (c *PanelClient) resetClientTrafficRaw(inboundID int64, email string) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	form := url.Values{}
	form.Set("id", fmt.Sprintf("%d", inboundID))
	form.Set("email", email)

	fullURL := c.base() + fmt.Sprintf("/panel/api/inbounds/%d/resetClientTraffic/%s", inboundID, email)
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.cfg.AuthMethod == AuthToken {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIToken)
	}
	c.mu.Lock()
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	c.mu.Unlock()

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("پاسخ reset نامعتبر: %s", string(data))
	}
	if !apiResp.Success {
		return fmt.Errorf("reset ناموفق: %s", apiResp.Msg)
	}
	return nil
}

func (c *PanelClient) getClientStatsRaw(inboundID int64) ([]ClientStats, error) {
	if err := c.ensureLogin(); err != nil {
		return nil, err
	}

	resp, err := c.do("GET", fmt.Sprintf("/panel/api/inbounds/getClientTraffics/%d", inboundID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var apiResp struct {
		Success bool          `json:"success"`
		Msg     string        `json:"msg"`
		Obj     []ClientStats `json:"obj"`
	}
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("پاسخ آمار نامعتبر: %s", string(data))
	}
	if !apiResp.Success {
		return nil, fmt.Errorf("خطا در آمار: %s", apiResp.Msg)
	}
	return apiResp.Obj, nil
}

func (c *PanelClient) getClientTrafficRaw(inboundID int64, email string) (*ClientStats, error) {
	if err := c.ensureLogin(); err != nil {
		return nil, err
	}

	resp, err := c.do("GET", fmt.Sprintf("/panel/api/inbounds/getClientTraffics/%s", url.PathEscape(email)), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var apiResp struct {
		Success bool         `json:"success"`
		Msg     string       `json:"msg"`
		Obj     *ClientStats `json:"obj"`
	}
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("پاسخ traffic نامعتبر: %s", string(data))
	}
	if !apiResp.Success || apiResp.Obj == nil {
		return nil, fmt.Errorf("traffic ناموفق: %s", apiResp.Msg)
	}
	return apiResp.Obj, nil
}

func buildLink(inb Inbound, cl InboundClient) string {
	host := extractHost(inb)
	switch inb.Protocol {
	case "vless":
		return fmt.Sprintf("vless://%s@%s:%d?type=tcp&security=none#%s",
			cl.ID, host, inb.Port, url.QueryEscape(inb.Remark))
	case "vmess":
		return fmt.Sprintf("vmess://%s@%s:%d#%s",
			cl.ID, host, inb.Port, url.QueryEscape(inb.Remark))
	case "trojan":
		return fmt.Sprintf("trojan://%s@%s:%d#%s",
			cl.ID, host, inb.Port, url.QueryEscape(inb.Remark))
	}
	return ""
}

func extractHost(inb Inbound) string {
	if inb.Listen != "" && inb.Listen != "0.0.0.0" && inb.Listen != "::" {
		return inb.Listen
	}
	return "YOUR_SERVER_IP"
}
