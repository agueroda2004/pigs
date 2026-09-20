package http

import (
	"encoding/json"
	"errors"
	"net/http"
)

const maxRequestBodySize = 1 << 20

// DecodeJSON decodes the request body into destination with a size limit.
// It rejects unknown fields and returns a generic validation error.
func DecodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("El cuerpo de la solicitud no es válido")
	}
	return nil
}

// WriteJSON writes the value as JSON with the given status code.
// It also sets the Content-Type header.
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// WriteError writes an error message as a JSON error response.
// It uses WriteJSON with the given status code.
func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}
