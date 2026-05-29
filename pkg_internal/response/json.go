package response

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lissy93/who-dat/pkg/types"
)

// JSON sends a JSON response with the given status code
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// Error sends a standardized error response
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	JSON(w, statusCode, types.NewErrorResponse(code, message))
}

// BadRequest sends a 400 Bad Request error
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, types.ErrCodeInvalidDomain, message)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, types.ErrCodeUnauthorized, message)
}

// NotFound sends a 404 Not Found error
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, types.ErrCodeDomainNotFound, message)
}

// TooManyRequests sends a 429 Too Many Requests error
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, types.ErrCodeRateLimit, message)
}

// InternalError sends a 500 Internal Server Error
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, types.ErrCodeInternal, message)
}

// GatewayTimeout sends a 504 Gateway Timeout error
func GatewayTimeout(w http.ResponseWriter, message string) {
	Error(w, http.StatusGatewayTimeout, types.ErrCodeTimeout, message)
}
