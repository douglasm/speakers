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
	"strings"
)

func groupsHandler(w http.ResponseWriter, r *http.Request) {
	readGroups(nil, w, r)
}

func readGroups(theQuery bson.M, w http.ResponseWriter, r *http.Request) {
	type (
		detailsRec struct {
			Speakers template.HTML
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

	details.Speakers = groupList(nil, nil)
	details.Places = GroupPlaceList(nil)

	header.Title = fmt.Sprintf("%s: Speakers", initials)
	header.Name = siteName

	header.Scripts = append(header.Scripts, "/js/listplaces.js")

	data := struct {
		Header  types.HeaderRecord
		Details detailsRec
	}{
		header,
		details,
	}

	t, err := template.ParseFiles("templates/groups.html", types.ViewHeader, types.ViewNavbar)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func reviseGroupPostHandler(w http.ResponseWriter, r *http.Request) {
	type (
		topicsRec struct {
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

	parts := strings.Split(theTopics.Places, ",")
	for _, item := range parts {
		if len(item) > 0 {
			placeList = append(placeList, item)
		}
	}

	retStr := groupList(tagList, placeList)

	w.Write([]byte(retStr))
}

func groupList(tagList []int, placeList []string) template.HTML {
	var (
		theUser  User
		theQuery bson.M
		retStr   string
	)

	if len(placeList) > 0 {
		theQuery = bson.M{KFieldGroup: true, FieldLocations: bson.M{"$in": placeList}}
	} else {
		theQuery = bson.M{KFieldGroup: true}
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	iter := userCollection.Find(theQuery).Sort(FieldName).Iter()
	for iter.Next(&theUser) {
		if len(theUser.Name) > 0 {
			retStr += "<tr>\r\n"
			retStr += fmt.Sprintf("<td><a href=\"/user/%d\">", theUser.Id)
			retStr += html.EscapeString(theUser.Name)
			retStr += "</a></td>\r\n"
			retStr += "<td>"
			retStr += theUser.GroupSize
			retStr += "</td>\r\n"
			retStr += "<td>" + strings.Join(theUser.Locations, ", ") + "</td>\r\n"
			retStr += "</tr>\r\n"
		}
	}
	iter.Close()

	return template.HTML(retStr)
}
