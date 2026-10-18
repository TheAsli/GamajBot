// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package bot

import (
	"log"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/handlers"
	"github.com/theasli/GamajBot/internal/handlers/admin"
	"github.com/theasli/GamajBot/internal/handlers/notify"
	"github.com/theasli/GamajBot/internal/handlers/support"
	"github.com/theasli/GamajBot/internal/jobs"
	"github.com/theasli/GamajBot/internal/menus"
	"github.com/theasli/GamajBot/internal/payment"
	"github.com/theasli/GamajBot/internal/rate"
	"github.com/theasli/GamajBot/internal/webhook"

	AsliBot "gopkg.in/telebot.v4"
)

func New(cfg *config.Config) (*AsliBot.Bot, error) {
	db, err := database.New("data/wallets.json")
	if err != nil {
		return nil, err
	}

	if err := rate.FetchRates(); err != nil {
		log.Printf("⚠️ نرخ لحظه‌ای دریافت نشد: %v", err)
	}

	g := db.GetPaymentGateways()
	pc := payment.NewClient(g.TetraminatorAPIKey, g.TetraminatorBaseURL)
	if g.TetraminatorBaseURL == "" {
		pc = payment.NewClient(cfg.TETRAMINATOR_API_KEY, cfg.TETRAMINATOR_BASE_URL)
	}

	pref := AsliBot.Settings{
		Token:  cfg.BOT_TOKEN,
		Poller: &AsliBot.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := AsliBot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	n := notify.New(b, cfg, db)
	admin.InitNotifier(n)
	admin.SetBotInstance(b)
	admin.SetMainAdminID(cfg.ADMIN_ID)
	support.Init(n)

	handlers.InitWallet(db, pc, cfg)
	menus.Setup(cfg.ASLICODE_TEXT)
	handlers.Register(b, cfg, db)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			_ = rate.FetchRates()
		}
	}()

	admin.StartExpiryNotifier(b, cfg, db)

	jobs.StartBackupJob(b, cfg, db)
	jobs.StartSyncJob(b, cfg, db)
	jobs.StartTrafficJob(b, cfg, db)
	jobs.StartDailyReportJob(b, cfg, db)

	wh := webhook.NewServer(b, cfg, db)
	wh.Start()

	return b, nil
}
