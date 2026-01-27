package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// ParseJSON parses JSON from request body
func ParseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// GetQueryParam gets a query parameter with default value
func GetQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryParamInt gets a query parameter as integer with default value
func GetQueryParamInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// GetPaginationParams gets pagination parameters from request
func GetPaginationParams(r *http.Request) (page, limit int) {
	page = GetQueryParamInt(r, "page", 1)
	limit = GetQueryParamInt(r, "limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}
