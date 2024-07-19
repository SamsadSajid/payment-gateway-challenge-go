//go:build integration
// +build integration

package restclient

import (
	"net/http"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestBank_RequestPaymentCapture(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		bankReq            models.BankRequest
		want               *models.BankResponse
		wantHttpStatusCode int
		wantErr            *models.BankErrorResponse
	}{
		{
			name: "returns an authorized response for a valid card request",
			bankReq: models.BankRequest{
				CardNumber: "2222405343248877",
				ExpiryDate: "04/2025",
				Currency:   "GBP",
				Amount:     100,
				Cvv:        "123",
			},
			wantErr: nil,
			want: &models.BankResponse{
				Authorized:        true,
				AuthorizationCode: "0bb07405-6d44-4b50-a14f-7ae0beff13ad",
			},
			wantHttpStatusCode: http.StatusOK,
		},
		{
			name: "returns an declined response for an invalid card request",
			bankReq: models.BankRequest{
				CardNumber: "2222405343248112",
				ExpiryDate: "01/2026",
				Currency:   "USD",
				Amount:     60000,
				Cvv:        "456",
			},
			wantErr: nil,
			want: &models.BankResponse{
				Authorized:        false,
				AuthorizationCode: "",
			},
			wantHttpStatusCode: http.StatusOK,
		},
		{
			name: "returns a HTTP 400 response for an invalid card request",
			bankReq: models.BankRequest{
				CardNumber: "420695343248112", // this card is not supported by bank simulator
				ExpiryDate: "01/2026",
				Currency:   "USD",
				Amount:     60000,
				Cvv:        "456",
			},
			wantErr: &models.BankErrorResponse{
				Appcode:      models.ErrBankResponseStatusCodeNon200,
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: "Bank returned HTTP 400 code",
			},
			want:               nil,
			wantHttpStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := &Bank{
				client: http.DefaultClient,
			}
			got, err := b.RequestPaymentCapture(tt.bankReq)
			if err != nil && tt.wantErr != nil {
				if diff := cmp.Diff(&err, &tt.wantErr, cmpopts.IgnoreFields(models.BankErrorResponse{}, "Error")); diff != "" {
					t.Errorf("Bank.RequestPaymentCapture() error mismatch (-want +got):\n%s", diff)
				}
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Bank.RequestPaymentCapture() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
