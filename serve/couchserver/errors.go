package couchserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/flimzy/kivik"
)

func errorDescription(status int) string {
	switch status {
	case 401:
		return "unauthorized"
	case 400:
		return "bad_request"
	case 404:
		return "not_found"
	case 500:
		return "internal_server_error" // TODO: Validate that this is normative
	}
	panic(fmt.Sprintf("unknown status %d", status))
}

type couchError struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

// HandleError returns a CouchDB-formatted error. It does nothing if err is nil.
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status := kivik.StatusCode(err)
	w.WriteHeader(status)
	wErr := json.NewEncoder(w).Encode(couchError{
		Error:  errorDescription(status),
		Reason: kivik.Reason(err),
	})
	if wErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to send send error: %s", wErr)
	}
}
