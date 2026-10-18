<div align="center">

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./architecture.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](../fa/architecture.md)

[![README](https://img.shields.io/badge/README-English-0d6efd?style=flat-square&logo=readme&logoColor=white)](../../README.md)
[![Usage](https://img.shields.io/badge/Usage-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./usage.md)
[![API](https://img.shields.io/badge/API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./API.md)
[![Architecture](https://img.shields.io/badge/Architecture-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./architecture.md)
[![Setup](https://img.shields.io/badge/Setup-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./setup.md)
[![Structure](https://img.shields.io/badge/Structure-20c997?style=flat-square&logo=files&logoColor=white)](../../STRUCTURE.md)
[![Changelog](https://img.shields.io/badge/Changelog-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](../../CHANGELOG.md)
[![Contributing](https://img.shields.io/badge/Contributing-0d6efd?style=flat-square&logo=git&logoColor=white)](../../CONTRIBUTING.md)

</div>

# Architecture

## 📊 System overview

```mermaid
graph TD
    subgraph Telegram
        TG[Telegram Bot API] -->|updates| BOT[Bot process]
        BOT -->|replies| TG
    end

    subgraph Application
        BOT --> MW[middleware<br/>ban · maintenance · force join]
        MW --> H[handlers<br/>user · admin · support]
        H --> DB[(SQLite<br/>data/bot.db)]
        H --> I18N[i18n]
        H --> PAY[payment drivers]
        H --> XUI[3x-ui panel client]
    end

    subgraph External
        PAY --> GW[payment gateways]
        GW -->|callbacks| WH[webhook server]
        WH --> H
        RATE[jobs / rate] --> API[dollar rate API]
        XUI --> PANEL[3x-ui panel]
    end

    subgraph Background
        JOBS[jobs] -->|backup| BK[data/backups]
        JOBS -->|sync · traffic · report| H
    end
```

## 📦 Package responsibilities

| Package | Responsibility | Depends on |
| :--- | :--- | :--- |
| `internal/bot` | Assemble the bot, start polling, webhook server and jobs | config, handlers, database, jobs |
| `internal/config` | Load and validate `.env` | godotenv |
| `internal/database` | Persistence facade and SQLite implementation | sqlite driver |
| `internal/handlers` | Translate updates into domain actions | database, i18n, payment, xui |
| `internal/i18n` | Message lookup for `fa` and `en` | — |
| `internal/menus` | Keyboard construction | i18n |
| `internal/middleware` | Pre-handler gates | database, config |
| `internal/payment` | Gateway drivers behind one interface | config |
| `internal/rate` | Dollar rate with cache and margin | config |
| `internal/jobs` | Scheduled work | database, notify |
| `internal/webhook` | HTTP callback receiver | handlers, payment |
| `internal/xui` | 3x-ui API client | config |
| `internal/notify` | Outbound notifications and audit log | database |

### Layering rule

`handlers` and `jobs` may depend on `database`, `payment`, `rate`, `i18n` and `xui`.
`database`, `payment`, `rate` and `i18n` must never import `handlers`.
`config` imports nothing from `internal` except `isenv`.

---

## 💳 Charge flow

```mermaid
sequenceDiagram
    participant U as User
    participant B as Bot
    participant H as handlers
    participant DB as SQLite
    participant G as Gateway
    participant W as Webhook

    U->>B: /charge
    B->>H: show gateways
    H->>DB: read enabled gateways
    DB-->>H: list
    H-->>U: gateway keyboard
    U->>H: choose gateway
    H-->>U: amount keyboard
    U->>H: choose amount
    H->>H: convert USD → Toman (rate × margin)
    H->>DB: create order (pending)
    H->>G: request payment
    G-->>H: payment URL + authority
    H-->>U: payment link
    U->>G: pay
    G->>W: POST /webhook/{gateway}
    W->>G: verify
    G-->>W: confirmed
    W->>DB: transaction + balance + order paid
    W->>U: confirmation message
```

Key guarantees:

- The order is persisted **before** the payment link is shown, so a lost reply never loses money.
- Balance updates run inside a transaction together with the `transactions` insert.
- Webhook processing is idempotent: a repeated callback for an already-paid order is acknowledged without double crediting.

---

## 🌐 Language flow

```mermaid
sequenceDiagram
    participant U as User
    participant B as Bot
    participant I as i18n
    participant DB as SQLite

    U->>B: tap language button
    B->>DB: load user.language
    DB-->>B: fa | en
    B->>I: T(lang, key, args)
    I-->>B: localized string
    B-->>U: message in selected language
    U->>B: pick other language
    B->>DB: save user.language
```

- Translations live in one map in `internal/i18n/i18n.go` keyed by message ID.
- Every handler renders through `T(lang, key, args...)`; no Persian or English literal is inlined in handlers.
- `LangLabel()` provides the human-readable language name for menus.

---

## 🔐 Security

### Access control

```mermaid
graph LR
    A[Sender ID] --> B{equals ADMIN_ID?}
    B -->|yes| OK[full admin access]
    B -->|no| C{in admins table?}
    C -->|yes| D[department-scoped access]
    C -->|no| E[user access only]
```

- Admin identity is verified on every admin callback, not only when the panel opens.
- A department-scoped admin only sees tickets belonging to their departments.
- Banned users are dropped in middleware before any handler runs.

### Input validation

- Amounts validated against `CHARGE_MIN`/`CHARGE_MAX`.
- Receipt photos validated for type and size before storing.
- User-supplied text escaped before HTML-format replies.
- Gateway selections validated against enabled gateways.

### Secrets

`BOT_TOKEN`, gateway keys and webhook secrets live only in `.env`. They are never written to the database, logged, or echoed in chat.

---

## 🗄️ Data model

```mermaid
erDiagram
    USERS ||--o{ TRANSACTIONS : has
    USERS ||--o{ ORDERS : places
    USERS ||--o{ TICKETS : opens
    USERS ||--o{ MANUAL_CHARGES : submits
    ADMINS ||--o{ TICKETS : answers
    DEPARTMENTS ||--o{ TICKETS : groups
    CATEGORIES ||--o{ PRODUCTS : contains
    PRODUCTS ||--o{ ORDERS : sold_as
    ORDERS ||--o| SUBSCRIPTIONS : creates
```

- WAL mode keeps readers non-blocking while the backup job snapshots the file.
- Migrations are applied in `migrate.go` on every start and are safe to re-run.

---

## ⚙️ Background jobs

| Job | Interval | Purpose |
| :--- | :--- | :--- |
| `backup` | scheduled | Copy `bot.db` into `data/backups/` |
| `sync` | scheduled | Refresh panel state and subscription data |
| `traffic` | periodic | Read usage and warn users near limits |
| `daily_report` | daily | Push the day's statistics to the main admin |

Each job runs in its own goroutine, recovers from panics, logs failures and never blocks update handling.

---

## 📈 Extending the bot

1. **New user command** — add a handler in `internal/handlers/` and register it in `handlers.go`.
2. **New admin screen** — add a file under `internal/handlers/admin/` and a button in `panel.go`.
3. **New strings** — add the key to both `fa` and `en` maps in `i18n.go`.
4. **New table** — extend `schema.sql`, add queries in `internal/database/sqlite/`, wire them in `migrate.go`.
5. **New gateway** — implement the `payment.gateway.go` interface and add a webhook branch.
6. **New language** — add a `Lang` constant, a translation map, and expose it in the language menu.

---

## 🧪 Testing strategy

- **Unit** — pure helpers: rate conversion, i18n lookup, validation.
- **Integration** — handler flows against a temporary SQLite file.
- **Webhook** — sign a payload with the test secret and assert idempotency.
- **Manual** — full charge flow against a gateway sandbox.

---

## 🔗 More

[Usage guide](./usage.md) · [API reference](./API.md) · [Setup](./setup.md) · [Structure](../../STRUCTURE.md) · [Persian version](../fa/architecture.md)
