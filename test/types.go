package types

import "time"

type ErrorData struct {
	StatusCode int       `json:"status_code,omitempty"`
	Message    string    `json:"message,omitempty"`
	Public     bool      `json:"public"`
	CreatedAt  time.Time `genType:"Date"`
	Deleted    bool
}

type Errors []ErrorData
