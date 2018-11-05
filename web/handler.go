package web

import (
	"fmt"
	"net/http"

	"github.com/jigurd/VirtualTabletop/tabletop"
)

// HandleRoot responds with 404
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r) // Respond with 404	"encoding/json"
}

/*
HandlerRegister handle registering a new user
*/
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	newUser := tabletop.User{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Email:    r.FormValue("email"),
	}

	if tabletop.UserDB.Exists(newUser) {
		fmt.Fprintln(w, "That username/email is taken.")
		return
	} else if newUser.Username == "" {
		fmt.Fprintln(w, "Please enter a username u dumb bitch")
		return
	}

	if newUser.Password != r.FormValue("pwdconfirm") {
		fmt.Fprintln(w, "Passwords arent the same lol")
		return
	}

	tabletop.UserDB.Add(newUser)
}
