package models

import (
	"net/http"
	"time"
)

// RawHTTPResponse is one public HTTP attempt before status handling or decoding.
// Body is the exact body delivered by net/http, after any automatic decompression.
// All byte slices and headers are owned by the receiver and may be retained.
// Requests without a response do not produce a record. Retries produce separate records.
type RawHTTPResponse struct {
	Service     string
	Method      string
	URL         string
	RequestBody []byte
	StatusCode  int
	Header      http.Header
	ReceivedAt  time.Time
	Body        []byte
	Complete    bool
	ReadError   string
}
