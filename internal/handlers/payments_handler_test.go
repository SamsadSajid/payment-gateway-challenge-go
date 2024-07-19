//go:build integration
// +build integration

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/restclient"
)

func TestGetPaymentHandler(t *testing.T) {
	ps := repository.NewPaymentsRepository()
	c := restclient.NewBankClient(http.DefaultClient)
	addSeedData(ps, 1)

	payments := NewPaymentsHandler(ps, c)

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", payments.GetHandler())

	httpServer := &http.Server{
		Addr:    ":8091",
		Handler: r,
	}

	go func() error {
		return httpServer.ListenAndServe()
	}()

	t.Run("PaymentFound", func(t *testing.T) {
		// Create a new HTTP request for testing
		req, _ := http.NewRequest("GET", "/api/payments/test-id", nil)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the body is not nil
		assert.NotNil(t, w.Body)

		// Check the HTTP status code in the response
		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
	t.Run("PaymentNotFound", func(t *testing.T) {
		// Create a new HTTP request for testing with a non-existing payment ID
		req, _ := http.NewRequest("GET", "/api/payments/NonExistingID", nil)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the HTTP status code in the response

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func addSeedData(ps *repository.PaymentsRepository, seed int) {
	for i := 0; i < seed; i++ {
		payment := models.PaymentRecord{
			Id:                 "test-id",
			PaymentStatus:      models.PaymentStatus(models.Authorized),
			CardNumberLastFour: "1234",
			ExpiryMonth:        10,
			ExpiryYear:         2035,
			Currency:           "GBP",
			Amount:             100,
		}
		ps.AddPayment(payment)
	}
}
