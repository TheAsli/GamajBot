<div align="center">

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./API.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](../fa/API.md)

[![README](https://img.shields.io/badge/README-English-0d6efd?style=flat-square&logo=readme&logoColor=white)](../../README.md)
[![Usage](https://img.shields.io/badge/Usage-6f42c1?style=flat-square&logo=bookstack&logoColor=white)](./usage.md)
[![API](https://img.shields.io/badge/API-0d6efd?style=flat-square&logo=swagger&logoColor=white)](./API.md)
[![Architecture](https://img.shields.io/badge/Architecture-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./architecture.md)
[![Setup](https://img.shields.io/badge/Setup-fd7e14?style=flat-square&logo=rocket&logoColor=white)](./setup.md)
[![Structure](https://img.shields.io/badge/Structure-20c997?style=flat-square&logo=files&logoColor=white)](../../STRUCTURE.md)
[![Changelog](https://img.shields.io/badge/Changelog-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](../../CHANGELOG.md)
[![Contributing](https://img.shields.io/badge/Contributing-0d6efd?style=flat-square&logo=git&logoColor=white)](../../CONTRIBUTING.md)

</div>

# API Reference

GamajBot is a Telegram bot, not a REST service: users interact through the Bot API. The only HTTP surface the process exposes is the **webhook server** that receives payment-gateway callbacks. This document describes both that surface and the internal operations the handlers perform.

---

## 1. Webhook server

Started by `internal/webhook/server.go` on `WEBHOOK_PORT`.

### `POST /webhook/{gateway}`

Callback endpoint used by payment providers.

| Part | Value |
| :--- | :--- |
| Method | `POST` |
| Path | `/webhook/tetraminator`, `/webhook/cubepay`, `/webhook/aban` |
| Content-Type | `application/json` (provider dependent) |
| Auth | Provider signature / secret token in the payload or query |

**Tetraminator example payload**

```json
{
  "order_id": "TRX-20261018-0001",
  "status": "paid",
  "amount": 250000,
  "authority": "A0000000000000000000000000000000abcd",
  "signature": "hex-hmac"
}
```

**Flow**

1. The gateway driver validates the signature.
2. The order is looked up and its status is compared with the callback.
3. A confirmed payment writes a `transactions` row, updates the balance and marks the order paid.
4. The user receives a confirmation message through `internal/handlers/notify`.

**Responses**

| Status | Meaning |
| :--- | :--- |
| `200 OK` | Callback accepted and processed |
| `400 Bad Request` | Malformed body or unknown order |
| `401 Unauthorized` | Signature validation failed |
| `404 Not Found` | Unknown gateway path |
| `500 Internal Server Error` | Persistence failure (the provider will retry) |

### `GET /health`

Lightweight liveness probe returning `200 OK` with `{"status":"ok"}`.

---

## 2. Internal operations

These are executed by handlers, not exposed as routes. They are listed so integrators can follow the data model.

### Users

| Operation | Input | Effect |
| :--- | :--- | :--- |
| Get user | `user_id` | Returns profile, balance, language, ban state |
| Update balance | `user_id`, `delta`, `reason` | Writes a `transactions` row and updates the balance atomically |
| Set language | `user_id`, `fa` \| `en` | Persists the user language preference |
| Ban / unban | `user_id`, `bool` | Toggles the ban flag checked by `internal/middleware` |

### Payments

| Operation | Input | Effect |
| :--- | :--- | :--- |
| Create charge | `user_id`, `amount`, `gateway` | Returns gateway payment URL and order ID |
| Verify charge | `order_id` | Queries the gateway and finalizes the order |
| Manual receipt | `user_id`, `photo`, `amount` | Creates a `manual_charges` row pending admin review |
| Approve receipt | `charge_id`, `admin_id` | Credits the wallet and notifies the user |

### Admin

| Operation | Input | Effect |
| :--- | :--- | :--- |
| List users | pagination, filter | Paginated user list |
| Broadcast | `text` | Sends a message to every user, returns delivery counts |
| Add admin | `user_id`, `role`, `departments` | Creates an `admins` row |
| Create department | `name` | Creates a department |
| Assign ticket | `ticket_id`, `department_id` | Routes the ticket |
| Update settings | `key`, `value` | Writes to the `settings` table |

---

## 3. Response envelope

Internal helpers return a consistent shape:

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

| Code | English | فارسی | Meaning |
| :--- | :--- | :--- | :--- |
| 400 | Invalid parameter | پارامتر نامعتبر | Input failed validation |
| 401 | Unauthorized | دسترسی غیرمجاز | Missing or wrong credentials |
| 403 | Forbidden | دسترسی ممنوع | Authenticated but not allowed |
| 404 | Not found | یافت نشد | Unknown resource |
| 409 | Conflict | تداخل | Duplicate order or receipt |
| 429 | Too many requests | درخواست بیش از حد | Rate limit hit |
| 500 | Server error | خطای سرور | Internal failure |

---

## 4. Rate limits

| Scope | Limit |
| :--- | :--- |
| Per user, per minute | 30 requests |
| Burst | 5 rapid requests |
| Webhook per gateway | provider-defined retry window |

---

## 5. Authentication

- **Telegram users** are identified by the numeric ID that arrives inside every update; there is no token to manage.
- **Admins** are authorized by comparing the sender ID with `ADMIN_ID` and with the `admins` table.
- **Webhooks** are authorized by the signature or shared secret configured per gateway.
- **External integrations** (if you add one) should use a dedicated bearer token stored in `.env`; never reuse `BOT_TOKEN`.

---

## 6. Versioning

- API surface version: `v1`
- Bot version: `is.0.0.1`
- Additive changes do not bump the version; breaking payload changes do and are documented in the changelog.

---

## 🔗 More

[Usage guide](./usage.md) · [Architecture](./architecture.md) · [Setup](./setup.md) · [Structure](../../STRUCTURE.md) · [Persian version](../fa/API.md)
