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

	r.HandleFunc("/logout", web.HandlerLogout)
	r.HandleFunc("/createChar", web.HandlerCreate)
	r.HandleFunc("/editChar", web.HandlerEdit)
	r.HandleFunc("/viewChar",web.HandlerView)
	r.HandleFunc("/api/usercount", web.HandleAPIUserCount)
	r.HandleFunc("/profile", web.HandlerProfile)
	r.HandleFunc("/login", web.HandlerLogin)
	r.HandleFunc("/register", web.HandlerRegister)
	r.HandleFunc("/board", web.HandlerBoard)
	r.HandleFunc("/", web.HandleRoot)
	r.HandleFunc("/chat/", web.HandleChat)
	r.HandleFunc("/ws", web.HandleChatConnections)
	r.HandleFunc("/newgame", web.HandleNewGame)
	r.HandleFunc("/gamebrowser", web.HandleGameBrowser)
	r.HandleFunc("/game/{id}", web.HandleGame)
	r.HandleFunc("/game/{id}/board", web.HandleGameBoard)
	r.HandleFunc("/u/{id}", web.HandleU)
	r.HandleFunc("/i/{id}", web.HandleI)
	r.HandleFunc("/playerdirectory", web.HandlePlayerDirectory)
	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":"+port, nil)
}
