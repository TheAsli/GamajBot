// SPDX-License-Identifier: AGPL-3.0-only
// Copyright (C) 2026 AsliCode
// Part of GamajBot — licensed under the GNU Affero General Public License v3.0.
// Coded by AsliCode

package sqlite

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTicketNotFound = errors.New("تیکت پیدا نشد")

type TicketStatus string

const (
	TicketOpen        TicketStatus = "open"
	TicketPending     TicketStatus = "pending"
	TicketAnswered    TicketStatus = "answered"
	TicketWaitingUser TicketStatus = "waiting_user"
	TicketClosed      TicketStatus = "closed"
)

type TicketPriority string

const (
	PriorityLow    TicketPriority = "low"
	PriorityNormal TicketPriority = "normal"
	PriorityHigh   TicketPriority = "high"
	PriorityUrgent TicketPriority = "urgent"
)

type TicketMessage struct {
	FromID    int64
	IsAdmin   bool
	Text      string
	CreatedAt string
}

type Ticket struct {
	ID            int64
	UserID        int64
	UserFirstName string
	UserUsername  string
	DeptID        int64
	Subject       string
	Status        TicketStatus
	Priority      TicketPriority
	Messages      []TicketMessage
	AssignedTo    int64
	CreatedAt     string
	UpdatedAt     string
	ClosedAt      string
}

func (d *DB) CreateTicket(userID int64, firstName, username string, deptID int64, subject, text string, priority TicketPriority) (*Ticket, error) {
	if priority == "" {
		priority = PriorityNormal
	}
	now := time.Now().Format(time.RFC3339)

	tx, err := d.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO tickets (user_id, user_first_name, user_username, dept_id, subject, status, priority, assigned_to, created_at, updated_at, closed_at)
		 VALUES (?, ?, ?, ?, ?, 'open', ?, 0, ?, ?, '')`,
		userID, firstName, username, deptID, subject, string(priority), now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	if _, err := tx.Exec(
		"INSERT INTO ticket_messages (ticket_id, from_id, is_admin, text, created_at) VALUES (?, ?, 0, ?, ?)",
		id, userID, text, now,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &Ticket{
		ID: id, UserID: userID, UserFirstName: firstName, UserUsername: username,
		DeptID: deptID, Subject: subject, Status: TicketOpen, Priority: priority,
		Messages:  []TicketMessage{{FromID: userID, IsAdmin: false, Text: text, CreatedAt: now}},
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (d *DB) GetTicket(id int64) (*Ticket, error) {
	var t Ticket
	err := d.QueryRow(
		`SELECT id, user_id, user_first_name, user_username, dept_id, subject, status, priority, assigned_to, created_at, updated_at, closed_at
		 FROM tickets WHERE id = ?`, id,
	).Scan(&t.ID, &t.UserID, &t.UserFirstName, &t.UserUsername, &t.DeptID, &t.Subject, &t.Status, &t.Priority, &t.AssignedTo, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTicketNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := d.Query(
		"SELECT from_id, is_admin, text, created_at FROM ticket_messages WHERE ticket_id = ? ORDER BY id", id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m TicketMessage
		var isAdmin int
		if err := rows.Scan(&m.FromID, &isAdmin, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.IsAdmin = isAdmin == 1
		t.Messages = append(t.Messages, m)
	}
	return &t, nil
}

func (d *DB) AddTicketMessage(ticketID, fromID int64, isAdmin bool, text string) error {
	now := time.Now().Format(time.RFC3339)
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"INSERT INTO ticket_messages (ticket_id, from_id, is_admin, text, created_at) VALUES (?, ?, ?, ?, ?)",
		ticketID, fromID, boolToInt(isAdmin), text, now,
	); err != nil {
		return err
	}

	newStatus := "waiting_user"
	if isAdmin {
		newStatus = "answered"
	}
	if _, err := tx.Exec("UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?", newStatus, now, ticketID); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *DB) SetTicketStatus(id int64, status TicketStatus) error {
	now := time.Now().Format(time.RFC3339)
	closedAt := ""
	if status == TicketClosed {
		closedAt = now
	}
	_, err := d.Exec(
		"UPDATE tickets SET status = ?, updated_at = ?, closed_at = ? WHERE id = ?",
		string(status), now, closedAt, id,
	)
	return err
}

func (d *DB) CloseTicket(id int64) error {
	return d.SetTicketStatus(id, TicketClosed)
}

func (d *DB) AssignTicket(ticketID, adminID int64) error {
	_, err := d.Exec("UPDATE tickets SET assigned_to = ?, updated_at = ? WHERE id = ?", adminID, time.Now().Format(time.RFC3339), ticketID)
	return err
}

func (d *DB) SetTicketPriority(ticketID int64, priority TicketPriority) error {
	_, err := d.Exec("UPDATE tickets SET priority = ?, updated_at = ? WHERE id = ?", string(priority), time.Now().Format(time.RFC3339), ticketID)
	return err
}

func (d *DB) TransferTicket(ticketID, newDeptID int64) error {
	_, err := d.Exec("UPDATE tickets SET dept_id = ?, updated_at = ? WHERE id = ?", newDeptID, time.Now().Format(time.RFC3339), ticketID)
	return err
}

func (d *DB) GetUserTickets(userID int64) ([]*Ticket, error) {
	rows, err := d.Query(
		`SELECT id, user_id, user_first_name, user_username, dept_id, subject, status, priority, assigned_to, created_at, updated_at, closed_at
		 FROM tickets WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(d, rows)
}

func (d *DB) GetAllTickets(onlyOpen bool) ([]*Ticket, error) {
	q := `SELECT id, user_id, user_first_name, user_username, dept_id, subject, status, priority, assigned_to, created_at, updated_at, closed_at FROM tickets`
	if onlyOpen {
		q += " WHERE status != 'closed'"
	}
	q += " ORDER BY created_at DESC"
	rows, err := d.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(d, rows)
}

func (d *DB) GetTicketsByDept(deptID int64, onlyOpen bool) ([]*Ticket, error) {
	q := `SELECT id, user_id, user_first_name, user_username, dept_id, subject, status, priority, assigned_to, created_at, updated_at, closed_at FROM tickets WHERE dept_id = ?`
	if onlyOpen {
		q += " AND status != 'closed'"
	}
	q += " ORDER BY created_at DESC"
	rows, err := d.Query(q, deptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(d, rows)
}

func scanTickets(d *DB, rows *sql.Rows) ([]*Ticket, error) {
	var out []*Ticket
	for rows.Next() {
		var t Ticket
		if err := rows.Scan(&t.ID, &t.UserID, &t.UserFirstName, &t.UserUsername, &t.DeptID, &t.Subject, &t.Status, &t.Priority, &t.AssignedTo, &t.CreatedAt, &t.UpdatedAt, &t.ClosedAt); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	rows.Close()

	for _, t := range out {
		msgRows, err := d.Query(
			"SELECT from_id, is_admin, text, created_at FROM ticket_messages WHERE ticket_id = ? ORDER BY id", t.ID,
		)
		if err != nil {
			continue
		}
		for msgRows.Next() {
			var m TicketMessage
			var isAdmin int
			if err := msgRows.Scan(&m.FromID, &isAdmin, &m.Text, &m.CreatedAt); err == nil {
				m.IsAdmin = isAdmin == 1
				t.Messages = append(t.Messages, m)
			}
		}
		msgRows.Close()
	}
	return out, nil
}

type TicketStats struct {
	Open        int
	Pending     int
	Answered    int
	WaitingUser int
	Closed      int
	Total       int
}

func (d *DB) GetTicketStats() (*TicketStats, error) {
	s := &TicketStats{}
	rows, err := d.Query("SELECT status, COUNT(*) FROM tickets GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var cnt int
		if err := rows.Scan(&status, &cnt); err != nil {
			return nil, err
		}
		s.Total += cnt
		switch TicketStatus(status) {
		case TicketOpen:
			s.Open = cnt
		case TicketPending:
			s.Pending = cnt
		case TicketAnswered:
			s.Answered = cnt
		case TicketWaitingUser:
			s.WaitingUser = cnt
		case TicketClosed:
			s.Closed = cnt
		}
	}
	return s, nil
}
