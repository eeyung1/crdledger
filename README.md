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

- [ ] Go project structure (`cmd/`, `internal/handler`, `internal/service`, `internal/repository`, `internal/models`, `internal/middleware`, `static/`, `templates/`)
- [ ] SQLite connection on startup
- [ ] `Users` table created (id, username, password_hash, display_name, created_at)
- [ ] `Transactions` table created (id, seller_id, buyer_id, amount, description, status, created_at, paid_at)
- [ ] `.gitignore` in place and verified — `.env` never committed
- [ ] Session secret read from environment variable, startup fails loudly if missing

---

## Milestone 2 — Registration + login
*The one login flow every screen after this depends on.*

- [ ] Registration page (username, password, display name)
- [ ] Passwords hashed with bcrypt before storage
- [ ] Login page (username + password)
- [ ] Session created on successful login
- [ ] Logged-in user sees their own name on the dashboard
- [ ] Logout

---

## Milestone 3 — Record a credit transaction
*The write side of "no shared record."*

- [ ] "Record transaction" form: buyer's username, amount, description
- [ ] Amount validated as positive
- [ ] Transaction saved with status `pending`
- [ ] Only a logged-in user can record a transaction (as the seller)

---

## Milestone 4 — View balances
*The read side of "no shared record," plus at-a-glance totals.*

- [ ] Dashboard: seller's total receivable
- [ ] Dashboard: buyer's total owed
- [ ] Transaction list — role-aware (each row shows correctly whether you're the seller or the buyer on it)
- [ ] Search/filter the transaction list by counterpart's name or username

---

## Milestone 5 — Mark as paid
*Confirmation when debt is settled.*

- [ ] "Mark as paid" action, visible only to the seller on that transaction
- [ ] Status updates to `paid`, `paid_at` timestamp recorded
- [ ] Only the original seller can trigger this (enforced in the service layer, not just hidden in the UI)
- [ ] Buyer sees the updated status next time they view the list

---

## Milestone 6 — Mobile-friendly UI
*The real users are on basic phone browsers, one thumb.*

- [ ] Viewport meta tag on every template
- [ ] Responsive, legible layout on a small phone screen
- [ ] Fast load — no build step, no heavy assets

---

## Milestone 7 — Deploy
*A real URL, real users, first real transaction this week.*

- [ ] Deployed as a single binary to Render or Railway
- [ ] Environment variables configured on the host (session secret, DB path)
- [ ] First real transaction recorded between one real seller and one real buyer