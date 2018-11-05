package web

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/jigurd/VirtualTabletop/tabletop"
)

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
	switch r.Method {
	case http.MethodPost:
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

	default: // hacky shitty solution for when it comes here the first time xd
		tpl, err := template.ParseFiles("register.html")
		if err != nil {
			fmt.Println("Error parsing register.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing")
		}
	}
}

/*
HandlerLogin handles users logging in
*/
func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		tpl, err := template.ParseFiles("login.html")
		if err != nil {
			fmt.Println("Error parsing login.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing")
		}

		r.ParseForm()

		uName := r.FormValue("username")
		pwd := r.FormValue("password")

		user, err := tabletop.UserDB.Get(uName)
		if err != nil {
			fmt.Fprintf(w, "Couldn't log in: %s", err.Error())
			return
		}

		if pwd != user.Password {
			fmt.Fprintf(w, "Incorrect password. You used: %s, should be: %s", pwd, user.Password)
		}

	default: // hacky shitty solution for when it comes here the first time xd
		tpl, err := template.ParseFiles("login.html")
		if err != nil {
			fmt.Println("Error parsing login.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing")
		}
	}
}
