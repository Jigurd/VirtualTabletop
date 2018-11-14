package web

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/sessions"
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

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

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

func setCookie(w http.ResponseWriter, cookie http.Cookie) {
	http.SetCookie(w, &cookie)
}

// HandleRoot responds with 404
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	message := ""

	cookie, err := r.Cookie("user")
	if err != http.ErrNoCookie { // If a cookie was found
		message = "<h1>Hello, " + cookie.Value + "</h1>"
	}

	htmlStr, err := ioutil.ReadFile("html/index.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
	}

	html := string(htmlStr)
	bodyEnd := strings.Index(html, "</body>")
	html = html[:bodyEnd] + message + html[bodyEnd:] // Inset the message at the end of the body

	io.WriteString(w, html)
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
	/*cookie, err := r.Cookie("user")
	if err == http.ErrNoCookie { // If no cookie was found we create a new one
		cookie = &http.Cookie{
			Name: "user",
		}
	}*/

	htmlByte, err := ioutil.ReadFile("html/login.html") // The html file is read as a []byte "string"
	if err != nil {
		fmt.Println("Error reading html file")
		return
	}

	html := string(htmlByte) // Conversion from []byte to string

	var message string // The message to be output to the user

	if r.Method == "POST" {
		r.ParseForm()

		uName := r.FormValue("username")
		password := md5Hash(r.FormValue("password"))

		user, err := tabletop.UserDB.Get(uName)
		if err != nil {
			message = fmt.Sprintf("Couldn't log in: %s", err.Error())
		}

		if password == user.Password {
			cookie := &http.Cookie{
				Name:    "user",
				Value:   user.Username,
				Expires: time.Now().Add(15 * time.Minute),
			}
			http.SetCookie(w, cookie)

			http.Redirect(w, r, "/", http.StatusMovedPermanently) // I guess this code?
		} else {
			message = fmt.Sprintf("Couldn't log in")
		}
	} else if r.Method != "GET" { // In Postman it will write this first and then the html, but who cares
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}

	bodyEnd := strings.Index(html, "</body>")                           // Find the position of the closing body tag
	html = html[:bodyEnd] + "<h3>" + message + "</h3>" + html[bodyEnd:] // Inserts the message to the html at the end of the body

	io.WriteString(w, html)
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

func addCookie(w http.ResponseWriter, name, cookieName string) {
	cookie := http.Cookie{
		Name:  cookieName,
		Value: name,
	}
	http.SetCookie(w, &cookie)
	fmt.Println("Cookie:", cookie)
}
