package common

// ErrorResponse creates a standard error response
func ErrorResponse(message string) map[string]string {
	return map[string]string{"error": message}
}

// DataResponse wraps data in a standard response format
func DataResponse(data any) map[string]any {
	return map[string]any{"data": data}
}

// BalancesResponse creates a standard balances response
func BalancesResponse(balances any) map[string]any {
	return map[string]any{"balances": balances}
}

// NotFoundResponse creates a standard not found response
func NotFoundResponse() map[string]any {
	return map[string]any{"found": false}
}

// FoundResponse creates a response with found status and data
func FoundResponse(key string, data any) map[string]any {
	return map[string]any{
		"found": true,
		key:     data,
	}
}
