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