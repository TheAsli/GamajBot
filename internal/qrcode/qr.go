// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package qrcode

import (
	"bytes"
	"image/png"

	qrcode "github.com/skip2/go-qrcode"
)

func GeneratePNG(text string, size int) ([]byte, error) {
	if size <= 0 {
		size = 512
	}
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	return q.PNG(size)
}

func GenerateRawPNG(text string, size int) ([]byte, error) {
	if size <= 0 {
		size = 512
	}
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	img := q.Image(size)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
