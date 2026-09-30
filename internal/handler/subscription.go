package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"
)

type SubscriptionHandler struct { users *repository.UserRepository; paystack *service.PaystackService; templates *template.Template }
func NewSubscriptionHandler(users *repository.UserRepository, paystack *service.PaystackService, templates *template.Template) *SubscriptionHandler { return &SubscriptionHandler{users:users,paystack:paystack,templates:templates} }

func (h *SubscriptionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w,"method not allowed",http.StatusMethodNotAllowed); return }
	id, ok := middleware.UserIDFromContext(r); if !ok { http.Redirect(w,r,"/login",http.StatusSeeOther); return }
	u, err := h.users.GetByID(id); if err != nil { http.Error(w,"failed to load account",http.StatusInternalServerError); return }
	plan := u.SubscriptionPlan
	if u.AccountType == "buyer" {
		plan = r.FormValue("subscription_plan")
		if plan != "monthly" && plan != "yearly" { http.Error(w,"Choose a monthly or yearly seller plan.",http.StatusBadRequest); return }
	} else if plan != "monthly" && plan != "yearly" {
		plan = r.FormValue("subscription_plan")
		if plan != "monthly" && plan != "yearly" { http.Error(w,"Choose a monthly or yearly seller plan.",http.StatusBadRequest); return }
	}
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
	callback := "https://" + r.Host + "/subscription/callback"
	checkout, err := h.paystack.Initialize(email,plan,callback,ref); if err != nil { http.Error(w,"Could not start payment. Please try again.",http.StatusBadGateway); return }
	if err := h.users.CreateSubscriptionPayment(id,ref,plan); err != nil { http.Error(w,"Could not save payment.",http.StatusInternalServerError); return }
	http.Redirect(w,r,checkout,http.StatusSeeOther)
}

func (h *SubscriptionHandler) Callback(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r); if !ok { http.Redirect(w,r,"/login",http.StatusSeeOther); return }
	ref := r.URL.Query().Get("reference"); if ref == "" { http.Error(w,"missing payment reference",http.StatusBadRequest); return }
	plan, err := h.users.PendingSubscriptionPayment(id,ref); if err != nil { http.Error(w,"Unknown payment reference.",http.StatusBadRequest); return }
	if err := h.paystack.Verify(ref,plan); err != nil { http.Error(w,"Payment could not be verified.",http.StatusBadRequest); return }
	if err := h.users.ActivateSellerSubscription(id,ref,plan); err != nil { http.Error(w,"Could not activate subscription.",http.StatusInternalServerError); return }
	http.Redirect(w,r,"/dashboard?subscription=active",http.StatusSeeOther)
}
