package users

import (
	"speakers/crypto"
	"speakers/db"
	"speakers/types"

	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func usersHandler(w http.ResponseWriter, r *http.Request) {
	readUsers(nil, w, r)
}

func readUsers(theQuery bson.M, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("read users"))
}

func setPlacePostHandler(w http.ResponseWriter, r *http.Request) {
	type (
		placeRec struct {
			Place    string `schema:"place"`
			Selected bool   `schema:"sel"`
		}
	)
	var (
		details placeRec
		userRec db.SessionRecord
		theUser User
		action  string
		// 	errorMessage string
		err error
	)

	if !UserLoggedIn(r, &userRec) {
		w.Write([]byte("-"))
		return
	}

	err = r.ParseForm()
	err = decoder.Decode(&details, r.PostForm)
	if err != nil {
		log.Println(err)
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	err = userCollection.FindId(userRec.UserNumber).One(&theUser)
	if err != nil {
		log.Println(err)
		w.Write([]byte(""))
		return
	}

	if details.Selected {
		action = "$addToSet"
	} else {
		action = "$pull"

	}
	err = userCollection.UpdateId(userRec.UserNumber, bson.M{action: bson.M{"loc": details.Place}})

	if err != nil {
		log.Println(err)
	}

	w.Write([]byte("+"))
}

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	type (
		talkRec struct {
			Number int
			Title  template.HTML
			Topics string
		}

		detailsRec struct {
			Name        string
			HasPlaces   bool
			HasTalks    bool
			Group       bool
			Description template.HTML
			Size        string
			When        string
			Talks       []talkRec
			Places      []string
		}
	)
	var (
		header     types.HeaderRecord
		sessionRec db.SessionRecord
		theUser    User
		details    detailsRec
		theTopic   db.TopicRecord
		allTopics  = map[int]string{}
		err        error
	)

	header.Loggedin = UserLoggedIn(r, &sessionRec)
	if !header.Loggedin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	topicsCollection := session.DB(db.MainDB).C(db.CollectionTopics)
	iter := topicsCollection.Find(nil).Iter()

	for iter.Next(&theTopic) {
		allTopics[theTopic.Id] = theTopic.Name
	}
	iter.Close()

	err = userCollection.FindId(sessionRec.UserNumber).One(&theUser)
	if err != nil {
		log.Println("Error: details read user", err)
	}

	details.Group = theUser.Group

	details.Name = theUser.Name

	tempStr := html.EscapeString(theUser.Description)
	tempStr = strings.Replace(tempStr, "\r\n", "\r", -1)
	tempStr = strings.Replace(tempStr, "\r\r", "\r", -1)
	tempStr = strings.Replace(tempStr, "\r\r", "\r", -1)
	tempStr = strings.Replace(tempStr, "\r", "<br />", -1)
	details.Description = template.HTML(tempStr)
	details.Size = theUser.GroupSize
	details.When = theUser.GroupWhen

	header.Title = fmt.Sprintf("%s: Your details", initials)
	header.Name = siteName

	if len(theUser.Talks) > 0 {
		details.HasTalks = true
		for _, item := range theUser.Talks {
			topicStr := ""
			for i, num := range item.Tags {
				if i > 0 {
					topicStr += ", "
				}
				topicStr += db.TopicsMap[num]
			}
			tr := talkRec{Number: item.Number, Title: template.HTML(html.EscapeString(item.Title)), Topics: topicStr}
			details.Talks = append(details.Talks, tr)
		}
	}

	details.Places = theUser.Locations

	header.Scripts = append(header.Scripts, "/js/talks.js")
	data := struct {
		Header  types.HeaderRecord
		Details detailsRec
	}{
		header,
		details,
	}

	t, err := template.ParseFiles("templates/details.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	type (
		talkRec struct {
			Number int
			Title  template.HTML
			Link   string
			Topics string
		}

		detailsRec struct {
			Name        string
			UserNumber  int
			HasPlaces   bool
			HasTalks    bool
			Group       bool
			Description template.HTML
			Size        string
			When        string
			Talks       []talkRec
			Places      []string
		}
	)
	var (
		header  types.HeaderRecord
		theUser User
		userNum int
		details detailsRec
		err     error
	)

	params := mux.Vars(r)
	number := GetParamVal(params, "number", 0)

	context := r.Context().Value("logged")
	if context != nil {
		header.Loggedin = context.(types.Session).LoggedIn
		header.God = context.(types.Session).IsGod
		userNum = context.(types.Session).UserNumber
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	// topicsCollection := session.DB(db.MainDB).C(db.CollectionTopics)

	err = userCollection.FindId(number).One(&theUser)
	if err != nil {
		log.Println("Error: user read user", err)
	}

	details.Group = theUser.Group

	details.Name = theUser.Name
	details.UserNumber = number

	tempStr := html.EscapeString(theUser.Description)
	tempStr = strings.Replace(tempStr, "\r\n", "\r", -1)
	tempStr = strings.Replace(tempStr, "\r\r", "\r", -1)
	tempStr = strings.Replace(tempStr, "\r\r", "\r", -1)
	tempStr = strings.Replace(tempStr, "\r", "<br />", -1)
	details.Description = template.HTML(tempStr)
	details.Size = theUser.GroupSize
	details.When = theUser.GroupWhen

	header.Title = fmt.Sprintf("%s: %s", initials, theUser.Name)
	header.Name = siteName

	if len(theUser.Talks) > 0 {
		details.HasTalks = true
		for _, item := range theUser.Talks {
			topicStr := ""
			for i, num := range item.Tags {
				if i > 0 {
					topicStr += ", "
				}
				topicStr += db.TopicsMap[num]
			}
			tr := talkRec{Number: item.Number, Title: template.HTML(html.EscapeString(item.Title)), Topics: topicStr}
			if userNum == theUser.Id {
				tr.Link = "talkedit"
			} else {
				tr.Link = "talk"
			}

			details.Talks = append(details.Talks, tr)
		}
	}

	details.Places = theUser.Locations

	header.Scripts = append(header.Scripts, "/js/talks.js")
	data := struct {
		Header  types.HeaderRecord
		Details detailsRec
	}{
		header,
		details,
	}

	t, err := template.ParseFiles("templates/user.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func userEditGetHandler(w http.ResponseWriter, r *http.Request) {
	var (
		header  types.HeaderRecord
		userRec db.SessionRecord
		theUser User
		// places     []db.PlaceRecord
		userEditRec userEditRecord
		// details    detailsRec
		// selected   bool
		err error
	)

	header.Loggedin = UserLoggedIn(r, &userRec)
	if !header.Loggedin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	err = userCollection.FindId(userRec.UserNumber).One(&theUser)
	if err != nil {
		log.Println("Error: details read user", err)
	}

	userEditRec.setup(theUser)
	outputGroupEditForm(userEditRec, "", header.Loggedin, w)
}

func userEditPostHandler(w http.ResponseWriter, r *http.Request) {
	var (
		header  types.HeaderRecord
		userRec db.SessionRecord
		// theUser     User
		userEditRec userEditRecord
		err         error
	)

	header.Loggedin = UserLoggedIn(r, &userRec)
	if !header.Loggedin {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	err = decoder.Decode(&userEditRec, r.PostForm)
	if err != nil {
		log.Println(err)
	}

	err = userEditRec.check()
	if err != nil {
		userEditRec.PlaceCheck = getPlaces(userEditRec.Places)
		outputGroupEditForm(userEditRec, err.Error(), header.Loggedin, w)
		return
	}

	userEditRec.save()
	UpdatePlaces()
	http.Redirect(w, r, "/details", http.StatusSeeOther)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	var (
		theUser      User
		details      messageRec
		header       types.HeaderRecord
		errorMessage string
	)
	params := mux.Vars(r)

	number := GetParamVal(params, "number", 0)

	context := r.Context().Value("logged")
	if context == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	senderNum := context.(types.Session).UserNumber
	details.Num = number

	session := db.MongoSession.Copy()
	defer session.Close()

	usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	err := usersCollection.FindId(senderNum).One(&theUser)
	if err != nil {
		log.Println("Error: messageHandler read sender", err)
	}

	details.Site = siteName
	details.Sender = theUser.Name
	details.SendNum = senderNum

	err = usersCollection.FindId(details.Num).One(&theUser)
	recpEmail := crypto.Decrypt(theUser.EMail)

	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		err = r.ParseForm()
		err = decoder.Decode(&details, r.PostForm)
		if err != nil {
			log.Println(err)
		}
		if len(details.Message) > 10 {
			sendMessage(recpEmail, details.Message, details.Sender)
			theStr := fmt.Sprintf("/user/%d", details.Num)
			http.Redirect(w, r, theStr, http.StatusSeeOther)
			return
		}
		errorMessage = ErrorMessage.Error()
	}

	header.Loggedin = context.(types.Session).LoggedIn
	header.God = context.(types.Session).IsGod
	header.Title = fmt.Sprintf("%s: Send message", initials)

	details.Name = theUser.Name
	details.Site = siteName

	data := struct {
		Header       types.HeaderRecord
		Details      messageRec
		ErrorMessage string
	}{
		header,
		details,
		errorMessage,
	}

	t, err := template.ParseFiles("templates/message.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func passwordHandler(w http.ResponseWriter, r *http.Request) {
	var (
		theUser      User
		details      passwordRec
		header       types.HeaderRecord
		errorMessage string
		err          error
	)

	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		err = r.ParseForm()
		if err != nil {
			log.Println("Error: passwordHandler ParseForm", err.Error())
		}
		err = decoder.Decode(&details, r.PostForm)
		if err != nil {
			log.Println("Error: passwordHandler Decode", err.Error())
		}

		if len(details.Email) > 0 {
			if validateEmail(details.Email) {
				session := db.MongoSession.Copy()
				defer session.Close()

				usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)

				theHash, err := crypto.GetHash(strings.ToLower(details.Email), emSalt)

				err = usersCollection.Find(bson.M{FieldEMHash: theHash}).One(&theUser)
				if err == nil {
					theCode := getRandomCode()
					usersCollection.UpdateId(theUser.Id, bson.M{"$set": bson.M{FieldResetCode: theCode}})
					sendReset(details.Email, theCode)

				}
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			} else {
				errorMessage = errorNotMail.Error()
			}
		} else {
			errorMessage = errorEMail.Error()
		}
	}

	header.Title = fmt.Sprintf("%s: Reset password", initials)
	details.Checkfield = crypto.MakeNonce(resetFormID, "noone")

	data := struct {
		Header       types.HeaderRecord
		Details      passwordRec
		ErrorMessage string
	}{
		header,
		details,
		errorMessage,
	}

	t, err := template.ParseFiles("templates/resetpass.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func newPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var (
		// theUser      User
		details      passResetRec
		header       types.HeaderRecord
		theUser      User
		errorMessage string
		err          error
	)

	templateName := "templates/newpass.html"
	params := mux.Vars(r)

	code, ok := params["code"]
	if ok {
		details.Code = code
		templateName = "templates/newpasscode.html"
	}

	switch r.Method {
	case http.MethodGet:

	case http.MethodPost:
		err = r.ParseForm()
		if err != nil {
			log.Println("Error: newPasswordHandler ParseForm", err.Error())
		}
		err = decoder.Decode(&details, r.PostForm)
		if err != nil {
			log.Println("Error: newPasswordHandler Decode", err.Error())
		}

		err = details.check(&theUser)
		if err == nil {
			startSession(w, theUser, details.Remember)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			errorMessage = err.Error()
		}
	}

	header.Title = fmt.Sprintf("%s: Reset password", initials)
	header.Scripts = append(header.Scripts, "/js/passwordtoggle.js")

	details.Checkfield = crypto.MakeNonce(resetFormID, "noone")

	data := struct {
		Header       types.HeaderRecord
		Details      passResetRec
		ErrorMessage string
	}{
		header,
		details,
		errorMessage,
	}

	t, err := template.ParseFiles(templateName, types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func outputGroupEditForm(details userEditRecord, errorMessage string, loggedIn bool, w http.ResponseWriter) {
	var (
		header       types.HeaderRecord
		templateName string
	)

	header.Loggedin = loggedIn

	header.Title = fmt.Sprintf("%s: Edit Your details", initials)
	header.Name = siteName

	data := struct {
		Header       types.HeaderRecord
		Details      userEditRecord
		ErrorMessage string
	}{
		header,
		details,
		errorMessage,
	}

	if details.Group {
		templateName = "templates/usergroup.html"
	} else {
		templateName = "templates/userspeaker.html"
	}

	t, err := template.ParseFiles(templateName, types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func UserLoggedIn(r *http.Request, theUser *db.SessionRecord) bool {
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}

	splits := strings.Split(c.Value, "|")
	if len(splits) != 2 {
		log.Println("bad format session cookie", c.Value)
		return false
	}
	if splits[1] != makeSecureVal(splits[0]) {
		return false
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	sessionCollection := session.DB(db.MainDB).C(db.CollectionSessions)
	err = sessionCollection.FindId(splits[0]).One(theUser)

	if err != nil {
		log.Println("Could not find session cookie")
		return false
	}

	return true
}

func CountSpeakerTalks(mainPageRec *types.MainPageRecord) {
	var (
		users    []User
		speakers int
		talks    int
		err      error
	)

	mainPageRec.Speakers = 0
	mainPageRec.Talks = 0
	mainPageRec.Locations = 0
	mainPageRec.Groups = 0
	mainPageRec.Places = 0

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	placeCollection := session.DB(db.MainDB).C(db.CollectionPlaces)

	// get the number of speakers and talks
	err = userCollection.Find(bson.M{FieldTalks: existsQuery, FieldGroup: false}).All(&users)
	if err != nil {
		log.Println(err)
		return
	}

	for _, item := range users {
		speakers++
		talks += len(item.Talks)
	}
	mainPageRec.Speakers = speakers
	mainPageRec.Talks = talks

	// get the number of groups and places
	// err = userCollection.Find(bson.M{FieldGroup: existsQuery, FieldLocations: existsQuery}).All(&users)
	num, err := userCollection.Find(bson.M{FieldGroup: true, FieldLocations: existsQuery}).Count()
	if err != nil {
		log.Println(err)
		return
	}

	mainPageRec.Groups = num

	mainPageRec.Locations, err = placeCollection.Find(bson.M{FieldSpeaker: bson.M{"$gt": 0}}).Count()
	mainPageRec.Places, err = placeCollection.Find(bson.M{FieldGroup: bson.M{"$gt": 0}}).Count()
}

func UpdatePlaces() {
	type (
		groupRec struct {
			Place string `bson:"_id"`
			Count int    `bson:"count"`
		}
		speakerGroupRec struct {
			Speakers int
			Groups   int
		}
	)
	var (
		res      groupRec
		speakers = map[string]int{}
		groups   = map[string]int{}
	)

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	placeCollection := session.DB(db.MainDB).C(db.CollectionPlaces)

	// get the places speakers say they can speak at
	match := bson.M{"$match": bson.M{FieldTalks: existsQuery, KFieldGroup: false}}

	theProject1 := bson.M{}
	theProject1[FieldLocations] = 1
	theProject1[db.KFieldId] = 0
	project1 := bson.M{"$project": theProject1}

	unwind1 := bson.M{"$unwind": "$" + FieldLocations}

	theGroup := bson.M{db.KFieldId: "$" + FieldLocations, "count": countAgg}
	group1 := bson.M{"$group": theGroup}

	pipeline := []bson.M{match, project1, unwind1, group1}

	iter := userCollection.Pipe(pipeline).Iter()
	err := iter.Err()
	if err != nil {
		log.Println(err)
	}
	for iter.Next(&res) {
		speakers[res.Place] = res.Count
	}
	iter.Close()

	// get the places groups say are at
	match = bson.M{"$match": bson.M{FieldGroup: true}}

	theProject1 = bson.M{}
	theProject1[FieldLocations] = 1
	theProject1[db.KFieldId] = 0
	project1 = bson.M{"$project": theProject1}

	unwind1 = bson.M{"$unwind": "$" + FieldLocations}

	theGroup = bson.M{db.KFieldId: "$" + FieldLocations, "count": countAgg}
	group1 = bson.M{"$group": theGroup}

	pipeline = []bson.M{match, project1, unwind1, group1}
	iter = userCollection.Pipe(pipeline).Iter()
	err = iter.Err()
	if err != nil {
		log.Println(err)
	}
	for iter.Next(&res) {
		groups[res.Place] = res.Count
	}
	iter.Close()

	for i := range db.Places {
		num, ok := speakers[db.Places[i].Id]
		if !ok {
			num = 0
		}
		set := bson.M{}
		change := false
		if num != db.Places[i].Speakers {
			db.Places[i].Speakers = num
			set[db.KFieldPlaceSpeaker] = num
			change = true
		}

		num, ok = groups[db.Places[i].Id]
		if !ok {
			num = 0
		}
		if num != db.Places[i].Groups {
			db.Places[i].Groups = num
			set[db.KFieldPlaceGroup] = num
			change = true
		}

		if change {
			placeCollection.UpdateId(db.Places[i].Id, bson.M{"$set": set})
		}
	}
}

func (uer *userEditRecord) setup(theUser User) {
	uer.Name = theUser.Name
	uer.Description = theUser.Description
	uer.PlaceCheck = getPlaces(theUser.Locations)
	uer.Number = theUser.Id

	uer.EMail = crypto.Decrypt(theUser.EMail)
	uer.EMailBackup = crypto.Decrypt(theUser.BackupEMail)

	// uer.EMail       string              `schema:"email"`
	// uer.EMailBackup string              `schema:"backup"`

	if theUser.Group {
		uer.Group = true
		uer.Size = theUser.GroupSize
		uer.When = theUser.GroupWhen
	} else {
		uer.Group = false
	}
}

func (uer userEditRecord) check() error {
	if len(uer.Name) == 0 || len(uer.EMail) > 1000 {
		return errorName
	}

	if len(uer.Description) == 0 || len(uer.Description) > 10000 {
		return errorDescription
	}

	if len(uer.Places) == 0 || len(uer.Places) > 1000 {
		return errorPlaces
	}

	if len(uer.EMail) == 0 || len(uer.EMail) > 1000 {
		return errorEMail
	}

	if uer.Group {
		if len(uer.Size) == 0 || len(uer.Size) > 1000 {
			return ErrorSize
		}
		if len(uer.When) == 0 || len(uer.When) > 1000 {
			return ErrorWhen
		}
	}

	return nil
}

func (uer userEditRecord) save() {
	var (
		err error
	)
	sets := bson.M{}
	unsets := bson.M{}
	action := bson.M{}
	sets[FieldName] = uer.Name
	sets[FieldDescription] = uer.Description
	sets[FieldLocations] = uer.Places
	sets[FieldEMail] = string(crypto.Encrypt(uer.EMail))
	sets[FieldEMHash], err = crypto.GetHash(uer.EMail, emSalt)

	if uer.Group {
		sets[FieldSize] = uer.Size
		sets[FieldWhen] = uer.When
	}

	action["$set"] = sets
	if len(uer.EMailBackup) == 0 {
		unsets[FieldBackupEMail] = 1
		unsets[FieldBackupHash] = 1
		action["$unset"] = unsets
	} else {
		sets[FieldBackupEMail] = string(crypto.Encrypt(uer.EMailBackup))
		sets[FieldBackupHash], err = crypto.GetHash(uer.EMailBackup, emSalt)
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	err = usersCollection.UpdateId(uer.Number, action)
	if err != nil {
		log.Println(err)
	}
}

func getPlaces(locations []string) []types.CheckBoxRec {
	var (
		places    []db.PlaceRecord
		allPlaces []types.CheckBoxRec
		err       error
	)
	session := db.MongoSession.Copy()
	defer session.Close()

	placeCollection := session.DB(db.MainDB).C(db.CollectionPlaces)

	err = placeCollection.Find(nil).Sort(db.KFieldId).All(&places)
	if err != nil {
		log.Println("Error: getPlaces read placeCollection")
		return allPlaces
	}

	for _, item := range places {
		selected := false
		if inStringSlice(item.Id, locations) {
			selected = true
		}
		allPlaces = append(allPlaces, types.CheckBoxRec{Title: item.Id, Selected: selected})
	}
	return allPlaces
}

func SetEMail(userNum int, email, password string) {
	var (
		theUser User
		theHash []byte
		err     error
	)
	session := db.MongoSession.Copy()
	defer session.Close()

	usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	theHash, err = crypto.GetHash(email, emSalt)
	if err != nil {
		log.Println("EMail hash:", err)
	}

	theStr := string(crypto.Encrypt(email))

	salt := crypto.RandomChars(20)
	pw, err := crypto.GetHash(password, salt)

	err = usersCollection.UpdateId(userNum, bson.M{"$set": bson.M{FieldEMHash: theHash, FieldEMail: theStr, FieldSalt: salt, FieldPassword: pw}})
	if err != nil {
		log.Println("EMail save:", err)
	}

	err = usersCollection.FindId(userNum).One(&theUser)
	if err != nil {
		log.Println("Error user read:", err)
	}
}

func (prr *passResetRec) check(userRec *User) error {
	var (
		checkPass SignupRecord
		theHash   []byte
	)
	checkPass.EMail = prr.Email
	checkPass.Password = prr.PassWord1
	checkPass.PasswordAgain = prr.PassWord2
	err := checkPass.check(false)
	if err != nil {
		return err
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	theHash, err = crypto.GetHash(strings.ToLower(prr.Email), emSalt)

	err = usersCollection.Find(bson.M{FieldEMHash: theHash}).One(userRec)
	if err != nil {
		return errorCodeMismatch
	}

	if userRec.ResetCode != prr.Code {
		return errorCodeMismatch
	}

	err = usersCollection.UpdateId(userRec.Id, bson.M{"$unset": bson.M{FieldResetCode: 1}})

	return nil
}
