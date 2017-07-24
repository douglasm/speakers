package users

import (
	"speakers/crypto"
	"speakers/db"
	"speakers/types"

	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"gopkg.in/mgo.v2/bson"
)

const (
	KFieldGroup  = "group"
	KFieldTalks  = "talks"
	KFieldPlaces = "loc"
	KFieldTopics = "topics"
)

const (
	signupFormID = iota + 100
	loginFormID
	editFormID
	detailsFormID
	activateFormID
	activateSendFormID
	resendFormID
	resetFormID
	changePasswordFormID
)

const (
	FieldName         = "name"
	FieldUserName     = "user"
	FieldLower        = "lower"
	FieldDescription  = "desc"
	FieldJoinDate     = "join date"
	FieldJoined       = "joined"
	FieldSalt         = "salt"
	FieldPassword     = "pw"
	FieldEMail        = "email"
	FieldEMHash       = "emhash"
	FieldBackupEMail  = "backup"
	FieldBackupHash   = "buhash"
	FieldResetCode    = "code"
	FieldNotVerified  = "notv"
	FieldResetExpires = "exp"
	FieldNumTalks     = "num"
	FieldActive       = "active"
	FieldIsGod        = "god"
	FieldGroup        = "group"
	FieldSpeaker      = "speak"
	FieldLocations    = "loc"
	FieldTalks        = "talks"
	FieldSize         = "size"
	FieldWhen         = "when"
	FieldTags         = "tags"
	FieldTalkTags     = "talks.tags"
	FieldTalkTopics   = "talks.topics"
)

const (
	CfgName = iota
	CfgInitials
	CfgDomain
	CfgSecure
	CfgNewPassEMail
	CfgActivateEMail
	CfgSendMessageEMail
	CfgMailGunDomain
	CfgMailKey
	CfgMailSalt
)

type (
	loginRecord struct {
		EMail      string `schema:"email"`
		Password   string `schema:"password"`
		Remember   bool   `schema:"remember"`
		Checkfield string `schema:"checkfield"`
		Commit     string `schema:"commit"`
	}

	messageRec struct {
		Name       string `schema:"-"`
		Site       string `schema:"-"`
		Sender     string `schema:"-"`
		Message    string `schema:"message"`
		Num        int    `schema:"num"`
		SendNum    int    `schema:"send"`
		Checkfield string `schema:"checkfield"`
	}

	passwordRec struct {
		Email      string `schema:"email"`
		Checkfield string `schema:"checkfield"`
		Commit     string `schema:"commit"`
	}

	passResetRec struct {
		Email      string `schema:"email"`
		Code       string `schema:"resetcode"`
		PassWord1  string `schema:"password1"`
		PassWord2  string `schema:"password2"`
		Checkfield string `schema:"checkfield"`
		Remember   bool   `schema:"remember"`
		Commit     string `schema:"commit"`
	}

	SignupRecord struct {
		EMail         string `schema:"email"`
		Name          string `schema:"name"`
		Password      string `schema:"password"`
		PasswordAgain string `schema:"passagain"`
		Group         bool   `schema:"group"`
		Checkfield    string `schema:"checkfield"`
	}

	speakerListRec struct {
		UserNum int
		Name    string
		Count   int
		Places  []string
		Topics  []string
	}

	TalkRecord struct {
		Number      int    `bson:"num"`
		Title       string `bson:"title"`
		Description string `bson:"desc"`
		Tags        []int  `bson:"tags"`
	}

	User struct {
		Id           int          `bson:"_id"`
		Name         string       `bson:"name"`
		Description  string       `bson:"desc"`
		JoinDate     string       `bson:"join date"`
		Joined       int64        `bson:"joined"`
		Salt         string       `bson:"salt"`
		Password     []byte       `bson:"pw"`
		EMail        string       `bson:"email"`
		EMHash       []byte       `bson:"emhash,omitempty"`
		BackupEMail  string       `bson:"backup,omitempty"`
		BackupHash   []byte       `bson:"buhash,omitempty"`
		ResetCode    string       `bson:"code,omitempty"`
		ResetExpires float64      `bson:"exp,omitempty"`
		NotVerified  bool         `bson:"notv,omitempty"`
		IsGod        bool         `bson:"god,omitempty"`
		Group        bool         `bson:"group"`
		GroupSize    string       `bson:"size,omitempty"`
		GroupMembers string       `bson:"member,omitempty"`
		GroupWhen    string       `bson:"when,omitempty"`
		Locations    []string     `bson:"loc,omitempty"`
		Talks        []TalkRecord `bson:"talks,omitempty"`
	}

	userEditRecord struct {
		Name        string              `schema:"name"`
		Description string              `schema:"desc"`
		EMail       string              `schema:"email"`
		EMailBackup string              `schema:"backup"`
		Size        string              `schema:"size"`
		When        string              `schema:"when"`
		PlaceCheck  []types.CheckBoxRec `schema:"-"`
		Places      []string            `schema:"places"`
		Number      int                 `schema:"number"`
		Group       bool                `schema:"group"`
		Checkfield  string              `schema:"checkfield"`
	}
)

var (
	decoder = schema.NewDecoder()

	siteName      string
	initials      string
	domain        string
	isSecure      bool
	activateEMail string
	newPassEMail  string
	messageEMail  string
	mailKey       string
	mailGunDomain string
	emSalt        string

	errorUserName        = errors.New("No Username")
	errorEMail           = errors.New("No E Mail")
	errorPassword        = errors.New("No password")
	errorMatch           = errors.New("Passwords do not match")
	errorDuplicate       = errors.New("That E-Mail is already in use")
	errorCodeMismatch    = errors.New("That E-Mail and the reset code do not match")
	errorPassEMail       = errors.New("Your password looks like an e-mail address")
	errorNotMail         = errors.New("Not an e-mail address")
	errorPassLongAscii   = errors.New("Your password is not long enough")
	errorPassLongUnicode = errors.New("Your password is not long enough")
	errorLogin           = errors.New("Login/Password wrong")
	errorName            = errors.New("Please enter a name")
	errorDescription     = errors.New("Please enter a description")
	errorPlaces          = errors.New("Please select at least one location")
	ErrorSize            = errors.New("Please enter the number of people who attend")
	ErrorWhen            = errors.New("Please enter when your group meets")
	ErrorMessage         = errors.New("Please enter a longer message")

	activateCode string

	countAgg       = bson.M{"$sum": 1}
	existsQuery    = bson.M{"$exists": 1}
	notExistsQuery = bson.M{"$exists": 0}
)

func SetVars(params types.M) {
	var (
		ok bool
	)
	emSalt = params[CfgMailSalt].(string)
	siteName = params[CfgName].(string)
	domain = params[CfgDomain].(string)
	initials = params[CfgInitials].(string)
	isSecure = params[CfgSecure].(bool)

	activateEMail, ok = params[CfgActivateEMail].(string)
	if !ok {
		activateEMail = "noreply"
	}
	newPassEMail, ok = params[CfgNewPassEMail].(string)
	if !ok {
		newPassEMail = "noreply"
	}
	messageEMail, ok = params[CfgSendMessageEMail].(string)
	if !ok {
		messageEMail = "noreply"
	}

	mailGunDomain = params[CfgMailGunDomain].(string)
	mailKey = params[CfgMailKey].(string)
}

func SetHandlers(r *mux.Router) {
	r.HandleFunc("/users", usersHandler)
	r.HandleFunc("/user/{number:[0-9]+}", userHandler)
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")
	r.HandleFunc("/logout", logoutHandler)
	r.HandleFunc("/signup", signupGetHandler).Methods("GET")
	r.HandleFunc("/signup", signupPostHandler).Methods("POST")
	r.HandleFunc("/signedup", signedupHandler)
	r.HandleFunc("/activate", activateGetHandler).Methods("GET")
	r.HandleFunc("/activate", activatePostHandler).Methods("POST")
	r.HandleFunc("/activate/{code}", activateGetHandler).Methods("GET")
	r.HandleFunc("/details", detailsHandler)
	r.HandleFunc("/setplace", setPlacePostHandler).Methods("POST")
	r.HandleFunc("/useredit", userEditGetHandler).Methods("GET")
	r.HandleFunc("/useredit", userEditPostHandler).Methods("POST")
	r.HandleFunc("/activatesend", activateSendHandler)
	r.HandleFunc("/resend", resendGetHandler).Methods("GET")
	r.HandleFunc("/resend", resendPostHandler).Methods("POST")

	r.HandleFunc("/message/{number:[0-9]+}", messageHandler)
	// r.HandleFunc("/message/{number:[0-9]+}", messagePostHandler).Methods("POST")

	r.HandleFunc("/speakers", speakersHandler)
	r.HandleFunc("/groups", groupsHandler)
	r.HandleFunc("/passwordreset", passwordHandler)
	r.HandleFunc("/newpassword", newPasswordHandler)
	r.HandleFunc("/newpassword/{code}", newPasswordHandler)

	r.HandleFunc("/revisespeaker", reviseSpeakerPostHandler).Methods("POST")
	r.HandleFunc("/revisegroup", reviseGroupPostHandler).Methods("POST")
}

func inStringSlice(theString string, theSlice []string) bool {
	theString = strings.ToLower(theString)
	for _, str := range theSlice {
		str = strings.ToLower(str)
		if theString == str {
			return true
		}
	}
	return false
}

func inIntSlice(theInt int, theSlice []int) bool {
	for _, num := range theSlice {
		if theInt == num {
			return true
		}
	}
	return false
}

func TopicList(topicSearch []int) (checkboxes []types.CheckBoxRec) {
	for _, item := range db.Topics {
		if item.Count > 0 {
			cbr := types.CheckBoxRec{}
			cbr.Title = fmt.Sprintf("%s (%d)", item.Name, item.Count)
			cbr.Value = fmt.Sprintf("%d", item.Id)
			if inIntSlice(item.Id, topicSearch) {
				cbr.Selected = true
			}
			checkboxes = append(checkboxes, cbr)
		}
	}

	return checkboxes
}

func PlaceList(placeSearch []string) (checkboxes []types.CheckBoxRec) {
	for _, item := range db.Places {
		if item.Speakers > 0 {
			cbr := types.CheckBoxRec{}
			cbr.Title = fmt.Sprintf("%s (%d)", item.Id, item.Speakers)
			cbr.Value = fmt.Sprintf("%s", item.Id)
			if inStringSlice(item.Id, placeSearch) {
				cbr.Selected = true
			}
			checkboxes = append(checkboxes, cbr)
		}
	}

	return
}

func GroupPlaceList(placeSearch []string) (checkboxes []types.CheckBoxRec) {
	for _, item := range db.Places {
		if item.Groups > 0 {
			cbr := types.CheckBoxRec{}
			cbr.Title = fmt.Sprintf("%s (%d)", item.Id, item.Groups)
			cbr.Value = fmt.Sprintf("%s", item.Id)
			if inStringSlice(item.Id, placeSearch) {
				cbr.Selected = true
			}
			checkboxes = append(checkboxes, cbr)
		}
	}

	return
}

func GetParamVal(params map[string]string, paramName string, defaultVal int) int {
	var (
		val int
		err error
	)

	theStr, ok := params[paramName]
	if ok {
		val, err = strconv.Atoi(theStr)
		if err != nil {
			val = defaultVal
		}
	} else {
		val = defaultVal
	}
	return val
}

func getRandomCode() string {
	var (
		ch rune
	)
	code := crypto.RandomChars(17)
	total := 0
	for i, ch := range code {
		total += i * int(ch)
	}
	total %= 26
	ch = rune(total + 65)
	return code + string(ch)
}
