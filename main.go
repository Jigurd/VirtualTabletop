package main

import (
	"net/http"
)

func main() {

	http.HandleFunc("/", HandleRoot) // return 404
}
