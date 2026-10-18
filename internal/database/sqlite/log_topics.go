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

var ErrLogTopicNotFound = errors.New("تاپیک لاگ پیدا نشد")

type LogTopic struct {
	TopicKey  string
	ThreadID  int64
	Title     string
	CreatedAt string
}

func (d *DB) SetLogTopic(topicKey string, threadID int64, title string) error {
	_, err := d.Exec(
		`INSERT INTO log_topics (topic_key, thread_id, title, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(topic_key) DO UPDATE SET thread_id = excluded.thread_id, title = excluded.title`,
		topicKey, threadID, title, time.Now().Format(time.RFC3339),
	)
	return err
}

func (d *DB) GetLogTopic(topicKey string) (*LogTopic, error) {
	var t LogTopic
	err := d.QueryRow(
		"SELECT topic_key, thread_id, title, created_at FROM log_topics WHERE topic_key = ?", topicKey,
	).Scan(&t.TopicKey, &t.ThreadID, &t.Title, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrLogTopicNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *DB) GetAllLogTopics() ([]*LogTopic, error) {
	rows, err := d.Query("SELECT topic_key, thread_id, title, created_at FROM log_topics ORDER BY topic_key")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*LogTopic
	for rows.Next() {
		var t LogTopic
		if err := rows.Scan(&t.TopicKey, &t.ThreadID, &t.Title, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &t)
	}
	return out, nil
}

func (d *DB) DeleteLogTopic(topicKey string) error {
	_, err := d.Exec("DELETE FROM log_topics WHERE topic_key = ?", topicKey)
	return err
}

func (d *DB) DeleteAllLogTopics() error {
	_, err := d.Exec("DELETE FROM log_topics")
	return err
}
