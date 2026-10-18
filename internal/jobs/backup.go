// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package jobs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/notify"

	AsliBot "gopkg.in/telebot.v4"
)

func StartBackupJob(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	go func() {
		time.Sleep(2 * time.Minute)
		runBackup(bot, cfg, db)

		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			runBackup(bot, cfg, db)
		}
	}()
}

func runBackup(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) {
	dbPath := "data/bot.db"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return
	}

	backupDir := "data/backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("⚠️ خطا در ساخت پوشه بکاپ: %v", err)
		notify.AdminLog(bot, cfg, db, notify.LogTopicErrors,
			fmt.Sprintf("⚠️ <b>خطا در بکاپ</b>\n\n<code>%v</code>", err))
		return
	}

	backupName := fmt.Sprintf("bot_%s.db", time.Now().Format("20060102_150405"))
	backupPath := filepath.Join(backupDir, backupName)

	src, err := os.ReadFile(dbPath)
	if err != nil {
		log.Printf("⚠️ خطا در خواندن دیتابیس: %v", err)
		return
	}

	if err := os.WriteFile(backupPath, src, 0644); err != nil {
		log.Printf("⚠️ خطا در نوشتن بکاپ: %v", err)
		notify.AdminLog(bot, cfg, db, notify.LogTopicErrors,
			fmt.Sprintf("⚠️ <b>خطا در نوشتن بکاپ</b>\n\n<code>%v</code>", err))
		return
	}

	cleanOldBackups(backupDir, 7)

	log.Printf("✅ بکاپ ساخته شد: %s", backupPath)
}

func cleanOldBackups(dir string, keepDays int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
