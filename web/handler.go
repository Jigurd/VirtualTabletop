package web

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

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
Creates an MD5 hash out of the given string
*/
func md5Hash(val string) string {
	hashed := md5.Sum([]byte(val))
	return fmt.Sprintf("%x", hashed)
}

/*
Reads a file and returns it as a string, and eventual error
*/
func readFile(fileName string) (string, error) {
	htmlByte, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return string(htmlByte), nil
}

/*
HandleRoot handles root
*/
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/index.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		return
	}

	message := "" // Message to be output to the user

	cookie, err := r.Cookie("user") // Get the user cookie
	if err != http.ErrNoCookie {    // If a cookie was found we display a nice welcome message
		message = "<h1>Hello, " + cookie.Value + " :)</h1>"
	}

	bodyEnd := strings.Index(html, "</body>")
	html = html[:bodyEnd] + message + html[bodyEnd:] // Inset the message at the end of the body

	io.WriteString(w, html)
}

//HandlerCreate handle Character creation
func HandlerCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		html, err := readFile("html/create.html")
		if err != nil {
			fmt.Println("Error reading html file:", err.Error())
			return
		}

		io.WriteString(w, html)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error parsing form: %s\n", err.Error())
		}
		//Get values for the character
		//characterName := r.FormValue("charName")
		//system := r.FormValue("system")
		//cookie, err := r.Cookie("user")
		if err != nil {
			fmt.Printf("Error getting username: %s\n", err.Error())
		}
		//userName := cookie.Value

	} else {
		w.WriteHeader(501)
	}

}

/*
HandlerRegister handle registering a new user
*/
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/register.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		return
	}

	message := ""
	statusCode := 200

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error parsing form: %s\n", err.Error())
		}

		newUser := tabletop.User{ // Create a user based on the form values
			Username: r.FormValue("username"),
			Password: md5Hash(r.FormValue("password")),
			Email:    r.FormValue("email"),
		}

		if newUser.Username == "" {
			message = fmt.Sprint("Please enter a username.") // Status code 422: Unprocessable entity means
			statusCode = http.StatusUnprocessableEntity      // the syntax was understood but the data is bad
		} else if newUser.Email == "" {
			message = fmt.Sprint("Please enter an email.")
			statusCode = http.StatusUnprocessableEntity
		} else if tabletop.UserDB.Exists(newUser) {
			message = fmt.Sprint("That username/email is taken.")
			statusCode = http.StatusUnprocessableEntity
		} else { // OK username/Email
			if newUser.Password != md5Hash(r.FormValue("confirm")) {
				message = fmt.Sprint("Passwords don't match.")
				statusCode = http.StatusUnprocessableEntity
			} else { // OK password
				if tabletop.UserDB.Add(newUser) {
					message = fmt.Sprint("User created!")
					w.WriteHeader(http.StatusCreated)
				} else {
					message = fmt.Sprint("Unknonwn error in creating the user.")
					statusCode = http.StatusUnprocessableEntity
				}
			}
		}
	} else if r.Method != http.MethodGet {
		statusCode = http.StatusNotImplemented
	}

	bodyEnd := strings.Index(html, "</body>")
	html = html[:bodyEnd] + "<h3>" + message + "</h3>" + html[bodyEnd:] // Inset the message at the end of the body

	w.WriteHeader(statusCode)
	io.WriteString(w, html)
}

/*
HandlerLogin handles users logging in
*/
func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/login.html") // Conversion from []byte to string
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		return
	}

	message := ""
	statusCode := 200

	if r.Method == http.MethodPost {
		r.ParseForm()

		uName := r.FormValue("username")
		password := md5Hash(r.FormValue("password"))

		user, err := tabletop.UserDB.Get(uName)
		if err != nil {
			message = fmt.Sprintf("Couldn't log in: %s", err.Error())
			statusCode = 500 // Not sure if this code makes sense, but not sure what else to give
		}

		if password == user.Password {
			cookie := &http.Cookie{
				Name:    "user",
				Value:   user.Username,
				Expires: time.Now().Add(15 * time.Minute),
			}
			http.SetCookie(w, cookie)

			http.Redirect(w, r, "/profile", http.StatusMovedPermanently) // TODO: Redirect to my profile?
		} else {
			message = fmt.Sprintf("Couldn't log in")
			statusCode = http.StatusUnprocessableEntity
		}
	} else if r.Method != http.MethodGet { // In Postman it will write this first and then the html, but who cares
		statusCode = http.StatusNotImplemented
	}

	bodyEnd := strings.Index(html, "</body>")                           // Find the position of the closing body tag
	html = html[:bodyEnd] + "<h3>" + message + "</h3>" + html[bodyEnd:] // Inserts the message to the html at the end of the body

	w.WriteHeader(statusCode)
	io.WriteString(w, html)
}

/*
HandlerProfile handles "My Profile"
*/
func HandlerProfile(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/profile.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		return
	}

	message := ""

	userCookie, err := r.Cookie("user")
	if err != http.ErrNoCookie {
		message = "<h2>" + userCookie.Value + "'s profile.</h2>"
	} else {
		message = "<h3>Hmm.. Seems like you are not logged in. Head over to the log in page to change that!</h3>"
	}

	bodyEnd := strings.Index(html, "</body>")
	html = html[:bodyEnd] + message + html[bodyEnd:]

	io.WriteString(w, html)
}

/*
HandlerChatConnections handles chat connections
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

/*
HandleAPIUserCount returns the amount of users in the database
*/
func HandleAPIUserCount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		m := make(map[string]int)
		m["count"] = tabletop.UserDB.Count()
		json.NewEncoder(w).Encode(m)

	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}
