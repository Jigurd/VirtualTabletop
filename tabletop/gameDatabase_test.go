package tabletop

import (
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"

	mgo "gopkg.in/mgo.v2"
)

func setupGamesDB(t *testing.T) *GamesDB {
	db := &GamesDB{
		DatabaseURL:    "mongodb://localhost",
		DatabaseName:   "testgamesdb",
		CollectionName: "games",
	}

	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()

	if err != nil {
		t.Error(err)
	}

	return db
}

func tearDownGamesDB(t *testing.T, db *GamesDB) {
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

func Test_AddGame(t *testing.T) {
	db := setupGamesDB(t)
	db.Init()
	defer tearDownGamesDB(t, db)

	newGame := Game{
		GameId: bson.NewObjectId().Hex(),
		Name:   "Test Game",
		Owner:  "Test Owner",
		System: "FATAL",
		Players: []string{
			"Test Owner",
		},
		GameMasters: []string{
			"Test Owner",
		},
	}

	if !db.Add(newGame) {
		t.Error("Error adding new game.")
	}
}

func Test_GetGame(t *testing.T) {
	db := setupGamesDB(t)
	db.Init()
	defer tearDownGamesDB(t, db)

	newGame := Game{
		GameId: bson.NewObjectId().Hex(),
		Name:   "Test Game",
		Owner:  "Test Owner",
		System: "FATAL",
		Players: []string{
			"Test Owner",
		},
		GameMasters: []string{
			"Test Owner",
		},
	}

	if !db.Add(newGame) {
		t.Error("Error adding new game.")
		return
	}

	gameFromDB, err := db.Get(newGame.GameId)
	if err != nil {
		t.Error("Error getting game back from DB:", err.Error())
		return
	}

	if !reflect.DeepEqual(newGame, gameFromDB) {
		t.Error("Added and retrieved game are not equal.")
	}

}
