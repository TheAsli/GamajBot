<div align="center">


# Contributing to GamajBot

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./CONTRIBUTING.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./CONTRIBUTING.fa.md)

[![README](https://img.shields.io/badge/README-English-0d6efd?style=flat-square&logo=readme&logoColor=white)](./README.md)
[![Changelog](https://img.shields.io/badge/Changelog-e83e8c?style=flat-square&logo=keepachangelog&logoColor=white)](./CHANGELOG.md)
[![Structure](https://img.shields.io/badge/Structure-20c997?style=flat-square&logo=files&logoColor=white)](./STRUCTURE.md)
[![Architecture](https://img.shields.io/badge/Architecture-e83e8c?style=flat-square&logo=diagramsdotnet&logoColor=white)](./Docs/en/architecture.md)

</div>

Thanks for taking the time to improve GamajBot. This guide keeps the project
consistent and reviewable — please read it before opening a pull request.

---

## Requirements

| Tool | Version | Needed for |
| :--- | :--- | :--- |
| Go | 1.27.1+ | building and running the bot |
| Git | any recent | version control |
| Node.js | 20+ *(optional)* | documentation website only |

## Getting started

```bash
git clone https://github.com/theasli/GamajBot.git
cd GamajBot
go mod tidy
cp .env.example .env      # set BOT_TOKEN and ADMIN_ID at minimum
go run cmd/AsliCode/main.go
```

Before pushing, make sure these pass locally:

```bash
go build ./...
bash scripts/check-doc-links.sh
bash scripts/check-spdx-headers.sh
```

Docs-site changes only: `npm install && npm run docs:build`.

---

## Project layout

Start with [STRUCTURE.md](./STRUCTURE.md) — it lists every package and where new
code belongs. In short:

| You want to… | Go to |
| :--- | :--- |
| add a user command | `internal/handlers/` + register in `handlers.go` |
| add an admin screen | `internal/handlers/admin/` + button in `panel.go` |
| add a table or query | `internal/database/sqlite/` + `schema.sql` + `migrate.go` |
| add a payment gateway | `internal/payment/` implementing `gateway.go` |
| add a scheduled task | `internal/jobs/` + start it in `internal/bot/bot.go` |

---

## Coding conventions

### Go

- Run `gofmt` on every file you touch; keep the tree `go vet` clean.
- Respect the layering rule: `database`, `payment`, `rate` and `i18n` must never
  import `handlers`. `config` imports nothing from `internal` except `isenv`.
- Errors are logged or returned — never swallowed silently.
- Keep functions focused; long handlers should be split into helpers.

### Bilingual text (important)

The bot supports **exactly two languages: Persian (`fa`) and English (`en`)**.
No other language may be added.

- Every user-visible string lives in `internal/i18n/i18n.go`.
- A new key must be added to **both** maps (`LangFA` and `LangEN`) in the same commit.
- Handlers render text only through `T(lang, key, args...)` — never inline a
  Persian or English literal in a handler.
- The language menu is generated from the `Lang` constants; do not hardcode it.

### Database

- Update `internal/database/sqlite/schema.sql` together with `migrate.go`.
- Migrations must be safe to run repeatedly on an existing database.
- Never write destructive `DROP`/`DELETE` migrations without a backup step.

### Secrets and privacy

- Secrets belong in `.env` only (see `.env.example`). Never commit real tokens,
  gateway keys or webhook secrets, and never print them in logs or messages.
- `data/` holds live user data and is git-ignored. Files that must never be
  released go to the `نه/` quarantine folder — it is ignored as well.

---

## Documentation

Documentation ships in both languages:

- English: `README.md`, `STRUCTURE.md`, `Docs/en/*`
- Persian: `README.fa.md`, `STRUCTURE.fa.md`, `Docs/fa/*`

Rules:

1. Update **both** language versions in the same pull request.
2. Keep the language-switcher and navigation badges at the top of each page.
3. Run `bash scripts/check-doc-links.sh` — CI runs the same check and fails on
   any relative link that points to a missing file.

---

## Workflow

1. Fork the repository and create a branch: `feat/short-name`, `fix/short-name`
   or `docs/short-name`.
2. Make focused commits. Suggested prefixes: `feat:`, `fix:`, `docs:`, `refactor:`,
   `chore:` — the subject line should say *why*, not *what*.
3. Keep one logical change per pull request.
4. Open the pull request against `master` and fill in what changed and how you
   tested it.

### Pull-request checklist

- [ ] `go build ./...` succeeds
- [ ] `bash scripts/check-doc-links.sh` reports no broken links
- [ ] every new `.go` file starts with the AGPL-3.0 SPDX header (`bash scripts/check-spdx-headers.sh` passes)
- [ ] new user-facing text exists in **both** `fa` and `en`
- [ ] database changes include schema **and** migration updates
- [ ] no secrets, tokens or runtime data added to the repository
- [ ] docs updated in both languages when behaviour changed

---

## Reporting bugs

Include the bot version, Go version, the exact command or button pressed, the
startup log and (if relevant) the webhook log. Never paste your `.env` values —
redact tokens before sharing.

For anything security-related, please contact the maintainer privately through
[@AsliCode](https://t.me/AsliCode) instead of opening a public issue.

---

## License

By contributing you agree that your work is released under the
[AGPL-3.0](./LICENSE) license that covers this project.

<div align="center">

**Thank you for contributing ❤️**

[![English](https://img.shields.io/badge/lang-English-blue?style=flat-square)](./CONTRIBUTING.md)
[![فارسی](https://img.shields.io/badge/lang-%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-green?style=flat-square)](./CONTRIBUTING.fa.md)

</div>
