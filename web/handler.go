package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/jigurd/VirtualTabletop/tabletop"
)

// HandleRoot responds with 404
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println("Error parsing index.html")
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		fmt.Println("Error executing")
	}
}

/*
HandlerRegister handle registering a new user
*/
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("register.html")
	if err != nil {
		fmt.Println("Error parsing register.html")
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		fmt.Println("Error executing")
	}

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

	if newUser.Password != r.FormValue("confirm") {
		fmt.Fprintln(w, "Passwords arent the same lol")
		return
	}

	tabletop.UserDB.Add(newUser)
}
