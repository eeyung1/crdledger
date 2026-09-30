package handler

import (
	"fmt"
	"html/template"
	"net/http"
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
	u, err := h.users.GetByID(id); if err != nil || u.AccountType != "seller" { http.Error(w,"seller account required",http.StatusForbidden); return }
	if u.Email == "" { http.Error(w,"Add an email address before paying for a seller subscription.",http.StatusBadRequest); return }
	ref := fmt.Sprintf("crdledger-%d-%d", id, time.Now().UnixNano())
	callback := "https://" + r.Host + "/subscription/callback"
	checkout, err := h.paystack.Initialize(u.Email,u.SubscriptionPlan,callback,ref); if err != nil { http.Error(w,"Could not start payment. Please try again.",http.StatusBadGateway); return }
	if err := h.users.CreateSubscriptionPayment(id,ref,u.SubscriptionPlan); err != nil { http.Error(w,"Could not save payment.",http.StatusInternalServerError); return }
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
