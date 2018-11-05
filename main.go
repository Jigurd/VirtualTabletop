package main

import (
	"net/http"
	"os"

	"github.com/jigurd/VirtualTabletop/tabletop"
)

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	http.HandleFunc("/", tabletop.HandleRoot) // return 404

	http.ListenAndServe(":"+port, nil)
}
