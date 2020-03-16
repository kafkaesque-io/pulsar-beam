package route

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kafkaesque-io/pulsar-beam/src/db"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/pulsardriver"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
)

var singleDb db.Db
var superRoles []string

// Init initializes database
func Init() {
	singleDb = db.NewDbWithPanic(util.GetConfig().PbDbType)
	superRoleStr := util.AssignString(util.GetConfig().SuperRoles, "superuser")
	superRoles = strings.Split(superRoleStr, ",")
}

// TokenSubjectHandler issues new token
func TokenSubjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subject, ok := vars["sub"]
	if !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if util.StrContains(superRoles, util.AssignString(r.Header.Get("injectedSubs"), "BOGUSROLE")) {
		tokenString, err := util.JWTAuth.GenerateToken(subject)
		if err != nil {
			util.ResponseErrorJSON(errors.New("failed to generate token"), w, http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(tokenString))
		}
		return
	}
	util.ResponseErrorJSON(errors.New("incorrect subject"), w, http.StatusUnauthorized)
	return
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, topicFN, pulsarURL, err2 := util.ReceiverHeader(&r.Header)
	if err2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("topicFN %s pulsarURL %s", topicFN, pulsarURL)

	pulsarAsync := r.URL.Query().Get("mode") == "async"
	err = pulsardriver.SendToPulsar(pulsarURL, token, topicFN, b, pulsarAsync)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

// GetTopicHandler gets the topic details
func GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicKey, err := getTopicKey(r)
	if err != nil {
		util.ResponseErrorJSON(err, w, http.StatusUnprocessableEntity)
		return
	}

	doc, err := singleDb.GetByKey(topicKey)
	if err != nil {
		log.Printf("get topic error %v", err)
		util.ResponseErrorJSON(err, w, http.StatusNotFound)
		return
	}
	if !VerifySubject(doc.TopicFullName, r.Header.Get("injectedSubs")) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	resJSON, err := json.Marshal(doc)
	if err != nil {
		util.ResponseErrorJSON(err, w, http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resJSON)
	}

}

// UpdateTopicHandler creates or updates a topic
func UpdateTopicHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var doc model.TopicConfig
	err := decoder.Decode(&doc)
	if err != nil {
		util.ResponseErrorJSON(err, w, http.StatusUnprocessableEntity)
		return
	}

	if _, err = model.ValidateTopicConfig(doc); err != nil {
		util.ResponseErrorJSON(err, w, http.StatusUnprocessableEntity)
		return
	}

	if !VerifySubject(doc.TopicFullName, r.Header.Get("injectedSubs")) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	id, err := singleDb.Update(&doc)
	if err != nil {
		// log.Println(err)
		w.WriteHeader(http.StatusConflict)
		return
	}
	if len(id) > 1 {
		savedDoc, err := singleDb.GetByKey(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		resJSON, err := json.Marshal(savedDoc)
		if err != nil {
			log.Printf("marshal updated topic error %v", err)
			return
		}
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
		util.ResponseErrorJSON(err, w, http.StatusUnprocessableEntity)
		return
	}

	doc, err := singleDb.GetByKey(topicKey)
	if err != nil {
		log.Println(err)
		util.ResponseErrorJSON(err, w, http.StatusNotFound)
		return
	}
	if !VerifySubject(doc.TopicFullName, r.Header.Get("injectedSubs")) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	deletedKey, err := singleDb.DeleteByKey(topicKey)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resJSON, err := json.Marshal(deletedKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resJSON)
	}
}

func getTopicKey(r *http.Request) (string, error) {
	var err error
	vars := mux.Vars(r)
	topicKey, ok := vars["topicKey"]
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
		topicKey, err = model.GetKeyFromNames(topic.TopicFullName, topic.PulsarURL)
		if err != nil {
			return "", err
		}
	}
	return topicKey, err
}

// VerifySubject verifies the subject can meet the requirement.
func VerifySubject(topicFN, tokenSub string) bool {
	parts := strings.Split(topicFN, "/")
	if len(parts) < 3 {
		return false
	}
	tenant := parts[2]
	if len(tenant) < 1 {
		log.Printf(" auth verify tenant %s token sub %s", tenant, tokenSub)
		return false
	}
	subjects := append(superRoles, tenant)

	return util.StrContains(subjects, tokenSub)
}
