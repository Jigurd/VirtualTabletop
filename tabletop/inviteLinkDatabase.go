package tabletop

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
InviteLink contains an invite link to a specific game
*/
type InviteLink struct {
	URL    string `json:"url"`
	GameId string `json:"gameid"`
}

/*
InviteLinksDB connects the db with invite link info
*/
type InviteLinksDB struct {
	DatabaseURL    string `json:"databaseurl"`
	DatabaseName   string `json:"databasename"`
	CollectionName string `json:"collectionmame"`
}

var InviteLinkDB InviteLinksDB // the db containing all invite links

func init() {
	InviteLinkDB = InviteLinksDB{
		dbURL,
		"virtualtabletop",
		"invitelinks",
	}
	InviteLinkDB.Init()
}

func NewInviteLink(g Game) InviteLink {
	link := InviteLink{
		"127.0.0.1:8080/i/" + bson.NewObjectId().Hex(),
		g.GameId,
	}
	return link
}

/*//////////////////////////////////////////////////////////////////////////////////////////////////////
      			                     	DATABASE FUNCTIONS
//////////////////////////////////////////////////////////////////////////////////////////////////////*/

/*
Init initialises the invite links DB
*/
func (db *InviteLinksDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"url"},
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
Add adds a new invite link to the database, returns true if the adding was successful
*/
func (db *InviteLinksDB) Add(l InviteLink) bool {
	if !db.Exists(l) {
		session, err := mgo.Dial(db.DatabaseURL)
		if err != nil {
			panic(err)
		}
		defer session.Close()

		err = session.DB(db.DatabaseName).C(db.CollectionName).Insert(l)
		if err != nil {
			fmt.Printf("Error inserting invite link into the DB: %s", err.Error())
			return false
		}

		return true
	}

	return false
}

/*
Exists returns true if an invite link already exists
*/
func (db *InviteLinksDB) Exists(l InviteLink) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	links := []InviteLink{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(nil).All(&links)

	for _, link := range links {
		if link.URL == l.URL {
			return true
		}
	}

	return false
}

/*
HasLink returns true if the game already has an invite link
*/
func (db *InviteLinksDB) HasLink(g Game) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	links := []InviteLink{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(nil).All(&links)

	for _, link := range links {
		if link.GameId == g.GameId {
			return true
		}
	}

	return false
}

/*
Get returns an invite link with the given url
*/
func (db *InviteLinksDB) Get(url string) (InviteLink, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	link := InviteLink{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"url": url}).One(&link)

	return link, err
}

/*
GetByGame returns an invite link with the given game
*/
func (db *InviteLinksDB) GetByGame(g Game) (InviteLink, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	link := InviteLink{}
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"gameid": g}).One(&link)

	return link, err
}

/*
GetAll returns a slice containing all existing invite links
*/
func (db *InviteLinksDB) GetAll() []InviteLink {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var all []InviteLink

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{}).All(&all)
	if err != nil {
		return []InviteLink{}
	}

	return all
}
