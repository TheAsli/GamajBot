<div align="center">

<a href="https://t.me/AsliCode" target="_blank">
  <img src="https://img.shields.io/badge/telegram-AsliCode-blue?style=flat-square&logo=telegram" alt="کانال توسعه‌دهنده" />
</a>
<a href="https://t.me/TheGamaj" target="_blank">
  <img src="https://img.shields.io/badge/telegram-TheGamaj-blue?style=flat-square&logo=telegram" alt="کانال پروژه" />
</a>

# GamajBot

### ربات تلگرام حرفه‌ای · کیف پول · پرداخت · تیکت · چندزبانه

![نسخه](https://img.shields.io/badge/%D9%86%D8%B3%D8%AE%D9%87-is.0.0.1-red?style=flat-square)
![گو](https://img.shields.io/badge/Go-1.27.1-00ADD8?style=flat-square&logo=go&logoColor=white)
![مجوز](https://img.shields.io/badge/%D9%85%D8%AC%D9%88%D8%B2-AGPL--3.0-20c997?style=flat-square)
![زبان‌ها](https://img.shields.io/badge/%D8%B2%D8%A8%D8%A7%D9%86%E2%80%8C%D9%87%D8%A7-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C%20%7C%20English-a855f7?style=flat-square)

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./README.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./README.fa.md)

</div>

---

## 📚 مستندات

| راهنما | فارسی | English |
| :--- | :--- | :--- |
| **راهنمای استفاده** | [![استفاده](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D8%A7%D8%B3%D8%AA%D9%81%D8%A7%D8%AF%D9%87-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./Docs/fa/usage.md) | [![Usage](https://img.shields.io/badge/open-Usage-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./Docs/en/usage.md) |
| **مستندات API** | [![API](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./Docs/fa/API.md) | [![API](https://img.shields.io/badge/open-API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./Docs/en/API.md) |
| **معماری** | [![معماری](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D9%85%D8%B9%D9%85%D8%A7%D8%B1%DB%8C-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./Docs/fa/architecture.md) | [![Architecture](https://img.shields.io/badge/open-Architecture-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./Docs/en/architecture.md) |
| **نصب و راه‌اندازی** | [![نصب](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D9%86%D8%B5%D8%A8-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./Docs/fa/setup.md) | [![Setup](https://img.shields.io/badge/open-Setup-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./Docs/en/setup.md) |
| **ساختار پروژه** | [![ساختار](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D8%B3%D8%A7%D8%AE%D8%AA%D8%A7%D8%B1-20c997?style=flat-square&logo=files&logoColor=white)](./STRUCTURE.fa.md) | [![Structure](https://img.shields.io/badge/open-Structure-20c997?style=flat-square&logo=files&logoColor=white)](./STRUCTURE.md) |
| **تغییرات** | [![تغییرات](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D8%AA%D8%BA%DB%8C%DB%8C%D8%B1%D8%A7%D8%AA-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](./CHANGELOG.fa.md) | [![Changelog](https://img.shields.io/badge/open-Changelog-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](./CHANGELOG.md) |
| **راهنمای مشارکت** | [![مشارکت](https://img.shields.io/badge/%D8%A8%D8%A7%D8%B2%DA%A9%D8%B1%D8%AF%D9%86-%D9%85%D8%B4%D8%A7%D8%B1%DA%A9%D8%AA-0d6efd?style=flat-square&logo=git&logoColor=white)](./CONTRIBUTING.fa.md) | [![Contributing](https://img.shields.io/badge/open-Contributing-0d6efd?style=flat-square&logo=git&logoColor=white)](./CONTRIBUTING.md) |

> هر مستند به زبان خودش نوشته شده و به نسخهٔ زبان دیگر هم لینک دارد؛ با دکمه‌های بالای هر صفحه می‌توانید زبان را عوض کنید یا بین راهنماها جابه‌جا شوید.

📖 **سایت مستندات** — همان راهنماها به‌صورت سایت قابل جستجو با دکمهٔ تغییر زبان:

```bash
cd website && npm install
npm run docs:dev     # http://localhost:5173
npm run docs:build   # خروجی استاتیک در website/.vitepress/dist
```

---

## 📋 معرفی پروژه

**GamajBot** یک ربات تلگرام آمادهٔ اجرا در محیط عملیاتی است که با زبان Go نوشته شده. این ربات کیف پول کاربران، چند درگاه پرداخت، پنل مدیریت کامل، سیستم پشتیبانی دپارتمان‌محور، کارهای زمان‌بندی‌شده و رابط کاملاً دوزبانه (فارسی/انگلیسی) را در یک بسته ارائه می‌کند.

- **فارسی** زبان پیش‌فرض است و **انگلیسی** به‌طور کامل ترجمه شده است.
- هر کاربر می‌تواند از منوی «حساب من» زبان را تغییر دهد.
- همهٔ عملیات از طریق آپدیت‌های تلگرام و یک وب‌هوک کوچک برای کال‌بک درگاه‌ها انجام می‌شود.

---

## ✨ امکانات

### 👤 برای کاربران
- **کیف پول** — مشاهدهٔ موجودی، شارژ و تاریخچهٔ کامل تراکنش‌ها با صفحه‌بندی
- **پرداخت** — تترامیناتور، CubePay، AbanGateway و کارت‌به‌کارت دستی
- **شارژ دستی** — ارسال رسید و تأیید توسط ادمین
- **تیکت** — ثبت تیکت، انتخاب دپارتمان و پیگیری گفتگو
- **زبان** — تغییر لحظه‌ای بین فارسی و انگلیسی

### 👨‍💼 برای مدیران
- **پنل ادمین** — کاربران، ادمین‌ها، دپارتمان‌ها، تنظیمات و آمار
- **مدیریت نقش‌ها** — افزودن/حذف ادمین و محدودکردن آن‌ها به دپارتمان
- **پیام همگانی** — ارسال پیام به تمام کاربران ربات
- **شارژهای دستی** — بررسی، تأیید یا رد رسیدهای ارسال‌شده
- **تراکنش‌ها** — مشاهده و بازبینی فعالیت کیف پول‌ها
- **سرویس‌ها** — مدیریت پلن‌های VPN، اشتراک‌ها و لینک‌های اتصال

### ⚙️ در پشت صحنه
- **کارهای پس‌زمینه** — پشتیبان‌گیری، همگام‌سازی، پایش ترافیک و گزارش روزانه
- **سرور وب‌هوک** — دریافت و اعتبارسنجی کال‌بک درگاه‌های پرداخت
- **نرخ لحظه‌ای دلار** — دریافت از API، کش و اعمال درصد سود قابل تنظیم
- **حالت تعمیرات** — قفل‌کردن ربات برای همه به‌جز ادمین‌ها
- **جوین اجباری** — الزام عضویت در کانال پیش از استفاده از ربات

---

## 🧱 پشتهٔ فنی

| لایه | انتخاب |
| :--- | :--- |
| زبان | Go 1.27.1 |
| تلگرام | Bot API (Long Polling + Webhook) |
| ذخیره‌سازی | JSON (`data/wallets.json`) / SQLite |
| تنظیمات | فایل `.env` با `godotenv` |
| چندزبانه | مترجم داخلی مبتنی بر Map (`internal/i18n`) |
| زمان‌بندی | کارهای Goroutine درون‌فرایندی (`internal/jobs`) |

---

## 📁 چیدمان پروژه

```text
GamajBot/
├── cmd/AsliCode/main.go        # نقطهٔ ورود برنامه
├── internal/
│   ├── bot/                    # راه‌اندازی، اتصال اجزا، سرور وب‌هوک
│   ├── config/                 # بارگذاری و اعتبارسنجی .env
│   ├── database/               # کاربران، کیف پول، تیکت، ادمین، تنظیمات
│   ├── handlers/               # هندلرهای آپدیت تلگرام (admin/، support/، notify/)
│   ├── i18n/                   # ترجمه‌های فارسی و انگلیسی
│   ├── jobs/                   # پشتیبان‌گیری، همگام‌سازی، ترافیک، گزارش روزانه
│   ├── menus/                  # کیبوردها و منوها
│   ├── middleware/             # جوین اجباری، حالت تعمیرات
│   ├── payment/                # اتصال به درگاه‌های پرداخت
│   ├── rate/                   # دریافت و کش نرخ دلار
│   └── Services/               # سرویس‌های جانبی
├── data/                       # داده‌های زمان اجرا (wallets.json، بکاپ bot.db)
├── Docs/en · Docs/fa           # مستندات
└── STRUCTURE.fa.md             # مرجع کامل ساختار
```

توضیح کامل: [STRUCTURE.fa.md](./STRUCTURE.fa.md) · [STRUCTURE.md](./STRUCTURE.md)

---

## ⚙️ تنظیمات

فایل [.env.example](./.env.example) را به `.env` کپی کنید و مقادیر را تکمیل کنید.

| متغیر | توضیح | نمونه |
| :--- | :--- | :--- |
| `BOT_TOKEN` | توکن ربات از [@BotFather](https://t.me/BotFather) | `123456789:ABCdef...` |
| `ADMIN_ID` | آیدی عددی مدیر اصلی | `8263673526` |
| `TETRAMINATOR_API_KEY` | کلید درگاه تترامیناتور | `your_api_key` |
| `TETRAMINATOR_BASE_URL` | آدرس پایهٔ تترامیناتور | `https://example.com` |
| `CHARGE_MIN` / `CHARGE_MAX` | حداقل و حداکثر شارژ (تومان) | `50000` / `10000000` |
| `QUICK_AMOUNTS` | مبالغ آمادهٔ شارژ (دلار) | `1,5,10,50` |
| `FORCE_JOIN_ENABLED` | الزام عضویت در کانال | `true` |
| `FORCE_JOIN_CHANNEL` | یوزرنیم کانال | `@example` |
| `FORCE_JOIN_CHANNEL_URL` | لینک دعوت کانال | `https://t.me/example` |
| `DOLLAR_RATE_API` | آدرس API نرخ دلار | `https://api.priceto.day/v1/latest/irr/usd` |
| `DOLLAR_RATE_MARGIN` | درصد سود | `5.0` |
| `DOLLAR_RATE_CACHE` | مدت کش نرخ (ثانیه) | `30` |
| `WEBHOOK_PORT` | پورت وب‌هوک | `8080` |
| `WEBHOOK_BASE_URL` | آدرس عمومی کال‌بک | `http://YOUR_SERVER:8080` |

---

## 💳 درگاه‌های پرداخت

| درگاه | نوع |
| :--- | :--- |
| **تترامیناتور** | پرداخت آنلاین لحظه‌ای |
| **CubePay** | سیستم پرداخت دیجیتال |
| **AbanGateway** | درگاه بانکی |
| **کارت دستی** | آپلود رسید + تأیید ادمین |

---

## 🚀 شروع سریع

```bash
# ۱. دریافت سورس
git clone https://github.com/theasli/GamajBot.git
cd gamajbot

# ۲. نصب وابستگی‌ها
go mod tidy

# ۳. پیکربندی
cp .env.example .env
#    → مقدار BOT_TOKEN و ADMIN_ID و درگاه‌ها را وارد کنید

# ۴. اجرا
go run cmd/AsliCode/main.go
```

به ربات `/start` بفرستید؛ به‌صورت پیش‌فرض باید فارسی پاسخ دهد.

ساخت نسخهٔ اجرایی:

```bash
go build -o gamajbot cmd/AsliCode/main.go
./gamajbot
```

راهنمای گام‌به‌گام (systemd و Docker): [Docs/fa/setup.md](./Docs/fa/setup.md) · [Docs/en/setup.md](./Docs/en/setup.md)

---

## 📦 نکات استقرار

۱. فایل `.env` را هرگز در گیت قرار ندهید.
۲. ربات را زیر یک مدیر فرایند (`systemd`، Docker، PM2 و…) اجرا کنید.
۳. پورت `WEBHOOK_PORT` را باز کنید و `WEBHOOK_BASE_URL` را روی آدرس عمومی تنظیم کنید.
۴. از پوشهٔ `data/` به‌طور منظم بکاپ بگیرید؛ کار پشتیبان‌گیری داخلی در `data/backups/` ذخیره می‌کند.

---

## 🆘 پشتیبانی

- **توسعه‌دهنده**: [@AsliCode](https://t.me/AsliCode)
- **پروژه**: [@TheGamaj](https://t.me/TheGamaj)
- **باگ و درخواست**: در GitHub Issue ثبت کنید

---

## 📜 مجوز

کپی‌رایت © 2026 AsliCode.

این پروژه تحت [مجوز GNU Affero General Public License v3.0](./LICENSE) (AGPL-3.0) منتشر شده است.

<div align="center">

**Coded by AsliCode**

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./README.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./README.fa.md)

</div>
