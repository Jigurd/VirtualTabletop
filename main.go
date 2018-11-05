package main

import (
	"net/http"
	"os"

	"github.com/jigurd/VirtualTabletop/web"

	"github.com/gorilla/mux"
)

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	r := mux.NewRouter()
	r.HandleFunc("/", web.HandleRoot)
	http.Handle("/", r)

	http.ListenAndServe(":"+port, nil)
}
