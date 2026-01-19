package handlers

import (
	"net/http"
)

// BaseData holds common information needed by all templates (like layout bits)
type BaseData struct {
	User string
	Data any
}

// GetBaseData extracts the user from the session and returns a base data object
func GetBaseData(r *http.Request) BaseData {
	user := ""
	cookie, err := r.Cookie("session_token")
	if err == nil {
		user = cookie.Value
	}

	return BaseData{
		User: user,
	}
}
