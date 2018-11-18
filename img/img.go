package img

import (
	"bytes"
	"encoding/base64"
	"image"
	"log"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// _ "image/gif"
	// _ "image/png"
	"image/jpeg"
)

//le me trying to do image hosting stuff:

const dbURL = "mongodb://trap-lover:number1trap@ds063178.mlab.com:63178/tabletop_images"
const imgDbIDName = "imgId"

// ImgDB is the main DB connection for our image hosting
var ImgDB ImgsDB

// ImgsDB stores the details of the DB connection.
type ImgsDB struct {
	DatabaseURL    string
	DatabaseName   string
	CollectionName string
}

// ImageData contains the rgba values of an image
type ImageData struct {
	Filename  string   `bson:"filename"`
	ImgB64    string   `bson:"img"`
	TimeStamp int      `bson:"timestamp"`
	Tags      []string `bson:"tags"`
}

func setTimestamp() int {
	return int(time.Since(time.Date(1990, time.January, 1, 0, 0, 0, 1, time.UTC)).Nanoseconds())
}
func init() {
	ImgDB = ImgsDB{
		DatabaseURL:    dbURL,
		DatabaseName:   "tabletop_images",
		CollectionName: "alltraps",
	}

	ImgDB.Init(imgDbIDName)
}

/*
Init initializes the mongo storage.
*/
func (db *ImgsDB) Init(IDname string) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{IDname},
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

//FetchImageByID goes into the database and returns a struct of ImageData
func FetchImageByID(imageID string) (ImageData, bool) {
	session, err := mgo.Dial(ImgDB.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	img := ImageData{}
	allWasGood := true

	err = session.DB(ImgDB.DatabaseName).C(ImgDB.CollectionName).Find(bson.M{imgDbIDName: imageID}).One(&img)
	if err != nil {
		allWasGood = false
	}

	return img, allWasGood
}

//FetchImageByTime goes into the database and returns a struct of ImageData
func FetchImageByTime(timestamp int) (ImageData, bool) {
	session, err := mgo.Dial(ImgDB.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	img := ImageData{}
	allWasGood := true

	err = session.DB(ImgDB.DatabaseName).C(ImgDB.CollectionName).Find(bson.M{"timestamp": timestamp}).One(&img)
	if err != nil {
		allWasGood = false
	}

	return img, allWasGood
}

//PushImage uploads an image to the database
func PushImage(image ImageData) error {
	session, err := mgo.Dial(ImgDB.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(ImgDB.DatabaseName).C(ImgDB.CollectionName).Insert(image)

	if err != nil {
		log.Printf("error in Insert(): %v", err.Error())
		return err
	}

	return nil
}

//PrepImage takes an image file and returns an ImageData struct
func PrepImage(file os.File, tags []string) (ImageData, error) {
	//decode the file
	m, _, err := image.Decode(&file)
	if err != nil {
		return ImageData{file.Name(), "", 0, tags}, err
	}
	//buffer and encode to string
	buf2 := new(bytes.Buffer)
	err2 := jpeg.Encode(buf2, m, nil)
	if err2 != nil {
		return ImageData{file.Name(), "", 0, tags}, err2
	}
	imgBase64Str := base64.StdEncoding.EncodeToString(buf2.Bytes())

	imgData := ImageData{file.Name(), imgBase64Str, setTimestamp(), tags}

	//we're done!
	return imgData, nil
}
