package tabletop

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
Game is a game TRPG that has been set up by a GM
and contains information about the game's name, system,
players, description, etc
*/
type Game struct {
	GameId      bson.ObjectId `bson:"gameid"`
	Name        string        `json:"name"`    // the is the name given to the game
	Owner       string        `json:"owner"`   // the owner is the creator and "admin" of the game
	System      string        `json:"system"`  // the system ran eg Shadowrun 5e, Stars Without Number etc
	Players     []string      `json:"players"` // the people playing in the game, (including GMs?)
	GameMasters []string      `json:"payers"`  // the people running the game, can access all information in the game etc
}

/*
GamesDB connects the db with game info
*/
type GamesDB struct {
	DatabaseURL    string `json:"databaseurl"`
	DatabaseName   string `json:"databasename"`
	CollectionName string `json:"collectionmame"`
}

var GameDB GamesDB // the db containing all games

func init() {
	GameDB = GamesDB{
		dbURL,
		"virtualtabletop",
		"games",
	}
}

/*//////////////////////////////////////////////////////////////////////////////////////////////////////
      			                     	DATABASE FUNCTIONS
//////////////////////////////////////////////////////////////////////////////////////////////////////*/

/*
Init initialises the games DB
*/
func (db *GamesDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"gameid"},
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
