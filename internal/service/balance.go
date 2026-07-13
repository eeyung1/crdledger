package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"crdledger/internal/models"
	"crdledger/internal/repository"
)

var ErrNotParticipant = errors.New("you are not a participant in this transaction")

// sparklineWeeks controls how far back the dashboard trend line looks.
// Chosen to answer "trending up or down lately", not to be a full
// analytical chart — a handful of points is enough for direction.
const sparklineWeeks = 8

// TransactionView is a role-aware presentation of a transaction from the
// perspective of the currently logged-in user.
type TransactionView struct {
	ID                    int64
	CounterpartName       string
	CounterpartUsername   string
	Role                  string // "You are owed" or "You owe"
	Amount                float64
	Description           string
	Status                string
	IsSeller              bool
	CreatedAt             time.Time
	PaidAt                *time.Time
	PhotoPath             string  // optional receipt photo, empty if none was attached
	DaysPending           int     // 0 once paid; only meaningful while Status == "pending"
	ReminderText          string  // pre-written nudge, only set when a reminder is sensible to offer
	AmountPaid            float64 // running total paid so far
	RemainingBalance      float64 // Amount - AmountPaid; 0 once fully paid
	IsPartial             bool    // true when some (but not all) of the amount has been paid
	ConfirmationStatus    string  // "pending", "confirmed", or "rejected"
	IsPendingConfirmation bool    // true while the buyer hasn't responded yet
	IsRejected            bool    // true if the buyer rejected this record
	CanRespond            bool    // true when the viewer is the buyer and a response is still needed
}

// CounterpartyNet is one bar in the "who owes who the most" chart — pending
// transactions with the same person are already netted together here, so
// two contradictory line items (you owe them ₦500, they owe you ₦300)
// become one number (they owe you ₦200 net) before it ever reaches the UI.
type CounterpartyNet struct {
	Name        string
	Username    string
	Amount      float64 // always positive — direction is which slice it's in
	BarWidthPct float64 // 0-100, relative to the largest bar in its chart
}

// Sparkline is a trailing net-position trend, expressed as ready-to-render
// SVG polyline points (0-100 x, 0-30 y, y already flipped so "up" is up).
type Sparkline struct {
	Points    string
	TrendUp   bool
	TrendFlat bool
}

type Balance struct {
	TotalReceivable float64 // owed to this user as seller
	TotalOwed       float64 // this user owes as buyer
	NetPosition     float64 // TotalReceivable - TotalOwed, the single headline number
	Transactions    []TransactionView
	NeedsResponse   []TransactionView // pending transactions where this user is the buyer — awaiting Accept/Reject
	TopCreditors    []CounterpartyNet // people who owe this user, netted, sorted desc, capped
	TopDebtors      []CounterpartyNet // people this user owes, netted, sorted desc, capped
	Sparkline       Sparkline
}

type BalanceService struct {
	transactions *repository.TransactionRepository
	users        *repository.UserRepository
}

func NewBalanceService(transactions *repository.TransactionRepository, users *repository.UserRepository) *BalanceService {
	return &BalanceService{transactions: transactions, users: users}
}

func (s *BalanceService) GetBalance(userID int64) (*Balance, error) {
	txs, err := s.transactions.GetByUser(userID)
	if err != nil {
		return nil, err
	}

	balance := &Balance{}
	netByCounterparty := make(map[string]float64) // key: "name|username"

	for _, t := range txs {
		view, err := s.toView(t, userID)
		if err != nil {
			return nil, err
		}

		// Only confirmed transactions count toward anyone's numbers — a
		// pending or rejected entry is visible (for dispute history / so
		// the buyer can respond) but never moves the balance.
		if t.Status == "pending" && t.ConfirmationStatus == models.ConfirmationConfirmed {
			remaining := t.Amount - t.AmountPaid
			key := view.CounterpartName + "|" + view.CounterpartUsername
			if t.SellerID == userID {
				balance.TotalReceivable += remaining
				netByCounterparty[key] += remaining
			} else {
				balance.TotalOwed += remaining
				netByCounterparty[key] -= remaining
			}
		}

		if t.ConfirmationStatus == models.ConfirmationPending && t.BuyerID == userID {
			balance.NeedsResponse = append(balance.NeedsResponse, view)
		}

		balance.Transactions = append(balance.Transactions, view)
	}

	balance.NetPosition = balance.TotalReceivable - balance.TotalOwed
	balance.TopCreditors, balance.TopDebtors = splitAndRankNet(netByCounterparty)
	balance.Sparkline = buildSparkline(txs, userID, sparklineWeeks)

	return balance, nil
}

// splitAndRankNet turns "who owes who" into the two bar charts the status
// page needs — one per question ("who owes me most" / "who do I owe
// most"), each capped at 5 bars so the chart stays scannable rather than
// becoming a second transaction list.
func splitAndRankNet(net map[string]float64) (creditors, debtors []CounterpartyNet) {
	const maxBars = 5

	type entry struct {
		name, username string
		net            float64
	}
	var entries []entry
	for key, amount := range net {
		if amount == 0 {
			continue
		}
		name, username, _ := strings.Cut(key, "|")
		entries = append(entries, entry{name: name, username: username, net: amount})
	}

	var pos, neg []entry
	for _, e := range entries {
		if e.net > 0 {
			pos = append(pos, e)
		} else {
			neg = append(neg, e)
		}
	}

	sort.Slice(pos, func(i, j int) bool { return pos[i].net > pos[j].net })
	sort.Slice(neg, func(i, j int) bool { return neg[i].net < neg[j].net }) // most negative first

	toBars := func(es []entry, magnitude func(entry) float64) []CounterpartyNet {
		if len(es) > maxBars {
			es = es[:maxBars]
		}
		var max float64
		for _, e := range es {
			if m := magnitude(e); m > max {
				max = m
			}
		}
		bars := make([]CounterpartyNet, len(es))
		for i, e := range es {
			amt := magnitude(e)
			pct := 100.0
			if max > 0 {
				pct = (amt / max) * 100
			}
			bars[i] = CounterpartyNet{Name: e.name, Username: e.username, Amount: amt, BarWidthPct: pct}
		}
		return bars
	}

	creditors = toBars(pos, func(e entry) float64 { return e.net })
	debtors = toBars(neg, func(e entry) float64 { return -e.net })
	return creditors, debtors
}

// buildSparkline computes net position (receivable minus owed) at each of
// the last `weeks` week-boundaries, purely from data already captured —
// created_at and paid_at — with no extra table or history to maintain. A
// transaction counts toward the balance at time T if it existed by T and
// (was still pending at T, or hadn't been paid yet by T).
func buildSparkline(txs []models.Transaction, userID int64, weeks int) Sparkline {
	now := time.Now()
	values := make([]float64, weeks+1)

	for i := 0; i <= weeks; i++ {
		asOf := now.AddDate(0, 0, -7*(weeks-i))
		var net float64
		for _, t := range txs {
			if t.ConfirmationStatus != models.ConfirmationConfirmed {
				continue // never confirmed (or since rejected) — never part of the balance
			}
			if t.CreatedAt.After(asOf) {
				continue
			}
			if t.PaidAt != nil && !t.PaidAt.After(asOf) {
				continue // was already settled by this point in time
			}
			if t.SellerID == userID {
				net += t.Amount
			} else if t.BuyerID == userID {
				net -= t.Amount
			}
		}
		values[i] = net
	}

	return Sparkline{
		Points:    sparklinePoints(values),
		TrendUp:   values[len(values)-1] > values[0],
		TrendFlat: values[len(values)-1] == values[0],
	}
}

func sparklinePoints(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	spread := max - min
	var b strings.Builder
	for i, v := range values {
		x := 0.0
		if len(values) > 1 {
			x = float64(i) / float64(len(values)-1) * 100
		}
		y := 15.0 // flat middle line when every point is equal
		if spread > 0 {
			y = 30 - ((v-min)/spread)*30
		}
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%.1f,%.1f", x, y)
	}
	return b.String()
}

// GetTransactionView returns the up-to-date, role-aware view of a single
// transaction — used to re-render just one row after an HTMX action
// (e.g. mark-paid) instead of reloading the whole list.
func (s *BalanceService) GetTransactionView(transactionID, userID int64) (TransactionView, error) {
	t, err := s.transactions.GetByID(transactionID)
	if err != nil {
		return TransactionView{}, err
	}
	if t.SellerID != userID && t.BuyerID != userID {
		return TransactionView{}, ErrNotParticipant
	}
	return s.toView(*t, userID)
}

func (s *BalanceService) toView(t models.Transaction, userID int64) (TransactionView, error) {
	isSeller := t.SellerID == userID

	var counterpartID int64
	var role string
	if isSeller {
		counterpartID = t.BuyerID
		role = "Creditor (you're owed money)"
	} else {
		counterpartID = t.SellerID
		role = "Debtor (you owe money)"
	}

	counterpart, err := s.users.GetByID(counterpartID)
	if err != nil {
		return TransactionView{}, err
	}

	remaining := t.Amount - t.AmountPaid
	view := TransactionView{
		ID:                  t.ID,
		CounterpartName:     counterpart.DisplayName,
		CounterpartUsername: counterpart.Username,
		Role:                role,
		Amount:              t.Amount,
		Description:         t.Description,
		Status:              t.Status,
		IsSeller:            isSeller,
		CreatedAt:           t.CreatedAt,
		PaidAt:              t.PaidAt,
		AmountPaid:          t.AmountPaid,
		RemainingBalance:    remaining,
		IsPartial:           t.Status == "pending" && t.AmountPaid > 0,
		ConfirmationStatus:    t.ConfirmationStatus,
		IsPendingConfirmation: t.ConfirmationStatus == models.ConfirmationPending,
		IsRejected:            t.ConfirmationStatus == models.ConfirmationRejected,
		CanRespond:            !isSeller && t.ConfirmationStatus == models.ConfirmationPending,
	}
	if t.PhotoPath != nil {
		view.PhotoPath = *t.PhotoPath
	}

	if t.Status == "pending" {
		view.DaysPending = int(time.Since(t.CreatedAt).Hours() / 24)
		if isSeller {
			view.ReminderText = fmt.Sprintf(
				"Hey %s — friendly reminder, you still owe me %.2f for %s.",
				counterpart.DisplayName, t.Amount, t.Description,
			)
		}
	}

	return view, nil
}

func FilterTransactions(views []TransactionView, query string) []TransactionView {
	if query == "" {
		return views
	}

	query = strings.ToLower(query)
	var filtered []TransactionView
	for _, v := range views {
		if strings.Contains(strings.ToLower(v.CounterpartName), query) ||
			strings.Contains(strings.ToLower(v.CounterpartUsername), query) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func FilterByRole(views []TransactionView, isSeller bool) []TransactionView {
	var filtered []TransactionView
	for _, v := range views {
		if v.IsSeller == isSeller {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
