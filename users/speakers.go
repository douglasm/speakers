package users

import (
	// "speakers/crypto"
	"speakers/db"
	"speakers/types"

	"fmt"
	"gopkg.in/mgo.v2/bson"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func speakersHandler(w http.ResponseWriter, r *http.Request) {
	readSpeakers(nil, w, r)
}

func readSpeakers(theQuery bson.M, w http.ResponseWriter, r *http.Request) {
	type (
		detailsRec struct {
			Speakers template.HTML
			Topics   []types.CheckBoxRec
			Places   []types.CheckBoxRec
		}

		groupRec struct {
			Id     int      `bson:"_id"`
			Number int      `bson:"num"`
			Title  string   `bson:"title"`
			Tags   []string `bson:"tags"`
		}
	)
	var (
		// topics  []db.TopicRecord
		details detailsRec
		header  types.HeaderRecord
		// loggedIn types.Session
		err error
	)

	context := r.Context().Value("logged")
	if context != nil {
		header.Loggedin = context.(types.Session).LoggedIn
		header.God = context.(types.Session).IsGod
	}

	details.Speakers = speakerList(nil, nil)
	details.Topics = TopicList(nil)
	details.Places = PlaceList(nil)

	header.Title = fmt.Sprintf("%s: Speakers", initials)
	header.Name = siteName

	header.Scripts = append(header.Scripts, "/js/listtalks.js")

	data := struct {
		Header  types.HeaderRecord
		Details detailsRec
	}{
		header,
		details,
	}

	t, err := template.ParseFiles("templates/speakers.html", types.ViewHeader, types.ViewNavbar)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func reviseSpeakerPostHandler(w http.ResponseWriter, r *http.Request) {
	type (
		topicsRec struct {
			Topics string `schema:"topics"`
			Places string `schema:"places"`
		}
	)
	var (
		theTopics topicsRec
		tagList   []int
		placeList []string
	)

	err := r.ParseForm()
	err = decoder.Decode(&theTopics, r.PostForm)
	if err != nil {
		log.Println("Error: reviseSpeakerPostHandler decode", err)
	}

	parts := strings.Split(theTopics.Topics, ",")
	for _, item := range parts {
		if len(item) > 0 {
			num, err := strconv.Atoi(item)
			if err == nil {
				tagList = append(tagList, num)
			}
		}
	}

	parts = strings.Split(theTopics.Places, ",")
	for _, item := range parts {
		if len(item) > 0 {
			placeList = append(placeList, item)
		}
	}

	retStr := speakerList(tagList, placeList)

	w.Write([]byte(retStr))
}

func speakerList(tagList []int, placeList []string) template.HTML {
	var (
		theUser  User
		theQuery bson.M
		retStr   string
	)

	if len(placeList) > 0 {
		theQuery = bson.M{KFieldGroup: bson.M{"$ne": true}, FieldLocations: bson.M{"$in": placeList}}
	} else {
		theQuery = bson.M{KFieldGroup: bson.M{"$ne": true}}
	}
	theQuery[FieldTalks] = existsQuery

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	iter := userCollection.Find(theQuery).Sort(FieldName).Iter()
	for iter.Next(&theUser) {
		theTopics := []int{}
		count := 0
		gotTalk := false
		for _, talk := range theUser.Talks {
			gotOne := false
			if tagList != nil {
				for _, item := range tagList {
					if inIntSlice(item, talk.Tags) {
						gotOne = true
						break
					}
				}
			} else {
				gotOne = true
			}
			if gotOne {
				gotTalk = true
				count++
				for _, item := range talk.Tags {
					if !inIntSlice(item, theTopics) {
						theTopics = append(theTopics, item)
						break
					}
				}
			}
		}
		if gotTalk {
			retStr += "<tr>\r\n"
			retStr += fmt.Sprintf("<td><a href=\"/user/%d\">", theUser.Id)
			retStr += html.EscapeString(theUser.Name)
			retStr += "</a></td>\r\n"
			retStr += "<td>"
			retStr += fmt.Sprintf("%d", count)
			retStr += "</td>\r\n"
			retStr += "<td>" + db.GetTopicsString(theTopics) + "</td>\r\n"
			retStr += "<td>" + strings.Join(theUser.Locations, ", ") + "</td>\r\n"
			retStr += "</tr>\r\n"
		}
	}
	iter.Close()

	return template.HTML(retStr)
}
