package models

const (
	ErrUnmarshalBytesToStruct        = 1
	ErrConvertNetworkResponseToBytes = 2
	ErrCreateNewHTTPRequest          = 4
	ErrMarshalStructToBytes          = 5
	ErrDatastorePaymentCreation      = 6
	ErrBankRequest                   = 7
	ErrBankResponseStatusCodeNon200  = 8
	ErrRequestRejected               = 9
)
