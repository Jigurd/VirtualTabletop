package main

import (
	"github.com/gorilla/mux"
	"github.com/jigurd/VirtualTabletop/web"
	"net/http"
	"os"
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
