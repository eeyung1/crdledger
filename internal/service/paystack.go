package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var ErrPaymentNotSuccessful = errors.New("payment not successful")
var ErrPaymentAmountMismatch = errors.New("payment amount mismatch")

type PaystackService struct { secretKey string; client *http.Client }
type paystackInitResponse struct { Status bool `json:"status"`; Data struct { AuthorizationURL string `json:"authorization_url"`; Reference string `json:"reference"` } `json:"data"` }
type paystackVerifyResponse struct { Status bool `json:"status"`; Data struct { Status string `json:"status"`; Amount int64 `json:"amount"`; Reference string `json:"reference"`; Currency string `json:"currency"` } `json:"data"` }

func NewPaystackService(secretKey string) *PaystackService { return &PaystackService{secretKey: secretKey, client: &http.Client{Timeout: 15*time.Second}} }

func SellerPlanAmount(plan string) (int64, error) {
	switch plan { case "monthly": return 200000, nil; case "yearly": return 1500000, nil; default: return 0, errors.New("invalid seller plan") }
}

func (s *PaystackService) Initialize(email, plan, callbackURL, reference string) (string, error) {
	amount, err := SellerPlanAmount(plan); if err != nil { return "", err }
	body, _ := json.Marshal(map[string]any{"email":email,"amount":amount,"currency":"NGN","reference":reference,"callback_url":callbackURL})
	req, err := http.NewRequest(http.MethodPost, "https://api.paystack.co/transaction/initialize", bytes.NewReader(body)); if err != nil { return "", err }
	req.Header.Set("Authorization", "Bearer "+s.secretKey); req.Header.Set("Content-Type","application/json")
	resp, err := s.client.Do(req); if err != nil { return "", err }; defer resp.Body.Close()
	var out paystackInitResponse; if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return "", err }
	if resp.StatusCode != http.StatusOK || !out.Status || out.Data.AuthorizationURL == "" { return "", fmt.Errorf("paystack initialization failed") }
	return out.Data.AuthorizationURL, nil
}

func (s *PaystackService) Verify(reference, plan string) error {
	amount, err := SellerPlanAmount(plan); if err != nil { return err }
	req, _ := http.NewRequest(http.MethodGet, "https://api.paystack.co/transaction/verify/"+url.PathEscape(reference), nil)
	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	resp, err := s.client.Do(req); if err != nil { return err }; defer resp.Body.Close()
	var out paystackVerifyResponse; if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return err }
	if resp.StatusCode >= 500 { return fmt.Errorf("paystack verify unavailable: status %d", resp.StatusCode) }
	if resp.StatusCode != http.StatusOK || !out.Status || out.Data.Status != "success" { return ErrPaymentNotSuccessful }
	if out.Data.Reference != reference || out.Data.Amount != amount || out.Data.Currency != "NGN" { return ErrPaymentAmountMismatch }
	return nil
}
