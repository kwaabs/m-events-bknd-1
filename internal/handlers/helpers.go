package handlers

import (
	"encoding/json"
	"net/http"
)

// decodeJSON decodes r.Body into v.
func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
