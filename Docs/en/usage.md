<div align="center">

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./usage.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](../fa/usage.md)

[![README](https://img.shields.io/badge/README-English-0d6efd?style=flat-square&logo=readme&logoColor=white)](../../README.md)
[![Usage](https://img.shields.io/badge/Usage-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./usage.md)
[![API](https://img.shields.io/badge/API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./API.md)
[![Architecture](https://img.shields.io/badge/Architecture-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./architecture.md)
[![Setup](https://img.shields.io/badge/Setup-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./setup.md)
[![Structure](https://img.shields.io/badge/Structure-20c997?style=flat-square&logo=files&logoColor=white)](../../STRUCTURE.md)
[![Changelog](https://img.shields.io/badge/Changelog-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](../../CHANGELOG.md)
[![Contributing](https://img.shields.io/badge/Contributing-0d6efd?style=flat-square&logo=git&logoColor=white)](../../CONTRIBUTING.md)

</div>

# Usage Guide

Everything a user and an administrator can do with **GamajBot**.

---

## 👤 For users

### Starting the bot
Send `/start`. The bot greets you in your saved language (Persian by default) and shows the main keyboard.

### My account
`/myaccount` (or the **My Account** button) shows your numeric ID, current balance and the entry point to every personal setting.

### Charging your wallet
1. Send `/charge` or press **Charge**.
2. Pick a payment gateway from the list shown by the admin.
3. Choose a preset amount (`QUICK_AMOUNTS`) or type a custom value between `CHARGE_MIN` and `CHARGE_MAX`.
4. Confirm — the bot replies with the payment link and an order ID.
5. Pay through the gateway. When the callback arrives the balance is credited automatically.

### Manual card transfer
1. Choose **Manual card** as the gateway.
2. Transfer the exact amount to the card number shown by the bot.
3. Upload the receipt photo.
4. Your request enters the admin queue; you get a notification when it is approved or rejected.

### Transaction history
`/history` lists every wallet operation with pagination (older/newer buttons). Each row shows amount, type, status and date.

### Support tickets
Open the support menu → **New ticket** → choose a department → send the subject and the message. Replies arrive in the same chat; you can keep answering inside the ticket.

### Language
From **My Account** press the language button and pick **فارسی** or **English**. The choice is stored per user and applies immediately.

---

## 👨‍💼 For administrators

Admins reach the panel with `/admin` (or the hidden admin keyboard). Access is granted by `ADMIN_ID` or by an entry in the `admins` table.

### Users
- Search a user by ID or username
- View balance, orders, tickets and last activity
- Adjust balance manually (the change is written to `transactions`)
- Ban or unban a user

### Admins & roles
- Add or remove administrators
- Scope an admin to one or more departments so they only see their own tickets
- Every action is recorded in the admin audit log

### Departments
Create departments (for example *Sales*, *Technical*), rename them and attach admins to them. Tickets are routed by department.

### Tickets
Browse the inbox, filter by department, reply to the user, change priority (low / normal / high / urgent) or close the ticket.

### Broadcast
Write a message and send it to every user. The bot reports how many deliveries succeeded.

### Products & services
- Create categories and products with price, duration and delivery type
- Configure the fields a product needs from the buyer
- Fulfil orders manually or automatically through the 3x-ui panel
- Manage renewals and expiry reminders

### Gateways
Enable or disable Tetraminator, CubePay, AbanGateway and manual card, and edit their keys without restarting the bot.

### Statistics
- Analytics dashboard: registrations, revenue, charge volume
- Reports by date range
- Daily report pushed by the `daily_report` job

### Settings
Charge limits, force-join channel, dollar-rate margin and cache TTL, maintenance mode and other runtime options are all editable from the panel.

---

## ⚙️ How pricing works

The Toman price of a USD-priced product is calculated as:

```text
price_toman = usd_amount × dollar_rate × (1 + DOLLAR_RATE_MARGIN / 100)
```

`dollar_rate` is fetched from `DOLLAR_RATE_API` and cached for `DOLLAR_RATE_CACHE` seconds. An admin can override the rate manually from settings.

---

## 🛠️ Background behaviour

| Job | What it does |
| :--- | :--- |
| Backup | Writes a snapshot of `data/bot.db` into `data/backups/` |
| Sync | Refreshes remote panel state (3x-ui clients, traffic) |
| Traffic | Polls used traffic and warns users near their limit |
| Daily report | Sends the day's statistics to the main admin |

---

## 🆘 Troubleshooting

| Symptom | Check |
| :--- | :--- |
| Bot does not answer | `BOT_TOKEN` is valid and the process is running; look at the startup log |
| No payment gateways listed | At least one gateway must be enabled in the admin panel |
| Payment paid but balance unchanged | The webhook must be reachable at `WEBHOOK_BASE_URL`; inspect the webhook log |
| Language button does nothing | The user record must exist — send `/start` first |
| "Join the channel" loop | The bot must be an admin of `FORCE_JOIN_CHANNEL`, or set `FORCE_JOIN_ENABLED=false` |

---

## 🔗 More

[API reference](./API.md) · [Architecture](./architecture.md) · [Setup](./setup.md) · [Structure](../../STRUCTURE.md) · [Persian version](../fa/usage.md)
