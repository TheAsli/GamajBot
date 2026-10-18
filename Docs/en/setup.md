<div align="center">

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./setup.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](../fa/setup.md)

[![README](https://img.shields.io/badge/README-English-0d6efd?style=flat-square&logo=readme&logoColor=white)](../../README.md)
[![Usage](https://img.shields.io/badge/Usage-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./usage.md)
[![API](https://img.shields.io/badge/API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./API.md)
[![Architecture](https://img.shields.io/badge/Architecture-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./architecture.md)
[![Setup](https://img.shields.io/badge/Setup-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./setup.md)
[![Structure](https://img.shields.io/badge/Structure-20c997?style=flat-square&logo=files&logoColor=white)](../../STRUCTURE.md)
[![Changelog](https://img.shields.io/badge/Changelog-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](../../CHANGELOG.md)
[![Contributing](https://img.shields.io/badge/Contributing-0d6efd?style=flat-square&logo=git&logoColor=white)](../../CONTRIBUTING.md)

</div>

# Installation & Setup

## Prerequisites

| Requirement | Notes |
| :--- | :--- |
| Go **1.27.1+** | `go version` must print 1.27.1 or newer |
| Git | To clone the repository |
| Linux server *(production)* | Ubuntu 20.04+, Debian 11+ or any systemd distro |
| SQLite | Bundled — no separate server needed |

Optional: Docker, `systemd`, a reverse proxy (nginx / Caddy) for the webhook.

---

## 1. Clone and build

```bash
git clone https://github.com/theasli/GamajBot.git
cd gamajbot
go mod tidy
go build -o gamajbot cmd/AsliCode/main.go
```

## 2. Configure

```bash
cp .env.example .env
```

Fill in at least these values:

| Variable | How to get it |
| :--- | :--- |
| `BOT_TOKEN` | [@BotFather](https://t.me/BotFather) → `/newbot` |
| `ADMIN_ID` | Your numeric ID from [@userinfobot](https://t.me/userinfobot) |
| `TETRAMINATOR_API_KEY` | Panel of the gateway |
| `TETRAMINATOR_BASE_URL` | Panel of the gateway |
| `WEBHOOK_BASE_URL` | Public address of your server, e.g. `https://bot.example.com` |

Full variable list lives in the *Configuration* section of the [README](../../README.md).

> Never commit `.env`. It is already listed in `.gitignore`.

## 3. First run

```bash
./gamajbot
```

Expected startup log: config loaded → database migrated → jobs started → webhook listening → polling started.

Open your bot in Telegram, send `/start`, and log in as the admin. If the bot answers, setup is complete.

## 4. Create the data directory

```bash
mkdir -p data/backups
```

The process creates it on first run, but creating it yourself makes permissions explicit:

```bash
chown -R botuser:botuser data
chmod 750 data
```

---

## Production deployment

### systemd

```ini
# /etc/systemd/system/gamajbot.service
[Unit]
Description=GamajBot
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=botuser
WorkingDirectory=/opt/gamajbot
ExecStart=/opt/gamajbot/gamajbot
Restart=on-failure
RestartSec=5
EnvironmentFile=/opt/gamajbot/.env
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/gamajbot/data

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now gamajbot
sudo journalctl -u gamajbot -f
```

### Docker

```dockerfile
FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gamajbot cmd/AsliCode/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 bot
WORKDIR /app
COPY --from=build /out/gamajbot .
USER bot
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["/app/gamajbot"]
```

```bash
docker build -t gamajbot .
docker run -d --name gamajbot \
  --env-file .env \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  --restart unless-stopped \
  gamajbot
```

### Webhook / reverse proxy

The webhook must be reachable over HTTPS at `WEBHOOK_BASE_URL`.

```nginx
server {
    listen 443 ssl http2;
    server_name bot.example.com;

    ssl_certificate     /etc/letsencrypt/live/bot.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/bot.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Verification checklist

- [ ] `./gamajbot` starts without fatal errors
- [ ] `data/bot.db` is created
- [ ] `/start` answers in Persian
- [ ] Language switch to English works and persists
- [ ] Admin panel opens for `ADMIN_ID` only
- [ ] A test charge reaches the gateway and the callback credits the balance
- [ ] `curl -s https://bot.example.com/health` returns `{"status":"ok"}`
- [ ] The backup job produces a file in `data/backups/`

---

## Updating

```bash
git pull
go mod tidy
go build -o gamajbot cmd/AsliCode/main.go
sudo systemctl restart gamajbot
```

Migrations run automatically on start. Take a backup of `data/bot.db` before upgrading.

---

## Common problems

| Problem | Fix |
| :--- | :--- |
| `invalid token` | `BOT_TOKEN` has a stray quote or space |
| `database is locked` | Only one process may open `data/bot.db`; stop duplicates |
| Webhook never fires | `WEBHOOK_BASE_URL` must be publicly reachable and HTTPS |
| Force-join loop | Add the bot as an admin of the channel, or disable `FORCE_JOIN_ENABLED` |
| Timezone off in reports | Set `TZ=Asia/Tehran` in the environment |

---

## 🔗 More

[Usage guide](./usage.md) · [Architecture](./architecture.md) · [API reference](./API.md) · [Structure](../../STRUCTURE.md) · [Persian version](../fa/setup.md)
