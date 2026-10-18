CREATE TABLE IF NOT EXISTS users (
                                     id            INTEGER PRIMARY KEY,
                                     first_name    TEXT NOT NULL DEFAULT '',
                                     username      TEXT NOT NULL DEFAULT '',
                                     balance_usd   REAL NOT NULL DEFAULT 0,
                                     banned        INTEGER NOT NULL DEFAULT 0,
                                     ban_reason    TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE IF NOT EXISTS transactions (
                                            id            INTEGER PRIMARY KEY AUTOINCREMENT,
                                            user_id       INTEGER NOT NULL,
                                            pay_id        TEXT NOT NULL DEFAULT '',
                                            amount_usd    REAL NOT NULL DEFAULT 0,
                                            amount_toman  INTEGER NOT NULL DEFAULT 0,
                                            type          TEXT NOT NULL,
                                            status        TEXT NOT NULL,
                                            description   TEXT NOT NULL DEFAULT '',
                                            gateway       TEXT NOT NULL DEFAULT '',
                                            created_at    TEXT NOT NULL,
                                            expires_at    TEXT NOT NULL DEFAULT '',
                                            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    ) STRICT;

CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id, created_at);

CREATE TABLE IF NOT EXISTS admins (
                                      user_id       INTEGER PRIMARY KEY,
                                      role          TEXT NOT NULL,
                                      dept_ids      TEXT NOT NULL DEFAULT '[]',
                                      added_by      INTEGER NOT NULL DEFAULT 0,
                                      added_at      TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE IF NOT EXISTS departments (
                                           id            INTEGER PRIMARY KEY AUTOINCREMENT,
                                           name          TEXT NOT NULL,
                                           emoji         TEXT NOT NULL DEFAULT '',
                                           description   TEXT NOT NULL DEFAULT '',
                                           active        INTEGER NOT NULL DEFAULT 1
) STRICT;

CREATE TABLE IF NOT EXISTS tickets (
                                       id              INTEGER PRIMARY KEY AUTOINCREMENT,
                                       user_id         INTEGER NOT NULL,
                                       user_first_name TEXT NOT NULL DEFAULT '',
                                       user_username   TEXT NOT NULL DEFAULT '',
                                       dept_id         INTEGER NOT NULL DEFAULT 0,
                                       subject         TEXT NOT NULL DEFAULT '',
                                       status          TEXT NOT NULL DEFAULT 'open',
                                       priority        TEXT NOT NULL DEFAULT 'normal',
                                       assigned_to     INTEGER NOT NULL DEFAULT 0,
                                       created_at      TEXT NOT NULL,
                                       updated_at      TEXT NOT NULL,
                                       closed_at       TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE IF NOT EXISTS ticket_messages (
                                               id          INTEGER PRIMARY KEY AUTOINCREMENT,
                                               ticket_id   INTEGER NOT NULL,
                                               from_id     INTEGER NOT NULL,
                                               is_admin    INTEGER NOT NULL DEFAULT 0,
                                               text        TEXT NOT NULL DEFAULT '',
                                               created_at  TEXT NOT NULL,
                                               FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE
    ) STRICT;

CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket ON ticket_messages(ticket_id);

CREATE TABLE IF NOT EXISTS categories (
                                          id                INTEGER PRIMARY KEY AUTOINCREMENT,
                                          parent_id         INTEGER NOT NULL DEFAULT 0,
                                          name              TEXT NOT NULL,
                                          emoji             TEXT NOT NULL DEFAULT '',
                                          active            INTEGER NOT NULL DEFAULT 1,
                                          sort_order        INTEGER NOT NULL DEFAULT 0,
                                          delivery_fields   TEXT NOT NULL DEFAULT '[]',
                                          service_actions   TEXT NOT NULL DEFAULT '[]'
) STRICT;

CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);

CREATE TABLE IF NOT EXISTS products (
                                        id              INTEGER PRIMARY KEY AUTOINCREMENT,
                                        category_id     INTEGER NOT NULL,
                                        name            TEXT NOT NULL,
                                        description     TEXT NOT NULL DEFAULT '',
                                        price_usd       REAL NOT NULL DEFAULT 0,
                                        available       INTEGER NOT NULL DEFAULT 1,
                                        delivery_type   TEXT NOT NULL DEFAULT 'manual',
                                        active          INTEGER NOT NULL DEFAULT 1,
                                        product_type    TEXT NOT NULL DEFAULT '',
                                        xui_panel_id    INTEGER NOT NULL DEFAULT 0,
                                        xui_inbound_ids TEXT NOT NULL DEFAULT '[]',
                                        traffic_gb      INTEGER NOT NULL DEFAULT 0,
                                        duration_days   INTEGER NOT NULL DEFAULT 0,
                                        created_at      TEXT NOT NULL,
                                        FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
    ) STRICT;

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);

CREATE TABLE IF NOT EXISTS orders (
                                      id                TEXT PRIMARY KEY,
                                      user_id           INTEGER NOT NULL,
                                      user_name         TEXT NOT NULL DEFAULT '',
                                      product_id        INTEGER NOT NULL,
                                      product_name      TEXT NOT NULL DEFAULT '',
                                      category_id       INTEGER NOT NULL DEFAULT 0,
                                      price_usd         REAL NOT NULL DEFAULT 0,
                                      status            TEXT NOT NULL DEFAULT 'pending',
                                      service_info      TEXT NOT NULL DEFAULT '',
                                      delivery_data     TEXT NOT NULL DEFAULT '{}',
                                      reject_note       TEXT NOT NULL DEFAULT '',
                                      code_value        TEXT NOT NULL DEFAULT '',
                                      client_email      TEXT NOT NULL DEFAULT '',
                                      client_links      TEXT NOT NULL DEFAULT '[]',
                                      action_requests   TEXT NOT NULL DEFAULT '[]',
                                      created_at        TEXT NOT NULL,
                                      delivered_at      TEXT NOT NULL DEFAULT '',
                                      code_requested_at TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

CREATE TABLE IF NOT EXISTS manual_charge_requests (
                                                      id            TEXT PRIMARY KEY,
                                                      requester_id  INTEGER NOT NULL,
                                                      requester_name TEXT NOT NULL DEFAULT '',
                                                      target_user_id INTEGER NOT NULL,
                                                      target_name   TEXT NOT NULL DEFAULT '',
                                                      amount_usd    REAL NOT NULL DEFAULT 0,
                                                      amount_toman  INTEGER NOT NULL DEFAULT 0,
                                                      status        TEXT NOT NULL DEFAULT 'pending',
                                                      reject_reason TEXT NOT NULL DEFAULT '',
                                                      review_note   TEXT NOT NULL DEFAULT '',
                                                      gateway       TEXT NOT NULL DEFAULT '',
                                                      pay_id        TEXT NOT NULL DEFAULT '',
                                                      tx_ref        TEXT NOT NULL DEFAULT '',
                                                      receipt_id    TEXT NOT NULL DEFAULT '',
                                                      receipt_type  TEXT NOT NULL DEFAULT '',
                                                      created_at    TEXT NOT NULL,
                                                      reviewed_at   TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE INDEX IF NOT EXISTS idx_manual_charge_status ON manual_charge_requests(status);

CREATE TABLE IF NOT EXISTS xui_panels (
                                          id          INTEGER PRIMARY KEY AUTOINCREMENT,
                                          name        TEXT NOT NULL,
                                          base_url    TEXT NOT NULL,
                                          server_ip   TEXT NOT NULL DEFAULT '',
                                          auth_method TEXT NOT NULL DEFAULT 'token',
                                          api_token   TEXT NOT NULL DEFAULT '',
                                          username    TEXT NOT NULL DEFAULT '',
                                          password    TEXT NOT NULL DEFAULT '',
                                          active      INTEGER NOT NULL DEFAULT 1,
                                          created_at  TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS settings (
                                        key   TEXT PRIMARY KEY,
                                        value TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS payment_gateways (
                                                key   TEXT PRIMARY KEY,
                                                value TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS subscription_plans (
                                                  id              INTEGER PRIMARY KEY AUTOINCREMENT,
                                                  type            TEXT NOT NULL DEFAULT 'fixed',
                                                  name            TEXT NOT NULL,
                                                  description     TEXT NOT NULL DEFAULT '',
                                                  price_usd       REAL NOT NULL DEFAULT 0,
                                                  traffic_gb      INTEGER NOT NULL DEFAULT 0,
                                                  duration_days   INTEGER NOT NULL DEFAULT 0,
                                                  limit_ip        INTEGER NOT NULL DEFAULT 0,
                                                  xui_panel_id    INTEGER NOT NULL DEFAULT 0,
                                                  xui_inbound_ids TEXT NOT NULL DEFAULT '[]',
                                                  active          INTEGER NOT NULL DEFAULT 1,
                                                  created_at      TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS dynamic_prices (
                                              id              INTEGER PRIMARY KEY AUTOINCREMENT,
                                              traffic_gb      INTEGER NOT NULL,
                                              duration_days   INTEGER NOT NULL,
                                              limit_ip        INTEGER NOT NULL,
                                              price_usd       REAL NOT NULL DEFAULT 0,
                                              xui_panel_id    INTEGER NOT NULL DEFAULT 0,
                                              xui_inbound_ids TEXT NOT NULL DEFAULT '[]'
) STRICT;

CREATE TABLE IF NOT EXISTS coupons (
                                       id            INTEGER PRIMARY KEY AUTOINCREMENT,
                                       code          TEXT NOT NULL UNIQUE,
                                       type          TEXT NOT NULL DEFAULT 'percent',
                                       value         REAL NOT NULL DEFAULT 0,
                                       max_uses      INTEGER NOT NULL DEFAULT 0,
                                       used_count    INTEGER NOT NULL DEFAULT 0,
                                       min_purchase  REAL NOT NULL DEFAULT 0,
                                       expires_at    TEXT NOT NULL DEFAULT '',
                                       active        INTEGER NOT NULL DEFAULT 1,
                                       created_at    TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS coupon_usages (
                                             id          INTEGER PRIMARY KEY AUTOINCREMENT,
                                             coupon_id   INTEGER NOT NULL,
                                             user_id     INTEGER NOT NULL,
                                             order_id    TEXT NOT NULL DEFAULT '',
                                             amount_usd  REAL NOT NULL DEFAULT 0,
                                             created_at  TEXT NOT NULL,
                                             FOREIGN KEY (coupon_id) REFERENCES coupons(id) ON DELETE CASCADE
    ) STRICT;

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupon_usages_user ON coupon_usages(user_id, coupon_id);

CREATE TABLE IF NOT EXISTS log_topics (
                                          topic_key   TEXT PRIMARY KEY,
                                          thread_id   INTEGER NOT NULL,
                                          title       TEXT NOT NULL DEFAULT '',
                                          created_at  TEXT NOT NULL
) STRICT;