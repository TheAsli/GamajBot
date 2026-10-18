<div align="center">

<a href="https://t.me/AsliCode" target="_blank">
  <img src="https://img.shields.io/badge/telegram-AsliCode-blue?style=flat-square&logo=telegram" alt="Developer Channel" />
</a>
<a href="https://t.me/TheGamaj" target="_blank">
  <img src="https://img.shields.io/badge/telegram-TheGamaj-blue?style=flat-square&logo=telegram" alt="Project Channel" />
</a>

# GamajBot

### Feature-rich Telegram bot · wallets · payments · tickets · i18n

![version](https://img.shields.io/badge/version-is.0.0.1-red?style=flat-square)
![go](https://img.shields.io/badge/Go-1.27.1-00ADD8?style=flat-square&logo=go&logoColor=white)
![license](https://img.shields.io/badge/license-AGPL--3.0-20c997?style=flat-square)
![langs](https://img.shields.io/badge/languages-Persian%20%7C%20English-a855f7?style=flat-square)

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./README.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./README.fa.md)

</div>

---

## 📚 Documentation

| Guide | English | فارسی |
| :--- | :--- | :--- |
| **Usage** | [![Usage](https://img.shields.io/badge/open-Usage-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./Docs/en/usage.md) | [![استفاده](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D8%A7%D8%B3%D8%AA%D9%81%D8%A7%D8%AF%D9%87-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./Docs/fa/usage.md) |
| **API** | [![API](https://img.shields.io/badge/open-API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./Docs/en/API.md) | [![API](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./Docs/fa/API.md) |
| **Architecture** | [![Architecture](https://img.shields.io/badge/open-Architecture-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./Docs/en/architecture.md) | [![معماری](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D9%85%D8%B9%D9%85%D8%A7%D8%B1%DB%8C-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./Docs/fa/architecture.md) |
| **Setup** | [![Setup](https://img.shields.io/badge/open-Setup-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./Docs/en/setup.md) | [![نصب](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D9%86%D8%B5%D8%A8-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./Docs/fa/setup.md) |
| **Structure** | [![Structure](https://img.shields.io/badge/open-Structure-20c997?style=flat-square&logo=files&logoColor=white)](./STRUCTURE.md) | [![ساختار](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D8%B3%D8%A7%D8%AE%D8%AA%D8%A7%D8%B1-20c997?style=flat-square&logo=files&logoColor=white)](./STRUCTURE.fa.md) |
| **Changelog** | [![Changelog](https://img.shields.io/badge/open-Changelog-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](./CHANGELOG.md) | [![تغییرات](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D8%AA%D8%BA%DB%8C%DB%8C%D8%B1%D8%A7%D8%AA-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](./CHANGELOG.fa.md) |
| **Contributing** | [![Contributing](https://img.shields.io/badge/open-Contributing-0d6efd?style=flat-square&logo=git&logoColor=white)](./CONTRIBUTING.md) | [![مشارکت](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D9%85%D8%B4%D8%A7%D8%B1%DA%A9%D8%AA-0d6efd?style=flat-square&logo=git&logoColor=white)](./CONTRIBUTING.fa.md) |

> Every document is written in its own language and links to its counterpart — use the badges at the top of any page to switch language or jump between guides.

📖 **Documentation website** — the same guides as a searchable site with a language switcher:

```bash
cd website && npm install
npm run docs:dev     # http://localhost:5173
npm run docs:build   # static output in website/.vitepress/dist
```

---

## 📋 Overview

**GamajBot** is a production-oriented Telegram bot written in Go. It bundles a wallet system, multiple payment gateways, a full admin panel, a department-based support desk, scheduled background jobs and a complete Persian/English interface.

- **Persian** is the default language; **English** is fully translated.
- Every user can switch language from the *My Account* menu.
- All operations are driven by Telegram updates and a small webhook server for gateway callbacks.

---

## ✨ Features

### 👤 For users
- **Wallet** — balance, top-up and full transaction history with pagination
- **Payments** — Tetraminator, CubePay, AbanGateway and manual card transfer
- **Manual charge** — send a receipt and wait for admin verification
- **Tickets** — open a ticket, pick a department, follow the conversation
- **Language** — instant switch between فارسی and English

### 👨‍💼 For administrators
- **Admin panel** — users, admins, departments, settings and statistics
- **Role management** — promote/demote admins and scope them to departments
- **Broadcast** — send a message to every user of the bot
- **Manual charges** — review, approve or reject submitted receipts
- **Transactions** — inspect and audit wallet activity
- **Services** — manage VPN plans, subscriptions and connection links

### ⚙️ Under the hood
- **Background jobs** — backup, sync, traffic monitoring and daily reports
- **Webhook server** — receives and verifies payment callbacks
- **Live dollar rate** — fetched from an API, cached and adjusted by a configurable margin
- **Maintenance mode** — lock the bot for everyone except admins
- **Force join** — require channel membership before using the bot

---

## 🧱 Tech stack

| Layer | Choice |
| :--- | :--- |
| Language | Go 1.27.1 |
| Telegram | Bot API (long polling + webhook) |
| Storage | JSON (`data/wallets.json`) / SQLite |
| Config | `.env` via `godotenv` |
| i18n | In-repo map based translator (`internal/i18n`) |
| Scheduling | In-process goroutine jobs (`internal/jobs`) |

---

## 📁 Project layout

```text
GamajBot/
├── cmd/AsliCode/main.go        # entrypoint
├── internal/
│   ├── bot/                    # startup, wiring, webhook server
│   ├── config/                 # .env loading & validation
│   ├── database/               # users, wallets, tickets, admins, settings
│   ├── handlers/               # Telegram update handlers (admin/, support/, notify/)
│   ├── i18n/                   # Persian/English translations
│   ├── jobs/                   # backup, sync, traffic, daily report
│   ├── menus/                  # keyboards and menus
│   ├── middleware/             # force join, maintenance mode
│   ├── payment/                # gateway integrations
│   ├── rate/                   # dollar rate fetching & caching
│   └── Services/               # auxiliary services
├── data/                       # runtime data (wallets.json, bot.db backups)
├── Docs/en · Docs/fa           # documentation
└── STRUCTURE.md                # detailed structure reference
```

Full breakdown: [STRUCTURE.md](./STRUCTURE.md) · [STRUCTURE.fa.md](./STRUCTURE.fa.md)

---

## ⚙️ Configuration

Copy [.env.example](./.env.example) to `.env` and fill in the values.

| Variable | Description | Example |
| :--- | :--- | :--- |
| `BOT_TOKEN` | Telegram bot token from [@BotFather](https://t.me/BotFather) | `123456789:ABCdef...` |
| `ADMIN_ID` | Numeric ID of the main administrator | `8263673526` |
| `TETRAMINATOR_API_KEY` | Tetraminator gateway key | `your_api_key` |
| `TETRAMINATOR_BASE_URL` | Tetraminator base URL | `https://example.com` |
| `CHARGE_MIN` / `CHARGE_MAX` | Allowed charge range in Toman | `50000` / `10000000` |
| `QUICK_AMOUNTS` | Preset amounts in USD | `1,5,10,50` |
| `FORCE_JOIN_ENABLED` | Require channel membership | `true` |
| `FORCE_JOIN_CHANNEL` | Channel username | `@example` |
| `FORCE_JOIN_CHANNEL_URL` | Channel invite link | `https://t.me/example` |
| `DOLLAR_RATE_API` | Dollar rate endpoint | `https://api.priceto.day/v1/latest/irr/usd` |
| `DOLLAR_RATE_MARGIN` | Profit margin (%) | `5.0` |
| `DOLLAR_RATE_CACHE` | Rate cache TTL (seconds) | `30` |
| `WEBHOOK_PORT` | Webhook listener port | `8080` |
| `WEBHOOK_BASE_URL` | Public callback base URL | `http://YOUR_SERVER:8080` |

---

## 💳 Payment gateways

| Gateway | Type |
| :--- | :--- |
| **Tetraminator** | Instant online payment |
| **CubePay** | Digital payment system |
| **AbanGateway** | Bank gateway |
| **Manual card** | Receipt upload + admin approval |

---

## 🚀 Quick start

```bash
# 1. clone
git clone https://github.com/theasli/GamajBot.git
cd gamajbot

# 2. dependencies
go mod tidy

# 3. configuration
cp .env.example .env
#    → set BOT_TOKEN, ADMIN_ID and your gateways

# 4. run
go run cmd/AsliCode/main.go
```

Send `/start` to your bot — it should answer in Persian by default.

Build for production:

```bash
go build -o gamajbot cmd/AsliCode/main.go
./gamajbot
```

Step-by-step guide (systemd + Docker): [Docs/en/setup.md](./Docs/en/setup.md) · [Docs/fa/setup.md](./Docs/fa/setup.md)

---

## 📦 Deployment notes

1. Keep `.env` out of version control.
2. Run under a process manager (`systemd`, Docker, PM2…).
3. Expose `WEBHOOK_PORT` and set `WEBHOOK_BASE_URL` to your public address.
4. Back up `data/` regularly — the built-in backup job writes to `data/backups/`.

---

## 🆘 Support

- **Developer**: [@AsliCode](https://t.me/AsliCode)
- **Project**: [@TheGamaj](https://t.me/TheGamaj)
- **Bugs & requests**: open a GitHub issue

---

## 📜 License

Copyright © 2026 AsliCode.

Released under the [GNU Affero General Public License v3.0](./LICENSE) (AGPL-3.0).

<div align="center">

**Coded by AsliCode**

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./README.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./README.fa.md)

</div>
