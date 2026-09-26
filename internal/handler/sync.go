package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

// SyncHandler accepts durable client mutations after reconnect. Requests use
// ordinary form fields so the existing CSRF middleware protects this endpoint
// exactly like the server-rendered transaction forms.
type SyncHandler struct {
	transactions *service.TransactionService
}

func NewSyncHandler(transactions *service.TransactionService) *SyncHandler {
	return &SyncHandler{transactions: transactions}
}

type syncResponse struct {
	OperationID   string `json:"operation_id"`
	TransactionID int64  `json:"transaction_id,omitempty"`
	Status        string `json:"status"`
	Error         string `json:"error,omitempty"`
}

func (h *SyncHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(syncResponse{Status: "failed", Error: "method not allowed"})
		return
	}

	sellerID, ok := middleware.UserIDFromContext(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(syncResponse{Status: "failed", Error: "authentication required"})
		return
	}

	operationID := r.FormValue("operation_id")
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(syncResponse{OperationID: operationID, Status: "failed", Error: "invalid amount"})
		return
	}

	t, created, err := h.transactions.RecordIdempotent(
		sellerID,
		r.FormValue("buyer_username"),
		amount,
		r.FormValue("description"),
		operationID,
	)
	if err != nil {
		status := http.StatusInternalServerError
		message := "sync failed"
		switch {
		case errors.Is(err, service.ErrOperationIDRequired):
			status, message = http.StatusBadRequest, "operation id is required"
		case errors.Is(err, service.ErrInvalidAmount):
			status, message = http.StatusBadRequest, "amount must be positive"
		case errors.Is(err, service.ErrInvalidDescription):
			status, message = http.StatusBadRequest, "description is required"
		case errors.Is(err, service.ErrBuyerNotFound):
			status, message = http.StatusNotFound, "buyer not found"
		case errors.Is(err, service.ErrCannotRecordSelf):
			status, message = http.StatusBadRequest, "cannot record a transaction with yourself"
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(syncResponse{OperationID: operationID, Status: "failed", Error: message})
		return
	}

	state := "synced"
	if !created {
		state = "already_synced"
	}
	json.NewEncoder(w).Encode(syncResponse{OperationID: operationID, TransactionID: t.ID, Status: state})
}
