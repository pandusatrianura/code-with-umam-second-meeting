package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type APIResponse struct {
	Code    string      `json:"code"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing body request")
	}
	return json.NewDecoder(r.Body).Decode(payload)
}

func Success(w http.ResponseWriter, status int, code int, message string, v any) {
	WriteJSONResponse(w, status, APIResponse{
		Code:    strconv.Itoa(code),
		Message: message,
		Data:    v,
	})
}

func Error(w http.ResponseWriter, status int, code int, message string, err error) {
	var e interface{}
	if err != nil {
		e = err.Error()
	}
	WriteJSONResponse(w, status, APIResponse{
		Code:    strconv.Itoa(code),
		Message: fmt.Sprintf("%s: %s", message, e),
	})
}
