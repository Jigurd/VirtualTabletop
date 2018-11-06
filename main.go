package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jigurd/VirtualTabletop/web"
)

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	r := mux.NewRouter()
	r.HandleFunc("/login", web.HandlerLogin)
	r.HandleFunc("/register", web.HandlerRegister)
	r.HandleFunc("/", web.HandleRoot)
	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":"+port, nil)
}
