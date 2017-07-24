package users

import (
	"speakers/crypto"
	"speakers/db"
	"speakers/types"

	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	outputLoginForm(loginRecord{}, "", w)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	var (
		details      loginRecord
		theUser      User
		errorMessage string
		err          error
	)

	err = r.ParseForm()

	if err != nil {
		log.Println("error: loginPostHandler parse", err)
	}

	err = decoder.Decode(&details, r.PostForm)
	if err != nil {
		log.Println("error: loginPostHandler decode", err)
	}

	theUser, err = details.check()

	if err == nil {
		if theUser.NotVerified {
			http.Redirect(w, r, "/activatesend", http.StatusSeeOther)
			return
		}

		startSession(w, theUser, details.Remember)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	outputLoginForm(details, errorMessage, w)
}

func outputLoginForm(details loginRecord, errorMessage string, w http.ResponseWriter) {
	var (
		header types.HeaderRecord
	)

	header.Title = fmt.Sprintf("%s: Login Page", initials)
	header.Name = siteName
	header.Scripts = append(header.Scripts, "/js/passwordtoggle.js")

	data := struct {
		Header       types.HeaderRecord
		Details      loginRecord
		ErrorMessage string
	}{
		header,
		details,
		errorMessage,
	}

	t, err := template.ParseFiles("templates/login.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println("Error: outputLoginForm", err)
	}
	t.Execute(w, data)
}

func (lr loginRecord) check() (User, error) {
	var (
		userRec User
		theHash []byte
		err     error
	)
	email := strings.TrimSpace(lr.EMail)
	if len(email) == 0 {
		return userRec, errorLogin
	}
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	theHash, err = crypto.GetHash(email, emSalt)
	if err != nil {
		log.Println("EMail hash:", err)
	}

	session := db.MongoSession.Copy()
	defer session.Close()

	usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	err = usersCollection.Find(bson.M{FieldEMHash: theHash}).One(&userRec)
	if err != nil {
		log.Println("read user error:", err)
		return userRec, errorLogin
	}

	password, err := crypto.GetHash(lr.Password, userRec.Salt)

	if len(password) != len(userRec.Password) {
		return userRec, errorLogin
	}

	for i, ch := range userRec.Password {
		if ch != password[i] {
			return userRec, errorLogin
		}
	}

	return userRec, nil
}
