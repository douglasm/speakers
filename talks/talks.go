package talks

import (
	"speakers/db"
	"speakers/types"
	"speakers/users"
	// "speakers/utils"

	// "utilscrypto"

	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func talksHandler(w http.ResponseWriter, r *http.Request) {
	readTalks(nil, w, r)
}

func readTalks(theQuery bson.M, w http.ResponseWriter, r *http.Request) {
	type (
		detailsRec struct {
			Talks  template.HTML
			Topics []types.CheckBoxRec
			Places []types.CheckBoxRec
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
		// loggedIn = context.(types.Session)
		// header.Loggedin = loggedIn.LoggedIn
		header.Loggedin = context.(types.Session).LoggedIn
	}

	details.Talks = talkList(nil, nil)
	details.Topics = topicList(nil)
	details.Places = placeList(nil)

	header.Title = fmt.Sprintf("%s: Talks", initials)
	header.Name = siteName

	header.Scripts = append(header.Scripts, "/js/listtalks.js")

	data := struct {
		Header  types.HeaderRecord
		Details detailsRec
	}{
		header,
		details,
	}

	t, err := template.ParseFiles("templates/talks.html", types.ViewHeader, types.ViewNavbar)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func talkAddHandlerGet(w http.ResponseWriter, r *http.Request) {
	var (
		editRec TalkEditRec
		theUser db.SessionRecord
	)

	if !users.UserLoggedIn(r, &theUser) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	editRec.buildTopicChecks()
	outputTalkForm(editRec, "", w)
}

func talkAddHandlerPost(w http.ResponseWriter, r *http.Request) {
	var (
		editRec TalkEditRec
		theUser db.SessionRecord
		err     error
	)

	if !users.UserLoggedIn(r, &theUser) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	err = decoder.Decode(&editRec, r.PostForm)
	if err != nil {
		log.Println(err)
	}

	err = editRec.check()
	if err == nil {
		editRec.Save(theUser.UserNumber)
		go UpdateTopics()
		http.Redirect(w, r, "/details", http.StatusSeeOther)
		return
	}

	editRec.buildTopicChecks()
	outputTalkForm(editRec, err.Error(), w)
}

func talkEditHandlerGet(w http.ResponseWriter, r *http.Request) {
	var (
		editRec    TalkEditRec
		theSession db.SessionRecord
		theUser    users.User
	)

	if !users.UserLoggedIn(r, &theSession) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	params := mux.Vars(r)
	number := getParamVal(params, "number", 0)

	if number == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	err := userCollection.Find(bson.M{"talks.num": number}).One(&theUser)
	if err != nil {
		log.Println("Error: talk edit read", err)
	}

	if theUser.Id != theSession.UserNumber {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	gotOne := false
	for _, item := range theUser.Talks {
		if item.Number == number {
			editRec.Title = item.Title
			editRec.Description = item.Description
			editRec.Num = number
			editRec.Tags = item.Tags
			editRec.buildTopicChecks()
			gotOne = true
			break
		}
	}

	if gotOne {
		outputTalkForm(editRec, "", w)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func talkEditHandlerPost(w http.ResponseWriter, r *http.Request) {
	var (
		editRec    TalkEditRec
		theSession db.SessionRecord
		theUser    users.User
		err        error
	)

	if !users.UserLoggedIn(r, &theSession) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	params := mux.Vars(r)
	number := getParamVal(params, "number", 0)

	if number == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = decoder.Decode(&editRec, r.PostForm)
	if err != nil {
		log.Println("Error: talkedit decode", err)
	}

	//  Check the entered details are correct
	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	err = editRec.check()
	if err == nil {
		err = userCollection.FindId(theSession.UserNumber).One(&theUser)
		gotOne := false
		for i, item := range theUser.Talks {
			if item.Number == number {
				gotOne = true
				theUser.Talks[i].Title = editRec.Title
				theUser.Talks[i].Description = editRec.Description
				theUser.Talks[i].Tags = editRec.Tags
				break
			}
		}
		if gotOne {
			err = userCollection.UpdateId(theUser.Id, bson.M{"$set": bson.M{users.KFieldTalks: theUser.Talks}})
		}
		// if err != nil {
		// 	log.Println("Error: Update talks", err)
		// 	fmt.Println("Error: Update talks", err)
		// }
		// editRec.Save(theUser.Id)
		go UpdateTopics()
		go users.UpdatePlaces()
		http.Redirect(w, r, "/details", http.StatusSeeOther)
		return
	}

	editRec.buildTopicChecks()
	outputTalkForm(editRec, err.Error(), w)
}

func talkDeleteHandlerGet(w http.ResponseWriter, r *http.Request) {
	var (
		deleteRec  TalkDeleteRec
		theSession db.SessionRecord
		theUser    users.User
	)

	if !users.UserLoggedIn(r, &theSession) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	params := mux.Vars(r)
	number := getParamVal(params, "number", 0)

	if number == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	err := userCollection.Find(bson.M{"talks.num": number}).One(&theUser)
	if err != nil {
		log.Println("Error: talk edit read", err)
	}

	if theUser.Id != theSession.UserNumber {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	gotOne := false
	for _, item := range theUser.Talks {
		if item.Number == number {
			deleteRec.Title = item.Title
			deleteRec.Description = item.Description
			deleteRec.Num = number
			deleteRec.Topics = db.GetTopicsString(item.Tags)
			gotOne = true
			break
		}
	}

	if gotOne {
		outputDeleteForm(deleteRec, "", w)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func talkDeleteHandlerPost(w http.ResponseWriter, r *http.Request) {
	var (
		deleteRec  TalkDeleteRec
		theSession db.SessionRecord
		theUser    users.User
		talkList   []users.TalkRecord
		action     = bson.M{}
		err        error
	)

	if !users.UserLoggedIn(r, &theSession) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Println("Error: talkDeleteHandlerPost parse:", err)
	}

	params := mux.Vars(r)
	number := getParamVal(params, "number", 0)
	if number == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = decoder.Decode(&deleteRec, r.PostForm)
	if err != nil {
		log.Println("Error: talkDeleteHandlerPost decode:", err)
	}

	if deleteRec.Button != "Delete" {
		http.Redirect(w, r, "/details", http.StatusSeeOther)
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	err = userCollection.Find(bson.M{"talks.num": number}).One(&theUser)
	if err != nil {
		log.Println("Error: talk edit read", err)
		http.Redirect(w, r, "/details", http.StatusSeeOther)
		return
	}

	if theUser.Id != theSession.UserNumber {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	gotOne := false
	for _, item := range theUser.Talks {
		if item.Number == number {
			gotOne = true
		} else {
			talkList = append(talkList, item)
		}
	}
	if gotOne {
		if len(talkList) == 0 {
			action = bson.M{"$unset": bson.M{users.FieldTalks: 1}}
		} else {
			action = bson.M{"$set": bson.M{users.FieldTalks: talkList}}
		}
		err = userCollection.UpdateId(theUser.Id, action)
	}

	http.Redirect(w, r, "/details", http.StatusSeeOther)
}

func talkHandler(w http.ResponseWriter, r *http.Request) {
	type (
		tvr struct {
			Title       template.HTML
			Speaker     template.HTML
			Description template.HTML
			Topics      string
		}
	)
	var (
		theUser users.User
		header  types.HeaderRecord
		details tvr
	)

	params := mux.Vars(r)
	number := getParamVal(params, "number", 0)

	if number == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	err := userCollection.Find(bson.M{"talks.num": number}).One(&theUser)
	if err != nil {
		log.Println("Error: talk edit read", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	gotOne := false
	for _, item := range theUser.Talks {
		if item.Number == number {
			details.Title = template.HTML(html.EscapeString(item.Title))
			details.Speaker = template.HTML(html.EscapeString(theUser.Name))
			theDesc := strings.Replace(item.Description, "\r\n", "\r", -1)
			theDesc = strings.Replace(theDesc, "\r\r", "\r", -1)
			theDesc = strings.Replace(theDesc, "\r\r", "\r", -1)
			theDesc = strings.Replace(theDesc, "\r", "<br />", -1)
			details.Description = template.HTML(html.EscapeString(theDesc))
			details.Topics = db.GetTopicsString(item.Tags)
			gotOne = true
			break
		}
	}

	if !gotOne {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	header.Title = fmt.Sprintf("%s: View talk", initials)

	context := r.Context().Value("logged")
	if context != nil {
		header.Loggedin = context.(types.Session).LoggedIn
	}
	header.Name = siteName

	data := struct {
		Header  types.HeaderRecord
		Details tvr
	}{
		header,
		details,
	}

	t, err := template.ParseFiles(templateTalkView, types.ViewHeader, types.ViewNavbar)
	if err != nil {
		log.Println("error: outputTalkViewForm ParseFiles", err)
	}
	t.Execute(w, data)
}

func reviseTalkHandlerPost(w http.ResponseWriter, r *http.Request) {
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
		log.Println("Error: reviseTalkHandlerPost decode", err)
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

	retStr := talkList(tagList, placeList)

	w.Write([]byte(retStr))
}

func outputTalkForm(details TalkEditRec, errorMessage string, w http.ResponseWriter) {
	var (
		theLink string
		action  string
		header  types.HeaderRecord
	)

	if details.Num == 0 {
		header.Title = fmt.Sprintf("%s: Add a talk", initials)
	} else {
		header.Title = fmt.Sprintf("%s: Edit talk", initials)
	}
	header.Loggedin = true
	header.Name = siteName

	if details.Num == 0 {
		theLink = "talkadd"
		action = "Add new talk"
	} else {
		theLink = fmt.Sprintf("talkedit/%d", details.Num)
		action = "Edit talk"
	}

	data := struct {
		Header       types.HeaderRecord
		Details      TalkEditRec
		Link         string
		Action       string
		ErrorMessage string
	}{
		header,
		details,
		theLink,
		action,
		errorMessage,
	}

	t, err := template.ParseFiles(templateTalkEdit, types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println("error: outputTalkForm ParseFiles", err)
	}
	t.Execute(w, data)
}

func outputDeleteForm(details TalkDeleteRec, errorMessage string, w http.ResponseWriter) {
	var (
		header types.HeaderRecord
	)

	header.Title = fmt.Sprintf("%s: Delete talk", initials)
	header.Loggedin = true
	header.Name = siteName

	details.Checkfield = "Hello farty"
	data := struct {
		Header  types.HeaderRecord
		Details TalkDeleteRec
	}{
		header,
		details,
	}

	t, err := template.ParseFiles(templateTalkDelete, types.ViewHeader, types.ViewNavbar)
	if err != nil {
		log.Println("Error: outputDeleteForm ParseFiles", err)
	}
	t.Execute(w, data)
}

func talkList(tagList []int, placeList []string) template.HTML {
	var (
		talk      talkListRec
		topicRecs []db.TopicRecord
		allTopics = map[int]string{}
		theMatch  bson.M
		retStr    string
		err       error
	)

	if len(placeList) > 0 {
		theMatch = bson.M{users.KFieldGroup: bson.M{"$ne": true}, users.FieldLocations: bson.M{"$in": placeList}}
	} else {
		theMatch = bson.M{users.KFieldGroup: bson.M{"$ne": true}}
	}
	match := bson.M{"$match": theMatch}

	sort := bson.M{"$sort": bson.M{"usernum": 1}}

	theProject := bson.M{}
	theProject["talks"] = 1
	theProject["name"] = 1
	project := bson.M{"$project": theProject}
	unwind := bson.M{"$unwind": "$talks"}
	theProject = bson.M{}
	theProject["name"] = 1
	theProject["usernum"] = "$_id"
	theProject["num"] = "$talks.num"
	theProject["title"] = "$talks.title"
	theProject["tags"] = "$talks.tags"
	theProject["topics"] = "$talks.topics"

	project2 := bson.M{"$project": theProject}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	topicCollection := session.DB(db.MainDB).C(db.CollectionTopics)

	err = topicCollection.Find(nil).Sort(db.KFieldName).All(&topicRecs)
	if err != nil {
		log.Println("Error: topic read talks list", err)
	}
	for _, item := range topicRecs {
		allTopics[item.Id] = item.Name
	}

	pipeline := []bson.M{match, project, unwind, project2, sort}
	iter := userCollection.Pipe(pipeline).Iter()
	for iter.Next(&talk) {
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
			retStr += "<tr>\r\n"
			retStr += fmt.Sprintf("<td><a href=\"/talk/%d\">", talk.Number)
			retStr += html.EscapeString(talk.Title)
			retStr += "</a></td>\r\n"
			retStr += fmt.Sprintf("<td><a href=\"/user/%d\">", talk.UserNum)
			retStr += html.EscapeString(talk.Name)
			retStr += "</a></td>\r\n"
			retStr += "<td>" + db.GetTopicsString(talk.Tags) + "</td>\r\n"
			retStr += "</tr>\r\n"
		}
	}
	iter.Close()

	return template.HTML(retStr)
}
