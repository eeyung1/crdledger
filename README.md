# crdledger — Credit Tracker MVP

A shared credit ledger for an informal hostel economy (~200 fellows).
Sellers extend credit to buyers for goods; this replaces memory, notebooks,
and word-of-mouth with one shared record both parties can see.

**No real money moves through this platform.** It's a trust and
record-keeping tool, not a payments product.

The one rule every feature below is tested against:

> Does this make recording and viewing credit transactions simpler and
> more trustworthy than a notebook? If not, cut it.

Check items off in order as you build. One milestone at a time — don't
skip ahead, and don't start the next milestone until the current one
works end to end.

---

## Milestone 1 — Project structure + schema
*Foundation only. Nothing visible yet.*

- [x] Go project structure (`cmd/`, `internal/handler`, `internal/service`, `internal/repository`, `internal/models`, `internal/middleware`, `static/`, `templates/`)
- [x] SQLite connection on startup
- [x] `Users` table created (id, username, password_hash, display_name, created_at)
- [x] `Transactions` table created (id, seller_id, buyer_id, amount, description, status, created_at, paid_at)
- [x] `.gitignore` in place and verified — `.env` never committed
- [x] Session secret read from environment variable, startup fails loudly if missing

---

## Milestone 2 — Registration + login
*The one login flow every screen after this depends on.*

- [x] Registration page (username, password, display name)
- [x] Passwords hashed with bcrypt before storage
- [x] Login page (username + password)
- [x] Session created on successful login
- [x] Logged-in user sees their own name on the dashboard
- [x] Logout

---

## Milestone 3 — Record a credit transaction
*The write side of "no shared record."*

- [x] "Record transaction" form: buyer's username, amount, description
- [x] Amount validated as positive
- [x] Transaction saved with status `pending`
- [x] Only a logged-in user can record a transaction (as the seller)

---

## Milestone 4 — View balances
*The read side of "no shared record," plus at-a-glance totals.*

- [x] Dashboard: seller's total receivable
- [x] Dashboard: buyer's total owed
- [x] Transaction list — role-aware (each row shows correctly whether you're the seller or the buyer on it)
- [x] Search/filter the transaction list by counterpart's name or username

---

## Milestone 5 — Mark as paid
*Confirmation when debt is settled.*

- [x] "Mark as paid" action, visible only to the seller on that transaction
- [x] Status updates to `paid`, `paid_at` timestamp recorded
- [x] Only the original seller can trigger this (enforced in the service layer, not just hidden in the UI)
- [x] Buyer sees the updated status next time they view the list

---

## Milestone 6 — Mobile-friendly UI
*The real users are on basic phone browsers, one thumb.*

- [x] Viewport meta tag on every template
- [x] Responsive, legible layout on a small phone screen
- [x] Fast load — no build step, no heavy assets

---

## Milestone 7 — Deploy
*A real URL, real users, first real transaction this week.*

- [ ] Deployed as a single binary to Render or Railway
- [ ] Environment variables configured on the host (session secret, DB path)
- [ ] First real transaction recorded between one real seller and one real buyer
---

## Milestone 8 — Optimization: profile photos + creditor/debtor split
*Refining the MVP after initial real-world use.*

- [x] `users` table gains a photo column; users can upload a profile photo
- [x] Decide and document photo storage approach (local disk vs. external storage) given host filesystem persistence
- [x] Left sidebar navigation: Edit profile, Record a transaction, Transactions
- [x] "Edit profile" page — photo upload moved here, off the dashboard
- [x] Dashboard simplified to: welcome + photo, total receivable, total owed (sidebar handles navigation)
- [x] Transactions link leads to a page with two choices: "Creditors" (people who owe you) and "Debtors" (people you owe)
- [x] Each choice shows the existing role-aware, searchable transaction list, filtered to just that role
- [x] Mark-as-paid and existing search/filter behavior preserved on the split lists


# RUN CODE

SESSION_SECRET=temp-local-value go run ./cmd
---

## Production-readiness + UI pass (this update)

Everything below stays inside the original constraints in `Agent.md`:
plain Go templates, vanilla CSS/JS, **no framework, no build step**.

### Design
- New glass/ledger visual identity (dark navy + gold "wax seal" accent,
  serif numerals for money, monospace for amounts) — see
  `static/css/style.css`.
- **Desktop**: persistent glass sidebar nav.
  **Mobile**: sticky top bar + glass bottom tab bar with a raised
  "record transaction" button — genuinely different layouts per
  breakpoint, not just a squeezed desktop view.
- 16px+ input font sizes (no iOS auto-zoom), visible focus rings,
  `prefers-reduced-motion` respected, empty states, inline toasts.

### PWA (installable app)
- `static/manifest.json`, `static/service-worker.js`, app icons in
  `static/icons/`, offline fallback at `static/offline.html`.
- Service worker: network-first for pages (balances must never be
  stale), cache-first for static assets, so it's genuinely useful
  offline without ever showing a stale ledger.
- Install prompt wired up in `static/js/app.js` (progressive
  enhancement — everything still works with JS off).

### Production hardening
- CSRF protection (double-submit cookie) on every state-changing
  request — `internal/middleware/csrf.go`.
- Security headers (CSP, X-Frame-Options, nosniff, etc.) and optional
  HSTS — `internal/middleware/security.go`.
- Session cookies are now `Secure` (env-gated), `HttpOnly`,
  `SameSite=Lax`, and expire after 30 days instead of living forever
  in memory.
- Per-IP rate limiting on login/register —
  `internal/middleware/ratelimit.go`.
- Panic recovery + structured JSON request logging.
- Config read once in `main()` via `internal/config` (never a
  package-level `os.Getenv` — see the lesson already in `Agent.md`).
- Graceful shutdown on SIGTERM/SIGINT, `/healthz` endpoint, DB indexes
  on `seller_id`/`buyer_id`, WAL mode.
- `Dockerfile` for a single-binary deploy to Render/Railway/Fly, plus
  `.env.example`.

### Try it locally
```
cp .env.example .env      # edit SESSION_SECRET
go run ./cmd
```
Open http://localhost:8080 on your phone (same network) or desktop —
Chrome/Edge will offer an "Install" prompt once served over HTTPS (or
localhost, which counts as a secure context).

---

## HTMX layer (this update)

Self-hosted (`static/js/htmx.min.js`, no CDN, ~50KB) — one `<script>` tag,
zero build step, fits the same constraints as everything else here.

Wired into exactly three interactions, each chosen because it has a real
UX payoff over a full-page reload — not applied blanket-wide (no
`hx-boost`):

1. **Mark as paid** (`templates/tx_row.html`) — swaps just that one row
   in place instead of redirecting. `internal/handler/payment.go` detects
   `HX-Request` and re-renders the single row (success *or* failure —
   errors show inline on the row, not as a bare "500" flash).
2. **Live search** (`templates/transactions_list.html`) — debounced
   `hx-get` on the search input swaps `#tx-list-wrap`, with
   `hx-push-url` so the URL/back-button still works.
   `internal/handler/transactions_list.go` returns the
   `tx_list_fragment` partial instead of a full page when it's an HTMX
   request.
3. **Record a transaction** (`templates/record_form.html`) — submits
   inline; on success the form clears with a "Recorded ✓" banner so
   someone settling several debts in one sitting never leaves the page.
   `internal/handler/transaction.go`'s `RecordFormData` echoes back
   whatever was typed on a validation error, so a typo doesn't cost you
   the whole form.

**Every one of these still works with JavaScript off or htmx failing to
load**: each `hx-post`/`hx-get` sits on a real `<form action="..."
method="...">` (or `hx-push-url`'d GET), so the browser's native
submission is the fallback, and the handlers branch on the `HX-Request`
header to decide fragment vs. full-page response. Progressive
enhancement, not a requirement.

---

## Status page, netting, and features (this update)

Per the companion product brief. Same constraints: no framework, no
build step, hand-rolled inline SVG charts (no charting library).

### Status page spec

**Decision: folded into the Dashboard rather than a separate page.**
The brief's own spec for a status page — "headline number above the
fold, two supporting numbers, one chart chosen by the actual question"
— is what a dashboard's home screen should already be. Adding a second,
separate "/status" page would either duplicate the dashboard or steal
its reason to exist; a ruthless editor cuts the redundant screen, not
the redundant work. `templates/dashboard.html` / `internal/handler/dashboard.go`.

- **Headline, above the fold**: net position (`TotalReceivable -
  TotalOwed`) as one signed number — `+142.50` in emerald if you're owed
  more, `-80.00` in amber if you owe more, neutral/no-hue if exactly
  `0`. A trailing 8-week sparkline (`<polyline>`, no axes/gridlines)
  sits beside it purely to answer "trending up or down," not to be a
  full chart. Owed-to-you / you-owe sit directly below as smaller
  supporting numbers.
- **Two bar charts, only when they earn their place**: "Who owes you
  the most" / "who you owe the most" — each a `<rect>` per person with
  a server-computed width (`internal/service/balance.go:splitAndRankNet`),
  capped at 5 bars. Name + amount are real HTML text next to each bar,
  never color-only. **They only render once there's more than one
  counterparty** (`ShowCreditorChart`/`ShowDebtorChart`) — with a single
  person, the total above already says everything the bar would; per
  the brief, a chart comparing one thing to nothing is dead weight.
- **States**: zero balance ("You're all settled up," no chart); one
  counterparty (numbers only, no bar chart); many (both charts, top 5
  each); loading (full-page render is the loading state — no client
  fetch, so no skeleton needed here); error (banner + retry, balance
  numbers never fall back to a fabricated `0`).
- **Debt netting** happens before any of this reaches the template —
  `splitAndRankNet` collapses "Alice owes Bob ₦500 / Bob owes Alice
  ₦300" into one `+₦200` bar, never two contradictory line items. The
  underlying transaction ledger is untouched (each transaction still has
  its own pay/mark-paid lifecycle) — netting is a view, not a rewrite of
  the source of truth.

### Feature shortlist

| Feature | Precedent | Problem it solves | Verdict |
|---|---|---|---|
| **Debt netting** | Splitwise/Settle Up's core simplification | Two contradictory line items read as less certain than one net figure — directly serves "who owes who, how much" | **Built** — dashboard bar charts |
| **Gentle reminders (manual)** | Near-universal in IOU apps | Waiting to be reminded is the actual friction point | **Built, manual only** — "days pending" indicator + a "Copy reminder" button that fills the clipboard with a pre-written nudge, paste into whatever chat app you already use. No email/push infra, no new trust surface (frequency, tone, opt-out) — exactly the brief's "evaluate manual first" |
| **CSV export** | Nearly universal | Honest answer to "what if I need this outside the app" | **Built** — `/transactions/export.csv`, one query, stdlib `encoding/csv` |
| **Receipt/proof photo** | Splitwise itemized receipts | Reuses existing photo-upload infra; real payoff in a dispute | **Built, optional** — one extra `<input type="file">` on the record form, never required |
| **Audit info (who/when)** | Universal in ledger products | Byproduct of `created_at`/`paid_at`, already captured | **Built, lightweight** — "12 days pending" / "paid · Jan 2" surfaced on each row, no new table |

### Explicitly rejected

- **Automatic reminders** (scheduled emails/push) — bigger trust surface
  (frequency, tone, opt-out) than manual copy-and-paste; only worth
  building once manual is proven wanted. Deferred, not built.
- **Budgeting / spending categories / "where does my money go"** — a
  different product (Lunch Money, Halfway). crdledger answers "who owes
  who," not "what did I spend on."
- **Bank connections / auto-import** — the entire value of this app is
  money that *doesn't* move through a bank. Connecting one would blur
  the exact thing that makes it useful.
- **Social feed / activity feed / gamification streaks** — the category's
  most-cited way a focused IOU tool gets worse trying to become bigger.
- **Multi-currency** — no stated need yet; not building for a
  hypothetical.
- **A separate `/status` route** — see decision above; would have
  duplicated the dashboard.

### Build order (cheapest relative to how much it helps "who owes who")

1. Debt netting (pure computation over existing data, zero schema change)
2. CSV export (one query, stdlib only)
3. Manual reminders (clipboard only, zero backend)
4. Audit info surfacing (data already existed)
5. Receipt photos (small schema change + reused upload infra)
6. *(deferred)* Automatic reminders — only after manual usage justifies it
