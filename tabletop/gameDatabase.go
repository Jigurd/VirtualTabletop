package tabletop

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
Game is a game TRPG that has been set up by a GM
and contains information about the game's name, system,
players, description, etc
TODO: Add a public setting i.e. the owner can hide the game from the games listing
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

	GameDB.Init()
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

/*
Add adds a new game to the database, returns true if the adding was successful
*/
func (db *GamesDB) Add(g Game) bool {
	if !db.Exists(g) {
		session, err := mgo.Dial(db.DatabaseURL)
		if err != nil {
			panic(err)
		}
		defer session.Close()

		err = session.DB(db.DatabaseName).C(db.CollectionName).Insert(g)
		if err != nil {
			fmt.Printf("Error inserting game into the DB: %s", err.Error())
			return false
		}

		return true
	}

	return false
}

/*
Exists returns true if a game already exists
*/
func (db *GamesDB) Exists(g Game) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	games := []Game{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(nil).All(&games)

	for _, game := range games {
		if g.GameId == game.GameId {
			return true
		}
	}

	return false
}

/*
Get returns a game with the given game id
*/
func (db *GamesDB) Get(id string) (Game, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	game := Game{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"gameid": id}).One(&game)

	return game, err
}

/*
GetAll returns a slice containing all existing games.
*/
func (db *GamesDB) GetAll() []Game {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var all []Game

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{}).All(&all)
	if err != nil {
		return []Game{}
	}

	return all
}
