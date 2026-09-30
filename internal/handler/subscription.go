package handler

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"errors"
	"html/template"
	"net/http"
	"net/mail"
	"strings"
	"log/slog"
	"time"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"
)

type SubscriptionHandler struct { users *repository.UserRepository; paystack *service.PaystackService; templates *template.Template; secretKey string; appBaseURL string }
func NewSubscriptionHandler(users *repository.UserRepository, paystack *service.PaystackService, templates *template.Template, secretKey, appBaseURL string) *SubscriptionHandler { return &SubscriptionHandler{users:users,paystack:paystack,templates:templates,secretKey:secretKey,appBaseURL:strings.TrimRight(appBaseURL, "/")} }

func (h *SubscriptionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w,"method not allowed",http.StatusMethodNotAllowed); return }
	id, ok := middleware.UserIDFromContext(r); if !ok { http.Redirect(w,r,"/login",http.StatusSeeOther); return }
	u, err := h.users.GetByID(id); if err != nil { http.Error(w,"failed to load account",http.StatusInternalServerError); return }
	plan := r.FormValue("subscription_plan")
	if plan != "monthly" && plan != "yearly" {
		plan = u.SubscriptionPlan
	}
	if plan != "monthly" && plan != "yearly" { http.Error(w,"Choose a monthly or yearly seller plan.",http.StatusBadRequest); return }
	email := strings.TrimSpace(u.Email)
	if u.AccountType == "buyer" {
		email = strings.TrimSpace(r.FormValue("email"))
		if email == "" { http.Redirect(w,r,"/profile/edit?upgrade_error=Enter+your+email+before+continuing+to+payment.",http.StatusSeeOther); return }
		if _, err := mail.ParseAddress(email); err != nil { http.Redirect(w,r,"/profile/edit?upgrade_error=Enter+a+valid+email+address.",http.StatusSeeOther); return }
		if email != u.Email {
			if err := h.users.UpdateEmail(id,email); err != nil { http.Error(w,"Could not save email.",http.StatusInternalServerError); return }
		}
	}
	if email == "" { http.Error(w,"Add an email address before paying for a seller subscription.",http.StatusBadRequest); return }
	ref := fmt.Sprintf("crdledger-%d-%d", id, time.Now().UnixNano())
	callback := h.appBaseURL + "/subscription/callback"
	checkout, err := h.paystack.Initialize(email,plan,callback,ref); if err != nil { slog.Error("subscription checkout initialization failed","user_id",id,"plan",plan,"error",err); http.Error(w,"Could not start payment. Please try again.",http.StatusBadGateway); return }
	if err := h.users.CreateSubscriptionPayment(id,ref,plan); err != nil { http.Error(w,"Could not save payment.",http.StatusInternalServerError); return }
	http.Redirect(w,r,checkout,http.StatusSeeOther)
}

func (h *SubscriptionHandler) Callback(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r); if !ok { http.Redirect(w,r,"/login",http.StatusSeeOther); return }
	ref := r.URL.Query().Get("reference"); if ref == "" { http.Error(w,"missing payment reference",http.StatusBadRequest); return }
	plan, status, err := h.users.SubscriptionPaymentForUser(id,ref); if err != nil { http.Error(w,"Unknown payment reference.",http.StatusBadRequest); return }
	// The webhook can arrive before the customer's browser returns from Paystack.
	// If it already completed this payment, activation has succeeded and the
	// callback should simply finish the user journey instead of reporting an error.
	if status != "completed" {
		if err := h.paystack.Verify(ref,plan); err != nil {
			if errors.Is(err, service.ErrPaymentNotSuccessful) || errors.Is(err, service.ErrPaymentAmountMismatch) {
				_ = h.users.MarkSubscriptionPaymentFailed(id, ref)
				slog.Warn("subscription payment verification failed","user_id",id,"reference",ref,"plan",plan,"error",err)
				http.Redirect(w,r,"/profile/edit?payment_error=Payment+was+not+completed.+No+subscription+was+activated.",http.StatusSeeOther)
				return
			}
			slog.Error("subscription payment verification unavailable","user_id",id,"reference",ref,"error",err)
			http.Error(w,"Payment verification is temporarily unavailable. Please retry from your payment return link.",http.StatusBadGateway); return
		}
		if err := h.users.ActivateSellerSubscription(id,ref,plan); err != nil { slog.Error("subscription activation failed","user_id",id,"reference",ref,"error",err); http.Error(w,"Could not activate subscription.",http.StatusInternalServerError); return }
		slog.Info("subscription activated","user_id",id,"reference",ref,"plan",plan)
	}
	http.Redirect(w,r,"/dashboard?subscription=active",http.StatusSeeOther)
}


type paystackWebhookEvent struct {
	Event string `json:"event"`
	Data struct {
		Reference string `json:"reference"`
	} `json:"data"`
}

func (h *SubscriptionHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil { http.Error(w, "invalid body", http.StatusBadRequest); return }

	mac := hmac.New(sha512.New, []byte(h.secretKey))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	signature := r.Header.Get("x-paystack-signature")
	if signature == "" || !hmac.Equal([]byte(expected), []byte(signature)) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var event paystackWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil { http.Error(w, "invalid event", http.StatusBadRequest); return }
	if event.Event != "charge.success" || event.Data.Reference == "" { w.WriteHeader(http.StatusOK); return }

	userID, plan, err := h.users.SubscriptionPaymentByReference(event.Data.Reference)
	if err != nil { w.WriteHeader(http.StatusOK); return }

	// Do not trust webhook fields alone. Verify the transaction directly
	// with Paystack, including successful status, NGN currency and exact plan amount.
	if err := h.paystack.Verify(event.Data.Reference, plan); err != nil {
		// A signed success webhook is still verified directly with Paystack.
		// Permanent verification mismatches are acknowledged so Paystack does
		// not retry forever; transient provider/network failures return 502.
		if errors.Is(err, service.ErrPaymentNotSuccessful) || errors.Is(err, service.ErrPaymentAmountMismatch) {
			_ = h.users.MarkSubscriptionPaymentFailed(userID, event.Data.Reference)
			slog.Warn("webhook payment verification rejected","user_id",userID,"reference",event.Data.Reference,"plan",plan,"error",err)
			w.WriteHeader(http.StatusOK); return
		}
		slog.Error("webhook payment verification unavailable","user_id",userID,"reference",event.Data.Reference,"error",err)
		http.Error(w, "verification temporarily unavailable", http.StatusBadGateway); return
	}
	if err := h.users.ActivateSellerSubscription(userID, event.Data.Reference, plan); err != nil { slog.Error("webhook subscription activation failed","user_id",userID,"reference",event.Data.Reference,"error",err); http.Error(w, "activation failed", http.StatusInternalServerError); return }
	slog.Info("webhook subscription activated","user_id",userID,"reference",event.Data.Reference,"plan",plan)
	w.WriteHeader(http.StatusOK)
}
