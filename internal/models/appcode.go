package models

// The following codes are application level code that denotes different
// status of an operation. These codes are returned to the user.
// The Operations PRD describes what each of these codes mean and the necessary
// action points
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
