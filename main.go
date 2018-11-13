package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jigurd/VirtualTabletop/web"
)

func main() {
	port, portOk := os.LookupEnv("PORT")
	if !portOk {
		port = "8080" // 8080 is used as the default port
	}

	web.Clients = make(map[*websocket.Conn]bool)
	web.Broadcast = make(chan web.Message)
	web.Upgrader = websocket.Upgrader{}

	r := mux.NewRouter()
	r.HandleFunc("/profile", web.HandlerProfile)
	r.HandleFunc("/login", web.HandlerLogin)
	r.HandleFunc("/register", web.HandlerRegister)
	r.HandleFunc("/", web.HandleRoot)
	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/chat/", http.StripPrefix("/chat/", http.FileServer(http.Dir("chat"))))
	http.HandleFunc("/ws", web.HandleChatConnections)
	go web.HandleChatMessages()

	http.ListenAndServe(":"+port, nil)
}
