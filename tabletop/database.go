package tabletop

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
)

const (
	dbURL = "mongodb://admin:greatadminpassword69@ds151523.mlab.com:51523/virtualtabletop"
)

var (
	UserDB UsersDB // UserDB contains the users
)

func init() {
	UserDB = UsersDB{
		DatabaseURL:    dbURL,
		DatabaseName:   "virtualtabletop",
		CollectionName: "users",
	}

	UserDB.Init()
}

/*
UsersDB connects to the database with User information
*/
type UsersDB struct {
	DatabaseURL    string `json:"databaseurl"`
	DatabaseName   string `json:"databasename"`
	CollectionName string `json:"collectionmame"`
}

/*
Init initialises the users DB
*/
func (db *UsersDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"username", "email"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.CollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

/*
Add adds a new user to the database, returns if the adding was successful
*/
func (db *UsersDB) Add(u User) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.CollectionName).Insert(u)
	if err != nil {
		fmt.Printf("Error inserting user into the DB: %s", err.Error())
		return false
	}

	return true
}

/*
Exists checks if a user already exists
*/
func (db *UsersDB) Exists(u User) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	users := []User{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(nil).All(&users)

	for _, user := range users {
		if u.Username == user.Username {
			return true
		}
		if u.Email == user.Email {
			return true
		}
	}

	return false
}
