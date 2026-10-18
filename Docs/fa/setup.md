<div align="center">

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](../en/setup.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./setup.md)

[![README فارسی](https://img.shields.io/badge/README-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square&logo=readme&logoColor=white)](../../README.fa.md)
[![استفاده](https://img.shields.io/badge/%D8%A7%D8%B3%D8%AA%D9%81%D8%A7%D8%AF%D9%87-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./usage.md)
[![API](https://img.shields.io/badge/API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./API.md)
[![معماری](https://img.shields.io/badge/%D9%85%D8%B9%D9%85%D8%A7%D8%B1%DB%8C-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./architecture.md)
[![نصب](https://img.shields.io/badge/%D9%86%D8%B5%D8%A8-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./setup.md)
[![ساختار](https://img.shields.io/badge/%D8%B3%D8%A7%D8%AE%D8%AA%D8%A7%D8%B1-20c997?style=flat-square&logo=files&logoColor=white)](../../STRUCTURE.fa.md)
[![تغییرات](https://img.shields.io/badge/%D8%AA%D8%BA%DB%8C%DB%8C%D8%B1%D8%A7%D8%AA-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](../../CHANGELOG.fa.md)
[![مشارکت](https://img.shields.io/badge/%D9%85%D8%B4%D8%A7%D8%B1%DA%A9%D8%AA-0d6efd?style=flat-square&logo=git&logoColor=white)](../../CONTRIBUTING.fa.md)

</div>

# نصب و راه‌اندازی

## پیش‌نیازها

| نیاز | توضیح |
| :--- | :--- |
| Go **1.27.1+** | خروجی `go version` باید 1.27.1 یا بالاتر باشد |
| Git | برای کلون کردن مخزن |
| سرور لینوکس *(محیط عملیاتی)* | Ubuntu 20.04+، Debian 11+ یا هر توزیع دارای systemd |
| SQLite | همراه پروژه است — نیازی به سرور جداگانه نیست |

اختیاری: Docker، `systemd`، و یک ریورس‌پروکسی (nginx / Caddy) برای وب‌هوک.

---

## ۱. دریافت و ساخت

```bash
git clone https://github.com/theasli/GamajBot.git
cd gamajbot
go mod tidy
go build -o gamajbot cmd/AsliCode/main.go
```

## ۲. پیکربندی

```bash
cp .env.example .env
```

حداقل این مقادیر را پر کنید:

| متغیر | نحوهٔ دریافت |
| :--- | :--- |
| `BOT_TOKEN` | از [@BotFather](https://t.me/BotFather) → `/newbot` |
| `ADMIN_ID` | آیدی عددی خودتان از [@userinfobot](https://t.me/userinfobot) |
| `TETRAMINATOR_API_KEY` | پنل درگاه |
| `TETRAMINATOR_BASE_URL` | پنل درگاه |
| `WEBHOOK_BASE_URL` | آدرس عمومی سرور، مثلاً `https://bot.example.com` |

فهرست کامل متغیرها در بخش «تنظیمات» فایل [README فارسی](../../README.fa.md) آمده است.

> فایل `.env` را هرگز کامیت نکنید؛ از قبل در `.gitignore` قرار دارد.

## ۳. اجرای اول

```bash
./gamajbot
```

لاگ شروع مورد انتظار: خواندن تنظیمات → مهاجرت دیتابیس → شروع کارها → گوش‌دادن وب‌هوک → شروع polling.

ربات را در تلگرام باز کنید، `/start` بفرستید و به‌عنوان ادمین وارد شوید. اگر ربات پاسخ داد، نصب کامل است.

## ۴. ساخت پوشهٔ داده

```bash
mkdir -p data/backups
```

فرایند خودش آن را می‌سازد، اما ساخت دستی اجازه‌های دسترسی را شفاف می‌کند:

```bash
chown -R botuser:botuser data
chmod 750 data
```

---

## استقرار در محیط عملیاتی

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

### وب‌هوک و ریورس‌پروکسی

وب‌هوک باید روی HTTPS و در آدرس `WEBHOOK_BASE_URL` در دسترس باشد.

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

## چک‌لیست بررسی نصب

- [ ] `./gamajbot` بدون خطای مرگبار اجرا می‌شود
- [ ] فایل `data/bot.db` ساخته می‌شود
- [ ] `/start` به فارسی پاسخ می‌دهد
- [ ] تغییر زبان به انگلیسی کار می‌کند و ذخیره می‌شود
- [ ] پنل مدیریت فقط برای `ADMIN_ID` باز می‌شود
- [ ] یک شارژ آزمایشی به درگاه می‌رسد و کال‌بک موجودی را افزایش می‌دهد
- [ ] `curl -s https://bot.example.com/health` مقدار `{"status":"ok"}` برمی‌گرداند
- [ ] کار پشتیبان‌گیری فایلی در `data/backups/` می‌سازد

---

## به‌روزرسانی

```bash
git pull
go mod tidy
go build -o gamajbot cmd/AsliCode/main.go
sudo systemctl restart gamajbot
```

مهاجرت‌ها در شروع به‌صورت خودکار اجرا می‌شوند. پیش از ارتقا از `data/bot.db` بکاپ بگیرید.

---

## مشکلات رایج

| مشکل | راه‌حل |
| :--- | :--- |
| `invalid token` | در `BOT_TOKEN` کوتیشن یا فاصلهٔ اضافه وجود دارد |
| `database is locked` | فقط یک فرایند باید `data/bot.db` را باز کند؛ نمونه‌های تکراری را ببندید |
| وب‌هوک اجرا نمی‌شود | `WEBHOOK_BASE_URL` باید از بیرون در دسترس و HTTPS باشد |
| حلقهٔ جوین اجباری | ربات را ادمین کانال کنید یا `FORCE_JOIN_ENABLED` را غیرفعال کنید |
| ساعت گزارش‌ها اشتباه است | `TZ=Asia/Tehran` را در محیط تنظیم کنید |

---

## 🔗 بیشتر

[راهنمای استفاده](./usage.md) · [معماری](./architecture.md) · [مرجع API](./API.md) · [ساختار](../../STRUCTURE.fa.md) · [نسخهٔ انگلیسی](../en/setup.md)
