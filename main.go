package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jigurd/VirtualTabletop/tabletop"
)

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	r := mux.NewRouter()
	r.HandleFunc("/", tabletop.HandleRoot)
	http.Handle("/", r) // return 404

	http.ListenAndServe(":"+port, nil)
}
