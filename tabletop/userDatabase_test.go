package tabletop

import (
	"testing"

	mgo "gopkg.in/mgo.v2"
)

func setup(t *testing.T) *UsersDB {
	db := &UsersDB{
		DatabaseURL:    "mongodb://localhost",
		DatabaseName:   "testusersdb",
		CollectionName: "users",
	}

	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()

	if err != nil {
		t.Error(err)
	}

	return db
}

func tearDown(t *testing.T, db *UsersDB) {
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()

	if err != nil {
		t.Error(err)
	}

	err = session.DB(db.DatabaseName).DropDatabase()
	if err != nil {
		t.Error(err)
	}
}

/*
Tests if adding a user is possible. Also tests the Exists function.
*/
func TestAdd(t *testing.T) {
	db := setup(t)
	db.Init()
	defer tearDown(t, db)

	newUser := User{
		Username: "Testing",
		Password: "Password_test",
		Email:    "tester@testing.us.gov.org",
	}

	added := db.Add(newUser)
	if !added {
		t.Error("Error adding to database")
	}

	if !db.Exists(newUser) {
		t.Error("The user doesn't exist in the database.")
	}

	userFromDB, err := db.Get("Testing")
	if err != nil {
		t.Errorf("Error getting the user back from the DB: %s", err.Error())
	}

	if userFromDB.Username != newUser.Username {
		t.Errorf("Expected username '%s' differs from actual '%s'.", newUser.Username, userFromDB.Username)
	}
	if userFromDB.Password != newUser.Password {
		t.Errorf("Expected username '%s' differs from actual '%s'.", newUser.Password, userFromDB.Password)
	}
	if userFromDB.Email != newUser.Email {
		t.Errorf("Expected username '%s' differs from actual '%s'.", newUser.Email, userFromDB.Email)
	}
}

/*
Tests that adding the same user twice fails
*/
func TestAddDuplicateName(t *testing.T) {
	db := setup(t)
	db.Init()
	defer tearDown(t, db)

	newUser := User{
		Username: "Testing",
		Password: "Password_test",
		Email:    "tester@testing.us.gov.org",
	}

	db.Add(newUser)

	if db.Add(newUser) {
		t.Error("Managed to add same user twice.")
	}
}
