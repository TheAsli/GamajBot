<div align="center">

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](../en/API.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./API.md)

[![README فارسی](https://img.shields.io/badge/README-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square&logo=readme&logoColor=white)](../../README.fa.md)
[![استفاده](https://img.shields.io/badge/%D8%A7%D8%B3%D8%AA%D9%81%D8%A7%D8%AF%D9%87-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./usage.md)
[![API](https://img.shields.io/badge/API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./API.md)
[![معماری](https://img.shields.io/badge/%D9%85%D8%B9%D9%85%D8%A7%D8%B1%DB%8C-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./architecture.md)
[![نصب](https://img.shields.io/badge/%D9%86%D8%B5%D8%A8-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./setup.md)
[![ساختار](https://img.shields.io/badge/%D8%B3%D8%A7%D8%AE%D8%AA%D8%A7%D8%B1-20c997?style=flat-square&logo=files&logoColor=white)](../../STRUCTURE.fa.md)
[![تغییرات](https://img.shields.io/badge/%D8%AA%D8%BA%DB%8C%DB%8C%D8%B1%D8%A7%D8%AA-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](../../CHANGELOG.fa.md)
[![مشارکت](https://img.shields.io/badge/%D9%85%D8%B4%D8%A7%D8%B1%DA%A9%D8%AA-0d6efd?style=flat-square&logo=git&logoColor=white)](../../CONTRIBUTING.fa.md)

</div>

# مرجع API

GamajBot یک ربات تلگرام است، نه یک سرویس REST؛ کاربران از طریق Bot API با آن کار می‌کنند. تنها سطح HTTP که این فرایند ارائه می‌دهد، **سرور وب‌هوک** برای دریافت کال‌بک درگاه‌های پرداخت است. این سند هم همان سطح را توضیح می‌دهد و هم عملیات داخلی‌ای که هندلرها انجام می‌دهند.

---

## ۱. سرور وب‌هوک

توسط `internal/webhook/server.go` روی پورت `WEBHOOK_PORT` اجرا می‌شود.

### `POST /webhook/{gateway}`

نقطهٔ دریافت کال‌بک از ارائه‌دهندهٔ پرداخت.

| بخش | مقدار |
| :--- | :--- |
| متد | `POST` |
| مسیر | `/webhook/tetraminator`، `/webhook/cubepay`، `/webhook/aban` |
| نوع محتوا | `application/json` (بسته به درگاه) |
| احراز هویت | امضای ارائه‌دهنده یا توکن مشترک در بدنه یا query |

**نمونهٔ بدنهٔ تترامیناتور**

```json
{
  "order_id": "TRX-20261018-0001",
  "status": "paid",
  "amount": 250000,
  "authority": "A0000000000000000000000000000000abcd",
  "signature": "hex-hmac"
}
```

**جریان کار**

۱. درایور درگاه امضا را اعتبارسنجی می‌کند.
۲. سفارش خوانده و وضعیتش با کال‌بک مقایسه می‌شود.
۳. پرداخت تأییدشده یک ردیف در `transactions` می‌نویسد، موجودی را به‌روز می‌کند و سفارش را پرداخت‌شده علامت می‌زند.
۴. پیام تأیید از طریق `internal/handlers/notify` برای کاربر ارسال می‌شود.

**پاسخ‌ها**

| وضعیت | معنی |
| :--- | :--- |
| `200 OK` | کال‌بک پذیرفته و پردازش شد |
| `400 Bad Request` | بدنهٔ نامعتبر یا سفارش ناشناس |
| `401 Unauthorized` | اعتبارسنجی امضا شکست خورد |
| `404 Not Found` | مسیر درگاه ناشناس |
| `500 Internal Server Error` | خطای ذخیره‌سازی (ارائه‌دهنده دوباره تلاش می‌کند) |

### `GET /health`

بررسی سلامت ساده که `200 OK` با `{"status":"ok"}` برمی‌گرداند.

---

## ۲. عملیات داخلی

این عملیات توسط هندلرها اجرا می‌شوند و مسیر HTTP ندارند؛ برای درک مدل داده فهرست شده‌اند.

### کاربران

| عملیات | ورودی | اثر |
| :--- | :--- | :--- |
| دریافت کاربر | `user_id` | پروفایل، موجودی، زبان و وضعیت مسدودی |
| تغییر موجودی | `user_id`، `delta`، `reason` | نوشتن ردیف `transactions` و به‌روزرسانی اتمی موجودی |
| تعیین زبان | `user_id`، `fa` \| `en` | ذخیرهٔ زبان کاربر |
| مسدود/آزاد | `user_id`، `bool` | تغییر پرچم مسدودی که `internal/middleware` بررسی می‌کند |

### پرداخت

| عملیات | ورودی | اثر |
| :--- | :--- | :--- |
| ساخت شارژ | `user_id`، `amount`، `gateway` | برگرداندن لینک پرداخت و شناسهٔ سفارش |
| بررسی شارژ | `order_id` | کوئری از درگاه و نهایی‌کردن سفارش |
| رسید دستی | `user_id`، `photo`، `amount` | ساخت ردیف `manual_charges` در انتظار بررسی ادمین |
| تأیید رسید | `charge_id`، `admin_id` | افزایش موجودی و اطلاع به کاربر |

### مدیریت

| عملیات | ورودی | اثر |
| :--- | :--- | :--- |
| فهرست کاربران | صفحه‌بندی، فیلتر | فهرست صفحه‌بندی‌شده |
| پیام همگانی | `text` | ارسال به همهٔ کاربران و برگرداندن آمار ارسال |
| افزودن ادمین | `user_id`، `role`، `departments` | ساخت ردیف `admins` |
| ساخت دپارتمان | `name` | ایجاد دپارتمان |
| تخصیص تیکت | `ticket_id`، `department_id` | مسیریابی تیکت |
| تغییر تنظیمات | `key`، `value` | نوشتن در جدول `settings` |

---

## ۳. قالب پاسخ

کمک‌تابع‌های داخلی قالب یکسانی برمی‌گردانند:

```json
{
  "success": true,
  "data": { },
  "error": null
}
```

```json
{
  "success": false,
  "data": null,
  "error": "invalid parameter"
}
```

| کد | فارسی | English | معنی |
| :--- | :--- | :--- | :--- |
| 400 | پارامتر نامعتبر | Invalid parameter | ورودی اعتبارسنجی نشد |
| 401 | دسترسی غیرمجاز | Unauthorized | اعتبارنامهٔ نامعتبر یا ناموجود |
| 403 | دسترسی ممنوع | Forbidden | احراز هویت شده ولی مجاز نیست |
| 404 | یافت نشد | Not found | منبع ناشناس |
| 409 | تداخل | Conflict | سفارش یا رسید تکراری |
| 429 | درخواست بیش از حد | Too many requests | عبور از محدودیت نرخ |
| 500 | خطای سرور | Server error | خطای داخلی |

---

## ۴. محدودیت نرخ

| دامنه | محدودیت |
| :--- | :--- |
| هر کاربر در دقیقه | ۳۰ درخواست |
| انفجار (Burst) | ۵ درخواست سریع |
| وب‌هوک هر درگاه | پنجرهٔ تلاش مجدد تعیین‌شده توسط ارائه‌دهنده |

---

## ۵. احراز هویت

- **کاربران تلگرام** با آیدی عددی‌ای شناسایی می‌شوند که در هر آپدیت می‌آید؛ توکنی برای مدیریت وجود ندارد.
- **مدیران** با مقایسهٔ آیدی فرستنده با `ADMIN_ID` و جدول `admins` مجاز می‌شوند.
- **وب‌هوک‌ها** با امضا یا کلید مشترک تعریف‌شده برای هر درگاه مجاز می‌شوند.
- **یکپارچه‌سازی خارجی** (اگر افزودید) باید از توکن Bearer اختصاصی در `.env` استفاده کند؛ هرگز `BOT_TOKEN` را دوباره به کار نبرید.

---

## ۶. نسخه‌گذاری

- نسخهٔ سطح API: `v1`
- نسخهٔ ربات: `is.0.0.1`
- تغییرات افزودنی نسخه را بالا نمی‌برند؛ تغییرات شکننده در بدنه، نسخه را بالا می‌برد و در تغییرات (changelog) ثبت می‌شود.

---

## 🔗 بیشتر

[راهنمای استفاده](./usage.md) · [معماری](./architecture.md) · [نصب](./setup.md) · [ساختار](../../STRUCTURE.fa.md) · [نسخهٔ انگلیسی](../en/API.md)
