package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/pulsar-beam/src/route"
)

func TestStatusAPI(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(StatusPage)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

}
