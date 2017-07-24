package users

import (
	"speakers/crypto"
	"speakers/db"

	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func startSession(w http.ResponseWriter, theUser User, remember bool) {
	var (
		tempSession db.SessionRecord
		err         error
	)

	session := db.MongoSession.Copy()
	defer session.Close()

	sessionCollection := session.DB(db.MainDB).C(db.CollectionSessions)
	for len(tempSession.Id) != 12 {
		tempStr := crypto.RandomChars(12)
		err = sessionCollection.FindId(tempStr).One(&tempSession)
		if err != nil {
			tempSession.Id = tempStr
		}
	}

	tempStr := tempSession.Id + "|" + makeSecureVal(tempSession.Id)

	cookie := &http.Cookie{
		Name:     "session",
		Value:    tempStr,
		Path:     "/",
		HttpOnly: true,
	}

	if remember {
		cookie.Expires = time.Now().Add(31 * 24 * time.Hour)
	}

	http.SetCookie(w, cookie)

	tempSession.UserNumber = theUser.Id
	tempSession.Name = theUser.Name
	tempSession.Group = theUser.Group
	tempSession.IsGod = theUser.IsGod
	tempSession.Remember = remember
	if remember {
		tempSession.Expires = cookie.Expires.Unix()
	}

	sessionCollection.Insert(tempSession)
	// addLogin(theUser, tempSession.Id, true, r)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	var (
		userRec db.SessionRecord
	)

	if UserLoggedIn(r, &userRec) {
		session := db.MongoSession.Copy()
		defer session.Close()

		sessionCollection := session.DB(db.MainDB).C(db.CollectionSessions)
		err := sessionCollection.RemoveId(userRec.Id)
		if err != nil {
			log.Println("Error: logout fail:", userRec.Id, err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func makeSecureVal(theStr string) string {
	secret := "b46ecb25d1b1f3fcc7d44f863c8532d629bb458b2385fdb1b0731492e7f306a7"

	h := sha256.New()
	io.WriteString(h, theStr+secret)
	str := fmt.Sprintf("%x", h.Sum(nil))

	return str
}
