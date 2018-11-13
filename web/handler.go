package web

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jigurd/VirtualTabletop/tabletop"
)

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var Clients map[*websocket.Conn]bool
var Broadcast chan Message
var Upgrader websocket.Upgrader

/*
Removes empty strings from an array
*/
func removeEmpty(arr []string) []string {
	newArr := []string{}
	for _, val := range arr {
		if val != "" {
			newArr = append(newArr, val)
		}
	}
	return newArr
}

func md5Hash(val string) string {
	hashed := md5.Sum([]byte(val))
	return fmt.Sprintf("%x", hashed)
}

// HandleRoot responds with 404
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("html/index.html")
	if err != nil {
		fmt.Println("Error parsing index.html")
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		fmt.Println("Error executing index.html")
	}
}

/*
HandlerRegister handle registering a new user
*/
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Printf("Error parsing form: %s\n", err.Error())
	}

	switch r.Method {
	case http.MethodGet: // Shitty solution for when we're redirected from index.html
		tpl, err := template.ParseFiles("html/register.html")
		if err != nil {
			fmt.Println("Error parsing register.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing register.html")
		}

	case http.MethodPost:
		tpl, err := template.ParseFiles("html/register.html")
		if err != nil {
			fmt.Println("Error parsing register.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing register.html")
		}

		newUser := tabletop.User{
			Username: r.FormValue("username"),
			Password: md5Hash(r.FormValue("password")),
			Email:    r.FormValue("email"),
		}

		if tabletop.UserDB.Exists(newUser) {
			fmt.Fprintln(w, "That username/email is taken.")
			return
		} else if newUser.Username == "" {
			fmt.Fprintln(w, "Please enter a username.")
			return
		}

		if newUser.Password != md5Hash(r.FormValue("confirm")) {
			fmt.Fprintln(w, "Passwords don't match.")
			return
		}

		if tabletop.UserDB.Add(newUser) {
			fmt.Fprintln(w, "User created!")
		} else {
			fmt.Fprintln(w, "Unknonwn error in creating the user.")
		}

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerLogin handles users logging in
*/
func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tpl, err := template.ParseFiles("html/login.html")
		if err != nil {
			fmt.Println("Error parsing login.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing login.html")
		}

	case http.MethodPost:
		tpl, err := template.ParseFiles("html/login.html")
		if err != nil {
			fmt.Println("Error parsing login.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing login.html")
		}

		r.ParseForm()

		uName := r.FormValue("username")
		password := md5Hash(r.FormValue("password"))

		user, err := tabletop.UserDB.Get(uName)
		if err != nil {
			fmt.Fprintf(w, "Couldn't log in: %s", err.Error())
			return
		}

		if password == user.Password {
			fmt.Fprintf(w, "Welcome back, %s", user.Username)
			// Create a token
			CreateToken(uName)
		} else {
			fmt.Fprintf(w, "Incorrect password")
		}

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}

/*
HandlerProfile handles "My Profile"
*/
func HandlerProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tpl, err := template.ParseFiles("html/profile.html")
		if err != nil {
			fmt.Println("Error parsing profile.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing profile.html")
		}

		// Get user and shit from cookies I guess
	}
}

/*
HandlerConnections handles chat connections
*/
func HandleChatConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	Clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(Clients, ws)
			break
		}
		Broadcast <- msg
	}
}

/*
HandleChatMessages handles chat messages
*/
func HandleChatMessages() {
	for {
		msg := <-Broadcast
		for client := range Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(Clients, client)
			}
		}
	}
}
