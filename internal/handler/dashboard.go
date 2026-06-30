package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"
)

type DashboardHandler struct {
	authService      *service.AuthService
	transactionService *service.TransactionService
	sessionMgr       *middleware.SessionManager
}

func NewDashboardHandler(authService *service.AuthService, transactionService *service.TransactionService, sessionMgr *middleware.SessionManager) *DashboardHandler {
	return &DashboardHandler{
		authService:      authService,
		transactionService: transactionService,
		sessionMgr:       sessionMgr,
	}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)
	username := r.Context().Value("username").(string)

	_, err := h.authService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get user's transactions
	transactions, err := h.transactionService.GetTransactionsByUser(userID)
	if err != nil {
		http.Error(w, "Failed to load transactions", http.StatusInternalServerError)
		return
	}

	// Calculate totals
	var totalReceivable int64
	var totalPayable int64
	var receivableTransactions []repository.Transaction
	var payableTransactions []repository.Transaction

	for _, t := range transactions {
		if t.Status == "pending" {
			if t.SellerID == userID {
				// User is seller, so they are owed money
				totalReceivable += t.Amount
				receivableTransactions = append(receivableTransactions, t)
			} else if t.BuyerID == userID {
				// User is buyer, so they owe money
				totalPayable += t.Amount
				payableTransactions = append(payableTransactions, t)
			}
		}
	}
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>CRDLEDGER - Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 30px; }
        .welcome { font-size: 1.5em; color: #333; }
        .logout-btn { background: #f44336; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .logout-btn:hover { background: #da190b; }
        .tab-container { display: flex; margin-bottom: 20px; }
        .tab { flex: 1; padding: 12px; text-align: center; background: #f5f5f5; border: none; cursor: pointer; }
        .tab.active { background: #4CAF50; color: white; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .transaction-list { margin-top: 20px; }
        .transaction-item {
            display: flex;
            justify-content: space-between;
            padding: 12px;
            border: 1px solid #eee;
            border-radius: 4px;
            margin-bottom: 10px;
        }
        .transaction-info { flex: 1; }
        .transaction-amount { font-weight: bold; }
        .transaction-status {
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.9em;
        }
        .status-pending { background: #fff3cd; color: #856404; }
        .status-paid { background: #d4edda; color: #155724; }
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        .btn-primary { background: #2196F3; color: white; }
        .btn-primary:hover { background: #0b7dda; }
        .btn-success { background: #4CAF50; color: white; }
        .btn-success:hover { background: #45a049; }
        .btn-danger { background: #f44336; color: white; }
        .btn-danger:hover { background: #da190b; }
        .form-section {
            background: #f9f9f9;
            padding: 20px;
            border-radius: 8px;
            margin-top: 20px;
        }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .form-group input, .form-group select, .form-group textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        .form-group button {
            margin-top: 10px;
        }
        .balance-summary {
            display: flex;
            justify-content: space-between;
            margin-bottom: 20px;
        }
        .balance-box {
            flex: 1;
            text-align: center;
            padding: 15px;
            background: #e3f2fd;
            border-radius: 8px;
            margin: 0 5px;
        }
        .balance-label {
            font-size: 0.9em;
            color: #666;
        }
        .balance-amount {
            font-size: 1.5em;
            font-weight: bold;
            color: #1976d2;
        }
        .mark-paid-btn {
            background: #4CAF50;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9em;
        }
        .mark-paid-btn:hover {
            background: #45a049;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="welcome">Welcome, %s!</div>
        <form method="POST" action="/logout" id="logout-form" style="display: inline;">
            <button type="submit" class="logout-btn">Logout</button>
        </form>
    </div>

    <div class="tab-container">
        <button class="tab active" onclick="showTab('overview')">Overview</button>
        <button class="tab" onclick="showTab('transactions')">Transactions</button>
        <button class="tab" onclick="showTab('new-transaction')">New Credit</button>
    </div>

    <div id="overview" class="tab-content active">
        <div class="balance-summary">
            <div class="balance-box">
                <div class="balance-label">Total Receivable</div>
                <div class="balance-amount" id="receivable-amount">$%d.%02d</div>
            </div>
            <div class="balance-box">
                <div class="balance-label">Total Payable</div>
                <div class="balance-amount" id="payable-amount">$%d.%02d</div>
            </div>
        </div>

        <h3>Pending Payments You're Owed</h3>
        <div class="transaction-list" id="receivable-list">
            %s
        </div>

        <h3>Payments You Owe</h3>
        <div class="transaction-list" id="payable-list">
            %s
        </div>
    </div>

    <div id="transactions" class="tab-content">
        <h3>Transaction History</h3>
        <div class="transaction-list" id="transaction-list">
            %s
        </div>
    </div>

    <div id="new-transaction" class="tab-content">
        <div class="form-section">
            <h3>Record New Credit Transaction</h3>
            <form id="new-transaction-form">
                <div class="form-group">
                    <label>Transaction Type:</label>
                    <select id="transaction-type" onchange="toggleTransactionFields()">
                        <option value="sell">I sold something (they owe me)</option>
                        <option value="buy">I bought something (I owe them)</option>
                    </select>
                </div>

                <div class="form-group" id="counterparty-field">
                    <label for="counterparty">Counterparty Username:</label>
                    <input type="text" id="counterparty" name="counterparty" placeholder="Enter their username" required>
                </div>

                <div class="form-group" id="amount-label-container">
                    <label for="amount">Amount Owed to You ($):</label>
                    <input type="number" id="amount" name="amount" min="0.01" step="0.01" required>
                </div>

                <div class="form-group" id="amount-label-container-buy" style="display: none;">
                    <label for="amount-buy">Amount You Owe ($):</label>
                    <input type="number" id="amount-buy" name="amount" min="0.01" step="0.01" required>
                </div>

                <div class="form-group">
                    <label>Description:</label>
                    <input type="text" id="description" name="description" placeholder="What was bought/sold?" required>
                </div>

                <button type="submit" class="btn-primary">Record Transaction</button>
            </form>
        </div>
    </div>

    <script>
        function showTab(tabName) {
            // Hide all tabs
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });

            // Remove active class from all tab buttons
            document.querySelectorAll('.tab').forEach(tab => {
                tab.classList.remove('active');
            });

            // Show selected tab
            document.getElementById(tabName).classList.add('active');

            // Add active class to clicked button
            event.target.classList.add('active');
        }

        function toggleTransactionFields() {
            const type = document.getElementById('transaction-type').value;
            const amountLabelSell = document.getElementById('amount-label-container');
            const amountLabelBuy = document.getElementById('amount-label-container-buy');

            if (type === 'sell') {
                amountLabelSell.style.display = 'block';
                amountLabelBuy.style.display = 'none';
            } else {
                amountLabelSell.style.display = 'none';
                amountLabelBuy.style.display = 'block';
            }
        }

        // Form submission handling
        document.getElementById('new-transaction-form').addEventListener('submit', function(e) {
            e.preventDefault();

            const formData = new FormData(this);
            const data = {
                type: document.getElementById('transaction-type').value,
                counterparty: formData.get('counterparty'),
                amount: formData.get('amount'),
                description: formData.get('description')
            };

            fetch('/transaction/new', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data)
            })
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    alert('Transaction recorded successfully!');
                    // Reload the page to see updated data
                    location.reload();
                } else {
                    alert('Error: ' + result.error);
                }
            })
            .catch(error => {
                alert('Error: ' + error.message);
            });
        });

        // Mark as paid functionality
        function markAsPaid(transactionId) {
            if (!confirm('Are you sure you want to mark this transaction as paid?')) {
                return;
            }

            fetch('/transaction/' + transactionId + '/pay', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            })
            .then(response => response.json())
            .then(result => {
                if (result.success) {
                    alert('Transaction marked as paid!');
                    location.reload();
                } else {
                    alert('Error: ' + result.error);
                }
            })
            .catch(error => {
                alert('Error: ' + error.message);
            });
        }
    </script>
</body>
</html>
	`, username,
		totalReceivable/100, totalReceivable%100,
		totalPayable/100, totalPayable%100,
		renderTransactionList(receivableTransactions, true),
		renderTransactionList(payableTransactions, false),
		renderTransactionList(transactions, false))

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (h *DashboardHandler) NewTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	var request struct {
		Type      string `json:"type"`
		Counterparty string `json:"counterparty"`
		Amount    string `json:"amount"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	amountFloat, err := strconv.ParseFloat(request.Amount, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	amountCents := int64(amountFloat * 100)

	var sellerUsername, buyerUsername string
	if request.Type == "sell" {
		// Current user is seller
		sellerUsername = username
		buyerUsername = request.Counterparty
	} else if request.Type == "buy" {
		// Current user is buyer
		sellerUsername = request.Counterparty
		buyerUsername = username
	} else {
		http.Error(w, "Invalid transaction type", http.StatusBadRequest)
		return
	}

	err = h.transactionService.CreateTransaction(sellerUsername, buyerUsername, amountCents, request.Description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *DashboardHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("userID").(int64)

	// Extract transaction ID from URL path
	path := r.URL.Path
	var transactionID int64
	fmt.Sscanf(path, "/transaction/%d/pay", &transactionID)

	// Verify the transaction belongs to the current user (as seller)
	transaction, err := h.transactionService.GetTransactionByID(transactionID)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}
	if transaction == nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Only the seller can mark a transaction as paid
	if transaction.SellerID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	err = h.transactionService.MarkTransactionAsPaid(transactionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to mark transaction as paid: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func renderTransactionList(transactions []repository.Transaction, isReceivable bool) string {
	if len(transactions) == 0 {
		if isReceivable {
			return "<p>No pending payments owed to you</p>"
		}
		return "<p>No pending payments you owe</p>"
	}

	var html string
	for _, t := range transactions {
		var counterparty string
		var amountStr string
		var statusClass string
		var statusText string

		if isReceivable {
			// User is seller, so counterparty is buyer
			counterparty = fmt.Sprintf("User #%d", t.BuyerID) // In a real app, we'd fetch the username
			amountStr = fmt.Sprintf("$%d.%02d", t.Amount/100, t.Amount%100)
		} else {
			// User is buyer, so counterparty is seller
			counterparty = fmt.Sprintf("User #%d", t.SellerID) // In a real app, we'd fetch the username
			amountStr = fmt.Sprintf("$%d.%02d", t.Amount/100, t.Amount%100)
		}

		if t.Status == "pending" {
			statusClass = "status-pending"
			statusText = "Pending"
		} else {
			statusClass = "status-paid"
			statusText = "Paid"
		}

		html += fmt.Sprintf(`
		<div class="transaction-item">
			<div class="transaction-info">
				<div><strong>From/To:</strong> %s</div>
				<div><strong>Amount:</strong> %s</div>
				<div><strong>Description:</strong> %s</div>
				<div><strong>Date:</strong> %s</div>
			</div>
			<div>
				<span class="transaction-status %s">%s</span>
				%s
			</div>
		</div>
		`, counterparty, amountStr, t.Description, t.CreatedAt, statusClass, statusText,
			func() string {
				if t.Status == "pending" && isReceivable {
					return fmt.Sprintf(`<button class="btn btn-success mark-paid-btn" onclick="markAsPaid(%d)">Mark as Paid</button>`, t.ID)
				}
				return ""
			}())
	}

	return html
}