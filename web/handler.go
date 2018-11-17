package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jigurd/VirtualTabletop/tabletop"
	"gopkg.in/mgo.v2/bson"
)

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

var Clients map[*websocket.Conn]bool
var Broadcast chan Message
var Upgrader websocket.Upgrader

// HandleRoot loads index.html
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/index.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		_, err := r.Cookie("user") // check if the User is logged in
		if err != nil {            // if the user is not logged in
			http.Redirect(w, r, "/", 303) //Throw user back to the index
			return
		}
		html, err := readFile("html/create.html")
		if err != nil {
			fmt.Println("Error reading html file:", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		io.WriteString(w, html)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error parsing form: %s\n", err.Error())
			return
		}
		//Get values for the character
		cookie, err := r.Cookie("user")
		if err != nil {
			fmt.Printf("Error getting username: %s\n", err.Error())
			return
		}
		userName := cookie.Value
		characterName := r.FormValue("charName")
		system := r.FormValue("system")

		intId, errormsg := tabletop.CreateChar(characterName, userName, system)
		id := strconv.Itoa(intId)
		if errormsg != "" {
			fmt.Printf(errormsg)
			return
		} else {
			cookie = &http.Cookie{
				Name:    "char",
				Value:   id,
				Expires: time.Now().Add(5 * time.Minute),
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/editChar", 303)
		}

	} else {
		w.WriteHeader(501)
	}

}

func HandlerEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		html, err := readFile("html/edit.html")
		if err != nil {
			fmt.Printf("Error reading html file: %s", err.Error())
			return
		}
		cookie, err := r.Cookie("char")
		if err != nil {
			fmt.Printf("Error getting cookie: %s", err.Error())
			return
		}
		charId, _ := strconv.Atoi(cookie.Value)

		errmsg, userName := tabletop.CharDB.GetString(charId, "username")
		if errmsg != "" {
			fmt.Print(errmsg)
			return
		}
		errmsg, charName := tabletop.CharDB.GetString(charId, "charname")
		if errmsg != "" {
			fmt.Print(errmsg)
			return
		}
		errmsg, system := tabletop.CharDB.GetString(charId, "system")
		if errmsg != "" {
			fmt.Print(errmsg)
			return
		}

		page := "<!DOCTYPE html><html><body><h1>" + charName + "</h1><h3>" + userName + "</h3><h4>" + system + "</h4>" + html

		io.WriteString(w, page)

	} else if r.Method == "POST" {

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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	message := ""
	statusCode := http.StatusOK

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error parsing form: %s\n", err.Error())
		}

		newUser := tabletop.User{ // Create a user based on the form values
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
			Email:    r.FormValue("email"),
			Options: tabletop.UserOptions{
				VisibleInDirectory: true, // Visible by default
			},
		}

		if newUser.Username == "" {
			message = "Please enter a username."        // Status code 422: Unprocessable entity means
			statusCode = http.StatusUnprocessableEntity // the syntax was understood but the data is bad
		} else if newUser.Email == "" {
			message = "Please enter an email."
			statusCode = http.StatusUnprocessableEntity
		} else if !validEmail(newUser.Email) {
			message = "Email is invalid."
			statusCode = http.StatusUnprocessableEntity
		} else if !validPassword(newUser.Password) {
			message = "Password is invalid"
			statusCode = http.StatusUnprocessableEntity
		} else if tabletop.UserDB.Exists(newUser) {
			message = "That username/email is taken."
			statusCode = http.StatusUnprocessableEntity
		} else { // OK username/Email
			if newUser.Password != r.FormValue("confirm") {
				message = "Passwords don't match."
				statusCode = http.StatusUnprocessableEntity
			} else { // OK password, eveything is OK and the user is added.
				newUser.Password = md5Hash(newUser.Password) // Hash the password before storing it
				if tabletop.UserDB.Add(newUser) {
					message = "User created!"
					statusCode = http.StatusCreated
				} else {
					message = "Unknonwn error in creating the user."
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

			http.Redirect(w, r, "/profile", http.StatusMovedPermanently)
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
	userCookie, err := r.Cookie("user")
	if err != http.ErrNoCookie {
		tpl, err := template.ParseFiles("html/profile.html") // Parse the HTML template
		if err != nil {
			fmt.Println("Error reading profile.html:", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		type Option struct { // Options for select tag in the HTML file
			Value, Text string
			Selected    bool
		}

		htmlData := []Option{} // Data for the HTML template
		user, err := tabletop.UserDB.Get(userCookie.Value)
		if err != nil {
			fmt.Println("Error getting user.")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		visible := user.Options.VisibleInDirectory

		htmlData = []Option{
			Option{
				Value:    "visible",
				Text:     "Visible",
				Selected: true,
			},
			Option{
				Value:    "notvisible",
				Text:     "Not visisble",
				Selected: true,
			},
		}

		if visible { // Is visible, set notvisible to not be selected
			htmlData[1].Selected = false
		} else {
			htmlData[0].Selected = false
		}

		if r.Method == http.MethodPost { // Update the profile on POST
			r.ParseForm()

			if r.Form["visible"][0] == "visible" {
				user.Options.VisibleInDirectory = true
			} else {
				user.Options.VisibleInDirectory = false
			}

			if visible != user.Options.VisibleInDirectory { // Only update the database if there actually was a change
				tabletop.UserDB.UpdateVisibilityInDirectory(user)
				http.Redirect(w, r, "/profile", http.StatusMovedPermanently) // So it actually refreshes the value for you
			} // This is such a shitty way to do it but fuck it :) Who cares about data usage these days right
		}

		tpl.Execute(w, htmlData)
	} else { // For now we just 404 on non-logged in users
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

/*
HandleChat handles "Chat" (/chat)
*/
func HandleChat(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/chat.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	message := ""

	bodyEnd := strings.Index(html, "</body>")
	html = html[:bodyEnd] + message + html[bodyEnd:]

	io.WriteString(w, html)

	go HandleChatMessages()
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

/*
HandleNewGame handles the creation of a new game
*/
func HandleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		html, err := readFile("html/newgame.html")
		if err != nil {
			log.Fatal(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		message := ""

		_, err = r.Cookie("user")
		if err == http.ErrNoCookie {
			message = "<h3>Hmm.. Seems like you are not logged in. Head over to the log in page to change that!</h3>"
		}

		bodyEnd := strings.Index(html, "</body>")
		html = html[:bodyEnd] + message + html[bodyEnd:]

		io.WriteString(w, html)
	} else if r.Method == "POST" {
		cookie, err := r.Cookie("user")
		if err != nil || cookie.Value == "" {
			fmt.Fprintf(w, "You are not logged in you retard. You fucking imbecile.")
			return
		}
		fmt.Println(cookie.Value)
		err = r.ParseForm()
		if err != nil {
			fmt.Printf("Error parsing form: %s\n", err.Error())
		}

		newGame := tabletop.Game{
			bson.NewObjectId().Hex(),
			r.FormValue("name"),
			cookie.Value,
			r.FormValue("system"),
			[]string{},
			[]string{},
		}
		newGame.Players = append(newGame.Players, cookie.Value)
		newGame.GameMasters = append(newGame.GameMasters, cookie.Value)
		tabletop.GameDB.Add(newGame)
	}
}

/*
HandleGameBrowser shows available games
TODO: Cool html thing
*/
func HandleGameBrowser(w http.ResponseWriter, r *http.Request) {
	games := tabletop.GameDB.GetAll()
	for _, game := range games {
		fmt.Fprintln(w, "<div><a href=\"/game/"+game.GameId+"\">"+game.Name+"</a></div>")
	}
}

// HandlerBoard loads board.html
func HandlerBoard(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/board.html")
	if err != nil {
		fmt.Println("Error loading board.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, html)
}

/*
HandleGame handles the page of one game
*/
func HandleGame(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	game, err := tabletop.GameDB.Get(parts[2])
	if err != nil {
		fmt.Println("HandleGame error")
		return
	}
	fmt.Fprintln(w, game.Name)
	fmt.Fprintln(w, game.System)
	fmt.Fprintln(w, game.Owner)
	fmt.Fprintln(w, game.Players)
	fmt.Fprintln(w, game.GameMasters)

	user, err := r.Cookie("user")
	if err == nil && user.Value == game.Owner {
		// check if there is an invite link for this game,
		// if not create one
		// TODO: This should be prompted by the owner, not happening automatically
		l := tabletop.InviteLink{}
		if !tabletop.InviteLinkDB.HasLink(game) {
			l = tabletop.NewInviteLink(game)
			tabletop.InviteLinkDB.Add(l)
		}
		fmt.Fprintln(w, l)
	}
}

/*
HandlePlayerDirectory shows all players
*/
func HandlePlayerDirectory(w http.ResponseWriter, r *http.Request) {
	html, err := readFile("html/playerdirectory.html")
	if err != nil {
		fmt.Println("Error reading playerdirectory.html")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	message := ""

	users := tabletop.UserDB.GetAllVisibleInDirectory()

	if len(users) != 0 { // Make a list as long as there are players to make a list of
		message = "<ul>"
	}

	for _, user := range users {
		message += "<li><div><a href=\"/u/" + user.Username + "\">" + user.Username + "</a></div></li>"
	}

	if len(users) != 0 {
		message += "</ul>"
	}

	bodyEnd := strings.Index(html, "</body>")
	html = html[:bodyEnd] + message + html[bodyEnd:]

	io.WriteString(w, html)
}

/*
HandleU handles a user profile
*/
func HandleU(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	user, err := tabletop.UserDB.Get(parts[2])
	if err != nil {
		fmt.Println("HandleU error")
		return
	}
	fmt.Fprintln(w, user.Username+"\nDescription\nPreferred systems\nSend message (not implemented)\nInvite to game (not implemented)")
}

/*
HandleI handles invite links
*/
func HandleI(w http.ResponseWriter, r *http.Request) {
	user, err := r.Cookie("user")
	l, err := tabletop.InviteLinkDB.Get(r.URL.Path)
	if err != nil {
		fmt.Println("Oh no no no")
		return
	}
	g, err := tabletop.GameDB.Get(l.GameId)
	if err != nil {
		fmt.Println("No no no no no no")
	}
	fmt.Println(user.Value + " joined " + g.Name)
}
