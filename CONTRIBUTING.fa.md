<div align="center">


# راهنمای مشارکت در GamajBot

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./CONTRIBUTING.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./CONTRIBUTING.fa.md)

[![README فارسی](https://img.shields.io/badge/README-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square&logo=readme&logoColor=white)](./README.fa.md)
[![تغییرات](https://img.shields.io/badge/%D8%AA%D8%BA%DB%8C%DB%8C%D8%B1%D8%A7%D8%AA-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](./CHANGELOG.md)
[![ساختار](https://img.shields.io/badge/%D8%B3%D8%A7%D8%AE%D8%AA%D8%A7%D8%B1-20c997?style=flat-square&logo=files&logoColor=white)](./STRUCTURE.fa.md)
[![معماری](https://img.shields.io/badge/%D9%85%D8%B9%D9%85%D8%A7%D8%B1%DB%8C-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./Docs/fa/architecture.md)

</div>

ممنون که برای بهتر شدن GamajBot وقت می‌گذارید. این راهنما کمک می‌کند پروژه
یکدست بماند و مرور تغییرات ساده باشد — لطفاً قبل از ارسال Pull Request آن را بخوانید.

---

## پیش‌نیازها

| ابزار | نسخه | برای چه کاری |
| :--- | :--- | :--- |
| Go | 1.27.1 یا بالاتر | ساخت و اجرای ربات |
| Git | هر نسخهٔ جدید | کنترل نسخه |
| Node.js | ۲۰ یا بالاتر *(اختیاری)* | فقط برای سایت مستندات |

## شروع کار

```bash
git clone https://github.com/theasli/GamajBot.git
cd GamajBot
go mod tidy
cp .env.example .env      # حداقل BOT_TOKEN و ADMIN_ID را تنظیم کنید
go run cmd/AsliCode/main.go
```

قبل از ارسال تغییرات، مطمئن شوید این‌ها لوکال پاس می‌شوند:

```bash
go build ./...
bash scripts/check-doc-links.sh
bash scripts/check-spdx-headers.sh
```

اگر فقط سایت مستندات را تغییر داده‌اید: `npm install && npm run docs:build`.

---

## ساختار پروژه

از [STRUCTURE.fa.md](./STRUCTURE.fa.md) شروع کنید — همهٔ پکیج‌ها و محل قرارگیری
کد جدید را فهرست کرده است. خلاصه:

| می‌خواهید… | محل |
| :--- | :--- |
| دستور کاربر اضافه کنید | `internal/handlers/` + ثبت در `handlers.go` |
| صفحهٔ ادمین اضافه کنید | `internal/handlers/admin/` + دکمه در `panel.go` |
| جدول یا کوئری اضافه کنید | `internal/database/sqlite/` + `schema.sql` + `migrate.go` |
| درگاه پرداخت اضافه کنید | `internal/payment/` با پیاده‌سازی `gateway.go` |
| کار زمان‌بندی‌شده اضافه کنید | `internal/jobs/` + شروع آن در `internal/bot/bot.go` |

---

## قواعد کدنویسی

### Go

- روی همهٔ فایل‌هایی که دست می‌زنید `gofmt` اجرا کنید و درخت پروژه را از نظر `go vet` تمیز نگه دارید.
- قاعدهٔ لایه‌بندی را رعایت کنید: `database`، `payment`، `rate` و `i18n` هرگز نباید
  `handlers` را ایمپورت کنند. `config` به‌جز `isenv` چیزی از `internal` ایمپورت نمی‌کند.
- خطاها یا لاگ می‌شوند یا برگردانده می‌شوند — هرگز بی‌صدا بلعیده نمی‌شوند.
- توابع را کوچک و متمرکز نگه دارید؛ هندلرهای طولانی را به کمک‌تابع بشکنید.

### متن دوزبانه (مهم)

ربات **فقط دو زبان دارد: فارسی (`fa`) و انگلیسی (`en`)** و افزودن زبان دیگر مجاز نیست.

- هر متن قابل‌مشاهده برای کاربر در `internal/i18n/i18n.go` نگهداری می‌شود.
- کلید جدید باید در **هر دو** نقشهٔ `LangFA` و `LangEN` و در یک کامیت اضافه شود.
- هندلرها فقط از طریق `T(lang, key, args...)` متن را رندر می‌کنند — متن فارسی یا
  انگلیسی را هرگز مستقیم داخل هندلر ننویسید.
- منوی زبان از روی ثابت‌های `Lang` ساخته می‌شود؛ آن را هاردکد نکنید.

### دیتابیس

- `internal/database/sqlite/schema.sql` را همراه با `migrate.go` به‌روزرسانی کنید.
- مهاجرت‌ها باید روی دیتابیس موجود چندبار قابل اجرا باشند.
- بدون مرحلهٔ بکاپ، مهاجرت مخرب (`DROP` / `DELETE`) ننویسید.

### اسرار و حریم خصوصی

- اسرار فقط در `.env` قرار می‌گیرند (نمونه: `.env.example`). هرگز توکن واقعی، کلید
  درگاه یا secret وب‌هوک را کامیت نکنید و در لاگ یا پیام چاپ نکنید.
- پوشهٔ `data/` دادهٔ واقعی کاربران است و در گیت نادیده گرفته می‌شود. فایل‌هایی که
  نباید منتشر شوند به پوشهٔ قرنطینهٔ `نه/` می‌روند که آن هم ignore شده است.

---

## مستندات

مستندات در دو زبان منتشر می‌شود:

- انگلیسی: `README.md`، `STRUCTURE.md`، `Docs/en/*`
- فارسی: `README.fa.md`، `STRUCTURE.fa.md`، `Docs/fa/*`

قواعد:

۱. هر دو نسخهٔ زبانی را در همان Pull Request به‌روزرسانی کنید.
۲. دکمهٔ تغییر زبان و دکمه‌های ناوبری بالای هر صفحه را حفظ کنید.
۳. `bash scripts/check-doc-links.sh` را اجرا کنید — CI همین بررسی را انجام می‌دهد و
   روی هر لینک نسبی شکسته شکست می‌خورد.

---

## روند کار

۱. مخزن را Fork کنید و یک شاخه بسازید: `feat/نام-کوتاه`، `fix/نام-کوتاه` یا `docs/نام-کوتاه`.
۲. کامیت‌های متمرکز بزنید. پیشوندهای پیشنهادی: `feat:`، `fix:`، `docs:`، `refactor:`،
   `chore:` — متن کامیت باید بگوید *چرا*، نه فقط *چه*.
۳. هر Pull Request فقط یک تغییر منطقی داشته باشد.
۴. Pull Request را روی `master` باز کنید و بنویسید چه چیزی عوض شده و چطور تست شده است.

### چک‌لیست Pull Request

- [ ] `go build ./...` موفق است
- [ ] `bash scripts/check-doc-links.sh` لینک شکسته‌ای گزارش نمی‌کند
- [ ] هر فایل `.go` جدید با سرصفحهٔ SPDX لایسنس AGPL-3.0 شروع می‌شود (`bash scripts/check-spdx-headers.sh` پاس می‌دهد)
- [ ] متن جدید قابل‌مشاهده در **هر دو** زبان `fa` و `en` وجود دارد
- [ ] تغییرات دیتابیس شامل اسکیما **و** مهاجرت است
- [ ] هیچ secret، توکن یا دادهٔ زمان اجرا به مخزن اضافه نشده است
- [ ] مستندات هر دو زبان در صورت تغییر رفتار به‌روزرسانی شده‌اند

---

## گزارش باگ

نسخهٔ ربات، نسخهٔ Go، دستور یا دکمهٔ دقیق، لاگ شروع و (در صورت نیاز) لاگ وب‌هوک را
بفرستید. مقادیر `.env` را هرگز منتشر نکنید — توکن‌ها را قبل از اشتراک‌گذاری پاک کنید.

برای مسائل امنیتی، به‌جای Issue عمومی با [@AsliCode](https://t.me/AsliCode) خصوصی تماس بگیرید.

---

## مجوز

با مشارکت، موافقت می‌کنید که کار شما تحت [مجوز AGPL-3.0](./LICENSE) پروژه منتشر شود.

<div align="center">

**ممنون از مشارکت شما ❤️**

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./CONTRIBUTING.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./CONTRIBUTING.fa.md)

</div>
