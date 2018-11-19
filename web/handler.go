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
	"github.com/jigurd/VirtualTabletop/img"
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

		user, err := tabletop.UserDB.Get(userCookie.Value)
		if err != nil {
			fmt.Println("Error getting user:", err.Error())
			return
		}

		if len(user.PartOfGames) != 0 {
			htmlData["PartOfAnyGame"] = true
			htmlData["PartOfGames"] = user.PartOfGames

			games := []tabletop.Game{}
			for _, gameID := range user.PartOfGames {
				game, err := tabletop.GameDB.Get(gameID)
				if err != nil {

				}
				games = append(games, game)
			}

			htmlData["Games"] = games
		}
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
				Expires: time.Now().Add(20 * time.Minute),
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/editChar", 303)
		}

	} else {
		w.WriteHeader(501)
	}

}

func HandlerEdit(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("html/edit.html", "html/header.html")
	if err != nil {
		fmt.Println("Error reading register.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	htmlData := make(map[string]interface{})

	cookie, err := r.Cookie("char")
	if err != nil {
		fmt.Printf("Error getting cookie: %s", err.Error())
		return
	}
	charId, _ := strconv.Atoi(cookie.Value)

	var errmsg string
	var character tabletop.Character

	character, errmsg = tabletop.CharDB.FindChar(charId)
	if errmsg != "" {
		fmt.Print(errmsg)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	htmlData["charName"] = character.Charactername
	htmlData["userName"] = character.Username
	htmlData["system"] = character.System

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Printf("Error parsing form: %s\n", err.Error())
			return
		}
		if r.FormValue("Stat") != "" {
			var values []tabletop.NameDesc
			values = append(values, tabletop.NameDesc{r.FormValue("name"), r.FormValue("desc")})
			tabletop.CharDB.UpdateChar_nameDesc(charId, "stats", values)
		} else if r.FormValue("Skill") != "" {
			var values []tabletop.NameDesc
			values = append(values, tabletop.NameDesc{r.FormValue("name"), r.FormValue("desc")})
			tabletop.CharDB.UpdateChar_nameDesc(charId, "skills", values)
		} else if r.FormValue("Inventory") != "" {
			var values []string
			values = append(values, r.FormValue("item"))
			tabletop.CharDB.UpdateCharString(charId, "inventory", values)
		} else if r.FormValue("Money") != "" {
			var values []tabletop.NameDesc
			values = append(values, tabletop.NameDesc{r.FormValue("name"), r.FormValue("desc")})
			tabletop.CharDB.UpdateChar_nameDesc(charId, "money", values)
		} else if r.FormValue("Asset") != "" {
			var values []tabletop.NameDesc
			values = append(values, tabletop.NameDesc{r.FormValue("name"), r.FormValue("desc")})
			tabletop.CharDB.UpdateChar_nameDesc(charId, "assets", values)
		} else if r.FormValue("Tag") != "" {
			var values []string
			values = append(values, r.FormValue("item"))
			tabletop.CharDB.UpdateCharString(charId, "tags", values)
		} else if r.FormValue("Macro") != "" {
			var values []tabletop.NameDesc
			values = append(values, tabletop.NameDesc{r.FormValue("name"), r.FormValue("desc")})
			tabletop.CharDB.UpdateChar_nameDesc(charId, "macros", values)
		} else if r.FormValue("Abilities") != "" {
			var values []tabletop.NameDesc
			values = append(values, tabletop.NameDesc{r.FormValue("name"), r.FormValue("desc")})
			tabletop.CharDB.UpdateChar_nameDesc(charId, "abilities", values)
		}

	} else if r.Method != "GET" {
		w.WriteHeader(501)
		return
	}

	if len(character.Stats) != 0 {
		htmlData["stat"] = true
		htmlData["stats"] = character.Stats
	}
	if len(character.Skills) != 0 {
		htmlData["skill"] = true
		htmlData["skills"] = character.Skills
	}
	if len(character.Inventory) != 0 {
		htmlData["Item"] = true
		htmlData["Inventory"] = character.Inventory
	}
	if len(character.Money) != 0 {
		htmlData["cash"] = true
		htmlData["money"] = character.Money
	}
	if len(character.Assets) != 0 {
		htmlData["asset"] = true
		htmlData["assets"] = character.Assets
	}
	if len(character.Abilities) != 0 {
		htmlData["ability"] = true
		htmlData["abilities"] = character.Abilities
	}
	if len(character.Macros) != 0 {
		htmlData["macro"] = true
		htmlData["macros"] = character.Macros
	}
	if len(character.Tags) != 0 {
		htmlData["tag"] = true
		htmlData["tags"] = character.Tags
	}
	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error reading executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
			PartOfGames: []string{},
			Avatar:      img.ImageData{},
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
HandlerLogout logs a user out
*/
func HandlerLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "user",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
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
		htmlData["LoggedIn"] = true

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

			// get image
			/*
				file, header, err := r.FormFile("fileupload")
				if err != nil {
					fmt.Println("Error", err.Error())
				} else {
					fmt.Println(header.Filename)
				}
				defer file.Close()
			*/
			// dab on the haters

			tabletop.UserDB.UpdateVisibilityInDirectory(user)
			tabletop.UserDB.UpdateDescription(user)

			http.Redirect(w, r, "/profile", http.StatusMovedPermanently) // So it actually refreshes the value for you
			// Obviously a pretty terrible way to do it (way higher data usage), but hey, it works right
		}

	} else {
		htmlData["LoggedIn"] = false
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
		htmlData := make(map[string]interface{})

		_, err = r.Cookie("user")
		if err != http.ErrNoCookie {
			htmlData["LoggedIn"] = true
		}

		err = tpl.Execute(w, htmlData)
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

		gameId := bson.NewObjectId().Hex()
		newGame := tabletop.Game{
			gameId,
			r.FormValue("name"),
			cookie.Value,
			r.FormValue("system"),
			[]string{},
			[]string{},
			r.FormValue("name"),
			10,
		}
		newGame.Players = append(newGame.Players, cookie.Value)
		newGame.GameMasters = append(newGame.GameMasters, cookie.Value)

		tabletop.UserDB.AddGame(cookie.Value, gameId)
		tabletop.GameDB.Add(newGame)
	}
}

/*
HandleGameBrowser shows available games
*/
func HandleGameBrowser(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("html/gamebrowser.html", "html/header.html")
	if err != nil {
		fmt.Println("Error loading gamebrowser.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	htmlData := make(map[string]interface{})

	_, err = r.Cookie("user")
	if err != http.ErrNoCookie {
		htmlData["LoggedIn"] = true
	}

	games := tabletop.GameDB.GetAll()

	type GameData struct {
		Name, ID, Desc, Owner, System string
		PlayerCount, MaxPlayers       int
	}
	gamesData := []GameData{}

	for _, game := range games {
		gamesData = append(gamesData, GameData{
			game.Name,
			game.GameId,
			game.Description,
			game.Owner,
			game.System,
			len(game.Players),
			game.MaxPlayers,
		})
	}

	htmlData["Games"] = gamesData

	if len(games) != 0 {
		htmlData["AnyGames"] = true
	}

	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
	tpl, err := template.ParseFiles("html/game.html", "html/header.html")
	if err != nil {
		fmt.Println("Error loading game.html:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	parts := strings.Split(r.URL.Path, "/") // Try to find the game
	game, err := tabletop.GameDB.Get(parts[2])
	if err != nil {
		fmt.Println("HandleGame error")
		return
	}

	htmlData := make(map[string]interface{})

	htmlData["Name"] = game.Name
	htmlData["System"] = game.System
	htmlData["Owner"] = game.Owner
	htmlData["Players"] = game.Players
	htmlData["Masters"] = game.GameMasters
	htmlData["Desc"] = game.Description

	user, err := r.Cookie("user")
	if err == nil {
		htmlData["LoggedIn"] = true

		if r.Method == http.MethodPost {
			// Join the fucking game
			fmt.Println(r.FormValue("joingame"))
		}

		l := tabletop.InviteLink{}
		if tabletop.InviteLinkDB.HasLink(game) {
			l, err = tabletop.InviteLinkDB.GetByGame(game)
			if err != nil {
				fmt.Println("Link error")
				return
			}
		}

		if user.Value == game.Owner {
			// check if there is an invite link for this game,
			// if not create one
			// TODO: This should be prompted by the owner, not happening automatically
			if !tabletop.InviteLinkDB.HasLink(game) {
				l = tabletop.NewInviteLink(game)
				tabletop.InviteLinkDB.Add(l)
			} else {
				l, err = tabletop.InviteLinkDB.GetByGame(game)
				if err != nil {
					fmt.Println("Link error")
					return
				}
			}
		}
		htmlData["Link"] = l.URL
	}

	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	type PlayerData struct {
		Username, Desc string
	}
	playerData := []PlayerData{}

	htmlData := make(map[string]interface{})

	users := tabletop.UserDB.GetAllVisibleInDirectory()
	if len(users) != 0 {
		htmlData["AnyPlayers"] = true

		for _, user := range users {
			playerData = append(playerData, PlayerData{Username: user.Username, Desc: user.Description})
		}

		htmlData["Players"] = playerData
	}

	_, err = r.Cookie("user")
	if err != http.ErrNoCookie {
		htmlData["LoggedIn"] = true
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

	tpl, err := template.ParseFiles("html/user.html", "html/header.html")
	if err != nil {
		fmt.Println("Error reading user.html")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	htmlData := make(map[string]interface{})

	userCookie, err := r.Cookie("user")
	if err == nil {
		htmlData["LoggedInUsername"] = userCookie.Value
	} else {
		htmlData["LoggedInUsername"] = ""
	}

	htmlData["Username"] = user.Username
	htmlData["Desc"] = user.Description

	err = tpl.Execute(w, htmlData)
	if err != nil {
		fmt.Println("Error executing template:", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

/*
HandleI handles invite links
*/
func HandleI(w http.ResponseWriter, r *http.Request) {
	user, err := r.Cookie("user")
	fmt.Println("Not libtard: " + user.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	l, err := tabletop.InviteLinkDB.Get("127.0.0.1:8080" + r.URL.Path)
	if err != nil {
		return
	}
	g, err := tabletop.GameDB.Get(l.GameId)
	defer http.Redirect(w, r, "/game/"+g.GameId, http.StatusMovedPermanently)
	if err != nil || len(g.Players) >= g.MaxPlayers {
		return
	}
	for _, player := range g.Players {
		if player == user.Value {
			return
		}
	}
	g.Players = append(g.Players, user.Value)
	tabletop.GameDB.UpdatePlayers(g)
	tabletop.UserDB.AddGame(user.Value, g.GameId)
}
