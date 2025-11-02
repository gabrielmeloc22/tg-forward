package rules

import (
	"encoding/json"
	"net/http"
)

func Wrap[T any](h func(w http.ResponseWriter, r *http.Request) (T, *Error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		res, err := h(w, r)
		encoder := json.NewEncoder(w)

		if err != nil {
			w.WriteHeader(err.StatusCode)
			encoder.Encode(ApiErrorResponse{
				Code:    err.Code,
				Message: err.Message,
				Meta:    err.Meta,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		encoder.Encode(res)
	}
}

func WrapWithBody[Req any, Res any](h func(w http.ResponseWriter, r *http.Request, body *Req) (Res, *Error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var body Req
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ApiErrorResponse{
				Code:    "INVALID_REQUEST_BODY",
				Message: "Invalid request body",
			})
			return
		}

		res, apiErr := h(w, r, &body)
		encoder := json.NewEncoder(w)

		if apiErr != nil {
			w.WriteHeader(apiErr.StatusCode)
			encoder.Encode(ApiErrorResponse{
				Code:    apiErr.Code,
				Message: apiErr.Message,
				Meta:    apiErr.Meta,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		encoder.Encode(res)
	}
}
