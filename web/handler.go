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
		fmt.Println("Error executing index.html")
	}
}

/*
HandlerRegister handle registering a new user
*/
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: // Shitty solution for when we're redirected from index.html
		tpl, err := template.ParseFiles("register.html")
		if err != nil {
			fmt.Println("Error parsing register.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing register.html")
		}

	case http.MethodPost:
		tpl, err := template.ParseFiles("register.html")
		if err != nil {
			fmt.Println("Error parsing register.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing register.html")
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
			fmt.Fprintln(w, "Please enter a username.")
			return
		}

		if newUser.Password != r.FormValue("confirm") {
			fmt.Fprintln(w, "Passwords don't match.")
			return
		}

		tabletop.UserDB.Add(newUser)

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
		tpl, err := template.ParseFiles("login.html")
		if err != nil {
			fmt.Println("Error parsing login.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing login.html")
		}

	case http.MethodPost:
		tpl, err := template.ParseFiles("login.html")
		if err != nil {
			fmt.Println("Error parsing login.html")
		}

		err = tpl.Execute(w, nil)
		if err != nil {
			fmt.Println("Error executing login.html")
		}

		r.ParseForm()

		uName := r.FormValue("username")
		pwd := r.FormValue("password")

		user, err := tabletop.UserDB.Get(uName)
		if err != nil {
			fmt.Fprintf(w, "Couldn't log in: %s", err.Error())
			return
		}

		if pwd == user.Password {
			fmt.Fprintf(w, "Welcome back, %s", user.Username)
		} else {
			fmt.Fprintf(w, "Incorrect password. You used: %s, should be: %s", pwd, user.Password)
		}

	default:
		statusCode := http.StatusNotImplemented
		http.Error(w, http.StatusText(statusCode), statusCode)
	}
}
