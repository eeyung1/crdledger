# crdledger

**A shared, confirmed ledger for the credit that already happens informally — between roommates, hostel-mates, and small vendor–customer relationships — replacing memory, notebooks, and "he-said-she-said" with one record both sides agree to.**

---

## The problem

In any tight community — a hostel, a dorm, a group of friends who buy things for each other — informal credit happens constantly. Someone fronts someone else money for provisions, a meal, phone credit, a favor. Today that debt lives in exactly one of three places:

- **Someone's memory** — which fades, gets disputed, or simply forgets.
- **A physical notebook** — which can be lost, damaged, or only exists in one person's hands.
- **Word of mouth** — which has no evidence at all when two people remember it differently.

All three share the same fatal flaw: **the record is one-sided.** Only the person who is owed money keeps track, and the other party has no way to confirm, dispute, or even see it. That's not a bookkeeping inconvenience — it's a trust problem. Disputes over "I don't remember agreeing to that" or "I already paid you for that" damage real relationships, not just a ledger balance.

## The solution

crdledger turns a one-sided IOU into a **two-sided, confirmed record.** Nothing counts as a real debt until both people agree it happened. Every entry has to be accepted or rejected by the other party before it ever touches anyone's balance — so the ledger reflects what both sides actually agreed to, not just what one side wrote down.

**No real money moves through the platform.** This is deliberately not a payments product — it's a trust and record-keeping layer for credit that already exists between people. That keeps it simple, keeps it fast, and keeps it out of the regulatory and liability weight that comes with handling money.

## How it works

1. **Either side can start the record.** A seller can log a debt someone owes them ("Record a transaction"), or — just as easily — a buyer can self-report something they bought and owe for ("Add orders"). Both paths lead to the exact same place.
2. **The other party gets notified on their dashboard**, under "Needs your response," with the full detail: who, how much, what for, and an optional receipt photo as evidence.
3. **They Accept or Reject it.** Accept, and it instantly becomes a real, counted debt on both people's balances. Reject, and it's kept as a visible, timestamped record of the dispute — nothing gets silently deleted, but it never affects anyone's numbers.
4. **Balances update automatically** — both people see a live, running total of what they owe and are owed, with no arithmetic required from either side.
5. **When the debt is settled, the seller marks it paid** — in full or partially, with a real payment date — and it drops off the active balance while staying in history.

## What it does for people

| Problem today | What crdledger does |
|---|---|
| One-sided notebook entries anyone can dispute | Every entry requires the other party's explicit Accept/Reject before it counts |
| No record of a rejected or disputed claim | Rejected entries are kept, visible, and timestamped — a real dispute history |
| "How much do I owe everyone, total?" requires mental math | One headline **net position** number, plus a full breakdown of who you owe and who owes you |
| Confusing back-and-forth debts between the same two people | Automatic **netting** — if you owe someone ₦500 and they owe you ₦300, the ledger shows one clean ₦200, not two contradictory lines |
| Forgetting to chase someone for money | A "days pending" indicator per debt, plus a one-tap **Copy reminder** button with a pre-written, friendly nudge to paste into WhatsApp or wherever you already talk |
| "I paid but they say I didn't" | Receipt/proof photos can be attached to any entry, and every payment is timestamped |
| Needing this data outside the app | One-click **CSV export** of your full transaction history |
| Only being able to pay all at once | **Partial payments** are tracked — pay some now, some later, and the remaining balance updates in real time |
| A messy, undifferentiated list of transactions | Separate, searchable **Creditors** and **Debtors** views, so "who owes me" and "who I owe" are never mixed together |

## Everyday features

- **Simple accounts** — username, password, a display name people recognize you by, and an optional profile photo.
- **Two ways to record a debt** — the person owed money logs it, *or* the person who owes logs it themselves via **Add Orders**; either way, the other side has to confirm it.
- **Dashboard at a glance** — net position, total owed to you, total you owe, a trailing trend line, and top-5 bar charts of who you owe/are owed the most.
- **Mobile-first, installable app** — works like a native app on a phone (add-to-homescreen), loads fast with no heavy frameworks, and keeps working even with a spotty connection.
- **Every action confirmed, nothing silent** — accept, reject, mark paid, and record actions all give instant, inline feedback.

## Why it's trustworthy by design

- **No unilateral debt creation.** Nobody can add a debt to someone else's balance without that person explicitly agreeing to it.
- **Nothing is ever silently deleted.** Rejected entries stay as a visible record, not a black hole.
- **Only the real parties can act.** Only the person actually on the other side of a transaction can confirm or reject it, and only the original seller can mark it paid — enforced on the server, not just hidden in the interface.
- **Passwords are hashed, sessions are secure**, and every state-changing action is protected against cross-site request forgery — standard, boring, correct security practice.

## Who it's for

Originally built for an informal hostel economy of roughly 200 people extending each other small amounts of credit for everyday goods — provisions, phone credit, favors. The same shape fits any group of people who informally front each other money and currently rely on memory or a notebook to keep track: roommates, campus communities, small social lending circles, or a vendor with a regular set of customers who buy on credit.

## The one-line pitch

> **crdledger replaces "trust me, I remember" with "we both agreed to this" — a shared, confirmed ledger for the credit people already extend each other, with no money ever moving through the platform.**
