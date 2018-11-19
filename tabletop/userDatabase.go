package tabletop

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	if !db.Exists(u) { // MongoDB index keys need both to be there, so check if it exists first
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

	return false
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

/*
Get returns a user with the given name and a user
*/
func (db *UsersDB) Get(uName string) (User, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	user := User{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"username": uName}).One(&user)

	return user, err
}

/*
GetAll returns a slice containing all existing users.
*/
func (db *UsersDB) GetAll() []User {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var all []User

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{}).All(&all)
	if err != nil {
		return []User{}
	}

	return all
}

/*
Count returns the amount of users in the database
*/
func (db *UsersDB) Count() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.CollectionName).Count()
	if err != nil {
		fmt.Println("Error getting count.")
	}

	return count
}

/*
Remove removes a give user from the database, returns true if a user was removed
*/
func (db *UsersDB) Remove(username string) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.CollectionName).Remove(bson.M{"username": username})
	if err != nil {
		return false
	}
	return true
}

/*
GetAllVisibleInDirectory gets all the users that have enabled visibility in the player directory
*/
func (db *UsersDB) GetAllVisibleInDirectory() []User {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	users := []User{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"options": bson.M{"visibleindirectory": true}}).All(&users)
	if err != nil {
		fmt.Println("Error retrieving users (all visible)")
		return []User{}
	}

	return users
}

/*
UpdateVisibilityInDirectory updates the users visibility in the player directory
*/
func (db *UsersDB) UpdateVisibilityInDirectory(u User) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	updateQ := bson.M{"$set": bson.M{"options": bson.M{"visibleindirectory": u.Options.VisibleInDirectory}}}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Update(bson.M{"username": u.Username}, updateQ)
	if err != nil {
		fmt.Println("Error updating the visibility.")
	}
}

/*
UpdateBio updates the users bio/description
*/
func (db *UsersDB) UpdateDescription(u User) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	updateQ := bson.M{"$set": bson.M{"description": u.Description}}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Update(bson.M{"username": u.Username}, updateQ)
	if err != nil {
		fmt.Println("Error updating the visibility.")
	}
}

/*
AddGame adds a game to the users participating games
*/
func (db *UsersDB) AddGame(username, gameID string) error {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	pushQuery := bson.M{"$push": bson.M{"partofgames": gameID}}
	return session.DB(db.DatabaseName).C(db.CollectionName).Update(bson.M{"username": username}, pushQuery)
}
