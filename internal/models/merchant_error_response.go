package models

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrResponse type is what we send back to the merchant upon any error from their action
type ErrResponse struct {
	HTTPStatusCode int    `json:"status_code"`        // http response status code
	StatusText     string `json:"status_text"`        // user-level status message
	AppCode        int64  `json:"app_code,omitempty"` // application-specific error code
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}
