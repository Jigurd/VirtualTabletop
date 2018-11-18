package tabletop

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Character struct {
	//Character metadata
	Username      string `json:"username"` //Karl Gustav
	Charactername string `json:"charname"` //Grog Kragson
	System        string `json:"system"`   //DnD 5e
	CharId        int    `bson:"_id"`      //Autogeneres under lagring, ingen verdi før det
	Token         string `json:"token"`    //OrcViking.jpg

	//Statline and skills
	Stats  []nameDesc `json:"stats"`  //{Dex, 10}, {Str, 15}, etc
	Skills []nameDesc `json:"skills"` //{Acrobatics, 15}, {investigation, 10}, etc

	//Items and possessions
	Inventory []string   `json:"inventory"` //"1 Egg 0.5lbs" , "2 string 0.1lbs", etc
	Money     []nameDesc `json:"money"`     //{Gold, 15}, {Scheckels, 20}, etc
	Assets    []nameDesc `json:"assets"`    //{Yacht, "10m long, 3m wide, name: RosenSwinge"}, etc

	//Abilities on the field
	Abilities []nameDesc `json:"abilities"` //{Fireball, "8d6 fire dmg, 3rd level Spellslot"},{Lucky, "FUCK"},etc

	//User made macros & tags
	Macros []nameDesc `json:"macros"` //{Attack, "/r 1d20+DEX"}, etc
	Tags   []string   `json:"tags"`   //Fighter, human, idiot,etc
}

//nameDesc is a sub document which records data int the form of two strings, a Name and a Description
type nameDesc struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}

const (
	dbURL = "mongodb://admin:greatadminpassword69@ds151523.mlab.com:51523/virtualtabletop"
)

var (
	CharDB CharsDB // CharDB contains the users
)

/*
CharsDB connects to the database with User information
*/
type CharsDB struct {
	DatabaseURL    string `json:"databaseurl"`
	DatabaseName   string `json:"databasename"`
	CollectionName string `json:"collectionmame"`
}

func init() {
	CharDB = CharsDB{
		DatabaseURL:    dbURL,
		DatabaseName:   "virtualtabletop",
		CollectionName: "Characters",
	}

	CharDB.Init()
}

/*//////////////////////////////////////////////////////////////////////////////////////////////////////
      			                     	DATABASE FUNCTIONS
///////////////////////////////////////////////////////////////////////////////////////////////////////*

/*
Init initialises the users DB
*/
func (db *CharsDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"charId"},
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

//MakeChar functions sends a finished(!) char to the database, creates a unique ID for it and returns this Id as a
//int64 as well as an empty strng on a Success, or 0 and a errormessage on a failure.
func (db *CharsDB) MakeChar(char Character) (int, string) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var errmsg string
	charId := int(time.Since(time.Date(1990, time.January, 1, 0, 0, 0, 1, time.UTC)).Nanoseconds())
	char.CharId = charId

	err = session.DB(db.DatabaseName).C(db.CollectionName).Insert(char)
	if err != nil {
		errmsg = fmt.Sprintf("There was an error printing character %s to the database.", char.Charactername)
		return 0, errmsg
	}
	return charId, ""
}

//SearchTags function finds ALL characters which have the specified tags, and returns them, as well as an empty string,
//Alternatively, if there has been an error, it instead returns nil, as well as an error message.
func (db *CharsDB) SearchTags(tags []bson.M) ([]Character, string) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var foundChars []Character
	var errormsg string

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"$and": tags}).All(&foundChars)
	if err != nil {
		errormsg = fmt.Sprintf("There was an error finding a character with tags %v", tags)
		return nil, errormsg
	}
	return foundChars, ""
}

//FindChar Function finds a character with and Id equal to charId, Returns a character as well as an empty string
//on success, or an empty Character as well as an errormessage on a failure.
func (db *CharsDB) FindChar(charId int) (Character, string) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var foundChar Character
	var errormsg string

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"_id": charId}).One(&foundChar)
	if err != nil {
		errormsg = fmt.Sprintf("There was an error finding the character %s", charId)
		return Character{}, errormsg
	}
	return foundChar, ""
}

//UpdateCharString function finds a character with id equal to charId, finds a field which contains a string[] with name
//equal to field, and then updates that field with the new values given in values, before updating the database entry.
//returns an errormessage on a failure, or an empty string on a success
func (db *CharsDB) UpdateCharString(charId int, field string, values []string) string {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var errormsg string
	var char Character
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"_id": charId}).One(&char)
	if err != nil {
		errormsg = fmt.Sprintf("There was an error finding a character with %s", charId)
		return errormsg
	}
	switch field {
	case "inventory":
		for _, value := range values {
			char.Inventory = append(char.Inventory, value)
		}
		break
	case "tags":
		for _, value := range values {
			char.Tags = append(char.Tags, value)
		}
		break

	default:
		errormsg = fmt.Sprintf("The field %s does not match any fields in the character sheet", field)
		return errormsg
	}

	err = session.DB(db.DatabaseName).C(db.CollectionName).Update(bson.M{"_id": charId}, char)
	if err != nil {
		errormsg = fmt.Sprintf("There was a problem updating character %s's field %s with the value %v", charId, field, values)
		return errormsg
	}
	return ""
}

//UpdateChar_nameDesc function does the same as UpdateCharString, and returns the same errors, but takes fields and values
//containing nameDesc instead of strings
func (db *CharsDB) UpdateChar_nameDesc(charId int, field string, values []nameDesc) string {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var errormsg string
	var char Character
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"_id": charId}).One(&char)
	if err != nil {
		errormsg = fmt.Sprintf("There was an error finding a character with %s", charId)
		return errormsg
	}
	switch field {
	case "stats":
		for _, value := range values {
			char.Stats = append(char.Stats, value)
		}
		break
	case "skills":
		for _, value := range values {
			char.Skills = append(char.Skills, value)
		}
		break
	case "money":
		for _, value := range values {
			char.Money = append(char.Money, value)
		}
		break
	case "assets":
		for _, value := range values {
			char.Assets = append(char.Assets, value)
		}
		break
	case "abilities":
		for _, value := range values {
			char.Abilities = append(char.Abilities, value)
		}
		break
	case "macros":
		for _, value := range values {
			char.Macros = append(char.Macros, value)
		}
		break
	default:
		errormsg = fmt.Sprintf("The field %s does not match any fields in the character sheet", field)
		return errormsg
	}

	err = session.DB(db.DatabaseName).C(db.CollectionName).Update(bson.M{"_id": charId}, char)
	if err != nil {
		errormsg = fmt.Sprintf("There was a problem updating character %s's field %s with the value %v", charId, field, values)
		return errormsg
	}
	return ""

}

//DeleteChar function deletes a character with id charId. On success it returns and empty string on a success or
//an errormessage on a failure
func (db *CharsDB) DeleteChar(charId int) string {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.CollectionName).Remove(bson.M{"_id": charId})
	if err != nil {
		errmsg := fmt.Sprintf("There was an error deleting the Character with id %v", charId)
		return errmsg
	}
	return ""
}

//GetString function which finds a specific STRING-field on the character and returns either an empty string and the fields value if
//everything is ok, or an error message and an empty string if not
func (db *CharsDB) GetString(charId int, field string) (string, string) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var parsedChar Character
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"_id": charId}).Select(bson.M{"_id": 0, field: 1}).One(&parsedChar)
	if err != nil {
		errmsg := fmt.Sprintf("Something went wrong when fetching a character with id %v. Field %s not returned", charId, field)
		return errmsg, ""
	} else {
		switch field {
		case ("username"):
			return "", parsedChar.Username
		case ("charname"):
			return "", parsedChar.Charactername
		case ("system"):
			return "", parsedChar.System
		default:
			errmsg := fmt.Sprintf("The field %s does not exist in the database for character %v.", field, charId)
			return errmsg, ""
		}
	}
}

/*////////////////////////////////////////////////////////////////////////////////////////////////////
										META FUNCTIONS
////////////////////////////////////////////////////////////////////////////////////////////////////*/

//CreateChar Function makes a character using a characterName, a userName and a system. Used for Primary Character creation.
//Returns the characters new ID as well as an empty string upon success, or an empty string and a errormessage upon failure.
func CreateChar(charName string, userName string, system string) (int, string) {
	char := Character{userName, charName, system, 0, "", nil, nil, nil, nil, nil, nil, nil, nil}
	charId, err := CharDB.MakeChar(char)
	if err != "" {
		return 0, err
	} else {
		return charId, ""
	}
}

//TagRequest function takes in an array of tags and sends these to CHarDB.SearchTags. It then returns an array of all
//Characters that match the request, as well as an empty string. Alternatively it returns an empty character array and an
//errormessage if the operation did not succeed.
func TagRequest(tags []string) ([]Character, string) {
	var queries []bson.M

	for _, tag := range tags {
		query := bson.M{"tags": tag}
		queries = append(queries, query)
	}
	chars, err := CharDB.SearchTags(queries)
	if err != "" {
		return []Character{}, err
	}
	return chars, err
}
