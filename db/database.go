package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	MainDB             = "speakers"
	CollectionCounters = "counters"
	CollectionSessions = "sessions"
	CollectionUsers    = "users"
	CollectionPlaces   = "places"
	CollectionTopics   = "topics"

	KFieldNumDecimal   = "decimal"
	KFieldNumBinary    = "binary"
	KFieldPlaceSpeaker = "speak"
	KFieldPlaceGroup   = "group"

	KFieldId   = "_id"
	KFieldName = "name"
)

type (
	CounterRecord struct {
		Id    string `json:"_id" bson:"_id"`
		Value int    `json:"value" bson:"value"`
	}

	LoginRecord struct {
		Id       int    `bson:"_id"`
		Name     string `bson:"name"`
		Number   int    `bson:"number"`
		Date     string `bson:"date"`
		Secs     int64  `bson:"secs"`
		Login    bool   `bson:"login"`
		Customer int    `bson:"customer"`
		Cookie   string `bson:"cookie"`
		Action   string `bson:"action"`
		IpAddr   string `bson:"ipaddr"`
	}

	PlaceRecord struct {
		Id       string `bson:"_id"`
		Speakers int    `bson:"speak"`
		Groups   int    `bson:"group"`
	}

	SessionRecord struct {
		Id         string `bson:"_id"`
		UserNumber int    `bson:"number"`
		Name       string `bson:"name"`
		Expires    int64  `bson:"expires"`
		Remember   bool   `bson:"remember"`
		Group      bool   `bson:"group"`
		IsGod      bool   `bson:"isgod"`
	}

	TopicRecord struct {
		Id    int    `bson:"_id"`
		Name  string `bson:"name"`
		Count int    `bson:"count"`
	}

	UserLoginRecord struct {
		Id       int    `bson:"_id"`
		Name     string `bson:"name"`
		Lower    string `bson:"lower"`
		Customer int    `bson:"customer"`
		Password []byte `bson:"pw"`
		IsGod    bool   `bson:"god"`
		Salt     []byte `bson:"salt"`
	}
)

var (
	MongoSession *mgo.Session

	Topics    = []TopicRecord{}
	TopicsMap = map[int]string{}
	Places    = []PlaceRecord{}
	PlacesMap = map[string]PlaceRecord{}
)

func GetNextSequence(name string) int {
	var (
		theCounter CounterRecord
		err        error
	)

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"value": 1}},
		ReturnNew: true,
	}
	session := MongoSession.Copy()
	defer session.Close()

	countersCollection := session.DB(MainDB).C(CollectionCounters)

	_, err = countersCollection.FindId(name).Apply(change, &theCounter)
	if err != nil {
		theCounter.Id = name
		theCounter.Value = 1
		countersCollection.Insert(theCounter)
		return theCounter.Value
	}
	return theCounter.Value
}

func GetCurrentSequenceNumber(name string) int {
	var theCounter CounterRecord

	session := MongoSession.Copy()
	defer session.Close()

	countersCollection := session.DB(MainDB).C(CollectionCounters)

	err := countersCollection.FindId(name).One(&theCounter)
	if err == nil {
		return theCounter.Value
	}
	return 0
}

func AllTopics() {
	var (
		topic TopicRecord
	)
	Topics = []TopicRecord{}
	session := MongoSession.Copy()
	defer session.Close()

	topicCollection := session.DB(MainDB).C(CollectionTopics)

	iter := topicCollection.Find(nil).Sort(KFieldName).Iter()
	for iter.Next(&topic) {
		TopicsMap[topic.Id] = topic.Name
		Topics = append(Topics, topic)
	}
	iter.Close()
}

func GetTopicsString(tags []int) (topicStr string) {
	for i, num := range tags {
		if i > 0 {
			topicStr += ", "
		}
		topicStr += TopicsMap[num]
	}
	return
}

func AllPlaces() {
	var (
		place PlaceRecord
	)
	Places = []PlaceRecord{}
	session := MongoSession.Copy()
	defer session.Close()

	placeCollection := session.DB(MainDB).C(CollectionPlaces)

	iter := placeCollection.Find(nil).Sort(KFieldId).Iter()
	for iter.Next(&place) {
		PlacesMap[place.Id] = place
		Places = append(Places, place)
	}
	iter.Close()
}
