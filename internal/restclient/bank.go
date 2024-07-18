package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

type Bank struct {
	client http.Client
}

func NewBankClient() *Bank {
	return &Bank{
		client: http.Client{},
	}
}

const (
	bankURL = "http://localhost:8080/payments"
)

func (b *Bank) RequestPaymentCapture(bankReq models.BankRequest) (*models.BankResponse, *models.BankErrorResponse) {
	data, err := json.Marshal(bankReq)
	if err != nil {
		return nil, &models.BankErrorResponse{
			Appcode:      models.ErrMarshalStructToBytes,
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Error while converting models.BankRequest{} to bytes while requesting to bank",
			Error:        err,
		}
	}

	r, err := http.NewRequest("POST", bankURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, &models.BankErrorResponse{
			Appcode:      models.ErrCreateNewHTTPRequest,
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Error while creating a request for bank",
			Error:        err,
		}
	}

	resp, err := b.Do(r)
	if err != nil {
		return nil, &models.BankErrorResponse{
			Appcode:      models.ErrBankRequest,
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			Error:        err,
		}
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Bank returned HTTP %d code", resp.StatusCode)
		return nil, &models.BankErrorResponse{
			Appcode:      models.ErrBankResponseStatusCodeNon200,
			StatusCode:   resp.StatusCode,
			ErrorMessage: err.Error(),
			Error:        err,
		}
	}

	defer func(b io.ReadCloser) {
		err = b.Close()
		if err != nil {
			// TODO: Use std log
			log.Println("error while closing response body from bank. Error %w", err)
		}
	}(resp.Body)

	bankResp := models.BankResponse{}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &models.BankErrorResponse{
			Appcode:      models.ErrConvertNetworkResponseToBytes,
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Error while converting bank response to bytes",
			Error:        err,
		}
	}

	err = json.Unmarshal(bytes, &bankResp)
	if err != nil {
		return nil, &models.BankErrorResponse{
			Appcode:      models.ErrUnmarshalBytesToStruct,
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Error while unmarshalling bytes to BankResponse{}",
			Error:        err,
		}
	}

	return &bankResp, nil
}

func (b *Bank) Do(req *http.Request) (*http.Response, error) {
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while making POST request to bank. Error: %w", err)
	}

	// TODO: need to change this logic and implement exponential retry mechanism
	// If bank does not return http.StatusOK
	// what should I send to the merchant?
	// if resp.StatusCode != http.StatusOK {
	// 	return resp, fmt.Errorf("Bank returned HTTP %d code", resp.StatusCode)
	// }

	return resp, nil
}
