//go:build integration
// +build integration

package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/restclient"
)

func TestPostPaymentHandler(t *testing.T) {
	t.Parallel()

	ps := repository.NewPaymentsRepository()
	bc := http.DefaultClient
	c := restclient.NewBankClient(bc)

	payments := NewPaymentsHandler(ps, c)

	router := chi.NewRouter()
	router.Post("/api/payments", payments.PostHandler())

	httpServer := &http.Server{
		Addr:    ":8091",
		Handler: router,
	}

	go func() error {
		return httpServer.ListenAndServe()
	}()

	tests := []struct {
		name           string
		payload        *strings.Reader
		wantStatusCode int
		want           string
	}{
		{
			name: "payment authorized",
			payload: strings.NewReader(`{
				"card_number": "2222405343248877",
				"expiry_month": 4, 
				"expiry_year": 2025, 
				"currency": "GBP", 
				"amount": 100, 
				"cvv": "123"
			}`),
			wantStatusCode: http.StatusOK,
			want:           "\"card_number_last_four\":\"8877\",\"expiry_month\":4,\"expiry_year\":2025,\"currency\":\"GBP\",\"amount\":100,\"payment_status\":\"Authorized\",\"error_message\":\"\"",
		},
		{
			name: "payment declined",
			payload: strings.NewReader(`{
				"card_number": "2222405343248112",
				"expiry_month": 1, 
				"expiry_year": 2026, 
				"currency": "USD", 
				"amount": 60000, 
				"cvv": "456"
			}`),
			wantStatusCode: http.StatusOK,
			want:           "\"card_number_last_four\":\"8112\",\"expiry_month\":1,\"expiry_year\":2026,\"currency\":\"USD\",\"amount\":60000,\"payment_status\":\"Declined\",\"error_message\":\"\"",
		},
		{
			name: "payment rejected from bank",
			payload: strings.NewReader(`{
				"card_number": "2222405343242069",
				"expiry_month": 1, 
				"expiry_year": 2026, 
				"currency": "USD", 
				"amount": 60000, 
				"cvv": "456"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"An error occurred. Please contact customer support and provide the app_code\",\"app_code\":8",
		},
		{
			name: "payment rejected bad card number",
			payload: strings.NewReader(`{
				"card_number": "22224053432",
				"expiry_month": 1, 
				"expiry_year": 2026, 
				"currency": "USD", 
				"amount": 60000, 
				"cvv": "456"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"Payment request rejected! Error: card number should be between 14-19 characters long\",\"app_code\":9",
		},
		{
			name: "payment rejected card number is not numeric",
			payload: strings.NewReader(`{
				"card_number": "222240534324206a",
				"expiry_month": 1, 
				"expiry_year": 2026, 
				"currency": "USD", 
				"amount": 60000, 
				"cvv": "456"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"Payment request rejected! Error: card number must contain only numberic characters\",\"app_code\":9",
		},
		{
			name: "payment rejected expiration time is not in future",
			payload: strings.NewReader(`{
				"card_number": "2222405343248877",
				"expiry_month": 1, 
				"expiry_year": 2024, 
				"currency": "USD", 
				"amount": 60000, 
				"cvv": "456"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"Payment request rejected! Error: please provide a valid expiration date\",\"app_code\":9",
		},
		{
			name: "payment rejected not ISO currency code",
			payload: strings.NewReader(`{
				"card_number": "2222405343248877",
				"expiry_month": 8, 
				"expiry_year": 2024, 
				"currency": "British pound and sterling", 
				"amount": 60000, 
				"cvv": "456"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"Payment request rejected! Error: currency code must be ISO formatted\",\"app_code\":9",
		},
		{
			name: "payment rejected invalid character length in CVV",
			payload: strings.NewReader(`{
				"card_number": "2222405343248877",
				"expiry_month": 8, 
				"expiry_year": 2024, 
				"currency": "GBP", 
				"amount": 60000, 
				"cvv": "45"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"Payment request rejected! Error: CVV should be between 3-4 characters long\",\"app_code\":9",
		},
		{
			name: "payment rejected CVV is not numeric",
			payload: strings.NewReader(`{
				"card_number": "2222405343248877",
				"expiry_month": 8, 
				"expiry_year": 2024, 
				"currency": "GBP", 
				"amount": 60000, 
				"cvv": "45a"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"Payment request rejected! Error: CVV must contain only numberic characters\",\"app_code\":9",
		},
		{
			name: "payment rejected invalid CVV for amex",
			payload: strings.NewReader(`{
				"card_number": "3222405343248877",
				"expiry_month": 8, 
				"expiry_year": 2024, 
				"currency": "GBP", 
				"amount": 60000, 
				"cvv": "456"
			}`),
			wantStatusCode: http.StatusBadRequest,
			want:           "\"status_code\":400,\"status_text\":\"Payment request rejected! Error: invalid CVV\",\"app_code\":9",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest("POST", "/api/payments", tt.payload)
			r.Header.Add("Content-Type", "application/json")
			assert.Nil(t, err)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			assert.NotNil(t, w.Body)

			if tt.wantStatusCode != w.Code {
				t.Errorf("handler returned wrong status code: got %d want %d",
					w.Code, tt.wantStatusCode)
			}
			assert.Contains(t, w.Body.String(), tt.want)
		})
	}
}

func TestPostPaymentHandler_RetryBankExponentially(t *testing.T) {
	t.Parallel()

	ps := repository.NewPaymentsRepository()
	bc := mockBankClient{}
	c := restclient.NewBankClient(&bc)

	payments := NewPaymentsHandler(ps, c)

	router := chi.NewRouter()
	router.Post("/api/payments", payments.PostHandler())

	httpServer := &http.Server{
		Addr:    ":8091",
		Handler: router,
	}

	go func() error {
		return httpServer.ListenAndServe()
	}()

	tests := []struct {
		name           string
		payload        *strings.Reader
		wantStatusCode int
		want           string
	}{
		{
			name: "return HTTP 503 because bank server down after retry is exhausted",
			payload: strings.NewReader(`{
				"card_number": "2222405343248877",
				"expiry_month": 4, 
				"expiry_year": 2025, 
				"currency": "GBP", 
				"amount": 100, 
				"cvv": "123"
			}`),
			wantStatusCode: http.StatusServiceUnavailable,
			want:           "\"status_code\":503,\"status_text\":\"An error occurred. Please contact customer support and provide the app_code\",\"app_code\":8",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest("POST", "/api/payments", tt.payload)
			r.Header.Add("Content-Type", "application/json")
			assert.Nil(t, err)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			assert.NotNil(t, w.Body)

			if tt.wantStatusCode != w.Code {
				t.Errorf("handler returned wrong status code: got %d want %d",
					w.Code, tt.wantStatusCode)
			}
			assert.Contains(t, w.Body.String(), tt.want)

			// retry will happen at least 3 times
			assert.GreaterOrEqual(t, bc.cnt, 3)
		})
	}
}

type mockBankClient struct {
	cnt int
}

func (m *mockBankClient) Do(r *http.Request) (*http.Response, error) {
	m.cnt++
	return &http.Response{
		StatusCode: http.StatusServiceUnavailable,
	}, nil
}
