package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func ExecuteRequest(req *http.Request, router *chi.Mux) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func NewRequest(t *testing.T, method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	return req
}

func NewAuthenticatedRequest(t *testing.T, method string, url string, body io.Reader, token string) *http.Request {
	req := NewRequest(t, method, url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func UnmarshallReqBody[T any](t *testing.T, data *bytes.Buffer) *T {
	var body T
	err := json.Unmarshal(data.Bytes(), &body)
	if err != nil {
		t.Errorf("could not parse body into expected format, actual: \n%s", data.String())
		return nil
	}
	return &body
}

func MarshallBody(t *testing.T, body any) io.Reader {
	if body == nil {
		return nil
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	return bytes.NewBuffer(jsonBody)
}
