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

/*
HandleRoot loads index.html
*/
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("html/index.html", "html/header.html") // Parse the HTML files into a template
	if err != nil {
		fmt.Println("Error reading profile.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	htmlData := make(map[string]interface{})

	userCookie, err := r.Cookie("user") // Get the user cookie
	if err != http.ErrNoCookie {        // If a cookie was found we display a nice welcome message
		htmlData["LoggedIn"] = true
		htmlData["Message"] = "Hello, " + userCookie.Value + " :)"
	}

	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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

		page := "<!DOCTYPE html><html><body><h2>" + charName + "</h2><h3>" + userName + "</h3><h4>" + system + "</h4>" + html

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
	tpl, err := template.ParseFiles("html/register.html", "html/header.html")
	if err != nil {
		fmt.Println("Error reading register.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	htmlData := make(map[string]interface{})
	statusCode := http.StatusOK

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error parsing form: %s\n", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		newUser := tabletop.User{ // Create a user based on the form values
			Username:    r.FormValue("username"),
			Password:    r.FormValue("password"),
			Email:       r.FormValue("email"),
			Description: "",
			Options: tabletop.UserOptions{
				VisibleInDirectory: true, // Visible by default
			},
		}

		if newUser.Username == "" {
			htmlData["Message"] = "Please enter a username." // Status code 422: Unprocessable entity means
			statusCode = http.StatusUnprocessableEntity      // the syntax was understood but the data is bad
		} else if newUser.Email == "" {
			htmlData["Message"] = "Please enter an email."
			statusCode = http.StatusUnprocessableEntity
		} else if !validEmail(newUser.Email) {
			htmlData["Message"] = "Email is invalid."
			statusCode = http.StatusUnprocessableEntity
		} else if !validPassword(newUser.Password) {
			htmlData["Message"] = "Password is invalid"
			statusCode = http.StatusUnprocessableEntity
		} else if tabletop.UserDB.Exists(newUser) {
			htmlData["Message"] = "That username/email is taken."
			statusCode = http.StatusUnprocessableEntity
		} else { // OK username/Email
			if newUser.Password != r.FormValue("confirm") {
				htmlData["Message"] = "Passwords don't match."
				statusCode = http.StatusUnprocessableEntity
			} else { // OK password, eveything is OK and the user is added.
				newUser.Password = md5Hash(newUser.Password) // Hash the password before storing it
				if tabletop.UserDB.Add(newUser) {
					htmlData["Message"] = "User created!"
					statusCode = http.StatusCreated
				} else {
					htmlData["Message"] = "Unknonwn error in creating the user."
					statusCode = http.StatusUnprocessableEntity
				}
			}
		}
	} else if r.Method != http.MethodGet {
		statusCode = http.StatusNotImplemented
	}

	w.WriteHeader(statusCode)
	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error reading executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

/*
HandlerLogin handles users logging in
*/
func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("html/login.html", "html/header.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	htmlData := make(map[string]interface{})
	statusCode := 200

	if r.Method == http.MethodPost {
		r.ParseForm()

		uName := r.FormValue("username")
		password := md5Hash(r.FormValue("password"))

		user, err := tabletop.UserDB.Get(uName)
		if err != nil {
			htmlData["Message"] = fmt.Sprintf("Couldn't log in: %s", err.Error())
			statusCode = 500 // Not sure if this code makes sense, but not sure what else to give
		}

		if password == user.Password {
			cookie := &http.Cookie{
				Name:    "user",
				Value:   user.Username,
				Expires: time.Now().Add(60 * time.Minute),
			}
			http.SetCookie(w, cookie)

			http.Redirect(w, r, "/profile", http.StatusMovedPermanently)
		} else {
			htmlData["Message"] = fmt.Sprintf("Couldn't log in")
			statusCode = http.StatusUnprocessableEntity
		}
	} else if r.Method != http.MethodGet { // In Postman it will write this first and then the html, but who cares
		statusCode = http.StatusNotImplemented
	}

	w.WriteHeader(statusCode)
	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
	}
}

/*
HandlerProfile handles "My Profile"
*/
func HandlerProfile(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("html/profile.html", "html/header.html") // Parse the HTML template
	if err != nil {
		fmt.Println("Error reading profile.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	htmlData := make(map[string]interface{}) // The data that will be used with the HTML template
	userCookie, err := r.Cookie("user")
	if err != http.ErrNoCookie { // A user cookie was found
		htmlData["UserFound"] = true

		type Option struct { // Options for select tag in the HTML file
			Value, Text string
			Selected    bool
		}

		user, err := tabletop.UserDB.Get(userCookie.Value)
		if err != nil {
			fmt.Println("Error getting user.")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		visibleOptions := []Option{ // The options for the "visible" selector
			Option{
				Value:    "visible",
				Text:     "Visible",
				Selected: false,
			},
			Option{
				Value:    "notvisible",
				Text:     "Not visisble",
				Selected: false,
			},
		}

		if user.Options.VisibleInDirectory { // Set the values corresponding to the value from the DB
			visibleOptions[0].Selected = true
		} else {
			visibleOptions[1].Selected = true
		}

		htmlData["VisibleOptions"] = visibleOptions
		htmlData["Desc"] = user.Description

		if r.Method == http.MethodPost { // Update the profile on POST
			r.ParseForm()

			user.Description = r.FormValue("desc")

			if r.Form["visible"][0] == "visible" { // The first option is "Visible"
				user.Options.VisibleInDirectory = true
			} else if r.Form["visible"][0] == "notvisible" { // If the form has neither of the two
				user.Options.VisibleInDirectory = false // nothing is done
			}

			tabletop.UserDB.UpdateVisibilityInDirectory(user)
			tabletop.UserDB.UpdateDescription(user)

			http.Redirect(w, r, "/profile", http.StatusMovedPermanently) // So it actually refreshes the value for you
			// Obviously a pretty terrible way to do it (way higher data usage), but hey, it works right
		}

	} else {
		htmlData["UserFound"] = false
	}

	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error executing profile.html template:", err.Error())
	}
}

/*
HandleChat handles "Chat" (/chat)
*/
func HandleChat(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("html/chat.html", "html/header.html")
	if err != nil {
		fmt.Println("Error reading html file:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//message := ""

	//bodyEnd := strings.Index(html, "</body>")
	//html = html[:bodyEnd] + message + html[bodyEnd:]

	err = tpl.Execute(w, nil)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

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
		tpl, err := template.ParseFiles("html/newgame.html", "html/header.html")
		if err != nil {
			log.Fatal(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		//message := ""

		_, err = r.Cookie("user")
		if err == http.ErrNoCookie {
			//message = "<h3>Hmm.. Seems like you are not logged in. Head over to the log in page to change that!</h3>"
		}

		//bodyEnd := strings.Index(html, "</body>")
		//html = html[:bodyEnd] + message + html[bodyEnd:]

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing template:", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
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
	tpl, err := template.ParseFiles("html/board.html", "html/header.html")
	if err != nil {
		fmt.Println("Error loading board.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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
	tpl, err := template.ParseFiles("html/playerdirectory.html", "html/header.html")
	if err != nil {
		fmt.Println("Error reading playerdirectory.html")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	htmlData := make(map[string]interface{})

	users := tabletop.UserDB.GetAllVisibleInDirectory()
	if len(users) != 0 {
		htmlData["AnyPlayers"] = true

		players := []string{}

		for _, user := range users {
			players = append(players, user.Username)
		}

		htmlData["Players"] = players

	} else {
		htmlData["AnyPlayers"] = false
	}

	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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
