//go:build !integration
// +build !integration

package repository

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

func TestPaymentsRepository_AddPayment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		paymentsRepo func() map[string]models.PaymentRecord
		payment      models.PaymentRecord
		wantErr      bool
	}{
		{
			name: "returns nil error when adding payment is successful",
			paymentsRepo: func() map[string]models.PaymentRecord {
				return make(map[string]models.PaymentRecord)
			},
			payment: models.PaymentRecord{
				Id:                 "a",
				CardNumberLastFour: "7789",
			},
			wantErr: false,
		},
		{
			name: "returns an error when payment record already exists",
			paymentsRepo: func() map[string]models.PaymentRecord {
				p := make(map[string]models.PaymentRecord)
				p["a"] = models.PaymentRecord{
					Id:                 "a",
					CardNumberLastFour: "7789",
				}
				return p
			},
			payment: models.PaymentRecord{
				Id:                 "a",
				CardNumberLastFour: "7789",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps := &PaymentsRepository{
				payments: tt.paymentsRepo(),
			}
			if err := ps.AddPayment(tt.payment); (err != nil) != tt.wantErr {
				t.Errorf("PaymentsRepository.AddPayment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentsRepository_GetPayment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		paymentsRepo func() map[string]models.PaymentRecord
		id           string
		want         *models.PaymentRecord
	}{
		{
			name: "returns a payment record if it exists in the datastore",
			paymentsRepo: func() map[string]models.PaymentRecord {
				p := make(map[string]models.PaymentRecord)
				p["a"] = models.PaymentRecord{
					Id:                 "a",
					CardNumberLastFour: "7789",
				}
				return p
			},
			id: "a",
			want: &models.PaymentRecord{
				Id:                 "a",
				CardNumberLastFour: "7789",
			},
		},
		{
			name: "returns nil when a payment record does not exist in the datastore",
			paymentsRepo: func() map[string]models.PaymentRecord {
				p := make(map[string]models.PaymentRecord)
				p["a"] = models.PaymentRecord{
					Id:                 "a",
					CardNumberLastFour: "7789",
				}
				return p
			},
			id:   "b",
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ps := &PaymentsRepository{
				payments: tt.paymentsRepo(),
			}
			if got := ps.GetPayment(tt.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PaymentsRepository.GetPayment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPayment(t *testing.T) {
	t.Parallel()

	type args struct {
		req      *models.PostPaymentRequest
		bankResp *models.BankResponse
	}
	tests := []struct {
		name string
		args args
		want models.PaymentRecord
	}{
		{
			name: "returns a payment record with Authorized payment status if bank authorization is true",
			args: args{
				req: &models.PostPaymentRequest{
					CardNumber:  "2222405343248877",
					ExpiryMonth: 4,
					ExpiryYear:  2025,
					Currency:    "GBP",
					Amount:      100,
					Cvv:         "123",
				},
				bankResp: &models.BankResponse{
					Authorized:        true,
					AuthorizationCode: "ab-cd",
				},
			},
			want: models.PaymentRecord{
				CardNumberLastFour: "8877",
				ExpiryMonth:        4,
				ExpiryYear:         2025,
				Currency:           "GBP",
				Amount:             100,
				Cvv:                "123",
				PaymentStatus:      models.PaymentStatus(models.Authorized),
			},
		},
		{
			name: "returns a payment record with Declined payment status if bank authorization is false",
			args: args{
				req: &models.PostPaymentRequest{
					CardNumber:  "2222405343248877",
					ExpiryMonth: 4,
					ExpiryYear:  2025,
					Currency:    "GBP",
					Amount:      100,
					Cvv:         "123",
				},
				bankResp: &models.BankResponse{
					Authorized:        false,
					AuthorizationCode: "",
				},
			},
			want: models.PaymentRecord{
				CardNumberLastFour: "8877",
				ExpiryMonth:        4,
				ExpiryYear:         2025,
				Currency:           "GBP",
				Amount:             100,
				Cvv:                "123",
				PaymentStatus:      models.PaymentStatus(models.Declined),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewPayment(tt.args.req, tt.args.bankResp)
			if diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(models.PaymentRecord{}, "Id")); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
