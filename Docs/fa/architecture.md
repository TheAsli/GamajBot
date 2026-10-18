<div align="center">

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](../en/architecture.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./architecture.md)

[![README فارسی](https://img.shields.io/badge/README-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square&logo=readme&logoColor=white)](../../README.fa.md)
[![استفاده](https://img.shields.io/badge/%D8%A7%D8%B3%D8%AA%D9%81%D8%A7%D8%AF%D9%87-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./usage.md)
[![API](https://img.shields.io/badge/API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./API.md)
[![معماری](https://img.shields.io/badge/%D9%85%D8%B9%D9%85%D8%A7%D8%B1%DB%8C-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./architecture.md)
[![نصب](https://img.shields.io/badge/%D9%86%D8%B5%D8%A8-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./setup.md)
[![ساختار](https://img.shields.io/badge/%D8%B3%D8%A7%D8%AE%D8%AA%D8%A7%D8%B1-20c997?style=flat-square&logo=files&logoColor=white)](../../STRUCTURE.fa.md)
[![تغییرات](https://img.shields.io/badge/%D8%AA%D8%BA%DB%8C%DB%8C%D8%B1%D8%A7%D8%AA-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](../../CHANGELOG.fa.md)
[![مشارکت](https://img.shields.io/badge/%D9%85%D8%B4%D8%A7%D8%B1%DA%A9%D8%AA-0d6efd?style=flat-square&logo=git&logoColor=white)](../../CONTRIBUTING.fa.md)

</div>

# معماری

## 📊 نمای کلی سیستم

```mermaid
graph TD
    subgraph Telegram
        TG[Bot API تلگرام] -->|آپدیت‌ها| BOT[فرایند ربات]
        BOT -->|پاسخ‌ها| TG
    end

    subgraph Application
        BOT --> MW[میدل‌ور<br/>مسدودی · تعمیرات · جوین اجباری]
        MW --> H[هندلرها<br/>کاربر · ادمین · پشتیبانی]
        H --> DB[(SQLite<br/>data/bot.db)]
        H --> I18N[i18n]
        H --> PAY[درایورهای پرداخت]
        H --> XUI[کلاینت پنل 3x-ui]
    end

    subgraph External
        PAY --> GW[درگاه‌های پرداخت]
        GW -->|کال‌بک| WH[سرور وب‌هوک]
        WH --> H
        RATE[jobs / rate] --> API[API نرخ دلار]
        XUI --> PANEL[پنل 3x-ui]
    end

    subgraph Background
        JOBS[jobs] -->|بکاپ| BK[data/backups]
        JOBS -->|همگام‌سازی · ترافیک · گزارش| H
    end
```

## 📦 مسئولیت پکیج‌ها

| پکیج | مسئولیت | وابستگی |
| :--- | :--- | :--- |
| `internal/bot` | ساخت ربات، شروع polling، سرور وب‌هوک و کارها | config، handlers، database، jobs |
| `internal/config` | خواندن و اعتبارسنجی `.env` | godotenv |
| `internal/database` | واسط ذخیره‌سازی و پیاده‌سازی SQLite | درایور sqlite |
| `internal/handlers` | تبدیل آپدیت‌ها به عملیات دامنه | database، i18n، payment، xui |
| `internal/i18n` | جستجوی پیام برای `fa` و `en` | — |
| `internal/menus` | ساخت کیبوردها | i18n |
| `internal/middleware` | دروازه‌های پیش از هندلر | database، config |
| `internal/payment` | درایورهای درگاه پشت یک واسط | config |
| `internal/rate` | نرخ دلار با کش و درصد سود | config |
| `internal/jobs` | کارهای زمان‌بندی‌شده | database، notify |
| `internal/webhook` | دریافت HTTP کال‌بک | handlers، payment |
| `internal/xui` | کلاینت API پنل 3x-ui | config |
| `internal/notify` | اطلاع‌رسانی خروجی و لاگ حسابرسی | database |

### قاعدهٔ لایه‌بندی

`handlers` و `jobs` می‌توانند به `database`، `payment`، `rate`، `i18n` و `xui` وابسته باشند.
`database`، `payment`، `rate` و `i18n` هرگز نباید `handlers` را ایمپورت کنند.
`config` به‌جز `isenv` چیزی از `internal` ایمپورت نمی‌کند.

---

## 💳 جریان شارژ

```mermaid
sequenceDiagram
    participant U as کاربر
    participant B as ربات
    participant H as هندلرها
    participant DB as SQLite
    participant G as درگاه
    participant W as وب‌هوک

    U->>B: /charge
    B->>H: نمایش درگاه‌ها
    H->>DB: خواندن درگاه‌های فعال
    DB-->>H: فهرست
    H-->>U: کیبورد درگاه
    U->>H: انتخاب درگاه
    H-->>U: کیبورد مبلغ
    U->>H: انتخاب مبلغ
    H->>H: تبدیل دلار به تومان (نرخ × درصد سود)
    H->>DB: ساخت سفارش (در انتظار)
    H->>G: درخواست پرداخت
    G-->>H: لینک پرداخت + authority
    H-->>U: لینک پرداخت
    U->>G: پرداخت
    G->>W: POST /webhook/{gateway}
    W->>G: اعتبارسنجی
    G-->>W: تأیید
    W->>DB: تراکنش + موجودی + سفارش پرداخت‌شده
    W->>U: پیام تأیید
```

تضمین‌های کلیدی:

- سفارش **قبل از** نمایش لینک پرداخت ذخیره می‌شود تا از دست رفتن پاسخ باعث از بین رفتن پول نشود.
- به‌روزرسانی موجودی همراه با درج در `transactions` در یک تراکنش انجام می‌شود.
- پردازش وب‌هوک تکرارپذیر (idempotent) است: کال‌بک تکراری برای سفارش پرداخت‌شده بدون افزایش دوبارهٔ موجودی تأیید می‌شود.

---

## 🌐 جریان زبان

```mermaid
sequenceDiagram
    participant U as کاربر
    participant B as ربات
    participant I as i18n
    participant DB as SQLite

    U->>B: زدن دکمهٔ زبان
    B->>DB: خواندن زبان کاربر
    DB-->>B: fa | en
    B->>I: T(lang, key, args)
    I-->>B: رشتهٔ ترجمه‌شده
    B-->>U: پیام به زبان انتخاب‌شده
    U->>B: انتخاب زبان دیگر
    B->>DB: ذخیرهٔ زبان کاربر
```

- ترجمه‌ها در یک نقشه در `internal/i18n/i18n.go` با کلید شناسهٔ پیام قرار دارند.
- هر هندلر از `T(lang, key, args...)` استفاده می‌کند؛ هیچ متن فارسی یا انگلیسی‌ای مستقیم در هندلرها نوشته نمی‌شود.
- `LangLabel()` نام قابل‌نمایش زبان را برای منوها فراهم می‌کند.

---

## 🔐 امنیت

### کنترل دسترسی

```mermaid
graph LR
    A[آیدی فرستنده] --> B{برابر ADMIN_ID؟}
    B -->|بله| OK[دسترسی کامل مدیریت]
    B -->|خیر| C{در جدول admins؟}
    C -->|بله| D[دسترسی محدود به دپارتمان]
    C -->|خیر| E[فقط دسترسی کاربری]
```

- هویت ادمین در هر کال‌بک مدیریتی بررسی می‌شود، نه فقط هنگام باز کردن پنل.
- ادمین محدود به دپارتمان تنها تیکت‌های همان دپارتمان‌ها را می‌بیند.
- کاربران مسدود در میدل‌ور پیش از هر هندلری حذف می‌شوند.

### اعتبارسنجی ورودی

- مبالغ با `CHARGE_MIN` و `CHARGE_MAX` بررسی می‌شوند.
- تصویر رسید از نظر نوع و حجم پیش از ذخیره بررسی می‌شود.
- متن ورودی کاربر پیش از پاسخ HTML امن‌سازی می‌شود.
- انتخاب درگاه با درگاه‌های فعال اعتبارسنجی می‌شود.

### اسرار

`BOT_TOKEN`، کلیدهای درگاه و اسرار وب‌هوک فقط در `.env` قرار دارند. هرگز در دیتابیس نوشته نمی‌شوند، لاگ نمی‌شوند و در چت نمایش داده نمی‌شوند.

---

## 🗄️ مدل داده

```mermaid
erDiagram
    USERS ||--o{ TRANSACTIONS : دارد
    USERS ||--o{ ORDERS : ثبت‌می‌کند
    USERS ||--o{ TICKETS : بازمی‌کند
    USERS ||--o{ MANUAL_CHARGES : ارسال‌می‌کند
    ADMINS ||--o{ TICKETS : پاسخ‌می‌دهد
    DEPARTMENTS ||--o{ TICKETS : گروه‌بندی
    CATEGORIES ||--o{ PRODUCTS : شامل
    PRODUCTS ||--o{ ORDERS : فروخته‌می‌شود
    ORDERS ||--o| SUBSCRIPTIONS : می‌سازد
```

- حالت WAL باعث می‌شود خواننده‌ها هنگام بکاپ‌گیری بلاک نشوند.
- مهاجرت‌ها در `migrate.go` در هر استارت اعمال می‌شوند و اجرای مجددشان بی‌خطر است.

---

## ⚙️ کارهای پس‌زمینه

| کار | بازه | هدف |
| :--- | :--- | :--- |
| `backup` | زمان‌بندی‌شده | کپی `bot.db` در `data/backups/` |
| `sync` | زمان‌بندی‌شده | به‌روزرسانی وضعیت پنل و داده‌های اشتراک |
| `traffic` | دوره‌ای | خواندن مصرف و هشدار به کاربران نزدیک سقف |
| `daily_report` | روزانه | ارسال آمار روز به مدیر اصلی |

هر کار در گوروتین خودش اجرا می‌شود، پس از panic بازیابی می‌شود، خطاها را لاگ می‌کند و هرگز پردازش آپدیت‌ها را بلاک نمی‌کند.

---

## 📈 توسعهٔ ربات

۱. **دستور جدید کاربر** — هندلر در `internal/handlers/` اضافه و در `handlers.go` ثبت کنید.
۲. **صفحهٔ جدید مدیریت** — فایل در `internal/handlers/admin/` و دکمه در `panel.go`.
۳. **رشتهٔ جدید** — کلید را در هر دو نقشهٔ `fa` و `en` در `i18n.go` اضافه کنید.
۴. **جدول جدید** — `schema.sql` را گسترش دهید، کوئری‌ها را در `internal/database/sqlite/` بنویسید و در `migrate.go` متصل کنید.
۵. **درگاه جدید** — واسط `payment/gateway.go` را پیاده‌سازی و شاخهٔ وب‌هوک را اضافه کنید.
۶. **زبان جدید** — ثابت `Lang`، نقشهٔ ترجمه و گزینهٔ منوی زبان را اضافه کنید.

---

## 🧪 راهبرد تست

- **واحد** — کمک‌تابع‌های خالص: تبدیل ارز، جستجوی i18n، اعتبارسنجی.
- **یکپارچه** — جریان هندلرها روی یک فایل SQLite موقت.
- **وب‌هوک** — بدنه را با کلید تست امضا کنید و تکرارپذیری را بررسی کنید.
- **دستی** — جریان کامل شارژ روی سندباکس درگاه.

---

## 🔗 بیشتر

[راهنمای استفاده](./usage.md) · [مرجع API](./API.md) · [نصب](./setup.md) · [ساختار](../../STRUCTURE.fa.md) · [نسخهٔ انگلیسی](../en/architecture.md)
