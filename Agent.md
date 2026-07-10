# Agent.md — Credit Tracker MVP

You are assisting with development on the **Credit Tracker MVP**, a shared
credit ledger for an informal hostel economy of ~200 fellows. Sellers extend
credit to buyers for goods (provisions, perfume, clothes, etc.); this product
replaces memory, notebooks, and word-of-mouth with a single shared record
both parties can see. **No real money moves through this platform** — it is
a trust and record-keeping tool, not a payments product.

Read this entire file before making any changes. If a request is ambiguous
or not covered here, ask before writing code — do not assume.

---

## THE ONE RULE THAT OVERRIDES ALL OTHERS

> Does this make recording and viewing credit transactions simpler and more
> trustworthy than a notebook? If not, cut it.

Every feature suggestion, every abstraction, every "while we're at it" idea
gets tested against this question before it gets built. This product handles
trust between people who live together — friction, ambiguity, or confusion
here damages real relationships, not just a support ticket queue. When in
doubt, favor the simpler, more explicit option, even if it feels
under-engineered by normal software standards. Under-engineered and shipped
beats polished and theoretical for this project.

---

## TECH STACK (do not introduce a different one)

- **Language**: Go, standard library `net/http` only. No Gin, Echo, Fiber,
  or any other web framework — this is a deliberate constraint, not an
  oversight, even though other projects this developer works on use Gin.
- **Database**: SQLite. Zero-configuration, single-file, sufficient for 200
  users. Do not suggest Postgres/MySQL "for scalability" — that is solving a
  problem this product doesn't have yet.
- **Frontend**: Go templates (server-side rendering) + vanilla JavaScript +
  plain CSS. No React/Vue/frontend framework, no build step. Mobile-first —
  the real users are non-technical fellows on basic phone browsers, not
  developers on desktops. Every screen must be usable and legible on a small
  phone screen with one thumb.
- **Auth**: Simple session-based auth (username + password, bcrypt-hashed).
  No OAuth, no JWT, no third-party auth provider for the MVP — sessions are
  sufficient and simpler to reason about for this scale.
- **Deployment**: Render or Railway, single binary.

---

## ARCHITECTURE (strict layering — same discipline as this developer's other Go projects, done right from day one this time)

```
cmd/
  main.go
internal/
  handler/     — parses requests, calls service, writes responses. No SQL. No business logic.
  service/     — all business logic lives here. Framework-agnostic.
  repository/  — all database operations. No business logic.
  models/      — plain structs.
  middleware/  — session/auth checks.
static/
  css/
templates/
```

- **Handlers never touch the database directly.** They call a service
  function and translate the result into an HTTP response.
- **Repositories never contain business logic.** They execute SQL and
  return data or an error — nothing else.
- **Services hold every rule** (e.g. "a transaction can only be marked paid
  by the seller who created it," "amount must be positive") — this is the
  one layer that should be unit-testable without spinning up a database.

This is a stricter separation than the Convention Management System or
L2EStudyLink use (both mix repository directly into handlers in places) —
that's intentional here. Follow it exactly, every milestone, from
Milestone 1 onward. Don't let a later milestone quietly add SQL to a handler
"just this once."

---

## DATABASE SCHEMA (minimal, do not expand without explicit request)

```sql
Users:
  id, username, password_hash, display_name, created_at

Transactions:
  id, seller_id, buyer_id, amount, description,
  status (pending | paid), created_at, paid_at
```

No additional tables unless the project owner explicitly asks for one. In
particular, resist the urge to add: a notifications table, an audit log
table, a categories/items table, or a disputes table — none of these are
part of the three problems this MVP solves. If a future milestone seems to
need one, ask first; don't add it preemptively "for later."

---

## THE THREE PROBLEMS THIS MVP SOLVES (and only these three)

1. **No shared record** → a transaction ledger visible to both seller and buyer.
2. **No easy way to see total debt at a glance** → a dashboard: total owed /
   total receivable + a list of individual transactions.
3. **No confirmation when debt is settled** → seller marks paid, buyer sees
   the updated status.

If a feature request doesn't map to one of these three problems, stop and
ask whether it actually belongs in the MVP, rather than building it.

---

## MILESTONES — BUILD IN THIS ORDER, ONE AT A TIME

1. **Project structure + schema** — Go program starts, connects to SQLite,
   creates tables. Nothing visible yet; foundation only.
2. **Registration + login** — session management working; logged-in user
   sees their name on a dashboard.
3. **Record a credit transaction** — seller records buyer's username,
   amount, description; saved to DB.
4. **View balances** — both parties see the transaction; seller sees total
   receivables, buyer sees total owed.
5. **Mark as paid** — seller marks paid, status updates, both parties see
   the update.
6. **Mobile-friendly UI** — clean, fast, usable on a phone browser.
7. **Deploy** — real URL, real users, first real transaction this week.

**Do not skip ahead or combine milestones.** Each milestone must produce
working, testable software before the next one starts. Explain *why* a
milestone exists before building it (this file already gives the reasoning
— restate it briefly, don't skip it). After each milestone, stop and wait
for the project owner to confirm it works before proceeding to the next.

---

## LESSONS FROM THIS DEVELOPER'S OTHER PROJECTS — APPLY PROACTIVELY HERE

This project owner has hit the same categories of mistakes across three
other codebases (a hostel admin dashboard, a church convention system, and
a tutoring platform). Since this project is greenfield, prevent these from
the start rather than fixing them later:

1. **`.gitignore` must correctly exclude `.env` from commit #1** — not just
   contain the line, but actually be verified with
   `git log --all --full-history -- .env` before any real credentials ever
   touch the file. One prior project had `.env` committed with `.gitignore`
   present but the relevant lines commented out; another had a live
   production database password pasted in plaintext during a debugging
   session. Treat any credential shared in chat, a screenshot, or a terminal
   paste as compromised — prompt the project owner to redact secrets before
   sharing, and rotate anything that does slip through.
2. **Never hardcode a session/auth secret in source.** Read it from an
   environment variable, and fail startup loudly if it's missing — don't
   let a placeholder value silently ship (this exact mistake — a hardcoded
   JWT secret — was a live auth bypass in production with 150 real users on
   another of this developer's projects).
3. **Don't let one broad `if err != nil` swallow every possible error into
   one misleading message.** Distinguish "no such user" from "something
   actually broke" — a masked permission/connection error was mistaken for
   a business-logic error for over an hour on another project because of
   this exact pattern. Check for the specific error condition you mean
   (e.g. `sql.ErrNoRows`), and let genuine failures surface as genuine
   failures.
4. **Don't commit backup files, stale drafts, or dead template files** —
   delete them as soon as they're superseded, and add patterns like
   `*.backup*` to `.gitignore` immediately, not after they've accumulated.
5. **Add the mobile viewport meta tag to every template from the start** —
   `<meta name="viewport" content="width=device-width, initial-scale=1.0">`.
   Given the real users here are phone-only, a missing viewport tag isn't a
   cosmetic bug, it's the product not working for its actual audience.
6. **Package-level variables initialized from `os.Getenv()` at program
   startup will capture an empty value if `.env` hasn't loaded yet** — read
   environment variables inside the function that needs them, not as a
   `var` at package scope, unless you're certain of initialization order.
7. **Keep schema files and actual running migrations in sync.** If the code
   ever references a table not in the schema file, that's a bug to fix
   immediately, not a note for later.

---

## WORKFLOW RULES (strict — follow exactly)

1. **One step at a time.** Give one file, or one small paired edit, per
   response. Do not bundle a whole milestone into one giant reply.
2. After each step, **stop and wait** for confirmation that the change was
   made and tested before giving the next step.
3. **Never assume unstated requirements.** If a request is ambiguous, ask a
   clarifying question before writing code.
4. **Follow the strict handler/service/repository layering** described
   above, from Milestone 1 onward — no exceptions, no "just this once."
5. When giving code, give the exact full block to paste — not a vague diff
   — unless the edit is a small, precise change to one existing file.
6. **Push back on scope creep, including from the project owner.** If a
   request would add a feature, table, or abstraction not required by the
   three core problems or the current milestone, name that plainly and ask
   whether it should really be in the MVP, rather than quietly building it.
7. Optimize every decision for: shipping a working product to one real
   seller and one real buyer this week — not for theoretical future scale.

Wait for the project owner to confirm they're ready to start Milestone 1
before writing any code.
