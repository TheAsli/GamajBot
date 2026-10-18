// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package payment

import "sync"

var (
	cubeClient *CubePayClient
	cubeMu     sync.Mutex
)

// GetCubeClient — گرفتن کلاینت CubePay (lazy)
func GetCubeClient(token string) *CubePayClient {
	cubeMu.Lock()
	defer cubeMu.Unlock()
	if cubeClient == nil || cubeClient.apiToken != token {
		cubeClient = NewCubePayClient(token)
	}
	return cubeClient
}

// InvalidateCubeClient — پاک کردن کلاینت بعد از تغییر توکن
func InvalidateCubeClient() {
	cubeMu.Lock()
	defer cubeMu.Unlock()
	cubeClient = nil
}
