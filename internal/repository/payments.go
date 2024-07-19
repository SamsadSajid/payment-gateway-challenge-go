package repository

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

type PaymentsRepository struct {
	payments map[string]models.PaymentRecord
	l        *sync.RWMutex
}

func NewPaymentsRepository() *PaymentsRepository {
	return &PaymentsRepository{
		payments: make(map[string]models.PaymentRecord),
		l:        &sync.RWMutex{},
	}
}

func (ps *PaymentsRepository) GetPayment(id string) *models.PaymentRecord {
	ps.l.RLock()
	defer ps.l.RUnlock()
	payment, ok := ps.payments[id]
	if !ok {
		return nil
	}
	return &payment
}

func (ps *PaymentsRepository) AddPayment(payment models.PaymentRecord) error {
	ps.l.Lock()
	defer ps.l.Unlock()
	if _, ok := ps.payments[payment.Id]; ok {
		return fmt.Errorf("payment record already exists for payment id %s", payment.Id)
	}

	ps.payments[payment.Id] = payment
	return nil
}

// NewPayment creates a datastore struct
func NewPayment(req *models.PostPaymentRequest, bankResp *models.BankResponse) models.PaymentRecord {
	paymentRecord := models.PaymentRecord{
		Id:                 uuid.New().String(),
		CardNumberLastFour: req.CardNumber[len(req.CardNumber)-4:],
		ExpiryMonth:        req.ExpiryMonth,
		ExpiryYear:         req.ExpiryYear,
		Currency:           req.Currency,
		Amount:             req.Amount,
		Cvv:                req.Cvv,
	}

	if bankResp != nil {
		paymentRecord.BankAuthorizationCode = bankResp.AuthorizationCode

		if bankResp.Authorized {
			paymentRecord.PaymentStatus = models.PaymentStatus(models.Authorized)
		} else {
			paymentRecord.PaymentStatus = models.PaymentStatus(models.Declined)
		}
	}

	return paymentRecord
}
