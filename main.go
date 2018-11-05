package main

import (
	"net/http"
	"os"
)

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	http.HandleFunc("/", HandleRoot) // return 404

	http.ListenAndServe(":"+port, nil)
}
