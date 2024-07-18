package models

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/render"
)

// PostPaymentRequest type is what merchant sends for a payment request
type PostPaymentRequest struct {
	CardNumber  string `json:"card_number"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	Currency    string `json:"currency"`
	Amount      int    `json:"amount"`
	Cvv         string `json:"cvv"`
}

// Bind validates the request. It returns an error if validation fails.
func (req *PostPaymentRequest) Bind(r *http.Request) error {
	// In the call chain, r will be decoded to PostPaymentRequest and then
	// Bind will be called - this is handled via go-chi/render.

	if len(req.CardNumber) < 14 || len(req.CardNumber) > 19 {
		return fmt.Errorf("card number should be between 14-19 characters long")
	}

	re := regexp.MustCompile(`^[0-9]+$`)
	if !re.MatchString(req.CardNumber) {
		return fmt.Errorf("card number must contain only numberic characters")
	}

	if req.isInvalidExpirationTime() {
		return fmt.Errorf("please provide a valid expiration date")
	}

	// TODO also validates on currency code - waiting on recruiter
	if len(req.Currency) < 3 || len(req.Currency) > 3 {
		return fmt.Errorf("currency code must be 3 characters long")
	}

	if len(req.Cvv) < 3 || len(req.Cvv) > 4 {
		return fmt.Errorf("CVV should be between 3-4 characters long")
	}
	if !re.MatchString(req.Cvv) {
		return fmt.Errorf("CVV must contain only numberic characters")
	}

	return nil
}

func (req *PostPaymentRequest) isInvalidExpirationTime() bool {
	currYear, currMonth, _ := time.Now().Date()

	return (req.ExpiryYear < currYear) || (currYear == req.ExpiryYear && time.Month(req.ExpiryMonth) < currMonth)
}

type omit *struct{}

// PostPaymentResponse type is what the payment gw sends back to the merchant
type PostPaymentResponse struct {
	*PaymentRecord
	ErrorMsg string `json:"error_message"`

	// we don't want the following two fields to send to the merchant
	Cvv                   omit `json:"cvv,omitempty"`
	BankAuthorizationCode omit `json:"bank_authorization_code,omitempty"`
}

// Render returns the PostPaymentResponse to the client
func (resp *PostPaymentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// If error message is not empty, for all cases it means PostPaymentRequest was not valid
	if resp.ErrorMsg != "" {
		render.Status(r, http.StatusBadRequest)
	} else {
		render.Status(r, http.StatusOK)
	}
	return nil
}

type GetPaymentResponse struct {
	*PaymentRecord

	// we don't want the following two fields to send to the merchant
	Cvv                   omit `json:"cvv,omitempty"`
	BankAuthorizationCode omit `json:"bank_authorization_code,omitempty"`
}

func (resp *GetPaymentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// If struct is nil, the payment record does not exist in the datastore
	if resp.PaymentRecord == nil {
		render.Status(r, http.StatusNoContent)
	} else {
		render.Status(r, http.StatusOK)
	}
	return nil
}

type PaymentStatus string

const (
	Authorized PaymentStatus = "Authorized"
	Declined   PaymentStatus = "Declined"
	Rejected   PaymentStatus = "Rejected"
)

// PaymentRecord is what we save in the datastore and use throughout
// the app's lifecycle
type PaymentRecord struct {
	Id                    string        `json:"id"`
	CardNumberLastFour    string        `json:"card_number_last_four"`
	ExpiryMonth           int           `json:"expiry_month"`
	ExpiryYear            int           `json:"expiry_year"`
	Currency              string        `json:"currency"`
	Amount                int           `json:"amount"`
	Cvv                   string        `json:"cvv"`
	PaymentStatus         PaymentStatus `json:"payment_status"`
	BankAuthorizationCode string        `json:"bank_authorization_code"`
}
