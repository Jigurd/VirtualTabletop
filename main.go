package main

import (
	"net/http"
	"os"

	"github.com/fogleman/gg"
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

	r.HandleFunc("/createChar", web.HandlerCreate)
	r.HandleFunc("/api/usercount", web.HandleAPIUserCount)
	r.HandleFunc("/profile", web.HandlerProfile)
	r.HandleFunc("/login", web.HandlerLogin)
	r.HandleFunc("/register", web.HandlerRegister)
	r.HandleFunc("/", web.HandleRoot)
	r.HandleFunc("/chat/", web.HandleChat)
	r.HandleFunc("/ws", web.HandleChatConnections)
	r.HandleFunc("/newgame", web.HandleNewGame)
	r.HandleFunc("/gamebrowser", web.HandleGameBrowser)
	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// http.Handle("/chat/", http.StripPrefix("/chat/", http.FileServer(http.Dir("chat"))))

	http.ListenAndServe(":"+port, nil)

	// testing out graphics stuff
	dc := gg.NewContext(1000, 1000)
	dc.DrawCircle(500, 500, 400)
	dc.SetRGB(0, 0, 0)
	dc.Fill()
	dc.SavePNG("out.png")
}
