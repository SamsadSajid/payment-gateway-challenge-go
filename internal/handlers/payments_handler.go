package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/restclient"
)

type PaymentsHandler struct {
	storage    *repository.PaymentsRepository
	bankClient *restclient.Bank
}

func NewPaymentsHandler(storage *repository.PaymentsRepository, bankClient *restclient.Bank) *PaymentsHandler {
	return &PaymentsHandler{
		storage:    storage,
		bankClient: bankClient,
	}
}

const (
	idUrlParamName = "id"
)

// GetHandler returns an http.HandlerFunc that handles HTTP GET requests.
// It retrieves a payment record by its ID from the storage.
// The ID is expected to be part of the URL.
func (ph *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, idUrlParamName)
		paymentRecord := ph.storage.GetPayment(id)

		render.Render(w, r, &models.GetPaymentResponse{
			PaymentRecord: paymentRecord,
		})
	}
}

// PostHandler returns an http.HandlerFunc that handles HTTP POST requests.
// It sends the payment request to the bank and returns the reponse to the
// merchant.
func (ph *PaymentsHandler) PostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &models.PostPaymentRequest{}
		if err := render.Bind(r, req); err != nil {
			render.Render(w, r, &models.ErrResponse{
				HTTPStatusCode: http.StatusBadRequest,
				StatusText:     fmt.Sprintf("Payment request rejected! Error: %v", err.Error()),
				AppCode:        models.ErrRequestRejected,
			})
			return
		}

		// TODO: If bank returns
		// - 4XX: payment request is not valid
		// - 5XX: we should retry with exponential backoff;
		//		  if retry has exhausted return a retriable
		//		  error code to the merchant
		bankResp, bankErr := requestBank(req, ph.bankClient)

		paymentRecord := repository.NewPayment(req, bankResp)
		if err := ph.storage.AddPayment(paymentRecord); err != nil {
			render.Render(w, r, &models.ErrResponse{
				HTTPStatusCode: http.StatusInternalServerError,
				StatusText:     err.Error(),
				AppCode:        models.ErrDatastorePaymentCreation,
			})
			return
		}

		if bankErr != nil {
			render.Render(w, r, &models.ErrResponse{
				HTTPStatusCode: bankErr.StatusCode,
				StatusText:     "An error occurred. Please contact customer support and provide the app_code",
				AppCode:        bankErr.Appcode,
			})
			return
		}

		render.Render(w, r, &models.PostPaymentResponse{
			PaymentRecord: &paymentRecord,
		})
	}
}

func requestBank(req *models.PostPaymentRequest, bankClient *restclient.Bank) (*models.BankResponse, *models.BankErrorResponse) {
	bankReq := models.BankRequest{
		CardNumber: req.CardNumber,
		ExpiryDate: formatExpiryDateForBankRequest(req),
		Currency:   req.Currency,
		Amount:     req.Amount,
		Cvv:        req.Cvv,
	}

	return bankClient.RequestPaymentCapture(bankReq)
}

func formatExpiryDateForBankRequest(req *models.PostPaymentRequest) string {
	expiryMonth := fmt.Sprint(req.ExpiryMonth)

	// For single digit month i.e., <January-September>, we append an extra "0" before the month's index
	// as this is the format the bank accepts
	if len(expiryMonth) == 1 {
		expiryMonth = "0" + expiryMonth
	}

	expiryYear := fmt.Sprint(req.ExpiryYear)

	return fmt.Sprintf("%s/%s", expiryMonth, expiryYear)
}
