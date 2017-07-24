package talks

import (
	"speakers/db"
	"speakers/types"
	"speakers/users"

	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"gopkg.in/mgo.v2/bson"
)

const (
	templateTalkView   = "templates/talk.html"
	templateTalkEdit   = "templates/talkedit.html"
	templateTalkDelete = "templates/talkdelete.html"
)

type (
	TalkRec struct {
		Num         int      `bson:"num"`
		Title       string   `bson:"title"`
		Description string   `bson:"desc"`
		Tags        []string `bson:"tags"`
	}

	tagOutputRec struct {
		Tag      string
		Value    string
		Selected bool
	}

	TalkDeleteRec struct {
		Num         int    `schema:"num"`
		Title       string `schema:"-"`
		Description string `schema:"-"`
		Topics      string `schema:"-"`
		Checkfield  string `schema:"checkfield"`
		Button      string `schema:"button"`
	}

	TalkEditRec struct {
		Num         int            `schema:"num"`
		Title       string         `schema:"title"`
		Description string         `schema:"desc"`
		Tags        []int          `schema:"tags"`
		TagChecks   []tagOutputRec `schema:"-"`
		Checkfield  string         `schema:"checkfield"`
	}

	talkListRec struct {
		Id      int    `bson:"_id"`
		UserNum int    `bson:"usernum"`
		Number  int    `bson:"num"`
		Name    string `bson:"name"`
		Title   string `bson:"title"`
		Tags    []int  `bson:"tags"`
	}
)

var (
	decoder = schema.NewDecoder()

	siteName string
	initials string

	countAgg    = bson.M{"$sum": 1}
	existsQuery = bson.M{"$exists": 1}

	ErrorTitle       = errors.New("No title")
	ErrorDescription = errors.New("No description")
	ErrorTags        = errors.New("No tags selected")
)

func topicList(topicSearch []int) (checkboxes []types.CheckBoxRec) {
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

func placeList(placeSearch []string) (checkboxes []types.CheckBoxRec) {
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

func (ter *TalkEditRec) Save(userNumber int) error {
	var (
		userRec users.User
		newTalk users.TalkRecord
		num     int
		looping bool
		err     error
	)

	rand.Seed(time.Now().Unix())

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)

	looping = true
	for looping {
		num = rand.Intn(1000000)
		err = userCollection.Find(bson.M{"talks.num": num}).One(&userRec)
		if err != nil {
			looping = false
		}
	}

	newTalk.Number = num
	newTalk.Title = ter.Title
	newTalk.Description = ter.Description
	newTalk.Tags = ter.Tags

	err = userCollection.UpdateId(userNumber, bson.M{"$addToSet": bson.M{"talks": newTalk}})
	if err != nil {
		log.Println("Error: Update talk", err)
	}
	return nil
}

func (ter *TalkEditRec) check() error {
	ter.Title = strings.TrimSpace(ter.Title)
	if len(ter.Title) < 2 {
		return ErrorTitle
	}

	ter.Description = strings.Trim(ter.Description, "\r\n! ")
	if len(ter.Description) < 2 {
		return ErrorDescription
	}

	if len(ter.Tags) == 0 {
		return ErrorTags
	}
	return nil
}

func (ter *TalkEditRec) buildTopicChecks() {
	var (
		theTopic db.TopicRecord
		selected bool
		// err      error
	)
	session := db.MongoSession.Copy()
	defer session.Close()

	topicCollection := session.DB(db.MainDB).C(db.CollectionTopics)

	iter := topicCollection.Find(nil).Sort("_id").Iter()
	for iter.Next(&theTopic) {
		selected = false
		if inIntSlice(theTopic.Id, ter.Tags) {
			selected = true
		}
		tr := tagOutputRec{Tag: theTopic.Name, Value: fmt.Sprintf("%d", theTopic.Id), Selected: selected}
		ter.TagChecks = append(ter.TagChecks, tr)
	}
	iter.Close()
}

func UpdateTopics() {
	var (
		countRec db.TopicRecord
		topicMap = map[int]int{}
	)

	session := db.MongoSession.Copy()
	defer session.Close()

	userCollection := session.DB(db.MainDB).C(db.CollectionUsers)
	topicCollection := session.DB(db.MainDB).C(db.CollectionTopics)

	match := bson.M{"$match": bson.M{"talks": bson.M{"$exists": 1}}}

	theProject1 := bson.M{}
	theProject1["talks"] = 1
	theProject1["_id"] = 0
	project1 := bson.M{"$project": theProject1}

	unwind1 := bson.M{"$unwind": "$talks"}

	theProject2 := bson.M{}
	theProject2["tags"] = "$talks.tags"
	project2 := bson.M{"$project": theProject2}

	unwind2 := bson.M{"$unwind": "$tags"}

	theGroup := bson.M{"_id": "$tags", "count": countAgg}
	group1 := bson.M{"$group": theGroup}

	pipeline := []bson.M{match, project1, unwind1, project2, unwind2, group1}
	iter := userCollection.Pipe(pipeline).Iter()
	err := iter.Err()
	if err != nil {
		log.Println("Error: Update topics iter", err)
	}
	for iter.Next(&countRec) {
		topicMap[countRec.Id] = countRec.Count
	}
	iter.Close()

	for i, item := range db.Topics {
		num := topicMap[item.Id]
		if num != item.Count {
			db.Topics[i].Count = num
			err = topicCollection.UpdateId(db.Topics[i].Id, bson.M{"$set": bson.M{"count": num}})
		}
	}
}

func getTopics() map[int]string {
	var (
		topic db.TopicRecord
	)
	topicMap := map[int]string{}
	session := db.MongoSession.Copy()
	defer session.Close()

	topicCollection := session.DB(db.MainDB).C(db.CollectionTopics)

	iter := topicCollection.Find(nil).Iter()
	for iter.Next(&topic) {
		topicMap[topic.Id] = topic.Name
	}
	iter.Close()
	return topicMap
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

func SetVars(sName, sInits string) {
	siteName = sName
	initials = sInits

}

func getParamVal(params map[string]string, paramName string, defaultVal int) int {
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

func SetHandlers(r *mux.Router) {
	r.HandleFunc("/talks", talksHandler)
	r.HandleFunc("/talkadd", talkAddHandlerGet).Methods("GET")
	r.HandleFunc("/talkadd", talkAddHandlerPost).Methods("POST")
	r.HandleFunc("/talkedit", talkEditHandlerGet).Methods("GET")
	r.HandleFunc("/talkedit/{number:[0-9]+}", talkEditHandlerGet).Methods("GET")
	r.HandleFunc("/talkedit/{number:[0-9]+}", talkEditHandlerPost).Methods("POST")
	r.HandleFunc("/talkdelete/{number:[0-9]+}", talkDeleteHandlerGet).Methods("GET")
	r.HandleFunc("/talkdelete/{number:[0-9]+}", talkDeleteHandlerPost).Methods("POST")

	r.HandleFunc("/revisetalk", reviseTalkHandlerPost).Methods("POST")
	r.HandleFunc("/talk/{number:[0-9]+}", talkHandler)

	// r.HandleFunc("/talkslist", talksListHandler).Methods("POST")

	// r.HandleFunc("/addtalk/{number:[0-9]+}", customersHandler)
	// r.HandleFunc("/edittalk/{number}", custEditHandlerGet).Methods("GET")
	// r.HandleFunc("/edittalk/{number}", custEditHandlerPost).Methods("POST")
	// r.HandleFunc("/deletetalk/{number}", customerHandler)
	// r.HandleFunc("/searchtalk", custSearchHandlerPost).Methods("POST")
}
