package users

import (
	"speakers/db"
	// "speakers/login"
	"speakers/crypto"
	"speakers/types"
	// "speakers/utils"

	// "utilscrypto"

	// "errors"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	// "github.com/gorilla/schema"
	// "golang.org/x/crypto/scrypt"
	"gopkg.in/mgo.v2/bson"
)

func signupGetHandler(w http.ResponseWriter, r *http.Request) {
	outputSignupForm(SignupRecord{}, "", w)
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	var (
		details        SignupRecord
		activationCode string
		errorMessage   string
		err            error
	)

	err = r.ParseForm()
	if err != nil {
		log.Println("error: signupPostHandler parse", err)
	}
	err = decoder.Decode(&details, r.PostForm)
	if err != nil {
		log.Println("error: signupPostHandler decode", err)
	}

	err = details.check(true)

	if err == nil {
		activationCode, err = details.Save()
		if err == nil {
			sendActivate(details.EMail, activationCode)

			http.Redirect(w, r, "/signedup", http.StatusSeeOther)
			return
		}
	}

	if err != nil {
		errorMessage = err.Error()
	}

	outputSignupForm(details, errorMessage, w)
}

func signedupHandler(w http.ResponseWriter, r *http.Request) {
	var (
		header types.HeaderRecord
	)

	header.Title = fmt.Sprintf("%s: You have signed up", initials)
	header.Name = siteName
	data := struct {
		Header types.HeaderRecord
		Code   string
	}{
		header,
		activateCode,
	}

	t, err := template.ParseFiles("templates/signedup.html", types.ViewHeader, types.ViewNavbar)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func outputSignupForm(details SignupRecord, errorMessage string, w http.ResponseWriter) {
	var (
		header types.HeaderRecord
	)

	header.Title = fmt.Sprintf("%s: Signup Page", initials)
	header.Name = siteName
	header.Scripts = append(header.Scripts, "/js/passwordtoggle.js")

	data := struct {
		Header       types.HeaderRecord
		Details      SignupRecord
		ErrorMessage string
	}{
		header,
		details,
		errorMessage,
	}

	t, err := template.ParseFiles("templates/signup.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println("error: outputSignupForm ParseFiles", err)
	}
	t.Execute(w, data)
}

func activateGetHandler(w http.ResponseWriter, r *http.Request) {
	var (
		good         bool
		errorMessage string
	)

	params := mux.Vars(r)
	theCode, ok := params["code"]

	good = false
	if ok {
		total := 0
		for i, ru := range theCode {
			if i < 17 {
				total += i * int(ru)
			} else {
				total %= 26
				ch := rune(total + 65)
				if ch == ru {
					good = true
				}
			}
		}

		if good {
			acceptCode(theCode, w, r)
		}
		errorMessage = "Incorrect code"
	}

	outputActivateForm(errorMessage, w)
}

func activatePostHandler(w http.ResponseWriter, r *http.Request) {
	var (
		theCode      string
		errorMessage string
		good         bool
	)

	err := r.ParseForm()
	if err != nil {
		log.Println("error: signupPostHandler parse", err)
	}

	theCode = r.Form.Get("code") // theCode will be "" if parameter is not set
	// checkField := r.Form.Get("checkfield")

	good = false
	if len(theCode) > 0 {
		total := 0
		for i, ru := range theCode {
			if i < 17 {
				total += i * int(ru)
			} else {
				total %= 26
				ch := rune(total + 65)
				if ch == ru {
					good = true
				}
			}
		}

		if good {
			acceptCode(theCode, w, r)
		}
		errorMessage = "Incorrect code"
	} else {
		errorMessage = "no code entered"
	}

	outputActivateForm(errorMessage, w)
}

func acceptCode(theCode string, w http.ResponseWriter, r *http.Request) {
	var (
		userRec User
		err     error
	)
	session := db.MongoSession.Copy()
	defer session.Close()

	usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	err = usersCollection.Find(bson.M{FieldResetCode: theCode}).One(&userRec)
	if err == nil {
		err = usersCollection.UpdateId(userRec.Id, bson.M{"$unset": bson.M{FieldResetCode: 1, FieldNotVerified: 1}})
		if err != nil {
			log.Println("Error: activate update error:", err)
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
}

func outputActivateForm(errorMessage string, w http.ResponseWriter) {
	var (
		header types.HeaderRecord
	)
	checkField := crypto.MakeNonce(activateFormID, "noone")
	header.Title = fmt.Sprintf("%s: Verify E Mail address", initials)
	header.Name = siteName
	data := struct {
		Header       types.HeaderRecord
		Code         string
		Checkfield   string
		ErrorMessage string
	}{
		header,
		activateCode,
		checkField,
		errorMessage,
	}

	t, err := template.ParseFiles("templates/activate.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println("Error: outputActivateForm ParseFiles", err)
	}
	t.Execute(w, data)
}

func activateSendHandler(w http.ResponseWriter, r *http.Request) {
	var (
		header types.HeaderRecord
	)

	context := r.Context().Value("logged")
	if context != nil {
		header.Loggedin = context.(types.Session).LoggedIn
		header.God = context.(types.Session).IsGod
	}

	checkField := crypto.MakeNonce(activateFormID, "noone")
	header.Title = fmt.Sprintf("%s: Verify E-Mail address", initials)
	header.Name = siteName
	data := struct {
		Header       types.HeaderRecord
		Checkfield   string
		ErrorMessage string
	}{
		header,
		checkField,
		"",
	}

	t, err := template.ParseFiles("templates/activatesend.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println("Error: activateGetHandler ParseFiles", err)
	}
	t.Execute(w, data)
}

func resendGetHandler(w http.ResponseWriter, r *http.Request) {
	var (
		header types.HeaderRecord
	)

	context := r.Context().Value("logged")
	if context != nil {
		header.Loggedin = context.(types.Session).LoggedIn
		header.God = context.(types.Session).IsGod
	}

	checkField := crypto.MakeNonce(resendFormID, "noone")
	header.Title = fmt.Sprintf("%s: Enter E-Mail address", initials)
	header.Name = siteName
	data := struct {
		Header       types.HeaderRecord
		EMail        string
		Checkfield   string
		ErrorMessage string
	}{
		header,
		"",
		checkField,
		"",
	}

	t, err := template.ParseFiles("templates/resend.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println("Error: resendGetHandler ParseFiles", err)
	}
	t.Execute(w, data)
}

func resendPostHandler(w http.ResponseWriter, r *http.Request) {
	var (
		email   string
		userRec User
		header  types.HeaderRecord
	)

	context := r.Context().Value("logged")
	if context != nil {
		header.Loggedin = context.(types.Session).LoggedIn
		header.God = context.(types.Session).IsGod
	}

	err := r.ParseForm()
	if err != nil {
		log.Println("error: signupPostHandler parse", err)
	}

	email = r.Form.Get("email") // theCode will be "" if parameter is not set

	if len(email) > 0 {
		email = strings.ToLower(email)
		session := db.MongoSession.Copy()
		defer session.Close()

		usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)
		theHash, err := crypto.GetHash(email, emSalt)

		err = usersCollection.Find(bson.M{FieldEMHash: theHash}).One(&userRec)
		if err == nil {
			sendActivate(email, userRec.ResetCode)
		}

		header.Title = fmt.Sprintf("%s: E-Mail Address verification sent", initials)
		header.Name = siteName
		data := struct {
			Header types.HeaderRecord
		}{
			header,
		}

		t, err := template.ParseFiles("templates/signedup.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
		if err != nil {
			log.Println("Error: resendPostHandler ParseFiles", err)
		}
		t.Execute(w, data)

		return
	}

	checkField := crypto.MakeNonce(resendFormID, "noone")

	header.Title = fmt.Sprintf("%s: Enter E-Mail address", initials)
	header.Name = siteName
	data := struct {
		Header       types.HeaderRecord
		EMail        string
		Checkfield   string
		ErrorMessage string
	}{
		header,
		email,
		checkField,
		"Please enter an e-mail address",
	}

	t, err := template.ParseFiles("templates/resend.html", types.ViewHeader, types.ViewNavbar, types.ViewErr)
	if err != nil {
		log.Println("Error: resendPostHandler ParseFiles", err)
	}
	t.Execute(w, data)
}

func (sr *SignupRecord) check(checkExists bool) error {
	var (
		theHash []byte
		userRec User
		err     error
	)

	sr.EMail = strings.TrimSpace(sr.EMail)
	if len(sr.EMail) == 0 {
		return errorEMail
	}

	if len(sr.Password) == 0 {
		return errorPassword
	}
	if len(sr.Password) != len(sr.PasswordAgain) {
		return errorMatch
	}

	if sr.Password != sr.PasswordAgain {
		return errorMatch
	}

	if validateEmail(sr.Password) {
		return errorPassEMail
	}

	if !checkPasswordLen(sr.Password) {
		return errorPassLongAscii
	}

	if len(sr.EMail) > 1000 {
		return errorEMail
	}
	if len(sr.Password) > 1000 {
		return errorPassword
	}
	if len(sr.PasswordAgain) > 1000 {
		return errorMatch
	}

	if checkExists {
		session := db.MongoSession.Copy()
		defer session.Close()

		usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)

		theHash, err = crypto.GetHash(strings.ToLower(sr.EMail), emSalt)

		err = usersCollection.Find(bson.M{FieldEMHash: theHash}).One(&userRec)
		if err == nil {
			return errorDuplicate
		}
	}

	return nil
}

func (sr *SignupRecord) Save() (string, error) {
	var (
		ch      rune
		userRec User
		err     error
	)

	session := db.MongoSession.Copy()
	defer session.Close()

	usersCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	userRec.Name = sr.Name

	userRec.EMail = string(crypto.Encrypt(sr.EMail))
	userRec.Salt = crypto.RandomChars(20)
	userRec.Password, err = crypto.GetHash(sr.Password, userRec.Salt)
	userRec.EMHash, err = crypto.GetHash(sr.EMail, emSalt)

	userRec.NotVerified = true
	userRec.Group = sr.Group

	code := crypto.RandomChars(17)
	total := 0
	for i, ch := range code {
		total += i * int(ch)
	}
	total %= 26
	ch = rune(total + 65)
	userRec.ResetCode = code + string(ch)

	if err != nil {
		log.Println("Error: Get hash", err)
	}

	userRec.Id = db.GetNextSequence(db.CollectionUsers)
	if userRec.Id == 1 {
		userRec.IsGod = true
	}
	err = usersCollection.Insert(userRec)
	return userRec.ResetCode, nil
}

func validateEmail(email string) bool {
	email = strings.Replace(email, "dot", ".", -1)
	email = strings.Replace(email, "at", "@", -1)
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func checkPasswordLen(password string) bool {
	var (
		ascii   int
		unicode int
	)
	for _, ch := range password {
		if int(ch) > 255 {
			unicode++
		} else {
			ascii++
		}
	}
	if unicode >= 6 {
		return true
	}
	if ascii >= 10 {
		return true
	}
	switch unicode {
	case 1:
		if ascii >= 8 {
			return true
		}
	case 2:
		if ascii >= 6 {
			return true
		}
	case 3:
		if ascii >= 4 {
			return true
		}
	case 4:
		if ascii >= 3 {
			return true
		}
	case 5:
		if ascii >= 2 {
			return true
		}
	}
	return false
}

func calcEntropy(theString string) (entropy float64) {
	var (
		strLen float64
		chars  = map[rune]float64{}
	)

	strLen = float64(len(theString))
	for _, ch := range theString {
		_, ok := chars[ch]
		if !ok {
			num := float64(strings.Count(theString, string(ch)))
			chars[ch] = num / strLen
		}
	}
	for _, v := range chars {
		entropy = entropy + v*math.Log2(v)
	}
	entropy = 0 - entropy
	return
}
