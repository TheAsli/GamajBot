// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/theasli/GamajBot/internal/config"
	"github.com/theasli/GamajBot/internal/database"
	"github.com/theasli/GamajBot/internal/notify"
	"github.com/theasli/GamajBot/internal/payment"

	AsliBot "gopkg.in/telebot.v4"
)

type Server struct {
	bot     *AsliBot.Bot
	cfg     *config.Config
	db      *database.DB
	aban    *payment.AbanClient
	httpSrv *http.Server
}

func NewServer(bot *AsliBot.Bot, cfg *config.Config, db *database.DB) *Server {
	g := db.GetPaymentGateways()
	s := &Server{
		bot: bot,
		cfg: cfg,
		db:  db,
	}
	if g.AbanAPIKey != "" {
		s.aban = payment.NewAbanClient(g.AbanAPIKey)
	}
	return s
}

func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/abangateway", s.handleAbanWebhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})

	port := s.cfg.WEBHOOK_PORT
	if port == "" {
		port = "8080"
	}

	s.httpSrv = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("🌐 Webhook server on :%s", port)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("⚠️ Webhook server error: %v", err)
		}
	}()
}

func (s *Server) Stop() {
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(ctx)
	}
}

func (s *Server) handleAbanWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if s.aban == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var event struct {
		InvoiceID string `json:"invoice_id"`
		Status    string `json:"status"`
		OrderID   string `json:"order_id"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, err := s.aban.Inquiry(r.Context(), event.InvoiceID)
	if err != nil {
		log.Printf("⚠️ Aban inquiry failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if status != payment.StatusPaid {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := s.aban.Verify(r.Context(), event.InvoiceID); err != nil {
		log.Printf("⚠️ Aban verify failed: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, _, err := s.db.ConfirmCharge(event.InvoiceID); err != nil {
		log.Printf("⚠️ Confirm charge failed: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	notify.AdminLog(s.bot, s.cfg, s.db, notify.LogTopicFinance,
		fmt.Sprintf("✅ پرداخت آبان گیت‌وی تایید شد\nInvoice: <code>%s</code>", event.InvoiceID))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
