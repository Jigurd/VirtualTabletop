package web

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"regexp"
)

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
Checks that an email is valid
*/
func validEmail(email string) bool {
	// This is obviously not a complete validator, but just to have some rules to follow on user creation
	validator, err := regexp.Compile("^[A-Za-z0-9]{3,}@[a-z]{2,}.[a-z]{2,}$")
	if err != nil {
		fmt.Println("Error compiling regex:", err.Error())
		return false
	}

	return validator.MatchString(email)
}

/*
Checks that a password is valid
*/
func validPassword(password string) bool {
	return len(password) >= 5 // Amazing validator
}
