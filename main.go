package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"speakers/crypto"
	"speakers/db"
	"speakers/talks"
	"speakers/types"
	"speakers/users"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

const (
	testPassword1 = "聂辰席主持召开央视分党组扩大会"
	testPassword2 = "🎍💝🎎🎒🎐🎑🎃"
)

type (
	Config struct {
		Port             int
		Name             string
		Initials         string
		Domain           string
		Secure           bool
		NewPassEMail     string
		ActivateEMail    string
		SendMessageEMail string
		MailGunDomain    string
		MailKey          string
	}
)

var (
	key string

	noLogPages = map[string]bool{"/css/barcodes.css": true,
		"/js/functions.js": true, "/images/gh-icons.png": true,
		"/favicon.ico": true}
	noLogExt = map[string]bool{".js": true, ".css": true, ".png": true, ".ico": true}

	config = ReadConfig()
)

func main() {
	var (
		err error
	)
	f, _ := os.OpenFile("./logs/speakers.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	log.SetOutput(f)
	log.Println("Application starting")

	db.MongoSession, err = mgo.Dial("127.0.0.1")
	if err != nil {
		log.Fatal("Bugger Mongo doesn't open")
	}

	parts := strings.Split(key, "|")
	if len(parts) != 2 {
		log.Fatal("Incomplete key and/or mail salt")
	}

	crypto.SetKey([]byte(parts[0]))

	defer db.MongoSession.Close()

	db.MongoSession.SetMode(mgo.Monotonic, true)
	db.AllTopics()
	db.AllPlaces()

	params := types.M{}
	params[users.CfgMailSalt] = parts[1]
	params[users.CfgName] = config.Name
	params[users.CfgDomain] = config.Domain
	params[users.CfgInitials] = config.Initials

	if config.Secure {
		params[users.CfgSecure] = config.Secure
	}
	if len(config.ActivateEMail) > 0 {
		params[users.CfgActivateEMail] = config.ActivateEMail
	}
	if len(config.NewPassEMail) > 0 {
		params[users.CfgNewPassEMail] = config.NewPassEMail
	}
	if len(config.SendMessageEMail) > 0 {
		params[users.CfgSendMessageEMail] = config.SendMessageEMail
	}

	params[users.CfgMailGunDomain] = config.MailGunDomain
	params[users.CfgMailKey] = config.MailKey

	users.SetVars(params)
	users.UpdatePlaces()
	talks.SetVars(config.Name, config.Initials)
	talks.UpdateTopics()

	r := mux.NewRouter()
	r.StrictSlash(true)
	cs := http.FileServer(http.Dir("css"))
	http.Handle("/css/", http.StripPrefix("/css/", cs))
	js := http.FileServer(http.Dir("js"))
	http.Handle("/js/", http.StripPrefix("/js/", js))
	images := http.FileServer(http.Dir("images"))
	http.Handle("/images/", http.StripPrefix("/images/", images))

	talks.SetHandlers(r)
	users.SetHandlers(r)

	r.HandleFunc("/", allHandler)

	http.Handle("/", r)
	portStr := fmt.Sprintf(":%d", config.Port)
	http.ListenAndServe(portStr, AddContext(http.DefaultServeMux))
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	var (
		header  types.HeaderRecord
		details types.MainPageRecord
		theUser db.SessionRecord
	)

	header.Loggedin = users.UserLoggedIn(r, &theUser)

	header.Title = fmt.Sprintf("%s: Main Page", config.Initials)
	header.Name = config.Name
	header.Scripts = append(header.Scripts, "/js/passwordtoggle.js")

	users.CountSpeakerTalks(&details)

	data := struct {
		Header  types.HeaderRecord
		Details types.MainPageRecord
	}{
		header,
		details,
	}

	t, err := template.ParseFiles("templates/main.html", types.ViewHeader, types.ViewNavbar, types.ViewLoginInsert)
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, data)
}

func AddContext(handler http.Handler) http.Handler {
	var (
		theUser db.SessionRecord
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.ToLower(r.URL.Path)
		if !noLogPages[urlPath] {
			log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		}

		theSession := types.Session{IsGod: false, UserNumber: 0}
		theSession.LoggedIn = users.UserLoggedIn(r, &theUser)

		if theSession.LoggedIn {
			//Add data to context
			theSession.IsGod = theUser.IsGod
			theSession.UserNumber = theUser.UserNumber
		}
		ctx := context.WithValue(r.Context(), "logged", theSession)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.ToLower(r.URL.Path)

		if !noLogPages[urlPath] {
			log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		}

		handler.ServeHTTP(w, r)
	})
}

func ReadConfig() Config {
	var (
		configfile = "speakers.cfg"
		config     Config
	)
	_, err := os.Stat(configfile)
	if err != nil {
		config.Port = 9003
		log.Println("Config file is missing: ", configfile)
		return config
	}

	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		config.Port = 9003
		log.Println(err)
	}

	return config
}
