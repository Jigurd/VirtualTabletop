package main

import (
	//"html/template"
	"net/http"
)

// HandleRoot responds with 404
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r) // Respond with 404	"encoding/json"
}
