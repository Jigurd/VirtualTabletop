package img

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	Filename string   `bson:"filename"`
	ImgB64   string   `bson:"img"`
	ImgID    int      `bson:"_id"`
	Tags     []string `bson:"tags"`
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

	ImgDB.Init()
}

/*//////////////////////////////////////////////////////////////////////////////////////////////////////
      			                     	DATABASE FUNCTIONS
//////////////////////////////////////////////////////////////////////////////////////////////////////*/

//Init initializes the mongo storage.
func (db *ImgsDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{imgDbIDName},
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
func FetchImageByID(imageID int) (ImageData, bool) {
	session, err := mgo.Dial(ImgDB.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	img := ImageData{}
	allWasGood := true

	err = session.DB(ImgDB.DatabaseName).C(ImgDB.CollectionName).Find(bson.M{"_id": imageID}).One(&img)
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

	err = session.DB(ImgDB.DatabaseName).C(ImgDB.CollectionName).Find(bson.M{"imgid": timestamp}).One(&img)
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

/*////////////////////////////////////////////////////////////////////////////////////////////////////
										META FUNCTIONS
////////////////////////////////////////////////////////////////////////////////////////////////////*/

//PrepImage takes an image file and returns an ImageData struct. Always turns image to jpeg.
func PrepImage(file os.File, tags []string) (ImageData, error) {
	//decode the file
	m, _, err := image.Decode(&file)
	if err != nil {
		return ImageData{file.Name(), "", 0, tags}, err
	}
	//encode to string and convert to jpeg
	imgBase64Str, err2 := EncodeImage(m)
	if err2 != nil {
		return ImageData{file.Name(), "", 0, tags}, err2
	}

	imgData := ImageData{file.Name(), imgBase64Str, setTimestamp(), tags}

	//we're done!
	return imgData, nil
}

//DecodeImage takes an image stuct and returns the image.Image data.
func DecodeImage(imgData ImageData) (image.Image, error) {
	back := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgData.ImgB64))
	p, _, err := image.Decode(back)
	if err != nil {
		return p, err
	}
	return p, nil
}

//EncodeImage takes an image and encodes it into a string using base64
func EncodeImage(m image.Image) (string, error) {
	//buffer and encode to bytes
	buf2 := new(bytes.Buffer)
	err2 := jpeg.Encode(buf2, m, nil)
	if err2 != nil {
		return "", err2
	}
	//encode bytes to string
	imgBase64Str := base64.StdEncoding.EncodeToString(buf2.Bytes())
	return imgBase64Str, nil
}

//ImageToHTML takes an image struct and prints an image as it is directly in the html writer.
func ImageToHTML(w io.Writer, im ImageData) {
	img2html := "<html><body><img src=\"data:image/jpeg;base64," + im.ImgB64 + "\" /></body></html>"

	w.Write([]byte(fmt.Sprintf(img2html)))
}
