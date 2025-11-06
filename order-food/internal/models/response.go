package models

// APIResponse represents a standard API response
type APIResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ErrorResponse creates an error API response
func ErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Type:    "error",
		Message: message,
	}
}

// SuccessResponse creates a success API response
func SuccessResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Type:    "success",
		Message: message,
	}
}
