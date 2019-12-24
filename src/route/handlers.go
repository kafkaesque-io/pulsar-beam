package route

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pulsar-beam/src/db"
	"github.com/pulsar-beam/src/model"
	"github.com/pulsar-beam/src/pulsardriver"
	"github.com/pulsar-beam/src/util"
)

var singleDb db.Db

// Init initializes database
func Init() {
	var err error
	singleDb, err = db.NewDb(util.GetConfig().PbDbType)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// StatusPage replies with basic status code
func StatusPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

// ReceiveHandler - the message receiver handler
func ReceiveHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader((http.StatusInternalServerError))
		return
	}

	vars := mux.Vars(r)
	tenant := vars["tenant"]
	token, topicFN, pulsarURL, err2 := util.ReceiverHeader(&r.Header)
	if err2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("tenant %s token %s topicFN %s puslarURL %s", tenant, token, topicFN, pulsarURL)

	err = pulsardriver.SendToPulsar(pulsarURL, token, topicFN, b)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

// GetTopicHandler gets the topic details
func GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicKey, err := getTopicKey(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		resJSON, _ := json.Marshal(ResponseErr{Error: err.Error()})
		w.Write(resJSON)
		return
	}

	doc, err := singleDb.GetByKey(topicKey)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resJSON, err := json.Marshal(doc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resJSON)
	}

}

// UpdateTopicHandler creates or updates a topic
func UpdateTopicHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var doc model.TopicConfig
	err := decoder.Decode(&doc)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		resErr := ResponseErr{Error: err.Error()}
		resJson, err := json.Marshal(resErr)
		w.Write(resJson)
		log.Println(err)
		return
	}

	id, err := singleDb.Create(&doc)
	if len(id) > 1 {
		savedDoc, err := singleDb.GetByKey(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		resJSON, err := json.Marshal(savedDoc)
		w.Write(resJSON)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	return
}

// DeleteTopicHandler deletes a topic
func DeleteTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicKey, err := getTopicKey(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		resJSON, _ := json.Marshal(ResponseErr{Error: err.Error()})
		w.Write(resJSON)
		return
	}
	log.Println(topicKey)

	doc, err := singleDb.DeleteByKey(topicKey)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resJSON, err := json.Marshal(doc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resJSON)
	}
	w.WriteHeader(http.StatusOK)
	return
}

func getTopicKey(r *http.Request) (string, error) {
	var err error
	vars := mux.Vars(r)
	topicKey, ok := vars["topicKey"]
	// log.Printf("topic key is %v", topicKey)
	if !ok {
		var topic model.TopicKey
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		err := decoder.Decode(&topic)
		switch {
		case err == io.EOF:
			return "", errors.New("missing topic key or topic names in body")
		case err != nil:
			return "", err
		}
		topicKey, err = db.GetKeyFromNames(topic.TopicFullName, topic.PulsarURL)
		if err != nil {
			return "", err
		}
	}
	return topicKey, err
}
